package rest

import (
	"github.com/hyper-ml/hyperml/server/pkg/base"
	"io"
	"net/http"
)

// raiseError : raise HTTP error
func raiseError(msg string) error {
	return base.HTTPErrorf(http.StatusInternalServerError, msg)
}

func raisePreconditionFailed(msg string) error {
	return base.HTTPErrorf(http.StatusPreconditionFailed, msg)
}

func raisePreconditionFailedErr(err error) error {
	return base.HTTPErrorf(http.StatusPreconditionFailed, err.Error())
}

func raiseAuthorizationErr(msg string) error {
	return base.HTTPErrorf(http.StatusUnauthorized, msg)
}

func raiseBadRequest(msg string) error {
	return base.HTTPErrorf(http.StatusBadRequest, msg)
}

// ValidateMethod :  check if HTTP method is permitted
func ValidateMethod(request, expected string) error {
	if expected != request {
		return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", request)
	}
	return nil
}

func (h *Handler) getReqBody() io.Reader {
	return h.rq.Body
}

func (h *Handler) handleRoot() error {
	response := map[string]interface{}{
		"hyperflow": "version 0.1",
	}

	h.writeJSON(response)
	return nil
}
