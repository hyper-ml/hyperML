package schedules

import (
	"fmt"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	"github.com/hyper-ml/hyperml/server/pkg/config"
	"github.com/hyper-ml/hyperml/server/pkg/pods"
	"github.com/hyper-ml/hyperml/server/pkg/qs"
	"github.com/hyper-ml/hyperml/server/pkg/types"
	"sync"
	"time"
)

// SaveStateInterval : Determines how often the state of botebook scheduler is backed up
var SaveStateInterval = 60 // seconds

// SweepInterval : Time interval to sync state of wait and active list
var SweepInterval = 15 // minutes

var emptyStruct struct{}

// NotebookScheduler : Responsible for monitoring scheduled notebooks
type NotebookScheduler struct {
	concurrency int

	podkeeper      *pods.Keeper
	notebookConfig *config.NotebookConfig
	schedulerConf  *config.NbSchedulerConfig
	jobConfig      *config.JobConfig

	// Active list of runnning notebooks
	Active map[uint64]interface{}
	alock  sync.Mutex

	// WaitList queue of notebook runs
	WaitList map[int]uint64
	wlock    sync.Mutex

	qh *qHandler

	dirty bool

	quit     chan int
	events   chan interface{}
	checkrun chan int

	// channel to record POD shutdown
	podShutdown chan struct{}
}

// NewNotebookScheduler : Returns New Instance of NotebookScheduler
func NewNotebookScheduler(q *qs.QueryServer, keeper *pods.Keeper, config *config.Config) *NotebookScheduler {
	pending := make(map[int]uint64)
	active := make(map[uint64]interface{})

	schConfig := config.GetNbScheduler()

	nbs := NotebookScheduler{
		concurrency:    schConfig.Concurrency,
		Active:         active,
		WaitList:       pending,
		qh:             newQueryHandler(q),
		podkeeper:      keeper,
		notebookConfig: config.GetNotebooks(),
		schedulerConf:  schConfig,
		jobConfig:      config.GetJobs(),
	}

	nbs.quit = make(chan int)
	nbs.events = make(chan interface{})
	nbs.checkrun = make(chan int)

	if keeper != nil {
		base.Out("Running Notebook Scheduler Daemon")
		go nbs.run()
	} else {
		base.Warn("Schedule Sweeper not initiated as k8s is not reachable ")
	}
	return &nbs
}

// Enqueue : schedule notebook
func (nb *NotebookScheduler) enqueue(podID uint64) error {
	if podID == 0 {
		return fmt.Errorf("Invalid Pod Record. No ID ")
	}

	nb.wlock.Lock()
	newIndex := len(nb.WaitList) + 1
	nb.WaitList[newIndex] = podID
	nb.dirty = true
	nb.wlock.Unlock()
	return nil
}

func (nb *NotebookScheduler) findWaitIndex(podID uint64) (int, error) {

	for k, id := range nb.WaitList {
		if id == podID {
			return k, nil
		}
	}
	return 0, fmt.Errorf("Pod does not exist in pending queue")
}

func (nb *NotebookScheduler) removeFromWL(podID uint64) error {
	if podID == 0 {
		return fmt.Errorf("Invalid Pod Record. No ID ")
	}

	index, err := nb.findWaitIndex(podID)
	if err != nil {
		base.Warn("Notebook scheduler failed to find pod " + fmt.Sprintf("%d", podID) + " in pending queue :" + err.Error())
		return nil
	}

	nb.wlock.Lock()
	delete(nb.WaitList, index)
	nb.dirty = true
	nb.wlock.Unlock()
	return nil
}

func (nb *NotebookScheduler) markActive(podID uint64) error {
	// move out of pendin queuq
	if podID == 0 {
		return fmt.Errorf("Invalid Pod ID")
	}

	if err := nb.removeFromWL(podID); err != nil {
		return nil
	}

	nb.alock.Lock()
	nb.Active[podID] = emptyStruct
	nb.dirty = true
	nb.alock.Unlock()
	return nil
}

func (nb *NotebookScheduler) markDone(podID uint64) error {
	if podID == 0 {
		return fmt.Errorf("Invalid Pod ID")
	}

	if _, ok := nb.Active[podID]; !ok {
		return nil
	}

	nb.alock.Lock()
	delete(nb.Active, podID)
	nb.dirty = true
	nb.alock.Unlock()
	return nil
}

