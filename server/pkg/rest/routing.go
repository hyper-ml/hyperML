package rest

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"net/http"
)

var upgrader = websocket.Upgrader{
	// TODO: this is only required for dev
	CheckOrigin: func(r *http.Request) bool {
		if r.Header.Get("Origin") == "http://localhost:3000" {
			return true
		}
		return false
	},
}

func createHandler(sc *ServerContext, privs HandlerPrivs) *mux.Router {
	router := mux.NewRouter()

	router.StrictSlash(true)

	router.Handle("/", makeHandler(sc, privs, (*Handler).handleRoot)).Methods("GET", "HEAD")

	//ws
	router.HandleFunc("/ws/log_stream/{taskId}", func(w http.ResponseWriter, r *http.Request) {
		_ = serveLogStream(sc, w, r)
		//todo: error catch
	})

	router.HandleFunc("/user/websocket", func(w http.ResponseWriter, r *http.Request) {
		_ = serveUserUpdates(sc, w, r)
		//todo: error catch
	})

	// auth
	router.Handle("/login", makeHandler(sc, privs, (*Handler).handleBasicAuth)).Methods("POST")

	// signup
	router.Handle("/signup", makeHandler(sc, privs, (*Handler).handleUserSignup)).Methods("POST")

	// resources
	router.Handle("/resources/group", makeHandler(sc, privs, (*Handler).handleCreateResourceGroup)).Methods("POST")
	router.Handle("/resources/group", makeHandler(sc, privs, (*Handler).handleListResourceGroup)).Methods("GET")
	router.Handle("/resources/group", makeHandler(sc, privs, (*Handler).handleDeleteResourceGroup)).Methods("DELETE")

	// resource profiles
	router.Handle("/resources/profiles", makeHandler(sc, privs, (*Handler).handleCreateResourceProfile)).Methods("POST")
	router.Handle("/resources/profiles", makeHandler(sc, privs, (*Handler).handleUpdateResourceProfiles)).Methods("PUT")

	router.Handle("/resources/profiles/{id}/disable", makeHandler(sc, privs, (*Handler).handleDisableResourceProfile)).Methods("PUT")
	router.Handle("/resources/profiles/{id}", makeHandler(sc, privs, (*Handler).handleDeleteResourceProfile)).Methods("DELETE")
	router.Handle("/resources/profiles", makeHandler(sc, privs, (*Handler).handleListResourceProfile)).Methods("GET")

	// container images
	router.Handle("/resources/containerimages", makeHandler(sc, privs, (*Handler).handleCreateContainerImage)).Methods("POST")
	router.Handle("/resources/containerimages/{id}", makeHandler(sc, privs, (*Handler).handleDeleteContainerImage)).Methods("DELETE")
	router.Handle("/resources/containerimages", makeHandler(sc, privs, (*Handler).handleListContainerImage)).Methods("GET")

	// User background notebooks
	router.Handle("/user/notebooks/bck/new", makeHandler(sc, privs, (*Handler).handleNewBckNotebook)).Methods("POST")
	router.Handle("/user/notebooks/bck", makeHandler(sc, privs, (*Handler).handleListBckNotebooks)).Methods("GET")
	router.Handle("/user/notebooks/bck/{id}", makeHandler(sc, privs, (*Handler).handleGetBckNotebook)).Methods("GET")
	router.Handle("/user/notebooks/bck/{id}/status", makeHandler(sc, privs, (*Handler).handleGetBckNotebookStatus)).Methods("GET")
	router.Handle("/user/notebooks/bck/{id}/stop", makeHandler(sc, privs, (*Handler).handleStopBckNotebook)).Methods("PUT")

	// jupyter lab
	router.Handle("/user/notebooks/new", makeHandler(sc, privs, (*Handler).handleStartNotebook)).Methods("POST")
	router.Handle("/user/notebooks", makeHandler(sc, privs, (*Handler).handleListNotebooks)).Methods("GET")
	router.Handle("/user/notebooks/{id}/stop", makeHandler(sc, privs, (*Handler).handleStopNotebook)).Methods("PUT")

	// object api
	//! router.Handle("/object", makeHandler(sc, privs, (*Handler).handleCreateObject)).Methods("POST")
	router.Handle("/object", makeHandler(sc, privs, (*Handler).handleGetObject)).Methods("GET")
	router.Handle("/object", makeHandler(sc, privs, (*Handler).handlePutObject)).Methods("PUT")

	// api getters for meta
	router.Handle("/repo", makeHandler(sc, privs, (*Handler).handlePostRepo)).Methods("POST")
	router.Handle("/repo", makeHandler(sc, privs, (*Handler).handleDeleteRepo)).Methods("DELETE")

	router.Handle("/repo/{repoId}", makeHandler(sc, privs, (*Handler).handleGetRepo)).Methods("GET")
	router.Handle("/repo/{repoId}/explode", makeHandler(sc, privs, (*Handler).handleExplodeRepo)).Methods("GET")

	// commit activities
	router.Handle("/commit", makeHandler(sc, privs, (*Handler).handleGetOrStartCommit)).Methods("GET")
	router.Handle("/commit_close", makeHandler(sc, privs, (*Handler).handleCloseCommit)).Methods("POST")

	// repo attr getters
	router.Handle("/repo_attrs", makeHandler(sc, privs, (*Handler).handleGetRepoAttrs)).Methods("GET")
	router.Handle("/repo_attrs/{repoId}/explode", makeHandler(sc, privs, (*Handler).handleExplodeRepoAttrs)).Methods("GET")
	router.Handle("/repo_attrs/explode", makeHandler(sc, privs, (*Handler).handleExplodeRepoAttrs)).Methods("GET")

	//  model by repo getter/setter
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

	//commit file
	//router.Handle("/file_content", makeHandler(sc, privs, (*Handler).handleGetContent)).Methods("GET")
	router.Handle("/file", makeHandler(sc, privs, (*Handler).handlePutFile)).Methods("PUT")
	router.Handle("/file_url", makeHandler(sc, privs, (*Handler).handleGetFileURL)).Methods("GET")
	router.Handle("/parts_url", makeHandler(sc, privs, (*Handler).handleGetFilePartsURL)).Methods("GET")
	router.Handle("/file_checkin", makeHandler(sc, privs, (*Handler).handleFileCheckIn)).Methods("POST")
	router.Handle("/parts_merge", makeHandler(sc, privs, (*Handler).handleFileMerge)).Methods("POST")

	// data repo
	router.Handle("/dataset", makeHandler(sc, privs, (*Handler).handlePostDataSet)).Methods("POST")

	// model repo
	router.Handle("/model", makeHandler(sc, privs, (*Handler).handlePostModelRepo)).Methods("POST")

	// out repo
	router.Handle("/out", makeHandler(sc, privs, (*Handler).handlePostOutRepo)).Methods("POST")

	// api for task and flows
	router.Handle("/flow/{flowId}", makeHandler(sc, privs, (*Handler).handleGetFlowAttrs)).Methods("GET")
	router.Handle("/flow/{flowId}/status", makeHandler(sc, privs, (*Handler).handleGetFlowStatus)).Methods("GET")
	router.Handle("/flow", makeHandler(sc, privs, (*Handler).handleLaunchFlow)).Methods("POST")

	// flow output
	router.Handle("/flow/{flowId}/output", makeHandler(sc, privs, (*Handler).handleGetOutputByFlow)).Methods("GET")
	router.Handle("/flow/{flowId}/output", makeHandler(sc, privs, (*Handler).handleGetOrCreateOutputByFlow)).Methods("POST")

	// model by flow id
	router.Handle("/flow/{flowId}/model", makeHandler(sc, privs, (*Handler).handleGetModelByFlow)).Methods("GET")
	router.Handle("/flow/{flowId}/model", makeHandler(sc, privs, (*Handler).handleGetOrCreateModelByFlow)).Methods("POST")

	// log getters
	router.Handle("/tasks/{taskId}/log", makeHandler(sc, privs, (*Handler).handleGetTaskLog)).Methods("GET")
	router.Handle("/flow/{taskId}/log", makeHandler(sc, privs, (*Handler).handleGetTaskLog)).Methods("GET")

	// command log
	router.Handle("/tasks/{taskId}/cmd_log", makeHandler(sc, privs, (*Handler).handleGetCommandLog)).Methods("GET")
	router.Handle("/flow/{taskId}/cmd_log", makeHandler(sc, privs, (*Handler).handleGetCommandLog)).Methods("GET")
	router.Handle("/tasks/{taskId}/cmd_log", makeHandler(sc, privs, (*Handler).handlePostCommandLog)).Methods("POST")
	router.Handle("/flow/{taskId}/cmd_log", makeHandler(sc, privs, (*Handler).handlePostCommandLog)).Methods("POST")

	// workers
	router.Handle("/worker/register", makeHandler(sc, privs, (*Handler).handleRegisterWorker)).Methods("POST")
	router.Handle("/worker/detach", makeHandler(sc, privs, (*Handler).handleDetachTaskWorker)).Methods("POST")
	router.Handle("/worker/{workerId}/task_status", makeHandler(sc, privs, (*Handler).handleUpdateTaskStatus)).Methods("PATCH")

	//vfs methods
	//!router.Handle("/vfs/list_dir", makeHandler(sc, privs, (*Handler).handleListDir)).Methods("GET")
	//!router.Handle("/vfs/lookup",  makeHandler(sc, privs, (*Handler).handleFileLookup)).Methods("GET")
	//!router.Handle("/vfs/put_file",  makeHandler(sc, privs, (*Handler).handleVfsPutFile)).Methods("PUT")
	//!router.Handle("/vfs/get_file",  makeHandler(sc, privs, (*Handler).handleVfsGetFile)).Methods("GET")

	return router
}

