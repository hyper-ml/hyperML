package types

import (
	"fmt"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	"path/filepath"
	"strings"
	"time"
)

// PodRequestMode : Pod Request Mode; can be Immediate or Scheduled
type PodRequestMode int

const (
	// IfNotPresent : Constant string
	IfNotPresent = "IfNotPresent"

	// Always : Constant
	Always = "Always"

	// OnFailure : Constant
	OnFailure = "OnFailure"

	// Notebook : PodType Notebook
	Notebook = "Notebook"

	// Lab : PodType LAB
	Lab = "Lab"

	// JobTypeNotebook : Job type of Notebook
	JobTypeNotebook = "Notebook"

	// JobTypeLab : Job type of LAB
	JobTypeLab = "Lab"

	// JobTypeSchedNotebook : Job type of scheduled notebook
	JobTypeSchedNotebook = "Scheduled Notebook"

	// JobTypeDefault : Default Job type if not Notebook or Lab
	JobTypeDefault = "Default"

	// PodReqModeSch : Pod Request Mode
	PodReqModeSch PodRequestMode = 1

	// PodReqModeImd : Pod Request Mode
	PodReqModeImd PodRequestMode = 2

	// PodReqModeBck : Pod Request Mode Background
	PodReqModeBck PodRequestMode = 3

	// DefaultExternalPort :
	DefaultExternalPort = 80

	// DefaultAppPort :
	DefaultAppPort = 8888

	// PodStatusUnknown : POD in Unknown Status
	PodStatusUnknown PodStatus = 0

	// PodReqScheduled :
	PodReqScheduled PodStatus = 10

	// PodRequested :
	PodRequested PodStatus = 100

	// PodReqInvalid :
	PodReqInvalid PodStatus = 200

	// PodReqInitiating :
	PodReqInitiating PodStatus = 300

	// PodReqInitiated :
	PodReqInitiated PodStatus = 400

	// PodReqInitiatedDeploy :
	PodReqInitiatedDeploy PodStatus = 410

	// PodReqServiceInitiated :
	PodReqServiceInitiated PodStatus = 420

	// PodReqIngressInitiated :
	PodReqIngressInitiated PodStatus = 430

	// PodReqCancelled :
	PodReqCancelled PodStatus = 500

	// PodInitiating :
	PodInitiating PodStatus = 600

	// PodRestarting :
	PodRestarting PodStatus = 700

	// PodRunning :
	PodRunning PodStatus = 1000

	// PodInit :
	PodInit PodStatus = 1100

	//PodDestroy :
	PodDestroy PodStatus = 1200

	// PodFailing :
	PodFailing PodStatus = 2000

	// PodCompleted :
	PodCompleted PodStatus = 2100

	// PodShuttingDown :
	PodShuttingDown PodStatus = 2200

	// PodTerminated :
	PodTerminated PodStatus = 2300

	// PodReqFailed :
	PodReqFailed PodStatus = 3000

	// PodInitFailed :
	PodInitFailed PodStatus = 3100

	// PodDestroyFailed :
	PodDestroyFailed PodStatus = 3200

	// PodRuntimeFailure :
	PodRuntimeFailure PodStatus = 3300

	// PodApplicationError :
	PodApplicationError PodStatus = 3400

	// PhaseScheduled :
	PhaseScheduled = 5

	// PhasePending :
	PhasePending = 10

	// PhaseRunning :
	PhaseRunning = 20

	// PhaseShutdown :
	PhaseShutdown = 40

	// PhaseFailed :
	PhaseFailed = 50

	// PhaseUnknown :
	PhaseUnknown = 0
)

// PodStatus : determines the state of POD. depends on phase
type PodStatus int

// PodStatusDesc : Returns meaning for a phase code
func PodStatusDesc(code PodStatus) string {
	switch code {
	case PodReqInitiating:
		return "Initiating"
	case PodRequested:
		return "Requested"
	case PodReqInvalid:
		return "Invalid"
	case PodReqInitiated:
		return "Initiated"
	case PodReqInitiatedDeploy:
		return "InitiatedDeploy"
	case PodReqServiceInitiated:
		return "ServiceInitiated"
	case PodReqIngressInitiated:
		return "Ingress Initiated"
	case PodReqCancelled:
		return "Cancelled"
	case PodInitiating:
		return "Initiating"
	case PodRestarting:
		return "Restarting"

	case PodRunning:
		return "Running"

	case PodInit:
		return "Initiating Pod"

	case PodDestroy:
		return "Destroying Pod"

	case PodFailing:
		return "Failing"

	case PodCompleted:
		return "Completed"

	case PodShuttingDown:
		return "Shutting Down"

	case PodTerminated:
		return "Terminated"

	case PodReqFailed:
		return "Request Failed"

	case PodInitFailed:
		return "Pod Init failed"

	case PodDestroyFailed:
		return "Pod Destroy Failed"

	case PodRuntimeFailure:
		return "Pod Runtime Failure"

	case PodApplicationError:
		return "Application Failure"
	default:
		return "Unknown status"
	}
}

// PodPhase : determines the current phase of POD
type PodPhase int

// PodPhaseDesc : Returns meaning for a phase code
func PodPhaseDesc(code PodPhase) string {
	switch code {
	case PhaseScheduled:
		return "Scheduled"
	case PhasePending:
		return "Pending"
	case PhaseRunning:
		return "Running"
	case PhaseShutdown:
		return "Shutdown"
	case PhaseFailed:
		return "Failed"
	}

	return "Unknown"
}

// POD : Interface used by other packages
//type POD interface {
//	GetID() uint64
//}

