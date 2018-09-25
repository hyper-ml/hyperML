package rest


import (
  "strings"
  "net/http"
  "github.com/gorilla/mux"
)


func createHandler(sc *ServerContext, privs HandlerPrivs) (*mux.Router) {
  router := mux.NewRouter()
  
  router.StrictSlash(true)
  router.Handle("/", makeHandler(sc, privs, (*Handler).handleRoot)).Methods("GET", "HEAD")
  router.Handle("/object", makeHandler(sc, privs, (*Handler).handleCreateObject)).Methods("POST")
  router.Handle("/object", makeHandler(sc, privs, (*Handler).handleGetObject)).Methods("GET")
  router.Handle("/object", makeHandler(sc, privs, (*Handler).handlePutObject)).Methods("PUT")
  
  // api getters for meta 
  router.Handle("/repo", makeHandler(sc, privs, (*Handler).handlePostRepo)).Methods("POST")
  router.Handle("/repo/{repoId}", makeHandler(sc, privs, (*Handler).handleGetRepo)).Methods("GET")
  router.Handle("/repo/{repoId}/explode", makeHandler(sc, privs, (*Handler).handleExplodeRepo)).Methods("GET")

  
  // commit activities
  router.Handle("/commit", makeHandler(sc, privs, (*Handler).handleGetOrStartCommit)).Methods("GET")
  router.Handle("/commit", makeHandler(sc, privs, (*Handler).handleCloseCommit)).Methods("POST")

  // repo attr getters
  router.Handle("/repo_attrs", makeHandler(sc, privs, (*Handler).handleGetRepoAttrs)).Methods("GET")
  router.Handle("/repo_attrs/{repoId}/explode", makeHandler(sc, privs, (*Handler).handleExplodeRepoAttrs)).Methods("GET")

  // model getter/setter
  router.Handle("/repo_attrs/{repoName}/branch/{branchName}/commit/{commitId}/model", makeHandler(sc, privs, (*Handler).handleGetModel)).Methods("GET")
  router.Handle("/repo_attrs/{repoName}/branch/{branchName}/commit/{commitId}/model", makeHandler(sc, privs, (*Handler).handleGetOrCreateModel)).Methods("POST")

  // commit attrs
  router.Handle("/repo_attrs/{repoName}/branch/{branchName}/commit/{commitId}/attrs", makeHandler(sc, privs, (*Handler).handleGetCommitAttrs)).Methods("GET")
  router.Handle("/repo/{repoName}/branch/{branchName}/commit/{commitId}/attrs", makeHandler(sc, privs, (*Handler).handleGetCommitAttrs)).Methods("GET")

  // other getters
  router.Handle("/branch_attr", makeHandler(sc, privs, (*Handler).handleGetBranchAttrs)).Methods("GET")
  router.Handle("/commit_attrs", makeHandler(sc, privs, (*Handler).handleGetCommitAttrs)).Methods("GET")
  router.Handle("/file_attrs", makeHandler(sc, privs, (*Handler).handleGetFileAttrs)).Methods("GET")
  router.Handle("/commit_map", makeHandler(sc, privs, (*Handler).handleGetCommitMap)).Methods("GET")
  
  // data repo 
  router.Handle("/dataset", makeHandler(sc, privs, (*Handler).handlePostDataSet)).Methods("POST")
  
  // api for task and flows
  router.Handle("/flow/{flowId}", makeHandler(sc, privs, (*Handler).handleGetFlowAttrs)).Methods("GET")
  router.Handle("/flow", makeHandler(sc, privs, (*Handler).handleLaunchFlow)).Methods("POST")

  router.Handle("/flow/{flowId}/output", makeHandler(sc, privs, (*Handler).handleGetFlowOutput)).Methods("GET")
  router.Handle("/flow/{flowId}/output", makeHandler(sc, privs, (*Handler).handleGetOrCreateFlowOutput)).Methods("POST")


  // log getters
  router.Handle("/tasks/{taskId}/log", makeHandler(sc, privs, (*Handler).handleGetTaskLog)).Methods("GET")
  router.Handle("/tasks/{taskId}/log", makeHandler(sc, privs, (*Handler).handlePostTaskLog)).Methods("POST")
    

  // workers
  router.Handle("/worker/register", makeHandler(sc, privs, (*Handler).handleRegisterWorker)).Methods("POST")
  router.Handle("/worker/detach", makeHandler(sc, privs, (*Handler).handleDetachTaskWorker)).Methods("POST")
  router.Handle("/worker/{workerId}/task_status", makeHandler(sc, privs, (*Handler).handleUpdateTaskStatus)).Methods("PATCH")
  
  //vfs methods
  router.Handle("/vfs/list_dir", makeHandler(sc, privs, (*Handler).handleListDir)).Methods("GET")
  router.Handle("/vfs/lookup",  makeHandler(sc, privs, (*Handler).handleFileLookup)).Methods("GET") 
  router.Handle("/vfs/put_file",  makeHandler(sc, privs, (*Handler).handleVfsPutFile)).Methods("PUT") 
  router.Handle("/vfs/get_file",  makeHandler(sc, privs, (*Handler).handleVfsGetFile)).Methods("GET") 
    
  return router
}


func CreatePublicHandler(sc *ServerContext) http.Handler {
  r := createHandler(sc, userPrivs)
  return topRouter(sc, userPrivs, r)
}

func topRouter(sc *ServerContext, privs HandlerPrivs, router *mux.Router) http.Handler {
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


func muxVars(rq *http.Request) map[string]string {
  return mux.Vars(rq)
}