// saveState: Saves the WaitList/active queues to disk
func (nb *NotebookScheduler) saveState() error {
	if nb.dirty {
		nbcopy := NotebookScheduler{}
		nbcopy.Active = make(map[uint64]interface{})
		for key, value := range nb.Active {
			nbcopy.Active[key] = value
		}

		nbcopy.WaitList = make(map[int]uint64)
		for key, value := range nb.WaitList {
			nbcopy.WaitList[key] = value
		}

		return nb.qh.UpsertSchedulerQueues(&nbcopy)
	}
	return nil
}

// restoreState: from database after server restart
func (nb *NotebookScheduler) restoreState() error {
	storedqueues, err := nb.qh.GetSchedulerQueues()

	if err != nil {
		if !qs.IsErrRecNotFound(err) {
			return fmt.Errorf("Failed to restore scheduler queues from database")
		}
		base.Warn("No Stored Notebook Schedules found. Skipping restore..")
		return nil
	}
	base.Log("Stored queues:", len(storedqueues.WaitList), len(storedqueues.Active))

	if len(storedqueues.WaitList) != 0 {
		nb.wlock.Lock()

		for k, v := range storedqueues.WaitList {
			if _, ok := nb.WaitList[k]; !ok {
				qsize := len(nb.WaitList)
				nb.WaitList[qsize+k] = v
			}
		}
		nb.wlock.Unlock()
	}

	if len(storedqueues.Active) != 0 {
		nb.alock.Lock()

		for k, v := range storedqueues.Active {
			if _, ok := nb.Active[k]; !ok {
				nb.Active[k] = v
			}
		}
		nb.alock.Unlock()
	}

	return nil
}

// Scheduler : Main go routine to keep track of notebooks schedules
func (nb *NotebookScheduler) run() {

	base.Log("Initiating Notebook Scheduler....")

	if err := nb.restoreState(); err != nil {
		base.Error("Failed to restore notebook scheduler queues: " + err.Error())
	}

	saveTicker := time.NewTicker(time.Duration(SaveStateInterval) * time.Second)
	sweepTicker := time.NewTicker(time.Duration(SweepInterval) * time.Minute)

	// channel to record POD shutdown
	nb.podShutdown = make(chan struct{})

	nb.qh.TrackPodChanges(nb.quit, nb.events)

	// go func() {
	//	for {
	//		select {
	//		case _, _ = <-nb.checkrun:
	//			fmt.Println("received checkrun")
	//			nb.checkAndStart()
	//
	//		}
	//	}
	//}()

	for {
		select {

		case event, ok := <-nb.events:
			if !ok {
				base.Error("Failed to capture pod change")
				continue
			}

			base.Log("[NB Scheduler POD changed] -> ", event)
			pod, ok := event.(*types.POD)

			if !ok {
				base.Error("Notebook Scheduler failed to read POD info from change event")
			}

			if pod != nil {
				if pod.RequestMode == types.PodReqModeImd {
					// skip for non-scheduled request
					continue
				}

				if pod.IsDone() {

					// cleanup from K8S
					if err := nb.podkeeper.CleanupJob(pod.PodType, pod.ID); err != nil {
						base.Error("Job Cleanup failure: " + err.Error())
					}

					// dequeue
					err := nb.markDone(pod.ID)
					if err != nil {
						base.Error("Notebook Scheduler failed to dequeue pod " + fmt.Sprintf("%d", pod.ID) + " from active list: " + err.Error())
						continue
					}

					// send slot open signal
					nb.podShutdown <- struct{}{}
				}

			}

		case <-nb.podShutdown:
			// call the method to pick next one in the pending
			// and launch if concurrency condition is met
			nb.checkAndStart()

		case <-saveTicker.C:
			if len(nb.WaitList) > 0 || len(nb.Active) > 0 {
				base.Log("Saving schedule queues: ", len(nb.WaitList), len(nb.Active))
				if err := nb.saveState(); err != nil {
					base.Error("Failed to save Notebook scheduler state: " + err.Error())
				}
			}
		case <-sweepTicker.C:
			go func() {
				if err := nb.sweepLists(); err != nil {
					base.Error("Notebook Scheduled failed to sweep through active/wait list: " + err.Error())
				}
			}()

		case <-nb.quit:
			return
		}
	}

	// Listen to POD status updates
	// 		if pod is done or terminated:
	// 			remove it from active list
	// 			initiate launch of new one

	// a new pod arrives
	// 		enqueue
	// 		initiate pod launcher

	// pod launcher
	// 		compare active queue length with concurrency
	// 		if slots are available then
	// 			initiate pod
	//			alter queues
	// 		else
	// 			do nothing. wait for next signal

	// Add timer tick to save state to database

	// Add a time to just go through each pending/active entry and confirm it makes sense

}

