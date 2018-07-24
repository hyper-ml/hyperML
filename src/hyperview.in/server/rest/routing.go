package rest


import (
  "strings"
  "net/http"
  "github.com/gorilla/mux"
)


func createHandler(sc *ServerContext, privs handlerPrivs) (*mux.Router) {
  router := mux.NewRouter()
  
  router.StrictSlash(true)
  router.Handle("/", makeHandler(sc, privs, (*handler).handleRoot)).Methods("GET", "HEAD")
  router.Handle("/object", makeHandler(sc, privs, (*handler).handleCreateObject)).Methods("POST")
  router.Handle("/object", makeHandler(sc, privs, (*handler).handleGetObject)).Methods("GET")
  router.Handle("/object", makeHandler(sc, privs, (*handler).handlePutObject)).Methods("PUT")
  
  // api getters for meta 
  router.Handle("/repo", makeHandler(sc, privs, (*handler).handleGetRepo)).Methods("GET")
  router.Handle("/repo_info", makeHandler(sc, privs, (*handler).handleGetRepoInfo)).Methods("GET")
  router.Handle("/branch_info", makeHandler(sc, privs, (*handler).handleGetBranchInfo)).Methods("GET")
  router.Handle("/commit_info", makeHandler(sc, privs, (*handler).handleGetCommitInfo)).Methods("GET")
  router.Handle("/file_info", makeHandler(sc, privs, (*handler).handleGetFileInfo)).Methods("GET")

  //vfs methods
  router.Handle("/vfs/list_dir", makeHandler(sc, privs, (*handler).handleListDir)).Methods("GET")
  router.Handle("/vfs/lookup",  makeHandler(sc, privs, (*handler).handleFileLookup)).Methods("GET") 
  router.Handle("/vfs/put_file",  makeHandler(sc, privs, (*handler).handleVfsPutFile)).Methods("PUT") 
  router.Handle("/vfs/get_file",  makeHandler(sc, privs, (*handler).handleVfsGetFile)).Methods("GET") 
  
  return router
}


func CreatePublicHandler(sc *ServerContext) http.Handler {
  r := createHandler(sc, userPrivs)
  return topRouter(sc, userPrivs, r)
}

func topRouter(sc *ServerContext, privs handlerPrivs, router *mux.Router) http.Handler {
  return http.HandlerFunc(func(response http.ResponseWriter, req *http.Request) {
    
    // TODO: add cors 
    response.Header().Add("Access-Control-Allow-Credentials", "true")
    
    var match mux.RouteMatch
    if router.Match(req, &match) {
      router.ServeHTTP(response, req)
    } else {
      h := newHandler(sc, privs, response, req)
      h.logRequest()
      var options []string
      for _, method := range []string{"GET", "HEAD", "POST", "PUT", "DELETE"} {
        if wouldMatch(router, req, method) {
          options = append(options, method)
        }
      }
      if len(options) == 0 {
        h.writeStatus(http.StatusNotFound, "unknown URL")
      } else {
        response.Header().Add("Allow", strings.Join(options, ", "))
        // TODO: add cors
        response.Header().Add("Access-Control-Allow-Methods", strings.Join(options, ", "))
      }
      if req.Method != "OPTIONS" {
        h.writeStatus(http.StatusMethodNotAllowed, "")
      } else {
        h.writeStatus(http.StatusNoContent, "")
      }
    }
  })
}


func wouldMatch(router *mux.Router, rq *http.Request, method string) bool {
  savedMethod := rq.Method
  rq.Method = method
  defer func() { rq.Method = savedMethod }()
  var matchInfo mux.RouteMatch
  return router.Match(rq, &matchInfo)
}


