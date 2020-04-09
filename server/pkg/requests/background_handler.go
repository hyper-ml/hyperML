package requests

import (
	"fmt"
	types "github.com/hyper-ml/hyperml/server/pkg/types"
)

// ProcBackgroundRequest : Submit notebook to scheduler
func (r *RequestHandler) ProcBackgroundRequest(user *types.User, resourceProfileID, containerImageID uint64, params *types.PODParams) (*types.POD, error) {
	cmd := r.conf.Notebooks.GetBackgroundCmd()
	pod, err := r.initiateRequest(user, types.Notebook, types.PodReqModeBck, cmd, resourceProfileID, containerImageID, params)
	if err != nil {
		return pod, err
	}
	return r.scheduler.ScheduleRequest(user, pod)
}

// StopBckNotebook : Stops background notebooks
func (r *RequestHandler) StopBckNotebook(user *types.User, podID uint64) (*types.POD, error) {
	return nil, fmt.Errorf("Unimplemented feature")
}

// GetBckNotebook : Get notebook
func (r *RequestHandler) GetBckNotebook(user *types.User, podID uint64) (*types.POD, error) {
	return r.getPodByUser(user, podID)
}

// GetBckNotebookStatus : Syncs status with K8s before returning POD Info
func (r *RequestHandler) GetBckNotebookStatus(user *types.User, podID uint64) (*types.POD, error) {
	fmt.Println("GetBckNotebookStatus: ")
	pod, err := r.getPodByUser(user, podID)
	if err != nil || pod == nil {
		return nil, nil
	}

	if !pod.IsDone() {
		fmt.Println("Is not done")
		return r.pk.SyncUserJob(pod.PodType, pod.ID)
	}

	return pod, nil
}
