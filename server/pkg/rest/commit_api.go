package rest

import (
	"github.com/hyper-ml/hyperml/server/pkg/base"
	"github.com/hyper-ml/hyperml/server/pkg/base/structs"
	ws "github.com/hyper-ml/hyperml/server/pkg/workspace"
	"net/http"
)

func isEmpty(v string) bool {
	if v == "" {
		return true
	}
	return false
}

func (h *Handler) handleGetCommitMap() error {

	var response map[string]interface{}
	repoName := h.getQuery("repoName")
	commitID := h.getQuery("commitID")

	if repoName == "" {
		return base.HTTPErrorf(http.StatusPreconditionFailed, "Invalid repo param - repoName")
	}

	if commitID == "" {
		return base.HTTPErrorf(http.StatusPreconditionFailed, "Invalid commitID param - commitID")
	}

	//TODO: handle error
	commitMap, err := h.server.wsAPI.GetCommitMap(repoName, commitID)

	if err == nil {
		response = structs.Map(commitMap)
	} else {
		return err
	}
	h.writeJSON(response)

	return nil
}

func (h *Handler) handleGetCommitAttrs() error {
	var commitAttrs *ws.CommitAttrs
	var err error

	if h.rq.Method != "GET" {
		return base.HTTPErrorf(http.StatusPreconditionFailed, "Invalid method %s", h.rq.Method)
	}

	var response map[string]interface{}
	repoName := h.getQuery("repoName")
	commitID := h.getQuery("commitID")
	branchName := h.getQuery("branchName")

	if isEmpty(repoName) {
		repoName, _ = h.getMandatoryURLParam("repoName")
	}

	if isEmpty(commitID) {
		commitID, _ = h.getMandatoryURLParam("commitID")
	}

	if isEmpty(repoName) || (isEmpty(commitID) && isEmpty(branchName)) {
		return base.HTTPErrorf(http.StatusBadRequest, "one of these params is missing - repoName, branchName, commitID: %v %v %v", repoName, branchName, commitID)
	}

	switch {
	case !isEmpty(commitID):
		commitAttrs, err = h.server.wsAPI.GetCommitAttrs(repoName, commitID)
	default:
		commitAttrs, err = h.server.wsAPI.GetCommitAttrsByBranch(repoName, branchName)
	}

	if err != nil {
		base.Error("[Handler.handleGetCommitAttrs] Error: ", repoName, commitID, err)
		return base.HTTPErrorf(http.StatusBadRequest, "Failed to retrieve commit attributes for repo, commit ID: %s %s %s", repoName, commitID, err.Error())
	}

	response = structs.Map(commitAttrs)
	h.writeJSON(response)
	return nil
}

func (h *Handler) handleGetOrStartCommit() error {

	if h.rq.Method != "GET" {
		return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
	}

	var response map[string]interface{}
	repoName := h.getQuery("repoName")
	branchName := h.getQuery("branchName")
	commitID := h.getQuery("commitID")

	//TODO: handle error
	commitAttrs, err := h.server.wsAPI.InitCommit(repoName, branchName, commitID)

	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, " Failed to initialize a commit: "+err.Error())
	}

	response = structs.Map(commitAttrs)
	h.writeJSON(response)
	return nil
}

func (h *Handler) handleCloseCommit() error {
	if h.rq.Method != "POST" {
		return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
	}

	var response map[string]interface{}
	repoName := h.getQuery("repoName")
	branchName := h.getQuery("branchName")
	commitID := h.getQuery("commitID")

	err := h.server.wsAPI.CloseCommit(repoName, branchName, commitID)

	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Failed to close commit: %s", err)
	}

	response = map[string]interface{}{
		"status": "success",
	}

	h.writeJSON(response)
	return nil
}

func (h *Handler) handleGetCommitSize() error {
	base.Info("[Handler.handleGetCommitSize]")
	if h.rq.Method != "GET" {
		return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
	}

	var response map[string]interface{}
	repoName, _ := h.getMandatoryURLParam("repoName")
	branchName, _ := h.getMandatoryURLParam("branchName")
	commitID, _ := h.getMandatoryURLParam("commitID")

	if repoName == "" || branchName == "" || commitID == "" {
		return base.HTTPErrorf(http.StatusBadRequest, "invalid_input: One of the params is missing: repoName commitID branchName")
	}
	commitSize, err := h.server.wsAPI.GetCommitSize(repoName, branchName, commitID)

	if err != nil {
		base.Warn("[Handler.handleGetCommitSize] Failed to get commit size: ", err)
		base.HTTPErrorf(http.StatusBadRequest, err.Error())
	}

	sizeResponse := &ws.CommitSizeResponse{}
	sizeResponse.Size = commitSize

	response = structs.Map(sizeResponse)
	h.writeJSON(response)

	return nil
}
