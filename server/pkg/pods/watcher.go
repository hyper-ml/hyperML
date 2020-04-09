package pods

import (
	"fmt"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	userTypes "github.com/hyper-ml/hyperml/server/pkg/types"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
)

// StartWatchers : watches a given namespace and deploy Type NOTEBOOK
// Todo handle failures
func (k *Keeper) StartWatchers(fail chan int) {
	go k.podMonitor(fail)
	go k.jobMonitor(fail)
}

func (k *Keeper) podMonitor(fail chan int) {
	base.Log("Initiating POD monitor for {userPodType: Notebook}")

	watcher, err := k.client.CoreV1().Pods(k.namespace).Watch(
		metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(
				metav1.SetAsLabelSelector(map[string]string{"userPodType": userTypes.Notebook})),
			Watch: true,
		},
	)

	if err != nil {
		return
	}

	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				base.Error("Watch stream is broken. Initiating Safe mode")
				fail <- 1
				return
			}

			if string(event.Type) == "" {
				continue
			}

			fmt.Println("Received POD event")
			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				continue
			}

			fmt.Println("**")
			fmt.Println("**")
			fmt.Println("**")

			// Get user Pod ID
			usrPodIDStr, ok := pod.ObjectMeta.Labels["userPodID"]
			if !ok {
				continue
			}
			fmt.Println("pod ID str: ", usrPodIDStr)

			upid, err := strconv.ParseUint(usrPodIDStr, 10, 64)
			if err != nil {
				continue
			}

			fmt.Println("POD Status: ", pod.Status)
			//if reqmode, ok := pod.ObjectMeta.Labels["requestMode"]; ok {
			//	if reqmode != string(userTypes.PodReqModeImd) {
			//		fmt.Println("skipping background or scheduled request")
			//		continue
			//	}
			//}
			k.updateUserPod(upid, pod)

		}
	}

}

func (k *Keeper) jobMonitor(fail chan int) {
	base.Log("Initiating Job monitor...")
	modeBck := fmt.Sprintf("%d", userTypes.PodReqModeBck)
	fmt.Println("Request mode monitoring: ", modeBck)

	watcher, err := k.client.BatchV1().Jobs(k.namespace).Watch(
		metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(
				metav1.SetAsLabelSelector(map[string]string{"requestMode": modeBck})),
			Watch: true,
		},
	)

	if err != nil {
		return
	}

	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				base.Error("Watch stream is broken. Initiating Safe mode")
				fail <- 1
				return
			}
			if string(event.Type) == "" {
				continue
			}
			fmt.Println("")
			fmt.Println("")

			fmt.Println("[Received Job event]")
			job, ok := event.Object.(*batchv1.Job)
			if !ok {
				continue
			}

			fmt.Println(job.Status)
			fmt.Println("**")
			fmt.Println("**")
			fmt.Println("**")

			// Get user Pod ID
			usrPodIDStr, ok := job.ObjectMeta.Labels["userPodID"]
			if !ok {
				continue
			}
			fmt.Println("pod ID str: ", usrPodIDStr)

			upid, err := strconv.ParseUint(usrPodIDStr, 10, 64)
			if err != nil {
				continue
			}

			fmt.Println("Job Status: ", job.Status)
			k.updateUserJob(upid, job)

		}
	}

}
