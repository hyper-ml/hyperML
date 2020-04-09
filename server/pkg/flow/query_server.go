package flow

import (
	"fmt"
	"golang.org/x/sync/errgroup"
	"time"

	"encoding/json"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	db_pkg "github.com/hyper-ml/hyperml/server/pkg/db"
	"github.com/hyper-ml/hyperml/server/pkg/utils"

	. "github.com/hyper-ml/hyperml/server/pkg/tasks"
)

type queryServer struct {
	db db_pkg.DatabaseContext
}

func NewQueryServer(db db_pkg.DatabaseContext) *queryServer {
	return &queryServer{
		db: db,
	}
}

func (qs *queryServer) flowKey(Id string) string {
	return "flow:" + Id
}

// TODO: raise error if flow Id is invalid
func (qs *queryServer) GetFlowAttr(flowId string) (*FlowAttrs, error) {
	flow_key := qs.flowKey(flowId)
	flowAttrs_raw, err := qs.db.Get(flow_key)

	if err != nil {
		base.Log("[queryServer.GetFlowAttr] Flow Record for this ID does not exist. ", flowId)
		return nil, err
	}

	FlowAttrs := FlowAttrs{}
	err = json.Unmarshal(flowAttrs_raw, &FlowAttrs)
	if err != nil {
		base.Log("[queryServer.GetFlowAttr] Failed to convert raw object to Flow Info", flowId)
		return nil, err
	}

	return &FlowAttrs, nil
}

func (qs *queryServer) InsertFlow(flowAttr *FlowAttrs) error {
	flow_key := qs.flowKey(flowAttr.Flow.Id)

	err := qs.db.Insert(flow_key, flowAttr)

	if err != nil {
		base.Log("[flowServer.InsertFlow] Failed to Insert flow:", err)
		return err
	}

	return nil
}

func (qs *queryServer) UpdateFlow(flowId string, flowAttr *FlowAttrs) error {

	flow_key := qs.flowKey(flowId)
	t := Flow{}

	err := qs.db.UpdateAndTrack(flow_key, flowAttr, t)

	if err != nil {
		base.Log("[flowServer.UpdateFlow] Failed to start flow:", err)
		return err
	}

	return nil
}

func validForDelete(status FlowStatus) bool {
	switch status {
	case FLOW_CANCELLED,
		FLOW_CREATED:
		return true
	}
	return false
}

func (qs *queryServer) DeleteFlow(flowId string) error {
	var err error
	flow_key := qs.flowKey(flowId)

	FlowAttrs, err := qs.GetFlowAttr(flowId)
	if err != nil {
		base.Log("[queryServer.DeleteFlow] Invalid flow Id: ", flowId)
		return err
	}

	if !validForDelete(FlowAttrs.Status) {
		base.Log("[queryServer.DeleteFlow] The status of this flow is invalid for delete. Check docs for valid statuses.", flowId)
		return fmt.Errorf("Invalid flow Status for deletion: %s", flowId)
	}

	err = qs.db.SoftDelete(flow_key)
	return err
}

func (qs *queryServer) GetTaskByFlowId(flowId, taskId string) (*TaskAttrs, error) {
	var task_attrs TaskAttrs
	var ok bool

	if flowId == "" {
		return nil, InvalidFlowID()
	}

	flow_attrs, err := qs.GetFlowAttr(flowId)
	if err != nil {
		return nil, err
	}
	if taskId == "" && len(flow_attrs.Tasks) == 1 {
		return flow_attrs.FirstTask(), nil
	} else {
		task_attrs, ok = flow_attrs.Tasks[taskId]
		if !ok {
			return nil, InvalidTaskError(taskId)
		}
	}

	return &task_attrs, nil
}

func (qs *queryServer) UpdateTaskByFlowId(flowId string, attrs TaskAttrs) error {
	task_id := attrs.Task.Id

	flow_attrs, err := qs.GetFlowAttr(flowId)

	if err != nil {
		return err
	}

	flow_attrs.Tasks[task_id] = attrs
	err = qs.UpdateFlow(flowId, flow_attrs)

	if err != nil {
		base.Log("[queryServer.UpdateTaskAttrsByFlowId] Failed to update flow with task attributes: ", err)
		return err
	}

	return nil
}

func taskWorkerKey(flowId, taskId string) string {
	return "flow:" + flowId + ":task:" + taskId + ":worker"
}

func (qs *queryServer) GetWorkerByTaskId(flowId, taskId string) *FlowTaskWorker {
	w_key := taskWorkerKey(flowId, taskId)

	ftw_raw, err := qs.db.Get(w_key)

	if err != nil {
		//base.Log("[queryServer.GetWorkerByTaskId] Failed to retrieve Flow Task Worker record: ", w_key, err)
		return nil
	}

	if ftw_raw == nil {
		return nil
	}

	ftw := FlowTaskWorker{}
	err = json.Unmarshal(ftw_raw, &ftw)
	return &ftw
}

