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

// handleStartLab : Initiates Jupyter Lab
// - ResourceProfile
// - ContainerImage
// Out:
// - NotebookInfo

func (h *Handler) handleStartNotebook() error {
	var err error
	var nbreq NotebookInfo

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

	nbInstance, err := h.server.usrReq.ProcessNotebookRequest(h.SessionUser(), nbreq.ResourceProfileID, containerImageID)

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

func serveUserUpdates(sc *ServerContext, w http.ResponseWriter, req *http.Request) error {

	c, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		base.Error("Failed to establish websocket: ", err)
		return base.HTTPErrorf(http.StatusBadRequest, "Failed to Initiate Websocket conn")
	}

	defer c.Close()

	var user *types.User
	for {
		_, authReq, err := c.ReadMessage()
		if err != nil {
			base.Error("Failed to read auth token for websocket:", err)
			return nil
		}

		if authReq != nil {
			wsMessage := make(map[string]interface{})
			err = json.Unmarshal(authReq, &wsMessage)
			if err != nil {
				base.Error("Invalid Auth request in Webscoket: ", err)
				continue
			}

			token := wsMessage["token"].(string)

			user, err = getUserFromToken(token)
			if user != nil {
				break
			}

			if user == nil || err != nil {
				authErr := `{"Error":"Invalid User"}`
				authErrB, _ := json.Marshal(authErr)
				c.WriteJSON(authErrB)
			}
		}
	}

	quit := make(chan int)
	events := make(chan interface{})
	sc.qs.TrackPodByUser(user.Name, quit, events)

	for {
		select {
		case event, ok := <-events:
			fmt.Println("Received a new POD event:", event)
			if !ok {
				fmt.Println("Event Channel failed")
				return nil
			}

			fmt.Println("received an event:", event)
			pod, ok := event.(*types.POD)

			if !ok {
				break
			}
			if pod != nil {
				nbInfo := newNotebookInfo(pod)
				c.WriteJSON(nbInfo)
			}

			fmt.Println("received an updated on POD:", pod)
		case _ = <-quit:
			return nil
		default:

		}
	}

}

func (h *Handler) handleListNotebooks() error {

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
			n := newNotebookInfo(nb)
			nblist = append(nblist, &n)
		}
	}
	h.writeJSON(nblist)
	return nil
}

func (h *Handler) handleStopNotebook() error {

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

	nbInstance, err := h.server.usrReq.StopNotebook(sessionUser, nbid)
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
