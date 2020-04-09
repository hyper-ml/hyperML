package rest

import (
	"fmt"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	"github.com/hyper-ml/hyperml/server/pkg/base/structs"
	"net/http"

	ws "github.com/hyper-ml/hyperml/server/pkg/workspace"
)

func (h *Handler) handleGetRepoAttrs() error {
	var response map[string]interface{}
	repoName := h.getQuery("repoName")

	if repoName == "" {
		return base.HTTPErrorf(http.StatusInternalServerError, "Invalid repo param - repoName")
	}

	//TODO: handle error
	repoAttrs, err := h.server.wsAPI.GetRepoAttrs(repoName)

	if err == nil {
		response = structs.Map(repoAttrs)
	} else {
		return err
	}

	h.writeJSON(response)

	return nil
}

func (h *Handler) handleGetBranchAttrs() error {
	var response map[string]interface{}
	repoName := h.getQuery("repoName")

	if repoName == "" {
		return base.HTTPErrorf(http.StatusInternalServerError, "Invalid repo param - repoName: %s", repoName)
	}

	branchName := h.getQuery("branchName")

	if branchName == "" {
		return base.HTTPErrorf(http.StatusInternalServerError, "Invalid branch param - branchName: %s", branchName)
	}

	//TODO: handle error
	branchAttr, err := h.server.wsAPI.GetBranchAttrs(repoName, branchName)

	if err == nil {
		response = structs.Map(branchAttr)
	} else {
		return err
	}

	fmt.Println("response on handleGetRepoAttrs: ", response)
	h.writeJSON(response)

	return nil
}

func (h *Handler) handleDeleteRepo() error {
	// to do:
	if h.rq.Method != "DELETE" {
		return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
	}

	var response map[string]interface{}
	repoName := h.getQuery("repoName")

	err := h.server.wsAPI.RemoveRepo(repoName)
	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Failed to create repo: "+err.Error())
	}

	response = map[string]interface{}{}
	h.writeJSON(response)
	return nil

}

func (h *Handler) handlePostModelRepo() error {
	if h.rq.Method != "POST" {
		return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
	}

	var response map[string]interface{}
	repoName := h.getQuery("repoName")

	_, err := h.server.wsAPI.InitTypedRepo(ws.MODEL_REPO, repoName)
	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Failed to create repo: "+err.Error())
	}

	response = map[string]interface{}{
		"name": repoName,
	}
	h.writeJSON(response)
	return nil
}

func (h *Handler) handlePostOutRepo() error {
	if h.rq.Method != "POST" {
		return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
	}

	var response map[string]interface{}
	repoName := h.getQuery("repoName")

	_, err := h.server.wsAPI.InitTypedRepo(ws.OUTPUT_REPO, repoName)
	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Failed to create repo: "+err.Error())
	}

	response = map[string]interface{}{
		"name": repoName,
	}
	h.writeJSON(response)
	return nil
}

func (h *Handler) handleGetRepo() error {
	var response map[string]interface{}
	repoName := h.getQuery("repoName")

	if repoName == "" {
		repoName, _ = h.getMandatoryURLParam("repoName")
		if repoName == "" {
			return base.HTTPErrorf(http.StatusBadRequest, "[GetRepo] Repo name is mandatory")
		}
	}

	repo, err := h.server.wsAPI.GetRepo(repoName)
	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Failed to get repo: "+err.Error())
	}
	response = structs.Map(repo)
	h.writeJSON(response)

	return nil
}

func (h *Handler) handlePostRepo() error {
	if h.rq.Method != "POST" {
		return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
	}

	var response map[string]interface{}
	repoName := h.getQuery("repoName")

	_, err := h.server.wsAPI.InitRepo(repoName)
	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Failed to create repo: "+err.Error())
	}

	response = map[string]interface{}{
		"name": repoName,
	}
	h.writeJSON(response)
	return nil
}

func (h *Handler) handleExplodeRepo() error {
	var response map[string]interface{}
	repoName, _ := h.getMandatoryURLParam("repoName")
	branchName := h.getQuery("branchName")

	if repoName == "" {
		return base.HTTPErrorf(http.StatusBadRequest, "[GetRepo] Repo name is mandatory")
	}

	repo, branch, commit, err := h.server.wsAPI.ExplodeRepo(repoName, branchName)

	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Request failed %s", err.Error())
	}

	repoResponse := &ws.RepoMessage{
		Repo:   repo,
		Branch: branch,
		Commit: commit,
	}

	response = structs.Map(repoResponse)
	h.writeJSON(response)

	return nil

}

func (h *Handler) handleExplodeRepoAttrs() error {
	var response map[string]interface{}
	repoName, _ := h.getMandatoryURLParam("repoName")

	if repoName == "" {
		repoName = h.getQuery("repoName")
	}

	branchName := h.getQuery("branchName")
	commitID := h.getQuery("commitID")

	if repoName == "" {
		return base.HTTPErrorf(http.StatusBadRequest, "[GetRepo] Repo name is mandatory")
	}

	repoAttrs, branchAttrs, commitAttrs, fileMap, err := h.server.wsAPI.ExplodeRepoAttrs(repoName, branchName, commitID)

	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Request failed %s", err.Error())
	}

	returnVal := &ws.RepoAttrsMessage{
		RepoAttrs:   repoAttrs,
		BranchAttrs: branchAttrs,
		CommitAttrs: commitAttrs,
		FileMap:     fileMap,
	}

	response = structs.Map(returnVal)
	h.writeJSON(response)

	return nil

}

func (h *Handler) handleGetModel() error {
	if h.rq.Method != "GET" {
		return base.HTTPErrorf(http.StatusMethodNotAllowed, "[Handler.handleGetModel] Invalid method %s", h.rq.Method)
	}

	repoName, _ := h.getMandatoryURLParam("repoName")
	branchName, _ := h.getMandatoryURLParam("branchName")
	commitID, _ := h.getMandatoryURLParam("commitID")
	if repoName == "" || commitID == "" {
		return base.HTTPErrorf(http.StatusBadRequest, "missing_id_param: repoName or commit id params is missing %s %s", repoName, commitID)
	}

	// create or get output repo for the flow
	repo, branch, commit, err := h.server.wsAPI.GetModel(repoName, branchName, commitID)

	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Request failed %s", err.Error())
	}

	modelResp := &ws.RepoMessage{
		Repo:   repo,
		Branch: branch,
		Commit: commit,
	}

	response := structs.Map(modelResp)
	h.writeJSON(response)

	return nil
}

func (h *Handler) handleGetOrCreateModel() error {

	if h.rq.Method != "POST" && h.rq.Method != "GET" {
		return base.HTTPErrorf(http.StatusMethodNotAllowed, "[Handler.handleGetOrCreateModel] Invalid method %s", h.rq.Method)
	}

	repoName, _ := h.getMandatoryURLParam("repoName")
	branchName, _ := h.getMandatoryURLParam("branchName")
	commitID, _ := h.getMandatoryURLParam("commitID")

	if repoName == "" || commitID == "" {
		return base.HTTPErrorf(http.StatusBadRequest, "missing_id_param: repoName or commit id params is missing %s %s", repoName, commitID)
	}

	// create or get output repo for the flow
	repo, branch, commit, err := h.server.wsAPI.GetOrCreateModel(repoName, branchName, commitID)

	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Request failed %s", err.Error())
	}

	modelResp := &ws.RepoMessage{
		Repo:   repo,
		Branch: branch,
		Commit: commit,
	}

	response := structs.Map(modelResp)
	h.writeJSON(response)

	return nil
}
