package rest

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hyper-ml/hyperml/server/pkg/base"
	"github.com/hyper-ml/hyperml/server/pkg/base/structs"

	ws "github.com/hyper-ml/hyperml/server/pkg/workspace"
)

func serveLogStream(sc *ServerContext, w http.ResponseWriter, r *http.Request) error {

	var err error
	reqParams := mux.Vars(r)
	taskID := reqParams["taskId"]

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		base.Error("websocket connection upgrade error:", err)
		return base.HTTPErrorf(http.StatusBadRequest, "Could not upgrade HTTP connection to websocket")
	}

	defer c.Close()
	reader, err := sc.flowAPI.LogStream(taskID)
	if err != nil {
		base.Error("failed to read log stream: ", err)
		return base.HTTPErrorf(http.StatusBadRequest, "Failed to read log stream: "+err.Error())
	}

	msgBuffer := make([]byte, 128)
	//todo: utils.GetBuffer()
	//defer utils.PutBuffer()

	for {
		n, err := reader.Read(msgBuffer)
		if err != nil {
			if err == io.EOF {

				//should handle any remaining bytes.
				if err = c.WriteMessage(websocket.TextMessage, msgBuffer[:n]); err != nil {
					break
				}

				break
			}

			base.Error(err.Error())

		}

		if err = c.WriteMessage(websocket.TextMessage, msgBuffer[:n]); err != nil {
			return base.HTTPErrorf(http.StatusBadRequest, "Failed to write to websocket: "+err.Error())
		}

	}

	return nil
}

func (h *Handler) handleGetTaskLog() error {
	if h.rq.Method != "GET" {
		return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
	}

	taskID, _ := h.getMandatoryURLParam("taskId")
	logPath := h.server.logAPI.GetObjectPath(h.server.flowAPI.GetTaskLogPath(taskID))

	rs, err := h.server.logAPI.ReadSeeker(logPath, 0, 0)
	if err != nil {
		if err != io.EOF {
			return base.HTTPErrorf(http.StatusBadRequest, "Error occurred while fetching log object.")
		}
	}

	h.setHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", logPath))
	http.ServeContent(h.response, h.rq, logPath, time.Time{}, rs)

	return nil

}

func errLogObjectMissing(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	if strings.Contains(errStr, "object doesn't exist") {
		return true
	}
	return false
}

func (h *Handler) handleGetCommandLog() error {

	if h.rq.Method != "GET" {
		return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
	}

	taskID, _ := h.getMandatoryURLParam("taskId")
	logPath := h.server.flowAPI.GetCommandLogPath(taskID)
	logPath = h.server.logAPI.GetObjectPath(logPath)
	logPath = logPath + ".log"

	rs, err := h.server.logAPI.ReadSeeker(logPath, 0, 0)

	if errLogObjectMissing(err) {
		logPath := h.server.logAPI.GetObjectPath(h.server.flowAPI.GetTaskLogPath(taskID))
		rs, err = h.server.logAPI.ReadSeeker(logPath, 0, 0)
	}

	if err != nil {
		if errLogObjectMissing(err) {
			return nil
		}

		if err != io.EOF {
			return base.HTTPErrorf(http.StatusBadRequest, "Error occurred while fetching log object.")
		}
	}

	h.setHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", logPath))
	http.ServeContent(h.response, h.rq, logPath, time.Time{}, rs)

	return nil
}

func (h *Handler) handlePostCommandLog() error {

	if h.rq.Method != "POST" {
		return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
	}
	var response map[string]interface{}
	taskID, err := h.getMandatoryURLParam("taskId")

	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "One of the params is missing: taskId")
	}

	if h.rq.Body != nil {
		logPath := h.server.flowAPI.GetCommandLogPath(taskID)
		logPath = h.server.logAPI.GetObjectPath(logPath)
		base.Info("[Handler.handlePostTaskLog] logPath: ", logPath)

		objPath, chksm, size, err := h.server.logAPI.SaveObject(logPath, "", h.rq.Body, false)
		if err != nil {
			base.Debug("[Handler.handlePostFlowLog] Error occurred writing log on server ", taskID, err)
			return base.HTTPErrorf(http.StatusBadRequest, "Failed to save log object: %v", err)
		}

		if size == 0 {
			base.Debug("[Handler.handlePostFlowLog] The input file is empty for flow log request: ", taskID)
			return base.HTTPErrorf(http.StatusBadRequest, "Input file is empty")
		}

		obj := &ws.Object{
			Path:     objPath,
			CheckSum: chksm,
			Size:     int(size)}

		response = structs.Map(obj)
		h.writeJSON(response)
		return nil

	}
	base.Debug("[Handler.handlePostFlowLog] Empty request body was sent for task: ", taskID)
	return base.HTTPErrorf(http.StatusBadRequest, "Request body is empty")
}
