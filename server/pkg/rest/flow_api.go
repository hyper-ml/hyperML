package rest

import (
	"encoding/json"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	"github.com/hyper-ml/hyperml/server/pkg/base/structs"
	flow_pkg "github.com/hyper-ml/hyperml/server/pkg/flow"
	ws "github.com/hyper-ml/hyperml/server/pkg/workspace"
	"io/ioutil"
	"net/http"
)

func (h *Handler) handleGetFlowAttrs() error {
	var err error
	var response map[string]interface{}

	if h.rq.Method != "GET" {
		return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
	}

	flowID, _ := h.getMandatoryURLParam("flowId")
	flowAttrs, err := h.server.flowAPI.GetFlowAttr(flowID)

	if err != nil {
		base.Log("[rest.flow.GetFlowAttrs] Failed to retrieve flow attributes. Please check params. ", flowID, err)
		return base.HTTPErrorf(http.StatusBadRequest, err.Error())
	}

	response = structs.Map(flowAttrs)
	h.writeJSON(response)

	return nil
}

// launches a flow for given repo -commit
// async submission

func (h *Handler) handleLaunchFlow() error {

	var response map[string]interface{}
	var rawInput []byte
	var err error

	if err := ValidateMethod(h.rq.Method, "POST"); err != nil {
		return err
	}

	rawInput, err = ioutil.ReadAll(h.rq.Body)
	if err != nil {
		return err
	}

	flowResponse := flow_pkg.FlowMessage{}
	if err := json.Unmarshal(rawInput, &flowResponse); err != nil {
		return err
	}

	if len(flowResponse.Repos) == 0 {
		return base.HTTPErrorf(http.StatusBadRequest, "Invalid method params: Empty repos")
	}

	repoResponse := flowResponse.Repos[0]
	repoName := repoResponse.Repo.Name
	branchName := repoResponse.Branch.Name
	commitID := repoResponse.Commit.Id
	cmdStr := flowResponse.CmdStr
	envVars := flowResponse.EnvVars

	switch {
	case repoName == "":
		return base.HTTPErrorf(http.StatusBadRequest, "Invalid method params: repoName missing")
	case flowResponse.CmdStr == "":
		return base.HTTPErrorf(http.StatusBadRequest, "Invalid method params: command string missing")

	}

	if branchName != "" {
		branchName = "master"
	}

	if commitID == "" {
		cattrs, err := h.server.wsAPI.StartCommit(repoName, branchName)
		if err != nil {
			return base.HTTPErrorf(http.StatusBadRequest, err.Error())
		}
		commitID = cattrs.Id()
	}

	flowAttrs, err := h.server.flowAPI.LaunchFlow(repoName, branchName, commitID, cmdStr, envVars)
	if err != nil {
		base.Log("[Handler.handleLaunchFlow] Failed to launch task: ", err)
		return base.HTTPErrorf(http.StatusBadRequest, err.Error())
	}

	flowResponse.Flow = &flowAttrs.Flow
	flowResponse.FlowStatusStr = flow_pkg.FlowStatusToString(flowAttrs.Status)
	flowResponse.Repos[0].Commit = &ws.Commit{Id: commitID}

	response = structs.Map(flowResponse)
	h.writeJSON(response)

	return nil
}

func (h *Handler) handleGetOutputByFlow() error {

	if h.rq.Method != "GET" {
		return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
	}
	flowID, _ := h.getMandatoryURLParam("flowId")
	base.Info("[Handler.handleGetOutputByFlow] FlowId: ", flowID)

	if flowID == "" {
		return base.HTTPErrorf(http.StatusBadRequest, "Invalid flowId params %s", flowID)
	}

	f := flow_pkg.Flow{Id: flowID}
	repo, branch, commit, err := h.server.flowAPI.GetOutput(f)

	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Request failed %s", err.Error())
	}

	if repo == nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Request failed ")
	}

	repoResponse := &ws.RepoMessage{
		Repo:   repo,
		Branch: branch,
		Commit: commit,
	}

	response := structs.Map(repoResponse)
	h.writeJSON(response)
	return nil
}

func (h *Handler) handleGetOrCreateOutputByFlow() error {
	if h.rq.Method != "POST" {
		return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
	}

	flowID, _ := h.getMandatoryURLParam("flowId")
	if flowID == "" {
		return base.HTTPErrorf(http.StatusBadRequest, "Invalid flowId params %s", flowID)
	}

	f := flow_pkg.FlowRef(flowID)
	repo, branch, commit, err := h.server.flowAPI.GetOrCreateOutput(f)

	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Request failed %s", err.Error())
	}

	if repo == nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Request failed ")
	}

	repoResponse := &ws.RepoMessage{
		Repo:   repo,
		Branch: branch,
		Commit: commit,
	}

	response := structs.Map(repoResponse)
	h.writeJSON(response)

	return nil
}

func (h *Handler) handleGetModelByFlow() error {

	if h.rq.Method != "GET" {
		return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
	}

	flowID, _ := h.getMandatoryURLParam("flowId")
	if flowID == "" {
		return base.HTTPErrorf(http.StatusBadRequest, "Invalid flowId params %s", flowID)
	}

	f := flow_pkg.Flow{Id: flowID}
	repo, branch, commit, err := h.server.flowAPI.GetModel(f)

	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Request failed %s", err.Error())
	}

	if repo == nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Request failed ")
	}

	modelResponse := &ws.RepoMessage{
		Repo:   repo,
		Branch: branch,
		Commit: commit,
	}

	response := structs.Map(modelResponse)
	h.writeJSON(response)
	return nil
}

func (h *Handler) handleGetOrCreateModelByFlow() error {
	if h.rq.Method != "POST" {
		return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
	}

	flowID, _ := h.getMandatoryURLParam("flowId")
	if flowID == "" {
		return base.HTTPErrorf(http.StatusBadRequest, "Invalid flowId params %s", flowID)
	}

	f := flow_pkg.FlowRef(flowID)
	repo, branch, commit, err := h.server.flowAPI.GetOrCreateModel(f)

	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Request failed %s", err.Error())
	}

	if repo == nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Request failed ")
	}

	modelResponse := &ws.RepoMessage{
		Repo:   repo,
		Branch: branch,
		Commit: commit,
	}

	response := structs.Map(modelResponse)
	h.writeJSON(response)

	return nil
}

func (h *Handler) handleGetFlowStatus() error {
	if h.rq.Method != "GET" {
		return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
	}

	flowID, _ := h.getMandatoryURLParam("flowId")
	flowAttrs, err := h.server.flowAPI.GetFlowAttr(flowID)

	if err != nil {
		base.Log("[rest.flow.GetFlowAttrs] Failed to retrieve flow attributes. Please check params. ", flowID, err)
		return base.HTTPErrorf(http.StatusBadRequest, err.Error())
	}

	flowResponse := &flow_pkg.FlowMessage{
		FlowStatusStr: flowAttrs.FlowStatus(),
	}

	response := structs.Map(flowResponse)
	h.writeJSON(response)

	return nil
}
