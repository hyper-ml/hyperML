package flow

import (
	"fmt"
	"io"
	"time"
	//"context"
	"golang.org/x/sync/errgroup"

	//"github.com/hyper-ml/hyperml/server/pkg/base/backoff"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	"github.com/hyper-ml/hyperml/server/pkg/config"
	db_pkg "github.com/hyper-ml/hyperml/server/pkg/db"
	"github.com/hyper-ml/hyperml/server/pkg/storage"
	task_pkg "github.com/hyper-ml/hyperml/server/pkg/tasks"
)

type FlowEngine interface {
	StartFlow(flowId, taskId string) (*FlowAttrs, error)
	LaunchFlow(repoName string, branchName string, commitId string, cmdString string, evars map[string]string) (*FlowAttrs, error)
	LogStream(flowId string) (io.ReadCloser, error)
}

type flowEngine struct {
	qs               *queryServer
	db               db_pkg.DatabaseContext
	wpool            WorkerPool
	namespace        string
	defaultImage     string
	dockerPullPolicy string

	masterIp      string
	masterPort    int32
	masterExtPort int32
	// storage details - add later

}

func NewFlowEngine(
	qs *queryServer,
	db db_pkg.DatabaseContext,
	logger storage.ObjectAPIServer,
	c *config.Config) (*flowEngine, error) {

	wp, err := NewWorkerPool(c, db, logger)
	if err != nil {
		return nil, err
	}
	masterIP, err := c.Get("MasterIP")
	if err != nil {
		return nil, err
	}

	// master port on which server listens
	masterPort, err := c.GetInt32("MasterPort")
	if err != nil {
		return nil, err
	}

	masterExternalPort, err := c.GetInt32("MasterExternalPort")
	if err != nil {
		return nil, err
	}

	return &flowEngine{
		qs:            qs,
		db:            db,
		wpool:         wp,
		masterIp:      masterIP,
		masterPort:    masterPort,
		masterExtPort: masterExternalPort,
	}, nil
}

func TaskWorkerExistsError(flowId, taskId string) error {
	return fmt.Errorf("[flowEngine] Worker already executing this flow task: %s %s", flowId, taskId)
}

func ErrTaskComplete() error {
	return fmt.Errorf("invalid_task_update: task already complete")
}

func InvalidFlowIdError(flowId string) error {
	return fmt.Errorf("[flowEngine] Invalid flow Id: %s", flowId)
}

func InvalidFlowParamsError(flowId string) error {
	return fmt.Errorf("[flowEngine] Invalid flow parameter: %s", flowId)
}

// monitor new messages from the worker pod or update on flow status
// end pods or mark flow completion

func (fe *flowEngine) master(quit chan int) {

	event_chan := make(chan interface{})
	lsnr := fe.db.GetListener()
	lsnr.RegisterObject(event_chan, Flow{})
	base.Println("[flowEngine.master] Starting Flow Master")

	w_chan := NewWorkPoolWatcher()

	go fe.wpool.Watch(w_chan)

	for {
		select {
		case evtval, ok := <-event_chan:
			if !ok {
				return
			}

			flow_attrs, ok := evtval.(*FlowAttrs)

			if !ok {
				break
			}

			// stop worker if flow is completed with success or error
			if flow_attrs.IsComplete() {
				_ = fe.wpool.SaveWorkerLog(Worker{}, flow_attrs.Flow)
				_ = fe.wpool.ReleaseWorker(flow_attrs.Flow)
			}

		case pool_evt, ok := <-w_chan:

			if !ok {
				break
			}

			flow := pool_evt.Flow

			switch {
			case pool_evt.Type == WorkerInitError:
				fe.processFlowError(flow.Id, flow.Id, string(pool_evt.Type))
			case pool_evt.Type == WorkerError:
				fe.processFlowError(flow.Id, flow.Id, string(pool_evt.Type))
			}

		case <-quit:
			base.Log("[flowEngine.master] Quiting flow Engine master..")
			return
		}
	}

	return
}

func (fe *flowEngine) StartFlow(flowId, taskId string) (*FlowAttrs, error) {
	flow_attrs, err := fe.qs.GetFlowAttr(flowId)

	if err != nil {
		return nil, InvalidFlowIdError(flowId)
	}

	if !fe.wpool.WorkerExists(flowId, taskId) {
		err = fe.wpool.AssignWorker(taskId, flow_attrs, fe.masterIp, fe.masterPort, fe.masterExtPort)

		if err != nil {
			if err = fe.processFlowError(flowId, taskId, err.Error()); err != nil {
				return flow_attrs, err
			}
			return flow_attrs, err
		}

		return flow_attrs, nil

	} else {
		// TODO: check if worker is active if not then flush it and restart a new worker
		return nil, TaskWorkerExistsError(flowId, taskId)
	}
	return flow_attrs, nil

}

