package pods

import (
	"fmt"

	"github.com/hyper-ml/hyperml/server/pkg/base"
	"github.com/hyper-ml/hyperml/server/pkg/config"
	qs_pkg "github.com/hyper-ml/hyperml/server/pkg/qs"
	types "github.com/hyper-ml/hyperml/server/pkg/types"
	"github.com/hyper-ml/hyperml/server/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwatch "k8s.io/apimachinery/pkg/watch"
	"strings"

	kns "k8s.io/client-go/kubernetes"
)

// Keeper : POD launcher
type Keeper struct {
	inactive   bool
	failReason string
	namespace  string
	client     *kns.Clientset
	config     *config.KubeConfig
	domain     string
	watcher    kwatch.Interface

	// used to signal checking k8s client connection
	// and initiate safe mode if k8s is unreachable
	checkConn chan int

	qs *qs_pkg.QueryServer
	lb *LoadBalancer
}

// NewKeeper :
func NewKeeper(conf *config.Config, qs *qs_pkg.QueryServer) (*Keeper, error) {
	namespace := conf.GetNS()
	if namespace == "" {
		namespace = DefaultNamespace
	}

	// Get serving domain for building POD URL
	domain := conf.GetDomain()

	client, err := getClient(conf.K8)
	if err != nil {
		return nil, err
	}

	if err := checkClient(client); err != nil {
		return nil, err
	}

	lb, _ := newLoadBalancer(namespace, client, domain)
	connCheck := make(chan int)

	k := &Keeper{
		namespace: namespace,
		client:    client,
		domain:    domain,
		config:    conf.K8,
		qs:        qs,
		lb:        lb,
		checkConn: connCheck,
	}

	if err := k.InitKeeper(); err != nil {
		return nil, fmt.Errorf("Failed to initiate POD Keeper: %v", err.Error())
	}

	return k, nil
}

// InitKeeper : Initialize ingress
func (k *Keeper) InitKeeper() error {

	if err := k.lb.Init(); err != nil {
		return err
	}
	// cache container images
	// set a flag in config to enable this

	// start listening to k8s failures
	go k.listenToConnFailures()

	// initiate POD Monitor
	k.StartWatchers(k.checkConn)

	return nil
}

func (k *Keeper) listenToConnFailures() {
	for {
		select {
		case <-k.checkConn:
			base.Warn("Received K8S connection failure")
			k.inactive = true
		}
	}
}

// IngressGC : Cleanup Ingress
func (k *Keeper) IngressGC() error {
	return nil
}

func genLabEndPoint(domain, userkey, token string) string {
	return domain + "/" + userkey + "/lab?token=" + token
}

// InitPOD : Initiates POD
func (k *Keeper) InitPOD(pod *types.POD) *types.POD {
	podID, _ := k.qs.NewPodID()
	pod.ID = podID

	if pod.PodType == types.Notebook && pod.RequestMode == types.PodReqModeImd {
		token, _ := utils.GetRandomStr(8)
		userKey, _ := utils.GetRandomStr(16)
		pod.UserKey = userKey
		pod.AuthToken = token

		pod.Endpoint = genLabEndPoint(k.domain, userKey, token)
		tokenMap := strings.NewReplacer("{token}", token, "{basePath}", userKey)
		addOnCmd := tokenMap.Replace(" --NotebookApp.token={token}  --NotebookApp.base_url={basePath}")
		pod.Config.Command = pod.Config.Command + " " + addOnCmd
	}

	return pod
}

