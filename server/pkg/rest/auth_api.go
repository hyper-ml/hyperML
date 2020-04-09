package rest

import (
	"io/ioutil"

	"encoding/json"
	"github.com/hyper-ml/hyperml/server/pkg/auth"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	"net/http"
)

func (h *Handler) handleUserSignup() error {
	if err := ValidateMethod(h.rq.Method, "POST"); err != nil {
		return err
	}

	signupData, err := ioutil.ReadAll(h.rq.Body)
	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Invalid signup request")
	}

	req := UserSignupInfo{}
	if err := json.Unmarshal(signupData, &req); err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Failed to understand signup request")
	}

	sessionAttrs, userAttrs, err := h.server.authAPI.CreateAndLoginUser(req.UserName, req.Email, req.Password)
	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, ": {"+err.Error()+"}")
	}

	sessionInfo := SessionInfo{
		UserName: userAttrs.Name,
		Email:    userAttrs.Email,
		JWT:      sessionAttrs.Session.JWT,
		Status:   string(sessionAttrs.Status),
	}

	h.writeJSON(sessionInfo)
	return nil
}

func (h *Handler) handleBasicAuth() error {
	if err := ValidateMethod(h.rq.Method, "POST"); err != nil {
		return err
	}

	authData, err := ioutil.ReadAll(h.rq.Body)
	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Invalid data input")
	}

	authReq := auth.LoginRequest{}
	if err := json.Unmarshal(authData, &authReq); err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Invalid json input")
	}

	sessionAttrs, err := h.server.authAPI.CreateSession(authReq.UserName, authReq.Password)
	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, ": {"+err.Error()+"}")
	}

	sessionInfo := SessionInfo{
		UserName: sessionAttrs.Session.User.Name,
		Email:    "",
		JWT:      sessionAttrs.Session.JWT,
		Status:   string(sessionAttrs.Status),
	}

	h.writeJSON(sessionInfo)
	return nil
}
