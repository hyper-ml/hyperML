package api_client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/hyper-ml/hyperml/worker/config"
	"github.com/hyper-ml/hyperml/worker/rest_client"
	"github.com/hyper-ml/hyperml/worker/utils"

	"github.com/hyper-ml/hyperml/server/pkg/base"
	flw "github.com/hyper-ml/hyperml/server/pkg/flow"
	ws "github.com/hyper-ml/hyperml/server/pkg/workspace"
)

const (
	RepoUriPath        = "repo"
	RepoAttrsUriPath   = "repo_attrs"
	BranchAttrsUriPath = "branch_attr"
	CommitUriPath      = "commit"
	CloseCommitUriPath = "commit_close"
	CommitAttrsUriPath = "commit_attrs"

	CommitMapUriPath    = "commit_map"
	FileUriPath         = "file"
	FileAttrsUriPath    = "file_attrs"
	ObjectUriPath       = "object"
	VfsUriPath          = "vfs"
	WorkerUriPath       = "worker"
	FlowUriPath         = "flow"
	FlowAttrsUriPath    = "flow"
	TaskAttrsUriPath    = "tasks"
	TasksUriPath        = "tasks"
	TaskStatusUriPath   = "task_status"
	CommandLogUriPath   = "/{task_id}/cmd_log"
	ModelUriPath        = "/{task_id}/model"
	CommitFileURL       = "file_url"
	CommitFileCheckIn   = "/file_checkin"
	CommitFilePartURL   = "parts_url"
	CommitFilePartMerge = "parts_merge"
)

type FsConfig struct {
	repoDir string
}

type ServerConfig struct {
	server_http string
	base_uri    string
	repo_path   string
}

type WorkerClient struct {
	BaseURL   *url.URL
	Config    *config.Config
	AuthToken string

	// Rest Client to fetch info from server
	RepoAttrs rest_client.Interface

	// Rest client for commit info
	CommitAttrs rest_client.Interface

	// Rest client for commit map
	CommitMap rest_client.Interface

	// Rest client for branch info
	BranchAttrs rest_client.Interface

	// File Info client
	FileAttrs rest_client.Interface

	// object client
	ContentIo rest_client.Interface

	// flow attributes
	FlowAttrs rest_client.Interface

	TaskAttrs rest_client.Interface

	TaskStatus rest_client.Interface

	//worker client

	WorkerAttrs rest_client.Interface

	// Virtual FS Client
	vfs rest_client.Interface

	//TODO: add stats
}

func NewWorkerClient(serverAddr string) (*WorkerClient, error) {
	var server_addr *url.URL

	c, err := config.ReadFromFile()

	if err != nil {
		base.Error("Failed to read config file")
		c = config.Default()
	}

	if serverAddr != "" {
		server_addr, err = url.Parse(serverAddr)

		if err != nil {
			base.Error("[NewWorkerClient] Invalid Server address string: ", serverAddr)
			return nil, err
		}

	} else {
		server_addr, err = url.Parse(c.DefaultServerAddr)
		if err != nil {
			base.Error("[NewWorkerClient] Invalid Default server address string: ", c.DefaultServerAddr)
			return nil, err
		}
	}

	return &WorkerClient{
		RepoAttrs:   newRestClient(server_addr, RepoAttrsUriPath),
		BranchAttrs: newRestClient(server_addr, BranchAttrsUriPath),
		CommitAttrs: newRestClient(server_addr, CommitAttrsUriPath),
		FileAttrs:   newRestClient(server_addr, FileAttrsUriPath),
		vfs:         newRestClient(server_addr, VfsUriPath),
		ContentIo:   newRestClient(server_addr, ObjectUriPath),
		FlowAttrs:   newRestClient(server_addr, FlowAttrsUriPath),
		TaskAttrs:   newRestClient(server_addr, TaskAttrsUriPath),
		WorkerAttrs: newRestClient(server_addr, WorkerUriPath),
		CommitMap:   newRestClient(server_addr, CommitMapUriPath),
		BaseURL:     server_addr,
		Config:      c,
		//TaskStatus: task_status,
	}, nil

}

func newRestClient(baseURL *url.URL, pathURL string) rest_client.Interface {
	c, err := rest_client.NewRESTClient(baseURL, pathURL, http.DefaultClient)
	if err != nil {
		return nil
	}
	return c
}

func (wc *WorkerClient) FetchBranchAttrs(repoName, branchName string) (*ws.BranchAttrs, error) {
	var branch_attr ws.BranchAttrs

	brq := wc.BranchAttrs.Get()
	brq.Param("repoName", repoName)
	brq.Param("branchName", branchName)

	resp := brq.Do()
	body, err := resp.Raw()
	if err != nil {
		return nil, err
	}

	branch_attr = ws.BranchAttrs{}
	err = json.Unmarshal(body, &branch_attr)

	return &branch_attr, nil
}

