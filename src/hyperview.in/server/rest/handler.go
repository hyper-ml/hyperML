package rest

import (
  "fmt"
  "time"
  "strings"
  "io"
  "net/http"
  "net/url"
  "sync/atomic"
  "encoding/json"
  "hyperview.in/server/base"
)

var lastUid uint64 = 0


type Handler struct {
	server *ServerContext
  rq *http.Request
  response http.ResponseWriter
  requestBody io.ReadCloser
  status int
  statusMessage string
  privs HandlerPrivs
  uid uint64
  startTime      time.Time
  query_params    url.Values 
}

type httpBody map[string]interface{}


type HandlerPrivs int
const handelerVersion = "0.1"

const ( userPrivs = iota
  publicPrivs
  adminPrivs
)

type HandlerMethod func(*Handler) error

func makeHandler(server *ServerContext, privs HandlerPrivs, method HandlerMethod) http.Handler {
  return http.HandlerFunc(func(response http.ResponseWriter, req *http.Request){
    h := newHandler(server, privs, response, req)
    err := h.invoke(method)
    h.writeError(err)
    })
}
 

// TODO: 
// 1. Add user details
func newHandler(server *ServerContext, privs HandlerPrivs, response http.ResponseWriter, request *http.Request ) *Handler{
  return &Handler{
    server: server,
    privs: privs,
    rq: request,
    response: response,
    status: http.StatusOK,
    startTime: time.Now(),
    uid: atomic.AddUint64(&lastUid, 1),
  }
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
  h.response.WriteHeader(status)
  h.setStatus(status, message)
  jsonOut, _ := json.Marshal(&httpBody{"error": errorStr, "reason": message})
  h.response.Write(jsonOut)
}

func (h *Handler) logRequest() {
  //queryParams := h.getQueryParams()

  //base.Log("HTTP Request Log:", h.uid, h.rq.Method, queryParams)
}

func (h *Handler) logRequestBody() {
  fmt.Println("Log request body. TODO")

}

func (h *Handler) getQueryParams() url.Values{
  if h.query_params == nil {
    h.query_params = h.rq.URL.Query()
  }
  return h.query_params
}

func (h *Handler) getQuery(query string) string {
  return h.getQueryParams().Get(query)
}



func (h *Handler) getUrlParams() map[string]string {
  return muxVars(h.rq)
}

func (h *Handler) getMandatoryUrlParam(pname string) (string, error) {
  vars := h.getUrlParams()
  param_value, ok := vars[pname]
  if !ok {
    return "", base.HTTPErrorf(http.StatusInternalServerError, "Invalid request parameter: " + pname )
  }

  return param_value, nil
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

 