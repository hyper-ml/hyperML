package rest

import (
	"encoding/json"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	"github.com/hyper-ml/hyperml/server/pkg/base/structs"
	ws "github.com/hyper-ml/hyperml/server/pkg/workspace"

	"io/ioutil"
	"net/http"
)

func (h *Handler) handleGetFileContent() error {
	return h.handleGetObject()
}

func (h *Handler) handleGetFileAttrs() error {
	var response map[string]interface{}
	repoName := h.getQuery("repoName")
	commitID := h.getQuery("commitID")
	fpath := h.getQuery("path")

	if repoName == "" {
		return base.HTTPErrorf(http.StatusInternalServerError, "Invalid repo param - repoName")
	}

	if commitID == "" {
		return base.HTTPErrorf(http.StatusInternalServerError, "Invalid repo param - commitID")
	}

	fileAttrs, err := h.server.wsAPI.GetFileAttrs(repoName, commitID, fpath)

	if err == nil {
		//TODO: handle nil file info
		response = structs.Map(fileAttrs)
	} else {
		return err
	}

	h.writeJSON(response)
	return nil
}

func (h *Handler) handleGetFileURL() error {
	var err error
	f := ws.FileMessage{}

	inputs, err := ioutil.ReadAll(h.rq.Body)
	if err := json.Unmarshal(inputs, &f); err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "failed to decode input FileMessage: %v", err)
	}

	if f.MessageType == "" || f.Branch == nil || f.Repo == nil || f.Commit == nil || f.File == nil {
		return base.HTTPErrorf(http.StatusBadRequest, "One or more required params are empty ")
	}

	switch f.MessageType {
	case http.MethodPut:
		f.PutURL, err = h.server.wsAPI.PutFileURL(f.Repo.Name, f.Branch.Name, f.Commit.Id, f.File.Path)
	case http.MethodGet:
		f.GetURL, err = h.server.wsAPI.GetFileURL(f.Repo.Name, f.Commit.Id, f.File.Path)
	}

	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "failed to generate put file signed url: %v", err)
	}

	response := structs.Map(f)
	h.writeJSON(response)
	return nil
}

func (h *Handler) handleGetFilePartsURL() error {
	var err error
	pm := ws.FilePartMessage{}

	inputs, err := ioutil.ReadAll(h.rq.Body)
	if err := json.Unmarshal(inputs, &pm); err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "failed to decode input FilePartMessage: %v", err)
	}

	if pm.MessageType == "" || pm.Branch == nil || pm.Repo == nil || pm.Commit == nil || pm.File == nil {
		return base.HTTPErrorf(http.StatusBadRequest, "One or more required params are empty ")
	}

	pm.PutURL, err = h.server.wsAPI.PutFilePartURL(pm.Seq, pm.Repo.Name, pm.Branch.Name, pm.Commit.Id, pm.File.Path)
	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "failed to generate URL for file part, err: %v", err)
	}

	response := structs.Map(pm)
	h.writeJSON(response)
	return nil
}

func (h *Handler) handleFileCheckIn() error {
	var err error

	inputs, err := ioutil.ReadAll(h.rq.Body)

	a := ws.FileAttrsMessage{}
	if err := json.Unmarshal(inputs, &a); err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "failed to decode input FilePartMessage: %v", err)
	}

	// Check mandatory params: RepoName, commitID, FilePath
	//
	if a.Repo == nil || a.Branch == nil || a.Commit == nil || a.FileAttrs == nil {
		return base.HTTPErrorf(http.StatusBadRequest, "One or more required params are empty ")
	}

	a.FileAttrs, err = h.server.wsAPI.FileCheckIn(a.Repo.Name, a.Branch.Name, a.Commit.Id, a.FileAttrs.File.Path, a.FileAttrs.SizeBytes, "")
	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "failed checking in file: %v", err)
	}

	response := structs.Map(a)
	h.writeJSON(response)
	return nil
}

func (h *Handler) handleFileMerge() error {

	inputs, err := ioutil.ReadAll(h.rq.Body)
	if err != nil {
		base.HTTPErrorf(http.StatusBadRequest, "failed to read request body, err: "+err.Error())
	}

	p := ws.FilePartsMessage{}
	if err := json.Unmarshal(inputs, &p); err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "failed to decode input FilePartMessage: %v", err)
	}

	// Check mandatory params:
	// Sequences, RepoName, Commit Id, FilePath, File Size

	if len(p.Sequences) == 0 || p.Repo == nil || p.Branch == nil || p.File == nil || p.Commit == nil || p.FileAttrs == nil {
		return base.HTTPErrorf(http.StatusBadRequest, "One or more required params are empty ")
	}

	p.FileAttrs, err = h.server.wsAPI.FileMergeAndCheckIn(p.Repo.Name, p.Branch.Name, p.Commit.Id, p.File.Path, p.Sequences, p.FileAttrs.SizeBytes)
	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "failed to merge and checkin, err: "+err.Error())
	}

	response := structs.Map(p)
	h.writeJSON(response)
	return nil
}

func (h *Handler) handlePutFile() error {

	var response map[string]interface{}
	var fileAttrs *ws.FileAttrs
	var err error
	var written int64

	repoName := h.getQuery("repoName")
	branchName := h.getQuery("branchName")
	commitID := h.getQuery("commitID")
	filePath := h.getQuery("path")
	objectHash := h.getQuery("hash")

	if h.rq.Body == nil {
		h.writeJSON(response)
		return nil
	}

	if objectHash == "" {
		fileAttrs, written, err = h.server.wsAPI.PutFile(repoName, branchName, commitID, filePath, h.rq.Body)
	} else {
		return base.HTTPErrorf(http.StatusBadRequest, "unimplemented")
	}

	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, err.Error())
	}

	putFileResponse := ws.PutFileResponse{
		FileAttrs: fileAttrs,
		Written:   written,
	}

	response = structs.Map(putFileResponse)
	h.writeJSON(response)
	return nil
}