func (wc *WorkerClient) FetchCommitAttrs(repoName, commitId string) (*ws.CommitAttrs, error) {
	var commit_attrs ws.CommitAttrs

	crq := wc.CommitAttrs.Get()
	crq.Param("repoName", repoName)
	crq.Param("commitId", commitId)
	resp := crq.Do()
	body, err := resp.Raw()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &commit_attrs)
	return &commit_attrs, nil
}

func (wc *WorkerClient) CloseCommit(repoName, branchName, commitId string) error {
	client, _ := rest_client.NewRESTClient(wc.BaseURL, CloseCommitUriPath, http.DefaultClient)
	rq := client.Verb("POST")

	rq.Param("repoName", repoName)
	rq.Param("branchName", branchName)
	rq.Param("commitId", commitId)
	//!

	resp := rq.Do()
	_, err := resp.Raw()

	if err == nil {
		return nil
	} else {
		base.Error("[WorkerClient.CloseCommit] HTTP request to close commit failed: ", repoName, commitId, err)
		return err
	}

	return nil
}

func (wc *WorkerClient) FetchCommitMap(repoName, commitId string) (*ws.FileMap, error) {
	var file_map ws.FileMap

	rq := wc.CommitMap.Get()
	rq.Param("repoName", repoName)
	rq.Param("commitId", commitId)
	resp := rq.Do()
	body, err := resp.Raw()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &file_map)
	return &file_map, nil
}

func (wc *WorkerClient) FetchFileAttrs(repoName string, commitId string, fpath string) (*ws.FileAttrs, error) {
	var file_attrs *ws.FileAttrs

	crq := wc.FileAttrs.Verb("GET")
	crq.Param("repoName", repoName)
	crq.Param("commitId", commitId)
	crq.Param("path", fpath)

	resp := crq.Do()
	body, err := resp.Raw()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &file_attrs)
	return file_attrs, nil

}

func (wc *WorkerClient) FetchFlowAttrs(flowId string) (*flw.FlowAttrs, error) {
	var flow_attrs *flw.FlowAttrs
	flow_rq := wc.FlowAttrs.VerbSp("GET", flowId)
	//flow_rq.Param("flowId", flowId)

	resp := flow_rq.Do()
	flow_service_resp, err := resp.Raw()

	if err != nil {
		base.Error("[WorkerClient.FetchFlowAttrs] Fetch Error: ", err)
		return nil, err
	}

	err = json.Unmarshal(flow_service_resp, &flow_attrs)
	return flow_attrs, nil
}

func (wc *WorkerClient) RegisterWorker(flowId string, taskId string, ip string) (*flw.WorkerAttrs, error) {
	req := wc.WorkerAttrs.VerbSp("POST", "register")
	req.Param("flowId", flowId)
	req.Param("taskId", taskId)
	req.Param("ip", ip)

	response, err := req.Do().Raw()

	if err != nil {
		base.Error("[WorkerClient.RegisterWorker] Failed to register worker for flow: ", flowId, taskId, err)
		return nil, err
	}

	worker_attrs := flw.WorkerAttrs{}
	err = json.Unmarshal(response, &worker_attrs)

	return &worker_attrs, err
}

func (wc *WorkerClient) DetachWorker(flowId string, taskId string, workerId string) error {
	req := wc.WorkerAttrs.VerbSp("POST", "detach")
	req.Param("flowId", flowId)
	req.Param("taskId", taskId)
	req.Param("workerId", workerId)

	_, err := req.Do().Raw()

	if err != nil {
		base.Error("[WorkerClient.UnRegisterWorker] Failed to detach worker from the flow: ", flowId, taskId, workerId)
		base.Error("[WorkerClient.UnRegisterWorker] Error: ", err)
		return err
	}

	return nil
}

func (wc *WorkerClient) GetOrCreateModelRepo(flowId string) (*ws.Repo, *ws.Branch, *ws.Commit, error) {

	client := newRestClient(wc.BaseURL, FlowUriPath)
	model_uri := strings.Replace(ModelUriPath, "{task_id}", flowId, -1)
	req := client.VerbSp("POST", model_uri)

	resp := req.Do()
	body, err := resp.Raw()

	switch {
	case err != nil:
		return nil, nil, nil, err
	case body == nil:
		base.Error("Received no model data")
		return nil, nil, nil, utils.ErrHttpEmptyResponse()
	}

	model_response := &ws.RepoMessage{}
	err = json.Unmarshal(body, model_response)

	switch {
	case err != nil:
		return nil, nil, nil, err
	case model_response.Repo == nil:
		base.Error("[WorkerClient.GetOrCreateModelRepo] Failed to generate model repo")
		return nil, nil, nil, utils.ErrNullRepo()
	}

	return model_response.Repo, model_response.Branch, model_response.Commit, nil
}

func (wc *WorkerClient) PostLogWriter(taskId string) (io.WriteCloser, error) {

	client := newRestClient(wc.BaseURL, TasksUriPath)
	r := client.VerbSp("POST", strings.Replace(CommandLogUriPath, "{task_id}", taskId, -1))

	hw := &httpObjectWriter{
		r: r,
	}

	return hw, nil

}