// POD :
type POD struct {
	ID          uint64
	Name        string
	PodType     string // NB for notebook, LAB for jupyter lab
	RequestMode PodRequestMode
	CreatedBy   *User
	AuthToken   string

	// User Key is reference used for identifying the pod by user
	UserKey     string
	ServiceName string
	ServicePort int32

	Config *PODConfig
	Phase  PodPhase
	Status PodStatus
	// Initiated        time.Time
	// InitiatedDeploy  time.Time
	// ServiceInitiated time.Time
	// IngressInitiated time.Time

	Started       time.Time
	Ended         time.Time
	Created       time.Time
	Updated       time.Time
	Failed        time.Time
	Cancelled     time.Time
	FailureReason string
	Endpoint      string
	Result        string
}

// NewPOD :
func NewPOD(user *User, jobType string, requestMode PodRequestMode, c *PODConfig) *POD {
	//TODO : generate ID

	if jobType == "" {
		jobType = JobTypeDefault
	}

	pod := &POD{
		Config:      c,
		CreatedBy:   user,
		PodType:     jobType,
		Created:     time.Now(),
		Updated:     time.Now(),
		Status:      PodRequested,
		Phase:       PhasePending,
		RequestMode: requestMode,
	}

	if jobType == Notebook || jobType == Lab {
		pod.ServicePort = 8888
	}

	return pod
}

// GetID : Returns POD ID
func (p *POD) GetID() uint64 {
	return p.ID
}

// IsPending : returns true if POD has not started yet
func (p *POD) IsPending() bool {
	if p.Phase == PhasePending {
		return true
	}
	return false
}

// IsDone : Returns true if pod has not ended yet
func (p *POD) IsDone() bool {
	if p.Phase == PhaseFailed ||
		p.Phase == PhaseShutdown {
		return true
	}
	return false
}

// Terminate : Initiate terminate status on POD
func (p *POD) Terminate() {
	p.SetStatus(PodTerminated, nil)
}

// SetRequestMode : Sets POD request Mode
func (p *POD) SetRequestMode(m PodRequestMode) {
	fmt.Println("p:", p)
	fmt.Println("m:", m)
	p.RequestMode = m
}

// SetStatus :
func (p *POD) SetStatus(s PodStatus, err error) {
	fmt.Println("SetStatus:", s, err)
	p.Updated = time.Now()

	if err != nil {
		p.FailureReason = p.FailureReason + "	" + err.Error()
	}

	switch s {
	case PodRequested,
		PodReqInitiated,
		PodReqInitiatedDeploy,
		PodReqServiceInitiated,
		PodReqIngressInitiated,
		PodInitiating,
		PodRestarting:
		p.Phase = PhasePending
		p.Status = s

	case PodRunning,
		PodInit,
		PodDestroy,
		PodInitFailed:
		p.Phase = PhaseRunning
		p.Status = s
	case PodFailing,
		PodCompleted,
		PodShuttingDown,
		PodTerminated,
		PodReqCancelled:
		p.Phase = PhaseShutdown
		p.Status = s
	case PodReqFailed,
		PodDestroyFailed,
		PodRuntimeFailure,
		PodApplicationError:
		p.Phase = PhaseFailed
		p.Status = s
	case PodReqScheduled:
		p.Phase = PhaseScheduled
		fmt.Println("phase scheduled:", s)
		p.Status = s
	default:
		base.Warn("Unhandled status: ", s)
		p.Phase = PhaseUnknown
		p.Status = s
	}

}

// PODConfig : represents a compute pod launched for notebooks
type PODConfig struct {
	*PODParams
	Resources       *ResourceProfile
	Image           *ContainerImage
	ImagePullPolicy string
	Command         string
	AppPort         uint16
	ExternalPort    uint16
	EnvVars         []*EnvVar
	PythonLibs      []*PythonLib
	Disk            *PersistentDisk
	RestartPolicy   string
}

// NewPODConfig : New POD Config
func NewPODConfig(rp *ResourceProfile, ci *ContainerImage, cmd string, dsk *PersistentDisk, params *PODParams) (*PODConfig, error) {

	return &PODConfig{
		PODParams:       params,
		Resources:       rp,
		Image:           ci,
		ImagePullPolicy: IfNotPresent,
		RestartPolicy:   OnFailure,
		Command:         cmd,
		Disk:            dsk,
		AppPort:         DefaultAppPort,
		ExternalPort:    DefaultExternalPort,
	}, nil
}

// PODParams : Additional Parameters for running job pods
type PODParams struct {

	// Input Notebook for scheduled / background notebooks
	InputPath string

	// Workspace Path (mostly s3 prefixed) to fetch s3 contents in the POD
	WorkspacePath string

	// Local Path where workspace contents need to be mounted
	MountPath string

	// Notebook output. Usually a notebook too
	OutputPath string

	// Persistent Disk host path in case volume persistence is enabled
	PersDiskHostPath string

	// Access / Secret key for saving or retrieving notebooks from s3
	// TODO: store in vault
	AwsAccessKey string
	AwsSecretKey string
}

// NewPodParams : Returns a new PodParams struct
func NewPodParams(inputpath, wspath, mountpath, outpath string) *PODParams {
	var out string
	if outpath != nullString {
		out = outpath
	} else {
		ext := filepath.Ext(inputpath)
		inputnoext := strings.Trim(inputpath, ext)
		out = inputnoext + OutPathPostfix + ext
	}
	return &PODParams{
		InputPath:     inputpath,
		WorkspacePath: wspath,
		MountPath:     mountpath,
		OutputPath:    out,
	}
}
