package rest

import (
	"encoding/json"
	"fmt"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	"github.com/hyper-ml/hyperml/server/pkg/types"
	"io/ioutil"
	"net/http"
	"strconv"
)

func (h *Handler) handleNewBckNotebook() (fnerr error) {
	var err error
	var nbreq NotebookInfo

	defer func() {
		fmt.Println("err:", fnerr)
	}()

	if err := ValidateMethod(h.rq.Method, "POST"); err != nil {
		return err
	}

	if !h.server.ReadyForRequests() {
		return raisePreconditionFailed("Server is not ready for new requests. Please contact system administrator.")
	}

	rawInput, err := ioutil.ReadAll(h.getReqBody())
	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Failed to read JSON inputbytes")
	}

	if err := json.Unmarshal(rawInput, &nbreq); err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Failed to unmarshal JSON input")
	}

	if errStr := validateNotebookReq(nbreq); errStr != "" {
		return base.HTTPErrorf(http.StatusPreconditionFailed, errStr)
	}

	containerImageID := nbreq.ContainerImageID
	if containerImageID == 0 {
		createdImage, _ := h.server.qs.GetOrCreateContainerImage(nbreq.ContainerImage)
		if createdImage != nil {
			containerImageID = createdImage.ID
		} else {
			return base.HTTPErrorf(http.StatusPreconditionFailed, "Failed to register new container image")
		}
	}

	nbInstance, err := h.server.usrReq.ProcBackgroundRequest(h.SessionUser(), nbreq.ResourceProfileID, containerImageID, nbreq.Params)

	if nbInstance == nil {
		return raisePreconditionFailed("Failed to Launch : " + err.Error())
	}

	if err != nil {
		if nbInstance.FailureReason == nullString {
			nbInstance.FailureReason = err.Error()
		}
	}

	h.writeJSON(newNotebookInfo(nbInstance))

	return nil
}

func (h *Handler) handleGetBckNotebook() error {
	if err := ValidateMethod(h.rq.Method, "GET"); err != nil {
		return err
	}

	sessionUser := h.SessionUser()

	if sessionUser == nil {
		return raiseAuthorizationErr("Invalid user")
	}

	idstr, err := h.getMandatoryURLParam("id")

	if err != nil || idstr == nullString {
		return raiseBadRequest("Invalid or empty Notebook ID")
	}

	nbid, err := strconv.ParseUint(idstr, 10, 64)
	if err != nil {
		return raiseBadRequest("Invalid or empty Notebook ID")
	}

	nbInstance, err := h.server.usrReq.GetBckNotebook(sessionUser, nbid)
	if err != nil {
		fmt.Println("Failed to get Notebook:", err)
		return err
	}

	if nbInstance == nil {
		return raisePreconditionFailed("Invalid request: Notebook does not exists")
	}

	nbInfo := newNotebookInfo(nbInstance)
	h.writeJSON(nbInfo)
	return nil

}

func (h *Handler) handleGetBckNotebookStatus() error {
	if err := ValidateMethod(h.rq.Method, "GET"); err != nil {
		return err
	}

	sessionUser := h.SessionUser()

	if sessionUser == nil {
		return raiseAuthorizationErr("Invalid user")
	}

	idstr, err := h.getMandatoryURLParam("id")

	if err != nil || idstr == nullString {
		return raiseBadRequest("Invalid or empty Notebook ID")
	}

	nbid, err := strconv.ParseUint(idstr, 10, 64)
	if err != nil {
		return raiseBadRequest("Invalid or empty Notebook ID")
	}
	nbInstance, err := h.server.usrReq.GetBckNotebookStatus(sessionUser, nbid)

	if err != nil {
		fmt.Println("Failed to get Notebook:", err)
		return err
	}

	if nbInstance == nil {
		return raisePreconditionFailed("Invalid request: Notebook does not exists")
	}

	nbInfo := newNotebookInfo(nbInstance)
	h.writeJSON(nbInfo)
	return nil
}

func (h *Handler) handleStopBckNotebook() error {

	if err := ValidateMethod(h.rq.Method, "PUT"); err != nil {
		return err
	}

	if !h.server.ReadyForRequests() {
		return raisePreconditionFailed("Server is not ready for new requests. Please contact system administrator.")
	}

	sessionUser := h.SessionUser()

	if sessionUser == nil {
		return raiseAuthorizationErr("Invalid user")
	}

	idstr, err := h.getMandatoryURLParam("id")

	if err != nil || idstr == nullString {
		return raiseBadRequest("Invalid or empty Notebook ID")
	}

	nbid, err := strconv.ParseUint(idstr, 10, 64)
	if err != nil {
		return raiseBadRequest("Invalid or empty Notebook ID")
	}

	nbInstance, err := h.server.usrReq.StopBckNotebook(sessionUser, nbid)
	if err != nil {
		fmt.Println("Failed to stop Notebook:", err)
		return err
	}

	if nbInstance == nil {
		return raisePreconditionFailed("Invalid request: Notebook does not exists")
	}

	nbInfo := newNotebookInfo(nbInstance)
	h.writeJSON(nbInfo)
	return nil
}

func (h *Handler) handleListBckNotebooks() error {

	var nblist []*NotebookInfo

	if err := ValidateMethod(h.rq.Method, "GET"); err != nil {
		return err
	}

	sessionUser := h.SessionUser()

	if sessionUser == nil {
		return raiseAuthorizationErr("Invalid user")
	}

	nbinstances, err := h.server.usrReq.ListNotebooks(sessionUser)

	if err != nil {
		if len(nbinstances) == 0 {
			return err
		}
		base.Error("Failed to retrieve all notebooks for user (" + sessionUser.Name + ") :" + err.Error())
	}

	if nbinstances != nil {
		for _, nb := range nbinstances {
			if nb.RequestMode == types.PodReqModeBck {
				n := newNotebookInfo(nb)
				nblist = append(nblist, &n)
			}
		}
	}
	h.writeJSON(nblist)
	return nil
}
