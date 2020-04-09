package rest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/hyper-ml/hyperml/server/pkg/base"
	"github.com/hyper-ml/hyperml/server/pkg/base/structs"
	flow_pkg "github.com/hyper-ml/hyperml/server/pkg/flow"
)

//These are worker specific API functions

func (h *Handler) handleUpdateTaskStatus() error {
	if h.rq.Method != "PATCH" {
		return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
	}

	// declare in/out variables
	var response map[string]interface{}
	var rawInput []byte
	var err error

	workerID, _ := h.getMandatoryURLParam("workerId")

	if workerID == "" {
		return base.HTTPErrorf(http.StatusBadRequest, "No worker Id")
	}

	if h.rq.Body == nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Empty status change request")
	}

	// read parameters
	rawInput, err = ioutil.ReadAll(h.rq.Body)

	// create a status change request from raw json body
	changeReq := flow_pkg.TaskStatusChangeRequest{}

	if err := json.Unmarshal(rawInput, &changeReq); err != nil {
		base.Log("[rest.flow.UpdateTaskStatus] Invalid JSON for TaskStatusChangeRequest: ", err)
		return err
	}

	//base.Log("[Handler.handleUpdateTaskStatus] Worker Id, New Status: ", workerId, changeReq.TaskStatus)

	// call internal API
	worker := flow_pkg.Worker{Id: workerID}
	taskStatusResponse, err := h.server.flowAPI.UpdateWorkerTaskStatus(worker, &changeReq)

	if err != nil {
		base.Log("[Handler.handleUpdateTaskStatus] Failed task update: ", changeReq)
		base.Log("[Handler.handleUpdateTaskStatus] ", err)
		return err
	}

	response = structs.Map(taskStatusResponse)
	h.writeJSON(response)

	return nil
}

func (h *Handler) handleRegisterWorker() error {

	if h.rq.Method != "POST" {
		return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
	}

	flowID := h.getQuery("flowId")
	taskID := h.getQuery("taskId")
	ip := h.getQuery("ip")

	var response map[string]interface{}

	workerAttrs, err := h.server.flowAPI.RegisterWorker(flowID, taskID, ip)
	if err != nil {
		return base.HTTPErrorf(http.StatusInternalServerError, err.Error())
	}

	response = structs.Map(workerAttrs)

	h.writeJSON(response)
	return nil

}

func (h *Handler) handleDetachTaskWorker() error {

	if h.rq.Method != "POST" {
		return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
	}

	flowID := h.getQuery("flowId")
	taskID := h.getQuery("taskId")
	workerID := h.getQuery("workerId")

	if flowID == "" || taskID == "" || workerID == "" {
		return base.HTTPErrorf(http.StatusInternalServerError, "One of the params is missing: flowId, taskId, workerId")
	}

	return h.server.flowAPI.DetachTaskWorker(workerID, flowID, taskID)
}