// CreatePOD : Creates a new POD in kubernetes
func (k *Keeper) CreatePOD(pod *types.POD) (created *types.POD, err error) {

	defer func() {
		_, _ = k.qs.UpsertPOD(pod)

		if err != nil {
			if err = k.CleanupPOD(pod.PodType, pod.RequestMode, pod.ID); err != nil {
				base.Error("POD Cleanup failed: ", err.Error())
			}

		}

		// if error then delete deployment, service and ingress point (any that exists)
	}()

	if pod.ID == 0 {
		// no matter what happens record pod request in DB
		pod = k.InitPOD(pod)
	}

	pod.SetStatus(types.PodReqInitiated, nil)

	// make POD template
	podSpec, err := makePODTemplateSpec(pod)

	if err != nil {
		pod.SetStatus(types.PodReqFailed, err)
		return pod, err
	}

	// make Deployment spec
	deploySpec, err := makeDeploySpec(pod, podSpec)
	if err != nil {
		pod.SetStatus(types.PodReqFailed, err)
		return pod, err
	}

	// make deployment
	_, err = k.client.AppsV1().Deployments(k.namespace).Create(deploySpec)
	if err != nil {
		pod.SetStatus(types.PodReqFailed, err)
		return pod, err
	}

	pod.SetStatus(types.PodReqInitiatedDeploy, nil)

	// make Service
	srvSpec, err := makeServiceSpec(pod)
	if err != nil {
		// TODO: ROLLBACK Deployment
		pod.SetStatus(types.PodReqFailed, err)
		return pod, err
	}

	_, err = k.client.CoreV1().Services(k.namespace).Create(&srvSpec)
	if err != nil {
		pod.SetStatus(types.PodReqFailed, err)
		return pod, err
	}

	pod.SetStatus(types.PodReqServiceInitiated, nil)

	// Add to ingress
	if err := k.lb.AddPod(pod); err != nil {
		// ROLLBACK: Service and Deployment both ??
		pod.SetStatus(types.PodReqFailed, err)
	} else {
		pod.SetStatus(types.PodReqIngressInitiated, nil)
	}

	return pod, nil
}

// CleanupJob: Clean up job from k8s
func (k *Keeper) CleanupJob(podType string, podID uint64) error {
	return k.DeleteJob(podType, podID)
}

// CleanupPOD :
// TODO: remove service from ingress
func (k *Keeper) CleanupPOD(podType string, mode types.PodRequestMode, podID uint64) error {
	var depl *appsv1.Deployment
	var service *corev1.Service
	var deplerr error
	var serverr error

	switch mode {
	case types.PodReqModeBck:
		return k.DeleteJob(podType, podID)

	default:

		service, serverr = k.DeleteService(podType, podID)
		depl, deplerr = k.DeleteDeployment(podType, podID)

		if depl == nil && deplerr == nil && service == nil && serverr == nil {
			// delete POD
			err := k.qs.DeletePOD(podID)
			if err != nil {
				return err
			}
		}

		if deplerr != nil {
			return deplerr
		}

		return serverr
	}

}

// GetDeployment : Get K8S Deployment for given POD type and ID
func (k *Keeper) GetDeployment(PodType string, PodID uint64) (*appsv1.Deployment, error) {
	selector := fmt.Sprintf("userPodID=%d", PodID)
	deployments, err := k.client.AppsV1().Deployments(k.namespace).List(metav1.ListOptions{LabelSelector: selector})

	if err != nil {
		return nil, err
	}

	for _, deployment := range deployments.Items {
		if deployment.ObjectMeta.Name != "" {
			return &deployment, nil
		}
	}

	return nil, nil
}

func (k *Keeper) deleteDeployment(depl *appsv1.Deployment) error {
	return k.client.AppsV1().Deployments(k.namespace).Delete(depl.ObjectMeta.Name, makeDeletePolicySpec())
}

// DeleteDeployment : Deletes deployment by USER POD ID
func (k *Keeper) DeleteDeployment(podType string, podID uint64) (*appsv1.Deployment, error) {
	depl, err := k.GetDeployment(podType, podID)
	if err != nil {
		return nil, err
	}

	if depl != nil {
		err = k.deleteDeployment(depl)
		return depl, err
	}

	return depl, nil
}

// GetService : Get Service Rec
func (k *Keeper) GetService(PodType string, PodID uint64) (*corev1.Service, error) {
	srvName := makeServiceName(PodType, PodID)

	service, err := k.client.CoreV1().Services(k.namespace).Get(srvName, metav1.GetOptions{})
	if err != nil {
		errstr := err.Error()
		if !(strings.HasPrefix(errstr, "services") &&
			strings.HasSuffix(errstr, "not found")) {
			return nil, fmt.Errorf("Unable to find service : %v", err.Error())
		}

		return nil, nil
	}

	return service, nil

}

