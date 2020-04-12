package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hyper-ml/hyperml/server/pkg/auth"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	"github.com/hyper-ml/hyperml/server/pkg/types"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

var lastUID uint64 = 0

// Handler : HTTP Request Handler
type Handler struct {
	server        *ServerContext
	session       *types.Session
	rq            *http.Request
	response      http.ResponseWriter
	requestBody   io.ReadCloser
	status        int
	statusMessage string
	privs         HandlerPrivs
	uid           uint64
	startTime     time.Time
	queryParams   url.Values
}

type httpBody map[string]interface{}

// HandlerPrivs  : HandlerPrivs
type HandlerPrivs int

const handelerVersion = "0.1"

const (
	userPrivs = iota
	publicPrivs
	adminPrivs
)

// HandlerMethod : Generic handler method
type HandlerMethod func(*Handler) error

func makeHandler(server *ServerContext, privs HandlerPrivs, method HandlerMethod) http.Handler {

	return http.HandlerFunc(func(response http.ResponseWriter, req *http.Request) {
		h, err := newHandler(server, privs, response, req)
		if err == nil {
			err = h.invoke(method)
		}

		h.writeError(err)
	})
}

// TODO:
// 1. Add user details
func newHandler(server *ServerContext, privs HandlerPrivs, response http.ResponseWriter, request *http.Request) (*Handler, error) {
	handler := &Handler{
		server:    server,
		privs:     privs,
		rq:        request,
		response:  response,
		status:    http.StatusOK,
		startTime: time.Now(),
		uid:       atomic.AddUint64(&lastUID, 1),
	}

	var user *types.User
	var userAttrs *types.UserAttrs
	var err error

	jwt := handler.getAuthToken()

	if jwt != "" {
		jwt = strings.TrimPrefix(jwt, "Bearer ")
		user, err = auth.VerifyToken(jwt)
		if user == nil || err != nil {
			return nil, base.HTTPErrorf(http.StatusUnauthorized, "Unauthorised User")
		}

	} else {

		apiKey := handler.getAPIKey()
		if apiKey != nullString {
			userAttrs, err = server.authAPI.GetUserByAPIKey(apiKey)
			if userAttrs == nil || err != nil {
				return nil, base.HTTPErrorf(http.StatusUnauthorized, "API user is invalid")
			}
			user = userAttrs.User
		}
	}

	if user == nil {
		if server.authAPI.IsAuthEnabled() {
			return nil, base.HTTPErrorf(http.StatusUnauthorized, "User Unauthorized")
		}

		userAttrs, _ = server.authAPI.GetGuestUser()
		user = userAttrs.User
	}

	handler.session = &types.Session{
		User: user,
	}

	return handler, nil
}

func (h *Handler) invoke(method HandlerMethod) error {
	//TODO : add stats logging

	h.setHeader("server", handelerVersion)

	//TODO: add auth privs check
	h.logRequest()
	return method(h)
}

// Utility functions
//

func (h *Handler) writeError(err error) {

	if err != nil {
		status, message := base.ErrorToHTTPStatus(err)
		h.writeStatus(status, message)
	}
}

func (h *Handler) writeStatus(status int, message string) {
	if status < 300 {
		h.response.WriteHeader(status)
		h.setStatus(status, message)
	}

	var errorStr string
	switch status {
	case http.StatusNotFound:
		errorStr = "not_found"
	case http.StatusConflict:
		errorStr = "conflict"
	default:
		errorStr = http.StatusText(status)
		if errorStr == "" {
			errorStr = fmt.Sprintf("%d", status)
		}
	}

	h.setHeader("Content-Type", "application/json")
	//h.response.WriteHeader(status)
	h.setStatus(status, message)
	jsonOut, _ := json.Marshal(&httpBody{"error": errorStr, "reason": message})
	h.response.Write(jsonOut)
}

func (h *Handler) logRequest() {
	var q *bytes.Buffer
	qparams := h.getQueryParams()
	if len(qparams) > 0 {
		q = new(bytes.Buffer)
		for k, val := range qparams {
			fmt.Fprintf(q, "%s=\"%s\"\n", k, val)
		}
	}

	base.Log(h.rq.Method + ": " + h.rq.URL.String() + " " + q.String())
}

func (h *Handler) logRequestBody() {
	fmt.Println("Log request body. TODO")

}

func (h *Handler) getQueryParams() url.Values {
	if h.queryParams == nil {
		h.queryParams = h.rq.URL.Query()
	}
	return h.queryParams
}

func (h *Handler) getQuery(query string) string {
	v, _ := url.PathUnescape(h.getQueryParams().Get(query))
	return v
}

func (h *Handler) getURLParams() map[string]string {
	return muxVars(h.rq)
}

func (h *Handler) getMandatoryURLParam(pname string) (string, error) {
	vars := h.getURLParams()
	pval, ok := vars[pname]
	if !ok {
		return "", base.HTTPErrorf(http.StatusInternalServerError, "Invalid request parameter: "+pname)
	}

	return url.PathUnescape(pval)
}

func (h *Handler) getAuthToken() string {
	return h.rq.Header.Get("Authorization")
}

func (h *Handler) getAPIKey() string {
	return h.rq.Header.Get("api-key")
}

func (h *Handler) requestAccepts(mimetype string) bool {
	accept := h.rq.Header.Get("Accept")
	return accept == "" || strings.Contains(accept, mimetype) || strings.Contains(accept, "*/*")
}

// Response methods

func (h *Handler) setHeader(name string, value string) {
	h.response.Header().Set(name, value)
}

func (h *Handler) setStatus(status int, message string) {
	h.status = status
	h.statusMessage = message
}

func (h *Handler) writeJSON(value interface{}) {
	h.writeJSONStatus(http.StatusOK, value)
}

func (h *Handler) writeJSONStatus(status int, value interface{}) {
	if !h.requestAccepts("application/json") {
		base.Log("client wont accept JSON. only ", h.rq.Header.Get("Accept"))
		h.writeStatus(http.StatusNotAcceptable, "only application/json available")
		return
	}

	jsonOut, err := json.Marshal(value)
	if err != nil {
		base.Log("Couldn't serialize JSON for %v : %s", value, err)
		h.writeStatus(http.StatusInternalServerError, "JSON serialization failed")
		return
	}

	h.setHeader("Content-Type", "application/json")

	if h.rq.Method != "HEAD" {
		//TODO: disable response compression
		h.setHeader("Content-Length", fmt.Sprintf("%d", len(jsonOut)))
		if status > 0 {
			h.response.WriteHeader(status)
			h.setStatus(status, "")
		}
		h.response.Write(jsonOut)
	} else if status > 0 {
		h.response.WriteHeader(status)
		h.setStatus(status, "")
	}
}

// SessionUser : Returns loggeed in User
func (h *Handler) SessionUser() *types.User {
	if h.session != nil {
		return h.session.User
	}
	return nil
}
