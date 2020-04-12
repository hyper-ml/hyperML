package rest

import (
	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/handlers"
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

type indexHandler struct {
	handler http.Handler
}

func (h indexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}

func registerWebRoutes(router *mux.Router) error {
	// register web router
	httpBox := rice.MustFindBox("../../browser/build").HTTPBox()
	staticBox := rice.MustFindBox("../../browser/build/static").HTTPBox()
	compressHandler := handlers.CompressHandler(http.StripPrefix("/static/", http.FileServer(staticBox)))
	router.PathPrefix("/static/").Handler(compressHandler)
	router.Handle("/login", http.FileServer(httpBox))
	router.Handle("/", http.FileServer(httpBox))
	return nil
}

func registerAPIRoutes(sc *ServerContext, privs HandlerPrivs, router *mux.Router) error {
	// register api router
	apiRouter := router.PathPrefix(apiPath).Subrouter()
	apiRouter.Handle("/version", makeHandler(sc, privs, (*Handler).handleRoot)).Methods("GET", "HEAD")

	// websocket
	apiRouter.HandleFunc("/ws/log_stream/{taskId}", func(w http.ResponseWriter, r *http.Request) {
		_ = serveLogStream(sc, w, r)
		//todo: error catch
	})

	apiRouter.HandleFunc("/user/websocket", func(w http.ResponseWriter, r *http.Request) {
		_ = serveUserUpdates(sc, w, r)
		//todo: error catch
	})

	// auth
	apiRouter.Handle("/login", makeHandler(sc, privs, (*Handler).handleBasicAuth)).Methods("POST")

	// signup
	apiRouter.Handle("/signup", makeHandler(sc, privs, (*Handler).handleUserSignup)).Methods("POST")

	// resources
	apiRouter.Handle("/resources/group", makeHandler(sc, privs, (*Handler).handleCreateResourceGroup)).Methods("POST")
	apiRouter.Handle("/resources/group", makeHandler(sc, privs, (*Handler).handleListResourceGroup)).Methods("GET")
	apiRouter.Handle("/resources/group", makeHandler(sc, privs, (*Handler).handleDeleteResourceGroup)).Methods("DELETE")

	// resource profiles
	apiRouter.Handle("/resources/profiles", makeHandler(sc, privs, (*Handler).handleCreateResourceProfile)).Methods("POST")
	apiRouter.Handle("/resources/profiles", makeHandler(sc, privs, (*Handler).handleUpdateResourceProfiles)).Methods("PUT")

	apiRouter.Handle("/resources/profiles/{id}/disable", makeHandler(sc, privs, (*Handler).handleDisableResourceProfile)).Methods("PUT")
	apiRouter.Handle("/resources/profiles/{id}", makeHandler(sc, privs, (*Handler).handleDeleteResourceProfile)).Methods("DELETE")
	apiRouter.Handle("/resources/profiles", makeHandler(sc, privs, (*Handler).handleListResourceProfile)).Methods("GET")

	// container images
	apiRouter.Handle("/resources/containerimages", makeHandler(sc, privs, (*Handler).handleCreateContainerImage)).Methods("POST")
	apiRouter.Handle("/resources/containerimages/{id}", makeHandler(sc, privs, (*Handler).handleDeleteContainerImage)).Methods("DELETE")
	apiRouter.Handle("/resources/containerimages", makeHandler(sc, privs, (*Handler).handleListContainerImage)).Methods("GET")

	// User background notebooks
	apiRouter.Handle("/user/notebooks/bck/new", makeHandler(sc, privs, (*Handler).handleNewBckNotebook)).Methods("POST")
	apiRouter.Handle("/user/notebooks/bck", makeHandler(sc, privs, (*Handler).handleListBckNotebooks)).Methods("GET")
	apiRouter.Handle("/user/notebooks/bck/{id}", makeHandler(sc, privs, (*Handler).handleGetBckNotebook)).Methods("GET")
	apiRouter.Handle("/user/notebooks/bck/{id}/status", makeHandler(sc, privs, (*Handler).handleGetBckNotebookStatus)).Methods("GET")
	apiRouter.Handle("/user/notebooks/bck/{id}/stop", makeHandler(sc, privs, (*Handler).handleStopBckNotebook)).Methods("PUT")

	// jupyter lab
	apiRouter.Handle("/user/notebooks/new", makeHandler(sc, privs, (*Handler).handleStartNotebook)).Methods("POST")
	apiRouter.Handle("/user/notebooks", makeHandler(sc, privs, (*Handler).handleListNotebooks)).Methods("GET")
	apiRouter.Handle("/user/notebooks/{id}/stop", makeHandler(sc, privs, (*Handler).handleStopNotebook)).Methods("PUT")

	// object api
	//! apiRouter.Handle("/object", makeHandler(sc, privs, (*Handler).handleCreateObject)).Methods("POST")
	apiRouter.Handle("/object", makeHandler(sc, privs, (*Handler).handleGetObject)).Methods("GET")
	apiRouter.Handle("/object", makeHandler(sc, privs, (*Handler).handlePutObject)).Methods("PUT")

	// api getters for meta
	apiRouter.Handle("/repo", makeHandler(sc, privs, (*Handler).handlePostRepo)).Methods("POST")
	apiRouter.Handle("/repo", makeHandler(sc, privs, (*Handler).handleDeleteRepo)).Methods("DELETE")

	apiRouter.Handle("/repo/{repoId}", makeHandler(sc, privs, (*Handler).handleGetRepo)).Methods("GET")
	apiRouter.Handle("/repo/{repoId}/explode", makeHandler(sc, privs, (*Handler).handleExplodeRepo)).Methods("GET")

	// commit activities
	apiRouter.Handle("/commit", makeHandler(sc, privs, (*Handler).handleGetOrStartCommit)).Methods("GET")
	apiRouter.Handle("/commit_close", makeHandler(sc, privs, (*Handler).handleCloseCommit)).Methods("POST")

	// repo attr getters
	apiRouter.Handle("/repo_attrs", makeHandler(sc, privs, (*Handler).handleGetRepoAttrs)).Methods("GET")
	apiRouter.Handle("/repo_attrs/{repoId}/explode", makeHandler(sc, privs, (*Handler).handleExplodeRepoAttrs)).Methods("GET")
	apiRouter.Handle("/repo_attrs/explode", makeHandler(sc, privs, (*Handler).handleExplodeRepoAttrs)).Methods("GET")

	//  model by repo getter/setter
	apiRouter.Handle("/repo_attrs/{repoName}/branch/{branchName}/commit/{commitId}/model", makeHandler(sc, privs, (*Handler).handleGetModel)).Methods("GET")
	apiRouter.Handle("/repo_attrs/{repoName}/branch/{branchName}/commit/{commitId}/model", makeHandler(sc, privs, (*Handler).handleGetOrCreateModel)).Methods("POST")

	// commit attrs
	apiRouter.Handle("/repo_attrs/{repoName}/branch/{branchName}/commit/{commitId}/attrs", makeHandler(sc, privs, (*Handler).handleGetCommitAttrs)).Methods("GET")
	apiRouter.Handle("/repo/{repoName}/branch/{branchName}/commit/{commitId}/attrs", makeHandler(sc, privs, (*Handler).handleGetCommitAttrs)).Methods("GET")

	// other getters
	apiRouter.Handle("/branch_attr", makeHandler(sc, privs, (*Handler).handleGetBranchAttrs)).Methods("GET")
	apiRouter.Handle("/commit_attrs", makeHandler(sc, privs, (*Handler).handleGetCommitAttrs)).Methods("GET")
	apiRouter.Handle("/file_attrs", makeHandler(sc, privs, (*Handler).handleGetFileAttrs)).Methods("GET")
	apiRouter.Handle("/commit_map", makeHandler(sc, privs, (*Handler).handleGetCommitMap)).Methods("GET")

	//commit file
	//apiRouter.Handle("/file_content", makeHandler(sc, privs, (*Handler).handleGetContent)).Methods("GET")
	apiRouter.Handle("/file", makeHandler(sc, privs, (*Handler).handlePutFile)).Methods("PUT")
	apiRouter.Handle("/file_url", makeHandler(sc, privs, (*Handler).handleGetFileURL)).Methods("GET")
	apiRouter.Handle("/parts_url", makeHandler(sc, privs, (*Handler).handleGetFilePartsURL)).Methods("GET")
	apiRouter.Handle("/file_checkin", makeHandler(sc, privs, (*Handler).handleFileCheckIn)).Methods("POST")
	apiRouter.Handle("/parts_merge", makeHandler(sc, privs, (*Handler).handleFileMerge)).Methods("POST")

	// data repo
	apiRouter.Handle("/dataset", makeHandler(sc, privs, (*Handler).handlePostDataSet)).Methods("POST")

	// model repo
	apiRouter.Handle("/model", makeHandler(sc, privs, (*Handler).handlePostModelRepo)).Methods("POST")

	// out repo
	apiRouter.Handle("/out", makeHandler(sc, privs, (*Handler).handlePostOutRepo)).Methods("POST")

	// api for task and flows
	apiRouter.Handle("/flow/{flowId}", makeHandler(sc, privs, (*Handler).handleGetFlowAttrs)).Methods("GET")
	apiRouter.Handle("/flow/{flowId}/status", makeHandler(sc, privs, (*Handler).handleGetFlowStatus)).Methods("GET")
	apiRouter.Handle("/flow", makeHandler(sc, privs, (*Handler).handleLaunchFlow)).Methods("POST")

	// flow output
	apiRouter.Handle("/flow/{flowId}/output", makeHandler(sc, privs, (*Handler).handleGetOutputByFlow)).Methods("GET")
	apiRouter.Handle("/flow/{flowId}/output", makeHandler(sc, privs, (*Handler).handleGetOrCreateOutputByFlow)).Methods("POST")

	// model by flow id
	apiRouter.Handle("/flow/{flowId}/model", makeHandler(sc, privs, (*Handler).handleGetModelByFlow)).Methods("GET")
	apiRouter.Handle("/flow/{flowId}/model", makeHandler(sc, privs, (*Handler).handleGetOrCreateModelByFlow)).Methods("POST")

	// log getters
	apiRouter.Handle("/tasks/{taskId}/log", makeHandler(sc, privs, (*Handler).handleGetTaskLog)).Methods("GET")
	apiRouter.Handle("/flow/{taskId}/log", makeHandler(sc, privs, (*Handler).handleGetTaskLog)).Methods("GET")

	// command log
	apiRouter.Handle("/tasks/{taskId}/cmd_log", makeHandler(sc, privs, (*Handler).handleGetCommandLog)).Methods("GET")
	apiRouter.Handle("/flow/{taskId}/cmd_log", makeHandler(sc, privs, (*Handler).handleGetCommandLog)).Methods("GET")
	apiRouter.Handle("/tasks/{taskId}/cmd_log", makeHandler(sc, privs, (*Handler).handlePostCommandLog)).Methods("POST")
	apiRouter.Handle("/flow/{taskId}/cmd_log", makeHandler(sc, privs, (*Handler).handlePostCommandLog)).Methods("POST")

	// workers
	apiRouter.Handle("/worker/register", makeHandler(sc, privs, (*Handler).handleRegisterWorker)).Methods("POST")
	apiRouter.Handle("/worker/detach", makeHandler(sc, privs, (*Handler).handleDetachTaskWorker)).Methods("POST")
	apiRouter.Handle("/worker/{workerId}/task_status", makeHandler(sc, privs, (*Handler).handleUpdateTaskStatus)).Methods("PATCH")

	return nil
}

func createRouter(sc *ServerContext, privs HandlerPrivs) *mux.Router {
	router := mux.NewRouter()

	router.StrictSlash(true)
	registerAPIRoutes(sc, privs, router)
	registerWebRoutes(router)

	// router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
	// 	t, err := route.GetPathTemplate()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	fmt.Println(t)
	// 	return nil
	// })

	return router
}

// CreatePublicHandler : Public Method handler
func CreatePublicHandler(sc *ServerContext) http.Handler {
	r := createRouter(sc, userPrivs)

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

		var match mux.RouteMatch
		if router.Match(req, &match) {
			router.ServeHTTP(response, req)
		} else {

			//todo : handle error from newHandler()
			h, _ := newHandler(sc, privs, response, req)

			h.logRequest()
			h.writeStatus(http.StatusNotFound, "unknown URL")
		}
	}))
}

func muxVars(rq *http.Request) map[string]string {
	return mux.Vars(rq)
}