func (fe *flowEngine) processFlowError(flowId, taskId, message string) error {

	var eg errgroup.Group

	f := Flow{
		Id: flowId,
	}

	//release worker from pool
	// check if worker is assigned. Release only then
	eg.Go(func() error {
		if task_worker := fe.qs.GetWorkerByTaskId(flowId, taskId); task_worker == nil {
			base.Error("[flowEngine.processFlowError] No task worker for this flow task :", flowId, taskId)
			return nil
		}

		err := fe.wpool.ReleaseWorker(f)
		return err
	})

	// update status of flow and remove worker assignment
	eg.Go(func() error {
		err := fe.updateTaskStatusWithText(flowId, taskId, task_pkg.TASK_FAILED, message)
		err = fe.workerCleanUp(f)
		return err
	})

	return eg.Wait()
}

func (fe *flowEngine) Logworker(flow Flow, worker Worker) error {

	//var w io.Writer
	err := fe.wpool.SaveWorkerLog(worker, flow)
	return err
}

func (fe *flowEngine) workerCleanUp(flow Flow) error {
	return fe.wpool.ReleaseWorker(flow)
}

// create flow with a single task
func (fe *flowEngine) createSimpleFlow(repoName string, branchName string, commitId string, cmdString string) (*FlowAttrs, error) {
	wdir := "/workspace"

	mount_map := task_pkg.NewMountConfig(repoName, branchName, commitId, wdir, 0)
	flow_config := &FlowConfig{
		MountMap: mount_map,
	}

	// new flow attr rec
	new_flow := NewFlowAttrs(flow_config)

	// insert task
	task_config := task_pkg.NewTaskConfig(cmdString, nil, wdir, mount_map)

	// add task
	_ = new_flow.AddTask(task_config)

	err := fe.qs.InsertFlow(new_flow)

	return new_flow, err
}

func (fe *flowEngine) LaunchFlow(repoName string, branchName string, commitId string, cmdString string, evars map[string]string) (*FlowAttrs, error) {
	// 1. create a new flow - task
	// 2. start flow

	var err error
	flow_attrs, err := fe.createSimpleFlow(repoName, branchName, commitId, cmdString)

	if err != nil {
		base.Error("[fe.LaunchFlow] Failed to create simple flow: ", err)
		return nil, err
	}

	task_attrs := flow_attrs.FirstTask()

	if task_attrs == nil {
		return nil, fmt.Errorf("[LaunchFlow] No task to run")
	}

	flow_attrs, err = fe.StartFlow(flow_attrs.Flow.Id, task_attrs.Task.Id)

	if err != nil {
		return nil, err
	}

	return flow_attrs, nil
}

/*
func (fe *flowEngine) updateWorkerTaskStatus(workerId string, flowId string, taskId string, newStatus  task_pkg.TaskStatus) (error) {

  w:= Worker { Id: workerId }
  f:= Flow { Id: flowId }
  _ = fe.Logworker(f, w)

  return fe.updateTaskStatus(flowId, taskId, newStatus)
}
*/

func (fe *flowEngine) updateTaskStatus(flowId string, taskId string, newStatus task_pkg.TaskStatus) error {
	return fe.updateTaskStatusWithText(flowId, taskId, newStatus, "")
}

func (fe *flowEngine) updateTaskStatusWithText(flowId string, taskId string, newStatus task_pkg.TaskStatus, completionText string) error {

	task_attrs, err := fe.qs.GetTaskByFlowId(flowId, taskId)
	if err != nil {
		return err
	}

	if task_attrs.Status == task_pkg.TASK_COMPLETED {
		return ErrTaskComplete()
	}

	task_attrs.Status = newStatus
	task_attrs.CompletionText = completionText

	switch s := newStatus; s {

	case task_pkg.TASK_CREATED:
		task_attrs.Created = time.Now()

	case task_pkg.TASK_COMPLETED:
		//TODO: should come in the request from worker
		task_attrs.Completed = time.Now()

	case task_pkg.TASK_INITIATED:
		if task_attrs.Completed.IsZero() {
			task_attrs.Started = time.Now()
		} else {
			return errorCompletedTask()
		}

	case task_pkg.TASK_FAILED:
		if task_attrs.Completed.IsZero() {
			task_attrs.Failed = time.Now()
		} else {
			return errorCompletedTask()
		}
	}

	if err := fe.qs.UpdateTaskByFlowId(flowId, *task_attrs); err == nil {
		return nil
	}

	return err
}

func (fe *flowEngine) LogStream(flowId string) (io.ReadCloser, error) {
	return fe.wpool.LogStream(flowId)
}