func (qs *queryServer) InsertTaskWorker(f Flow, t Task, w Worker) (*FlowTaskWorker, error) {
	w_key := taskWorkerKey(f.Id, t.Id)

	ftw := &FlowTaskWorker{
		Worker:  w,
		Flow:    f,
		Task:    t,
		Created: time.Now(),
	}

	err := qs.db.Insert(w_key, ftw)
	if err != nil {
		return nil, err
	}

	return ftw, nil
}

func (qs *queryServer) UpsertTaskWorker(f Flow, t Task, w Worker) (*FlowTaskWorker, error) {
	w_key := taskWorkerKey(f.Id, t.Id)

	ftw := &FlowTaskWorker{
		Worker:  w,
		Flow:    f,
		Task:    t,
		Created: time.Now(),
	}

	err := qs.db.Upsert(w_key, ftw)
	if err != nil {
		return nil, err
	}

	return ftw, nil
}

func workerKey(workerId string) string {
	return "worker:" + workerId
}

func (qs *queryServer) GetWorkerAttrs(workerId string) (*WorkerAttrs, error) {
	w_key := workerKey(workerId)

	wattrs_raw, err := qs.db.Get(w_key)
	if err != nil {
		base.Log("[queryServer.GetWorkerAttrs] Failed to fetch worker attributes: ", err)
		return nil, err
	}

	work_attrs := &WorkerAttrs{}
	err = json.Unmarshal(wattrs_raw, work_attrs)

	return work_attrs, nil
}

//register worker
func (qs *queryServer) registerW(flowId string, taskId string, ipAddress string) (*WorkerAttrs, error) {

	if taskId == "" {
		return nil, InvalidTaskID()
	}

	if flowId == "" {
		return nil, InvalidTaskID()
	}

	// check if a worker already exists agains this task
	// current flow task worker record

	var g errgroup.Group
	var cftw_rec *FlowTaskWorker
	var task_attrs *TaskAttrs
	var err error

	g.Go(func() error {
		task_attrs, err = qs.GetTaskByFlowId(flowId, taskId)
		if task_attrs.Status >= TASK_RUNNING {
			base.Log("[queryServer.registerW] Invalid task status: ", task_attrs.Status, flowId, taskId)
			return InvalidTaskStatusError()
		}

		return err
	})

	g.Go(func() error {
		cftw_rec = qs.GetWorkerByTaskId(flowId, taskId)
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	if cftw_rec != nil {
		if cftw_rec.Worker.Id != "" {

			// get current worker attributes
			cw_attrs, _ := qs.GetWorkerAttrs(cftw_rec.Worker.Id)

			if cw_attrs.Ip == ipAddress {
				// requester is registering again. May be after a failure
				// send worker attributes for sync
				return cw_attrs, nil
			} else {
				// requester is different that existing worker on task
				// This could be conflict or due to dead worker
				return nil, fmt.Errorf("[RegisterWorker] A worker is already assigned to this task. Wait for Reaper to collect the failed worker.")
			}
		}
	}

	// TODO: ID no worker for this task. Proceed
	newWorkerId := utils.NewUUID()
	w_key := workerKey(newWorkerId)

	// check if a worker us already registered
	w_attrs := &WorkerAttrs{
		Worker: Worker{
			Id: newWorkerId,
		},
		Flow: Flow{
			Id: flowId,
		},
		Task: Task{
			Id: taskId,
		},
		Started: time.Now(),
		Ip:      ipAddress,
		Status:  WORKER_REGISTERED,
	}

	err = qs.db.Insert(w_key, w_attrs)
	if err != nil {
		return nil, err
	}

	_, err = qs.UpsertTaskWorker(w_attrs.Flow, w_attrs.Task, w_attrs.Worker)
	if err != nil {
		// TODO: delete worker too
		return nil, err
	}

	return w_attrs, nil
}

func (qs *queryServer) DetachTaskWorker(workerId, flowId, taskId string) error {

	w_attrs, err := qs.GetWorkerAttrs(workerId)
	if err != nil {
		base.Log("[queryServer.DetachTaskWorker] Failed to retrieve worker: ", workerId)
		return err
	}

	if w_attrs.Flow.Id != flowId {
		base.Log("[queryServer.DetachTaskWorker] Invalid worker-flow-task combination: ", workerId, flowId, taskId)
		return InvalidWorkerFlowCombo()
	}

	// insert nil worker
	_, err = qs.UpsertTaskWorker(w_attrs.Flow, w_attrs.Task, Worker{})

	if err != nil {
		return err
	}

	return nil
}