// CreatePublicHandler : Public Method handler
func CreatePublicHandler(sc *ServerContext) http.Handler {
	r := createHandler(sc, userPrivs)
	return topRouter(sc, userPrivs, r)
}

func topRouter(sc *ServerContext, privs HandlerPrivs, router *mux.Router) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080"},
		AllowCredentials: false,
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		// Enable Debugging for testing, consider disabling in production
		Debug: false,
	})
	return c.Handler(http.HandlerFunc(func(response http.ResponseWriter, req *http.Request) {

		// TODO: add cors
		//response.Header().Add("Access-Control-Allow-Credentials", "true")
		//response.Header().Add("Access-Control-Allow-Headers", "*")

		var match mux.RouteMatch
		if router.Match(req, &match) {
			router.ServeHTTP(response, req)
		} else {

			//todo : handle error from newHandler()
			h, _ := newHandler(sc, privs, response, req)

			h.logRequest()

			//var options []string
			//for _, method := range []string{"GET", "HEAD", "POST", "PUT", "DELETE"} {
			//	if wouldMatch(router, req, method) {
			//		options = append(options, method)
			//	}
			//}
			//if len(options) == 0 {
			//	h.writeStatus(http.StatusNotFound, "unknown URL")
			//} else {
			//	fmt.Println("writing response headers: ", options)
			//response.Header().Add("Allow", strings.Join(options, ", "))
			// TODO: add cors
			//	response.Header().Add("Access-Control-Allow-Origin", "*")
			//	response.Header().Add("Access-Control-Allow-Methods", strings.Join(options, ", "))

			//}
			//if req.Method != "OPTIONS" {
			//	h.writeStatus(http.StatusMethodNotAllowed, "")
			//} else {
			//	h.writeStatus(http.StatusNoContent, "")
			//}
		}
	}))
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
