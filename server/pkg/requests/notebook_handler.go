package requests

import (
	"fmt"
	types "github.com/hyper-ml/hyperml/server/pkg/types"
)

// ProcessNotebookRequest : creates a new POD for jupyter LAB
func (r *RequestHandler) ProcessNotebookRequest(user *types.User, resourceProfileID, containerImageID uint64) (*types.POD, error) {
	// TODO : get it from config
	//cmd := "jupyter notebook --no-browser --port={port} --ip={ip}  --NotebookApp.port_retries=0 --NotebookApp.disable_check_xsrf=True" // "--NotebookApp.token=abc"
	//tokenMap := strings.NewReplacer("{port}", LabInternalPort, "{ip}", LabInternalIP)
	//cmd = tokenMap.Replace(cmd)
	cmd := r.conf.Notebooks.GetCommand()
	fmt.Println(cmd)
	return r.processRequest(user, types.Notebook, cmd, resourceProfileID, containerImageID, nil)
}

// StopNotebook : Stop a notebook POD owned by user
func (r *RequestHandler) StopNotebook(user *types.User, podID uint64) (*types.POD, error) {
	// check if POD ID belongs to user first
	pod, err := r.getPodByUser(user, podID)
	if err != nil {
		return pod, err
	}

	if pod == nil {
		return pod, fmt.Errorf("User does not have sufficient privileges to stop this notebook")
	}
	fmt.Println("pod:", pod)

	// Initiate termination status on POD
	pod.Terminate()
	fmt.Println("post terminate:", pod)
	pod, err = r.uds.UpdateUserPOD(pod)
	if err != nil {
		return pod, err
	}
	fmt.Println("After updating user POD ")
	err = r.pk.CleanupPOD(pod.PodType, pod.RequestMode, pod.ID)
	if err != nil {
		return pod, err
	}

	return pod, nil
}

// GetNotebook : Get POD Status and Object
func (r *RequestHandler) GetNotebook(user *types.User, id string) (pod types.POD, fnerr error) {
	// check POD belongs to user
	// send status of POD
	return
}

// ListNotebooks : Get list of user pods with status info
func (r *RequestHandler) ListNotebooks(user *types.User) ([]*types.POD, error) {
	// Get list of PODs for user
	// get objects from DB
	// send the PODs over
	return r.ListPODs(types.Notebook, user)
}
