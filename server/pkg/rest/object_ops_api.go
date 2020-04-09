package rest

import (
	"fmt"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	db_pkg "github.com/hyper-ml/hyperml/server/pkg/db"
	"io"
	"net/http"
	"strconv"
	"time"
)

func (h *Handler) handleGetObject() error {
	var err error
	if h.rq.Method != "GET" {
		return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
	}

	repoName := h.getQuery("repoName")
	commitID := h.getQuery("commitId")
	filePath := h.getQuery("filePath")
	offsetSr := h.getQuery("offset")
	sizeStr := h.getQuery("size")

	base.Info("[Handler.handleGetObject] Params: ", repoName, commitID, filePath, offsetSr, sizeStr)

	var offset int64
	var size int64

	if offsetSr != "" {
		//string to int64 base 10
		offset, err = strconv.ParseInt(offsetSr, 10, 64)
	}

	if sizeStr != "" {
		//string to int64 base 10
		size, err = strconv.ParseInt(sizeStr, 10, 64)
	}

	fileAttrs, err := h.server.wsAPI.GetFileAttrs(repoName, commitID, filePath)

	if err != nil {
		base.Error("[Handler.handleGetObject] Failed to retrieve file info: ", repoName, commitID, filePath, err)
		return base.HTTPErrorf(http.StatusBadRequest, err.Error())
	}

	objectHash := fileAttrs.Object.Path
	base.Info("[Handler.handleGetObject] objectHash: ", objectHash)
	rs, err := h.server.objAPI.ReadSeeker(objectHash, offset, size)

	if err != nil {
		if err != io.EOF {
			base.Log("Failed to fetch file object: ", objectHash, err)
			return base.HTTPErrorf(http.StatusBadRequest, "Error occurred while fetching the given object.")
		}
		base.Log("handleGetObject(): Reached EOF. Nothing to read. ", objectHash)
		return nil
	}

	h.setHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", filePath))
	h.setHeader("Content-Type", "application/octet-stream")
	http.ServeContent(h.response, h.rq, filePath, time.Time{}, rs)

	return nil
}

// create a new hash for first request from vfs server
// updae commit with hash

// if object hash is available then just check if commit is open and get on with write
// if obj has is null then call put file writer which may return existing or new obj hash
// need really bare minimum code here. Manage some how so other network methods can be
// added in future. Let API Server manage most of the workload ?? make calls depend on values of
// obj has and send body reader

// send obj has to client through file info so next time client
// can send hash and improve write time

func (h *Handler) handlePutObject() error {

	if h.rq.Method != "PUT" {
		return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
	}

	var response map[string]interface{}

	repoName := h.getQuery("repoName")
	commitID := h.getQuery("commitId")
	filePath := h.getQuery("path")

	// TODO: use commit map or raise error if commt is not open

	// TODO: create file info and object cache

	fileAttrs, err := h.server.wsAPI.GetFileAttrs(repoName, commitID, filePath)

	if err != nil {
		if !db_pkg.IsErrRecNotFound(err) {
			base.Log("failed to retrieve file info", err)
			return err
		}

		base.Debug("File not found in commit map. Creating a new entry", repoName, commitID, filePath)
	}

	objectHash := fileAttrs.Object.Path

	if h.rq.Body != nil {
		objPath, chksum, n, err := h.server.objAPI.SaveObject(objectHash, "objects", h.rq.Body, false)
		if err != nil {
			base.Log("Failed to update object on server", filePath)
			return err
		}
		response = map[string]interface{}{
			"obj_path": objPath,
			"size":     n,
			"checksum": chksum,
		}
		h.writeJSON(response)
		return nil
	}

	response = map[string]interface{}{
		"obj_path": "",
		"size":     0,
	}

	h.writeJSON(response)
	return nil
}
