package rest

import (
	"github.com/hyper-ml/hyperml/server/pkg/types"
)

// UserInfo : API Info Object for User
type UserInfo struct {
	Name string
}

func newUserInfo(user *types.User) UserInfo {
	return UserInfo{
		Name: user.Name,
	}
}

// UserSignupInfo :
type UserSignupInfo struct {
	UserName string
	Email    string
	Password string
}

// SessionInfo :
type SessionInfo struct {
	UserName string
	Email    string
	JWT      string
	Status   string
}

// ResourceGroupInfo :
type ResourceGroupInfo struct {
	ID              uint64
	Name            string
	ResourceProfile []ResourceProfInfo
}

// ResourceProfInfo : API Object for sending/receiving ResourceProfInfo
type ResourceProfInfo struct {
	*types.ResourceProfile
}

// ResourceProfsInfo : List of Resource Profile messages
type ResourceProfsInfo []types.ResourceProfile

func newResourceProfInfo(rp *types.ResourceProfile) ResourceProfInfo {
	return ResourceProfInfo{
		rp,
	}
}

func validateResourceProfInfo(rp *ResourceProfInfo) string {
	if rp.Name == "" {
		return "Profile Name is empty"
	}
	return ""
}

// ContainerImageInfo : API Object for sending / receiving ContainerImage
type ContainerImageInfo struct {
	*types.ContainerImage
}

// newContainerInfo : creates a new ContainerInfo object
func newContainerImageInfo(img *types.ContainerImage) ContainerImageInfo {
	return ContainerImageInfo{
		img,
	}
}

func validateContainerImageInfo(cii *ContainerImageInfo) string {
	if cii.Name == "" {
		return "Container Image name is empty"
	}

	return ""
}

// PodInfo : API Object for User POD
type PodInfo struct {
	// User POD ID
	ID uint64

	// Resource Profile ID
	ResourceProfileID uint64

	// Resources
	ResourceProfile *types.ResourceProfile

	// Container Image ID
	ContainerImageID uint64

	// ContainerImage
	ContainerImage *types.ContainerImage

	// Pod Workspace Params
	Params *types.PODParams

	// UserPOD (optional)
	POD *types.POD

	// POD Status message
	Status string

	// Pod Phase message
	Phase string
}

// newPodInfo : creates UserPodInfo object
func newPodInfo(p *types.POD) PodInfo {

	if string(p.Phase) == nullString {
		p.Phase = types.PhaseUnknown
	}

	if string(p.Status) == nullString {
		p.Status = types.PodStatusUnknown
	}

	return PodInfo{
		ID:              p.ID,
		POD:             p,
		ContainerImage:  p.Config.Image,
		ResourceProfile: p.Config.Resources,
		Params:          p.Config.PODParams,
		Phase:           types.PodPhaseDesc(p.Phase),
		Status:          types.PodStatusDesc(p.Status),
	}
}

// NotebookInfo : API Object for POD with jupyter LAB
type NotebookInfo struct {
	PodInfo
}

func newNotebookInfo(p *types.POD) NotebookInfo {
	podInfo := newPodInfo(p)
	return NotebookInfo{
		podInfo,
	}
}

func validateNotebookReq(nbInfo NotebookInfo) string {
	if nbInfo.ResourceProfileID == 0 {
		return "Resource Profile ID is required"
	}

	if nbInfo.ContainerImageID == 0 && nbInfo.ContainerImage.Name == "" {
		return "Container Image Name or ID is required"
	}
	return ""
}
