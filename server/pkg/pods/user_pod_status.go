package pods

import (
	"bytes"
	"fmt"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	"github.com/hyper-ml/hyperml/server/pkg/types"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

func (k *Keeper) updateUserPod(userpid uint64, pod *corev1.Pod) (userPod *types.POD) {

	switch pod.Status.Phase {
	case phasePending:

		for _, cond := range pod.Status.Conditions {
			switch cond.Type {
			case corev1.PodScheduled:
				fmt.Println("Pod Condition:", cond)
				if strings.Contains(cond.Reason, "Unschedulable") {

					userPod, _ = k.updateUserStatus(userpid, phaseFailed, podUnschedulable, fmt.Sprintf("%s %s", cond.Reason, cond.Message))
					return userPod
				}

			}
		}

		// check container status
		if len(pod.Status.ContainerStatuses) == 0 {
			// no container yet
			userPod, _ = k.updateUserStatus(userpid, phasePending, containersNone, "")
			return
		}

		cstate := pod.Status.ContainerStatuses[0].State
		if cstate.Waiting != nil {
			// ignore non-image related failures
			// let the container be restarted
			fmt.Println("container state:", cstate.Waiting)

			if !containerFailures[cstate.Waiting.Reason] {
				userPod, _ = k.updateUserStatus(userpid, phasePending, creatingContainer, "")
				return
			}

			// update user POD with image failure and delete deployment
			userPod, _ = k.updateUserStatus(userpid, phaseFailed, containerInitFailed, "")
			// cleanup pod?

		}

	case phaseRunning:

		for _, cond := range pod.Status.Conditions {
			switch cond.Type {
			case corev1.PodReady:
				if cond.Status == corev1.ConditionFalse {
					fmt.Println("pod not ready:", cond)
					if strings.Contains(cond.Reason, "ContainersNotReady") {

						userPod, _ = k.updateUserStatus(userpid, "", containersNotReady, fmt.Sprintf("%s %s", cond.Reason, cond.Message))

						for _, cs := range pod.Status.ContainerStatuses {

							if cs.State.Terminated != nil {
								t := cs.State.Terminated
								if containerFailures[t.Reason] {
									var message bytes.Buffer
									message.WriteString("\n")
									message.WriteString("Exit Code: " + int32toString(t.ExitCode))
									message.WriteString("\n")
									message.WriteString("Reason: " + t.Reason)
									message.WriteString("\n")
									message.WriteString("Message: " + t.Message)
									message.WriteString("\n")

									userPod, _ = k.updateUserStatus(userpid, "", containerFailed, message.String())
									return userPod
								}
							}
							if cs.State.Waiting != nil {
								w := cs.State.Waiting
								fmt.Println("got waiting", w)
								if containerFailures[w.Reason] {
									userPod, _ = k.updateUserStatus(userpid, "", containerFailed, fmt.Sprintf("%s %s", w.Reason, w.Message))
									return userPod
								}
							}

						}

					}
				}
				fmt.Println("Pod Condition:", cond)

			}
		}

		// no failures so far so report all going well
		userPod, _ = k.updateUserStatus(userpid, phaseRunning, containerRunning, "")

	case phaseFailed:
		// update user POD phase and status
		userPod, _ = k.updateUserStatus(userpid, phaseFailed, containerFailed, "")
	case phaseSucceeded:
		userPod, _ = k.updateUserStatus(userpid, phaseSucceeded, containerCompleted, "")
	default:
		base.Warn("Unhandled Condition for Kubernetes phase:", pod.Status.Phase)
	}
	return
}

// updateUserStatus : Will update user pod status to hyper ML database
func (k *Keeper) updateUserStatus(userpid uint64, newPhase string, newStatus string, reason string) (*types.POD, error) {
	if userpid == 0 {
		return nil, nil
	}

	fmt.Println("current phase:", newPhase)

	userpod, err := k.qs.GetPOD(userpid)
	if err != nil {
		return nil, err
	}
	if newPhase != "" {
		userpod.Phase = toUserPodPhase(newPhase)
	}
	s := toUserPodStatus(newStatus)
	userpod.SetStatus(s, fmt.Errorf(reason))

	userpod, err = k.qs.UpdateUserPOD(userpod)
	if err != nil {
		base.Error("Failed to updates status of " + uint64toString(userpid) + " : " + err.Error())
	}
	return userpod, err
}

// updateUserJob : Updates user pod request with recent job status from k8s
func (k *Keeper) updateUserJob(userpid uint64, job *batchv1.Job) (userPod *types.POD) {
	var newStatus string

	//if job.Status.Active > 0 {
	//	newStatus = jobActive
	//	userPod, _ = k.updateUserStatus(userpid, "", newStatus, "Worker Assigned")
	//	return
	//}

	for _, cond := range job.Status.Conditions {
		switch cond.Type {
		case batchv1.JobFailed:
			if cond.Status == corev1.ConditionTrue {
				// job failed

				newStatus = jobFailed

				userPod, _ = k.updateUserStatus(userpid, "", newStatus, cond.Reason)
				return
			}
		case batchv1.JobComplete:
			if cond.Status == corev1.ConditionTrue {
				// job completed
				newStatus = jobComplete
				userPod, _ = k.updateUserStatus(userpid, "", newStatus, cond.Reason)
				return
			}

		}
	}

	return
}

// SyncUserJob : Updates the status from Kubernetes POD to the user pod
func (k *Keeper) SyncUserJob(podType string, userpid uint64) (*types.POD, error) {
	job, err := k.GetJobInfo(podType, userpid)
	if err != nil {
		return nil, err
	}
	fmt.Println("job", job.Status)

	userPod := k.updateUserJob(userpid, job)
	//if userPod.IsDone() {
	//	if err = k.CleanupJob(userPod.PodType, userpid); err != nil {
	//		base.Error("Failed to cleanup job: ", err.Error())
	//	}
	//}
	return userPod, nil
}

// SyncUserPod : Updates the status from Kubernetes POD to the user pod
func (k *Keeper) SyncUserPod(podType string, userpid uint64) (*types.POD, error) {
	pod, err := k.GetPodInfo(podType, userpid)
	if err != nil {
		return nil, err
	}
	fmt.Println("pod status", pod.Status)
	return k.updateUserPod(userpid, pod), nil
}