// claimSlot : Adds pod ID to active queue removing from pending queue
func (nb *NotebookScheduler) claimSlot(podID uint64) error {
	return nb.markActive(podID)
}

func (nb *NotebookScheduler) slotAvailable() bool {
	if nb.concurrency == 0 {
		return true
	}

	if len(nb.Active) < nb.concurrency {
		return true
	}

	return false
}

func (nb *NotebookScheduler) nextPod() uint64 {
	for _, podID := range nb.WaitList {
		if podID != 0 {
			return podID
		}
	}

	return 0
}

// checkAndStart : Check if slot is available and launch POD if it is
func (nb *NotebookScheduler) checkAndStart() {
	var err error

	if !nb.slotAvailable() {
		return
	}

	if len(nb.WaitList) == 0 {
		return
	}

	processed := 0
	var nextPodID uint64
	for processed < 1 {
		nextPodID = nb.nextPod()

		if nextPodID == 0 {
			continue
		}

		err := nb.claimSlot(nextPodID)
		if err != nil {
			base.Warn("Notebook scheduler failed to claim Slot for POD " + toString(nextPodID) + ": " + err.Error())
		}
		processed = 1
	}

	// submitPod
	_, err = nb.processRequest(nextPodID)
	if err != nil {
		// TODO: this in parallel

		// remove from active list
		if err := nb.markDone(nextPodID); err == nil {

			// add to pending list
			_ = nb.enqueue(nextPodID)
		}

	}

	return

}

// ScheduleRequest : Adds notebook request in the queue
func (nb *NotebookScheduler) ScheduleRequest(user *types.User, pod *types.POD) (*types.POD, error) {
	var err error

	pod.SetRequestMode(types.PodReqModeBck)
	pod.SetStatus(types.PodReqScheduled, nil)

	pod, err = nb.qh.UpsertPOD(pod)
	if err != nil {
		pod.SetStatus(types.PodReqFailed, err)
		return pod, err
	}

	// 	queue up pod
	if err = nb.enqueue(pod.ID); err != nil {
		pod.SetStatus(types.PodReqFailed, err)
		return pod, err
	}

	nb.checkAndStart()
	//nb.checkrun <- 1
	return pod, nil
}

// processRequest : Submit POD requests when slot is ready
func (nb *NotebookScheduler) processRequest(podID uint64) (*types.POD, error) {

	pod, err := nb.qh.GetPOD(podID)
	if err != nil {
		if qs.IsErrRecNotFound(err) {
			return nil, fmt.Errorf("Failed to locate scheduled POD request. Please re-submit the request")
		}
		return pod, err
	}
	pod, err = nb.podkeeper.CreateJob(pod, nb.jobConfig)
	if err != nil {
		return pod, err
	}
	return pod, nil
}

func (nb *NotebookScheduler) sweepLists() error {

	base.Log("Sweeping through Scheduler Queues..", len(nb.WaitList), len(nb.Active))
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for _, pid := range nb.WaitList {
			pod, err := nb.qh.GetPOD(pid)

			if err != nil {
				base.Error("Notebook scheduler failed to get POD info: " + fmt.Sprintf("%d", pid) + ": " + err.Error())
			}

			if !pod.IsPending() {
				nb.events <- pod
			}
		}
	}()
	wg.Add(1)

	go func() {
		defer wg.Done()
		for pid := range nb.Active {
			pod, err := nb.qh.GetPOD(pid)
			if err != nil {
				base.Error("Notebook scheduler failed to get POD info: " + fmt.Sprintf("%d", pid) + ": " + err.Error())
			}

			if pod.IsDone() {
				nb.events <- pod
			}
		}
	}()

	wg.Wait()
	return nil
}