// DeleteService :
func (k *Keeper) DeleteService(PodType string, PodID uint64) (*corev1.Service, error) {
	srvName := makeServiceName(PodType, PodID)

	serviceObj, err := k.client.CoreV1().Services(k.namespace).Get(srvName, metav1.GetOptions{})
	if err != nil {
		if strings.HasSuffix(err.Error(), "not found") {
			return nil, nil
		}
		return nil, err
	}

	if serviceObj != nil || serviceObj.ObjectMeta.Name != "" {
		err = k.client.CoreV1().Services(k.namespace).Delete(srvName, makeDeletePolicySpec())
		fmt.Println("err:", err)

		if err != nil {
			//TODO : Queue failures for re-processing
			errstr := err.Error()
			if !(strings.HasPrefix(errstr, "services") &&
				strings.HasSuffix(errstr, "not found")) {
				return nil, fmt.Errorf("Unable to delete service : %v", err.Error())
			}
		}
	}

	return serviceObj, nil
}

// CreatePV : Creates persistent volume
func (k *Keeper) CreatePV() error {
	return nil
}

// CreateJob : Launches job pods
func (k *Keeper) CreateJob(pod *types.POD, jobConfig *config.JobConfig) (created *types.POD, err error) {

	defer func() {
		_, _ = k.qs.UpsertPOD(pod)

		if err != nil {
			if err = k.CleanupPOD(pod.PodType, pod.RequestMode, pod.ID); err != nil {
				base.Error("POD Cleanup failed: ", err.Error())
			}

		}

	}()

	if pod.ID == 0 {
		// no matter what happens record pod request in DB
		pod = k.InitPOD(pod)
	}

	pod.SetStatus(types.PodReqInitiated, nil)

	// make Job spec
	jobSpec, err := makeJobSpec(pod, jobConfig)
	if err != nil {
		pod.SetStatus(types.PodReqFailed, err)
		return pod, fmt.Errorf("Failed to make job spec: %v", err)
	}
	_, err = k.client.BatchV1().Jobs(k.namespace).Create(jobSpec)
	if err != nil {
		pod.SetStatus(types.PodReqFailed, err)
		return pod, err
	}

	pod.SetStatus(types.PodReqInitiated, nil)

	return pod, nil
}

// GetJob : Get job for a given POD
func (k *Keeper) GetJob(podType string, podID uint64) (*batchv1.Job, error) {
	selector := fmt.Sprintf("userPodID=%d", podID)
	jobs, err := k.client.BatchV1().Jobs(k.namespace).List(metav1.ListOptions{LabelSelector: selector})

	if err != nil {
		return nil, fmt.Errorf("Job not found: %v", err)
	}

	for _, job := range jobs.Items {
		if job.ObjectMeta.Name != "" {
			return &job, nil
		}
	}

	return nil, nil

}

// DeleteJob : Deletes job against a given POD
func (k *Keeper) DeleteJob(podType string, podID uint64) error {
	job, err := k.GetJob(podType, podID)
	if err != nil {
		return err
	}

	if job != nil {
		err = k.client.BatchV1().Jobs(k.namespace).Delete(job.ObjectMeta.Name, makeDeletePolicySpec())
		return err
	}

	return nil
}

// GetJobInfo : Returns a Job POD (k8s) tied to a given User Pod
func (k *Keeper) GetJobInfo(podType string, podID uint64) (*batchv1.Job, error) {
	selector := fmt.Sprintf("userPodID=%d", podID)
	jobs, err := k.client.BatchV1().Jobs(k.namespace).List(metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, err
	}

	for _, job := range jobs.Items {
		if job.ObjectMeta.Name != "" {
			return &job, nil
		}
	}
	return nil, fmt.Errorf("Not found")
}

// GetJobInfo : Returns a Job POD (k8s) tied to a given User Pod
func (k *Keeper) GetPodInfo(podType string, podID uint64) (*corev1.Pod, error) {
	selector := fmt.Sprintf("userPodID=%d", podID)
	pods, err := k.client.CoreV1().Pods(k.namespace).List(metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		if pod.ObjectMeta.Name != "" {
			return &pod, nil
		}
	}
	return nil, fmt.Errorf("Not found")
}
