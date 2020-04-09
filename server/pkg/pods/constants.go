package pods

import (
	"fmt"
	types "github.com/hyper-ml/hyperml/server/pkg/types"
)

const (
	// DefaultNamespace :
	DefaultNamespace = "hyperml"

	// DefaultIngressName : Default name for Ingress Controller
	DefaultIngressName = "hyperml-lb"

	nullString = ""

	// pod phases
	phaseRunning   = "Running"
	phaseFailed    = "Failed"
	phasePending   = "Pending"
	phaseSucceeded = "Succeeded"

	// pod conditions
	podScheduled     = "PodScheduled"
	podUnschedulable = "PodUnschedulable"

	// job conditions
	jobComplete = "JobComplete"
	jobFailed   = "JobFailed"
	jobActive   = "JobActive"

	// Container statuses
	containersNone         = "NoContainersYet"
	containerRunning       = "Container Running"
	containerInit          = "Container Initializing"
	containerCreatingImage = "Creating Image"
	creatingContainer      = "Creating Container"
	containerInitFailed    = "Container Init Failed"
	containerFailed        = "Container Failed"
	containerCompleted     = "Container succeeded"
	containerTerminated    = "Container Terminated"
	containerRestarting    = "Container Restarting"
	containersNotReady     = "Containers Not Ready"
)

var (
	// List of container serious failures that are not to be ignored
	// Pod needs to be terminated in all these scenarios
	containerFailures = map[string]bool{
		"CrashLoopBackOff":   true,
		"InvalidImageName":   true,
		"ErrImagePull":       true,
		"ContainerCannotRun": true,
		"RunContainerError":  true,
	}
)

func uint64toString(i uint64) string {
	return fmt.Sprintf("%d", i)
}

func int32toString(i int32) string {
	return fmt.Sprintf("%d", i)
}

// PhaseToUserPhase : Maps kubernetes phase to user pod phase
func toUserPodPhase(phase string) types.PodPhase {
	switch phase {
	case phasePending:
		return types.PhasePending
	case phaseRunning:
		return types.PhaseRunning
	case phaseFailed:
		return types.PhaseFailed
	case phaseSucceeded:
		return types.PhaseShutdown
	}
	return types.PhaseUnknown
}

// K8StatusToUserPodStatus : Maps kubernetes status to user pod status
func toUserPodStatus(kubestatus string) types.PodStatus {
	switch kubestatus {
	case podUnschedulable:
		return types.PodReqFailed
	case podScheduled:
		return types.PodReqScheduled
	case jobFailed:
		return types.PodReqFailed
	case jobComplete:
		return types.PodCompleted
	case jobActive:
		return types.PodRunning
	case containersNotReady, containerInitFailed:
		return types.PodInitFailed
	case containersNone:
		return types.PodInitiating
	case containerCreatingImage,
		creatingContainer,
		containerInit:
		return types.PodInitiating

	case containerRunning:
		return types.PodRunning
	case containerFailed:
		return types.PodReqFailed
	case containerCompleted:
		return types.PodCompleted
	case containerTerminated:
		return types.PodTerminated
	case containerRestarting:
		return types.PodRestarting
	default:
		return types.PodStatusUnknown
	}
}
