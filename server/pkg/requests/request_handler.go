package requests

import (
	"fmt"
	config "github.com/hyper-ml/hyperml/server/pkg/config"
	"github.com/hyper-ml/hyperml/server/pkg/pods"
	"github.com/hyper-ml/hyperml/server/pkg/qs"
	"github.com/hyper-ml/hyperml/server/pkg/schedules"
	types "github.com/hyper-ml/hyperml/server/pkg/types"
)

// RequestHandler : Handle user requests
type RequestHandler struct {
	uds       *UserDataStore
	conf      *config.Config
	pk        *pods.Keeper
	scheduler *schedules.NotebookScheduler
}

// NewRequestHandler :
func NewRequestHandler(cqs *qs.QueryServer, conf *config.Config, pk *pods.Keeper, scheduler *schedules.NotebookScheduler) *RequestHandler {
	uds := NewUserDataStore(cqs)
	return &RequestHandler{
		uds:       uds,
		conf:      conf,
		pk:        pk,
		scheduler: scheduler,
	}
}

// GetOrCreateUserDisk :
func (r *RequestHandler) GetOrCreateUserDisk(user *types.User, diskname, diskHostpath string, size uint64) (*types.UserPersistentDisk, error) {

	// get user persistent disk if exists
	pDisk, err := r.uds.GetUserDisk(user.Name, diskname)
	if IsUserDiskDoesntExist(err) {
		udisk := types.UserPersistentDisk{
			User: user,
			Disk: &types.PersistentDisk{
				Name:     diskname,
				HostPath: diskHostpath,
				Size:     size,
			},
		}
		pDisk, err = r.uds.InsertUserDisk(&udisk)
	}
	// create pv if not exists
	if pDisk == nil {
		return nil, fmt.Errorf("Failed to create user disk")
	}
	return pDisk, nil
}

// initiatePodRequest : Initiates a POD object
func (r *RequestHandler) initiateRequest(user *types.User, jobType string, requestMode types.PodRequestMode, cmd string, resourceProfileID uint64, containerImageID uint64, params *types.PODParams) (*types.POD, error) {
	var userDisk *types.UserPersistentDisk
	var podConfig *types.PODConfig
	var rprofile *types.ResourceProfile
	var image *types.ContainerImage
	var err error

	if resourceProfileID == 0 {
		return nil, fmt.Errorf("Resource Profile is mandatory")
	}

	rprofile, err = r.uds.GetResourceProfile(resourceProfileID)
	if err != nil {
		return nil, fmt.Errorf("Error fetching Resource Profile: %v", err)
	}

	if containerImageID == 0 {
		return nil, fmt.Errorf("Container Image is mandatory")
	}

	image, err = r.uds.GetContainerImage(containerImageID)
	if err != nil {
		return nil, fmt.Errorf("Error fetching Container Image: %v", err)
	}

	if r.conf.GetBool("EnableDefaultUserDisk") {
		userDisk, err = r.GetOrCreateUserDisk(user, user.Name+"-disk-0", params.PersDiskHostPath, 1024)
		if err != nil {
			return nil, fmt.Errorf("Failed to Create Disk: %v", err)
		}
		podConfig, err = types.NewPODConfig(rprofile, image, cmd, userDisk.Disk, params)
	} else {
		podConfig, err = types.NewPODConfig(rprofile, image, cmd, nil, params)
	}

	pod := types.NewPOD(user, jobType, requestMode, podConfig)
	fmt.Println("Pod:", *pod)
	return r.pk.InitPOD(pod), nil
}

// processRequest : Creates a new pod request for user
func (r *RequestHandler) processRequest(user *types.User, jobType string, cmd string, resourceProfileID uint64, containerImageID uint64, params *types.PODParams) (*types.POD, error) {

	pod, err := r.initiateRequest(user, jobType, types.PodReqModeImd, cmd, resourceProfileID, containerImageID, params)
	if err != nil {
		return pod, err
	}

	pod, err = r.pk.CreatePOD(pod)
	fmt.Println("pod: ", pod)

	return pod, nil
}

func (r *RequestHandler) getPodByUser(user *types.User, id uint64) (*types.POD, error) {
	return r.uds.GetPODByUser(user, id)
}

// ListPODs : Get all user pods
func (r *RequestHandler) ListPODs(kind string, user *types.User) (pods []*types.POD, fnerr error) {
	// get list from DB
	pods, fnerr = r.uds.ListPODsByUser(user)

	// filter on kind
	// TODO:
	return
}
