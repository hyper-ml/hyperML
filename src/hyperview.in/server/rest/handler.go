package rest

import (
  "fmt"
  "time"
  "strings"
  "net/http"
  "net/url"
  "sync/atomic"
  "encoding/json"
  "hyperview.in/server/base"
)

var lastUid uint64 = 0


type handler struct {
	server *ServerContext
  rq *http.Request
  response http.ResponseWriter
  status int
  statusMessage string
  privs handlerPrivs
  uid uint64
  startTime      time.Time
  query_params    url.Values 
}

type httpBody map[string]interface{}


type handlerPrivs int
const handelerVersion = "0.1"

const ( userPrivs = iota
  publicPrivs
  adminPrivs
)

type handlerMethod func(*handler) error

func makeHandler(server *ServerContext, privs handlerPrivs, method handlerMethod) http.Handler {
  return http.HandlerFunc(func(response http.ResponseWriter, req *http.Request){
    h := newHandler(server, privs, response, req)
    err := h.invoke(method)
    h.writeError(err)
    })
}

// TODO: 
// 1. Add user details
func newHandler(server *ServerContext, privs handlerPrivs, response http.ResponseWriter, request *http.Request ) *handler{
  return &handler{
    server: server,
    privs: privs,
    rq: request,
    response: response,
    status: http.StatusOK,
    startTime: time.Now(),
    uid: atomic.AddUint64(&lastUid, 1),
  }
}


func (h *handler) invoke(method handlerMethod) error {
  //TODO : add stats logging

  h.setHeader("server", handelerVersion)

  //TODO: add auth privs check
  h.logRequest()
  return method(h)
}


// Utility functions 
//

func (h *handler) writeError(err error) {
  if err != nil {
    status, message := base.ErrorToHTTPStatus(err)
    h.writeStatus(status, message)
  }
}

func (h *handler) writeStatus(status int, message string) {
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

func (h *handler) logRequest() {
  queryParams := h.getQueryParams()

  base.Log("HTTP Request Log:", h.uid, h.rq.Method, queryParams)
}

func (h *handler) logRequestBody() {
  fmt.Println("Log request body. TODO")
}

func (h *handler) getQueryParams() url.Values{
  if h.query_params == nil {
    h.query_params = h.rq.URL.Query()
  }
  return h.query_params
}

func (h *handler) getQuery(query string) string {
  return h.getQueryParams().Get(query)
}

func (h *handler) requestAccepts(mimetype string) bool {
  accept := h.rq.Header.Get("Accept")
  return accept == "" || strings.Contains(accept, mimetype) || strings.Contains(accept, "*/*")
}

// Response methods 


func (h *handler) setHeader(name string, value string) {
  h.response.Header().Set(name, value)
}

func (h *handler) setStatus(status int, message string) {
  h.status = status
  h.statusMessage = message
}

func (h *handler) writeJSON(value interface{}) {
  h.writeJSONStatus(http.StatusOK, value)
}


func (h *handler) writeJSONStatus(status int, value interface{}) {
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
