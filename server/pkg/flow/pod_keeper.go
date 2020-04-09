package flow

//TODO: add checks to check if pod IP is active. Remove task assignment if not and restart another worker.
// may interfere with kube adm.

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	kwatch "k8s.io/apimachinery/pkg/watch"
	kube "k8s.io/client-go/kubernetes"
	kuber "k8s.io/client-go/rest"
	kubecc "k8s.io/client-go/tools/clientcmd"

	apps_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/hyper-ml/hyperml/server/pkg/base"
	config "github.com/hyper-ml/hyperml/server/pkg/config"
	db_pkg "github.com/hyper-ml/hyperml/server/pkg/db"
	"github.com/hyper-ml/hyperml/server/pkg/storage"
	"github.com/hyper-ml/hyperml/server/pkg/utils/osutils"
)

const (
	defaultImage         = "hflow/worker:latest"
	defaultReplicas      = 1
	defaultPullPolicy    = "IfNotPresent"
	defaultNameSpace     = "hyperflow"
	defaultRestartPolicy = "Always"
	storageVolumeName    = "work_data"
	defaultWorkerPort    = 9999
	defaultFlowLogDir    = "flows"
	taskContainerPrefix  = "task-"
	defaultWorkspaceDir  = "/home/hflow"
)

var (
	podImageFailures = map[string]bool{
		//"CrashLoopBackOff": true,
		"InvalidImageName": true,
		"ErrImagePull":     true,
	}
)

// generate packages / functions for creating/destroying workers
type PodKeeper struct {
	namespace  string
	db         *db_pkg.DatabaseContext
	qs         *queryServer
	client     *kube.Clientset
	podWatcher kwatch.Interface
	logger     storage.ObjectAPIServer
	kubeConfig *config.KubeConfig
}

func GetClient(hfConfig *config.KubeConfig) (client *kube.Clientset, fnerr error) {
	var cluster_config *kuber.Config
	var err error

	if hfConfig.InCluster {
		cluster_config, err = kuber.InClusterConfig()
		if err != nil {
			return client, fmt.Errorf("failed to read incluster config: %v", err)
		}
		return kube.NewForConfig(cluster_config)
	}

	if hfConfig.Path != "" {
		if osutils.PathExists(hfConfig.Path) {

			cluster_config, err = kubecc.BuildConfigFromFlags("", hfConfig.Path)
			if err != nil {
				return client, fmt.Errorf("failed to create kubernetes client config with given path %s, err: %v", hfConfig.Path, err)
			}

		} else {
			return client, fmt.Errorf("invalid path in kubernetes config: %v", err)
		}

	} else {
		p, err := osutils.K8ConfigValidPath()
		if err != nil {
			return client, fmt.Errorf("failed to create kubernetes client with %s, err: %v", p, err)
		}

		cluster_config, err = kubecc.BuildConfigFromFlags("", p)
		if err != nil {
			return client, err
		}
	}

	return kube.NewForConfig(cluster_config)
}

func NewDefaultPodKeeper(
	config *config.Config,
	db db_pkg.DatabaseContext,
	logger storage.ObjectAPIServer) (*PodKeeper, error) {

	ns := config.K8.Namespace
	c, err := GetClient(config.K8)

	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes connection, err: ", err)
	}

	queryServer := NewQueryServer(db)

	if ns == "" {
		ns = defaultNameSpace
	}

	return &PodKeeper{
		namespace:  ns,
		client:     c,
		logger:     logger,
		qs:         queryServer,
		kubeConfig: config.K8,
	}, nil
}

// Worker/Pod Details to generate kubernetes namespace
type ComputeOptions struct {

	// kubernetes deployment
	deployId string

	flowAttrs     *FlowAttrs
	currentTaskId string

	labels    map[string]string
	annots    map[string]string
	nreplicas int32

	containerImage      string
	containerPullPolicy string

	//container volumes
	volumes      []core_v1.Volume
	volumeMounts []core_v1.VolumeMount

	resourceReq *core_v1.ResourceList
	resLimits   *core_v1.ResourceList
	envVars     []core_v1.EnvVar

	workerPort   int32
	workspaceDir string

	masterIp      string
	masterPort    int32
	masterExtPort int32

	// true when the master runs inside cluster
	InCluster bool
}

func errInvalidFlowAttrs() error {
	return fmt.Errorf("Passed flow Attributes are either null or Invalid. ")
}

func getDeployId(flowId, flowVersion string) string {
	/*var version = "0"
	  if flowVersion != "" {
	    version = flowVersion
	  }
	  return flowId + "." + version*/
	return flowId
}

func (pk PodKeeper) genOptions(taskId string, flowAttrs *FlowAttrs, masterIp string, masterPort int32, masterExtPort int32) (nsOpts *ComputeOptions, err error) {

	var image string
	var deploy_id string
	var in_cluster bool

	if flowAttrs.Flow.Id == "" || taskId == "" {
		return nil, errInvalidFlowAttrs()
	}

	task_attrs, ok := flowAttrs.Tasks[taskId]
	if !ok {
		return nil, errInvalidFlowAttrs()
	}

	image, ok = task_attrs.WorkerPref["ContainerImage"]
	if !ok {
		image = defaultImage
	}

	deploy_id = getDeployId(flowAttrs.Flow.Id, flowAttrs.Flow.Version)

	//TODO: add version to flows
	labels := map[string]string{}
	labels["deployId"] = deploy_id
	labels["type"] = "worker"

	var mounts []core_v1.VolumeMount
	var volume []core_v1.Volume
	var env_vars []core_v1.EnvVar

	volume = append(volume, core_v1.Volume{
		Name: "wh-data",
		VolumeSource: core_v1.VolumeSource{
			EmptyDir: &core_v1.EmptyDirVolumeSource{},
		},
	})

	if flowAttrs.EnvVars != nil {

		for k, v := range flowAttrs.EnvVars {
			env_vars = append(env_vars, core_v1.EnvVar{
				Name:  k,
				Value: v})
		}
	}

	mounts = append(mounts, core_v1.VolumeMount{
		Name:      "wh-data",
		MountPath: "/wh_data",
	})

	w_port := int32(defaultWorkerPort)

	if pk.kubeConfig != nil {
		in_cluster = pk.kubeConfig.InCluster
	}

	return &ComputeOptions{
		workspaceDir:   defaultWorkspaceDir,
		deployId:       deploy_id,
		flowAttrs:      flowAttrs,
		currentTaskId:  taskId,
		containerImage: image,
		nreplicas:      defaultReplicas,
		labels:         labels,
		volumes:        volume,
		volumeMounts:   mounts,
		workerPort:     w_port,
		masterIp:       masterIp,
		masterPort:     masterPort,
		envVars:        env_vars,
		InCluster:      in_cluster,
		masterExtPort:  masterExtPort,
	}, nil
}

// skip names space if exists
func (pk *PodKeeper) genDeploySpec(opts *ComputeOptions) *apps_v1.Deployment {

	//base.Log("[podKeeper.genDeploySpec] Deploy Id: ", opts.deployId)

	return &apps_v1.Deployment{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "v1",
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:        opts.deployId,
			Labels:      opts.labels,
			Annotations: opts.annots,
		},
		Spec: apps_v1.DeploymentSpec{
			Selector: &meta_v1.LabelSelector{
				MatchLabels: map[string]string{
					"deployId": opts.deployId,
				},
			},
			Replicas: &opts.nreplicas,
		},
	}
}

func imagePolicy(opts *ComputeOptions) string {
	policy := opts.containerPullPolicy

	if policy == "" {
		policy = defaultPullPolicy
	}

	return policy
}

func resourceBounds(opts *ComputeOptions) core_v1.ResourceRequirements {
	res_req := core_v1.ResourceRequirements{}

	if opts.resourceReq != nil {
		res_req.Requests = *opts.resourceReq
	}

	if opts.resLimits != nil {
		res_req.Limits = *opts.resLimits
	}

	return res_req
}

func (pk *PodKeeper) genWorkerServiceSpec(opts *ComputeOptions) (*core_v1.Service, error) {

	var ws_spec *core_v1.Service
	ws_spec = &core_v1.Service{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:   "worker-service-" + opts.deployId,
			Labels: opts.labels,
		},
		Spec: core_v1.ServiceSpec{
			Selector: opts.labels,
			Ports: []core_v1.ServicePort{
				{Port: opts.workerPort,
					Name: "worker-port",
				},
			},
		},
	}

	return ws_spec, nil
}

func getContainerName(taskId string) string {
	return taskContainerPrefix + taskId
}

func (pk *PodKeeper) genPodTemplate(opts *ComputeOptions) (core_v1.PodTemplateSpec, error) {

	pod_spec := core_v1.PodSpec{}
	container_name := getContainerName(opts.currentTaskId)
	zero_value := int64(0)

	policy := imagePolicy(opts)

	master_port := opts.masterPort

	if opts.masterExtPort > 0 {
		master_port = opts.masterExtPort
	}

	if master_port == 0 {
		master_port = config.DefaultMasterPort
	}

	base.Println(" master port: ", master_port)

	env := []core_v1.EnvVar{{
		Name:  "API_SERVER_PORT",
		Value: fmt.Sprint(master_port),
	}, {
		Name:  "API_SERVER_PROTOCOL",
		Value: "http://",
	}, {
		Name:  "FLOW_ID",
		Value: opts.flowAttrs.Flow.Id,
	}, {
		Name:  "TASK_ID",
		Value: opts.currentTaskId,
	}, {
		Name:  "PYTHONUNBUFFERED",
		Value: "1",
	}, {
		Name:  "WORKSPACE_DIR",
		Value: opts.workspaceDir,
	}, {
		Name: "WORKER_IP",
		ValueFrom: &core_v1.EnvVarSource{
			FieldRef: &core_v1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "status.podIP",
			},
		},
	}, {
		Name: "WORKER_NAME",
		ValueFrom: &core_v1.EnvVarSource{
			FieldRef: &core_v1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.name",
			},
		},
	}, {
		Name: "HOST_IP",
		ValueFrom: &core_v1.EnvVarSource{
			FieldRef: &core_v1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "status.hostIP",
			},
		},
	},
	}

	if opts.InCluster {
		env = append(env, core_v1.EnvVar{
			Name: "API_SERVER_IP",
			ValueFrom: &core_v1.EnvVarSource{
				FieldRef: &core_v1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.hostIP",
				},
			},
		})

	} else {
		env = append(env, core_v1.EnvVar{
			Name:  "API_SERVER_IP",
			Value: opts.masterIp,
		})
	}

	env = append(env, opts.envVars...)
	base.Info("[podKeeper.genPodTemplate] env variables added: ", env)

	pod_spec = core_v1.PodSpec{
		Containers: []core_v1.Container{
			{
				Name:            container_name,
				Image:           opts.containerImage,
				Command:         []string{"workhorse"},
				Env:             env,
				ImagePullPolicy: core_v1.PullPolicy(policy),
				VolumeMounts:    opts.volumeMounts,
			},
		},
		RestartPolicy:                 defaultRestartPolicy,
		TerminationGracePeriodSeconds: &zero_value,
		Volumes:                       opts.volumes,
		//    HostNetwork: true,
		//    DNSPolicy: "ClusterFirstWithHostNet",
		//TODO: add service account name from providers
	}

	pod_spec.Containers[0].Resources = resourceBounds(opts)

	return core_v1.PodTemplateSpec{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:        opts.deployId,
			Labels:      opts.labels,
			Annotations: opts.annots,
		},
		Spec: pod_spec,
	}, nil
}

func (pk *PodKeeper) deleteDeployment(deployId string) error {

	base.Info("[PodKeeper.deleteDeployment] Deleting flow deployment: ", deployId)

	selector := fmt.Sprintf("deployId=%s", deployId)
	deployer, err := pk.client.AppsV1().Deployments(pk.namespace).List(meta_v1.ListOptions{LabelSelector: selector})

	if err != nil {
		base.Log("[PodKeeper.deleteDeployment] Failed to find deployment record.", deployId)
		return err
	}

	if deployer == nil {
		base.Log("[PodKeeper.deleteDeployment] No deployment record found. The worker is probably already released.")
		return nil
	}

	delete_policy := meta_v1.DeletePropagationForeground
	//false_value := false
	options := &meta_v1.DeleteOptions{
		///OrphanDependents: &false_value,
		PropagationPolicy: &delete_policy,
	}

	for _, dn := range deployer.Items {
		if err := pk.client.AppsV1().Deployments(pk.namespace).Delete(dn.Name, options); err != nil {
			base.Log("[PodKeeper.deleteDeployment] Failed to delete deployer: ", err)
			return err
		}
	}
	return nil
}

// task Id?
func (pk *PodKeeper) ReleaseWorker(flow Flow) error {
	deploy_id := getDeployId(flow.Id, flow.Version)
	return pk.deleteDeployment(deploy_id)
}

func (pk *PodKeeper) regWorkerService(ns string, spec *core_v1.Service) error {

	if _, err := pk.client.CoreV1().Services(ns).Create(spec); scrubError(err) != nil {
		return err
	}

	return nil
}

// launch a new namespace config with K8s
//
func (pk *PodKeeper) AssignWorker(taskId string, flowAttrs *FlowAttrs, masterIp string, masterPort int32, masterExtPort int32) error {

	user_opts, err := pk.genOptions(taskId, flowAttrs, masterIp, masterPort, masterExtPort)

	if err != nil {
		base.Log("[PodKeeper.AssignWorker] Failed to create flow worker options: ", flowAttrs.Flow.Id, err)
		return err
	}
	rc_spec := pk.genDeploySpec(user_opts)
	pod_template, err := pk.genPodTemplate(user_opts)
	rc_spec.Spec.Template = pod_template
	ws_spec, err := pk.genWorkerServiceSpec(user_opts)

	result, err := pk.client.AppsV1().Deployments(pk.namespace).Create(rc_spec)

	if scrubError(err) != nil {
		base.Log("[PodKeeper.AssignWorker] Failed to create deployment for namespace: FlowId, TaskId", flowAttrs.Flow.Id, taskId)
		base.Log("[PodKeeper.AssignWorker] Error: ", err)
		return err
	}

	base.Log("[PodKeeper.AssignWorker] Created deployment .\n", result.GetObjectMeta().GetName())

	if err := pk.regWorkerService(pk.namespace, ws_spec); scrubError(err) != nil {
		base.Error("[PodKeeper.AssignWorker] Worker Service Creation failed: ", err)
		return err
	}

	return nil
}

func (pk *PodKeeper) WorkerExists(flowId, taskId string) bool {
	return false
}

func (pk *PodKeeper) Watch(eventCh chan WorkerEvent) {

	var pod_events <-chan kwatch.Event
	pod_watcher, err := pk.client.CoreV1().Pods(pk.namespace).Watch(
		meta_v1.ListOptions{
			LabelSelector: meta_v1.FormatLabelSelector(meta_v1.SetAsLabelSelector(
				map[string]string{
					"type": "worker",
				})),
			Watch: true,
		})

	if err != nil {
		return
	}

	pk.podWatcher = pod_watcher
	pod_events = pod_watcher.ResultChan()

	for {
		select {
		case pod_event := <-pod_events:
			wevt := WorkerEvent{}

			pod_evt_str := string(pod_event.Type)
			pod, ok := pod_event.Object.(*core_v1.Pod)
			if !ok || pod_evt_str == "" {
				continue
			}
			deploy_id := pod.ObjectMeta.Labels["deployId"]

			var container1_state core_v1.ContainerState
			if len(pod.Status.ContainerStatuses) > 0 {
				container1_state = pod.Status.ContainerStatuses[0].State
			}

			switch {
			case container1_state == core_v1.ContainerState{}:
				base.Error("Container state is empty for deploy Id %s", deploy_id)
				continue

			case container1_state.Waiting != nil:
				// container is waiting

				if !podImageFailures[container1_state.Waiting.Reason] {
					continue
				}

				failure_reason := "Image failure: " + container1_state.Waiting.Reason

				wevt.Flow = Flow{Id: deploy_id}
				wevt.Worker = Worker{PodId: pod.Name, PodPhase: string(pod.Status.Phase)}

				pk.SaveMessageToWorkerLog(failure_reason, wevt.Worker, wevt.Flow)

				wevt.Type = WorkerEventType(WorkerInitError)
				eventCh <- wevt

			case container1_state.Terminated != nil:
				// container failed
				deploy_id := pod.ObjectMeta.Labels["deployId"]

				wevt.Flow = Flow{Id: deploy_id}
				wevt.Worker = Worker{PodId: pod.Name, PodPhase: string(pod.Status.Phase)}
				pk.SaveWorkerLog(wevt.Worker, wevt.Flow)

				wevt.Type = WorkerEventType(WorkerError)
				eventCh <- wevt

			default:
				continue
			}
		}
	}
	return
}

func (pk *PodKeeper) CloseWatch() {
	pk.podWatcher.Stop()
}

func (pk *PodKeeper) SaveMessageToWorkerLog(s string, worker Worker, flow Flow) error {

	r := strings.NewReader(s)
	//TODO: find a way to append log
	_, _, _, err := pk.logger.SaveObject(flow.Id+".log", defaultFlowLogDir, r, false)
	return err
}

func (pk *PodKeeper) SaveWorkerLog(worker Worker, flow Flow) error {

	if worker.PodId == "" {
		return pk.saveLogWithFlowId(flow.Id)
	}
	return pk.SavePodLog(worker.PodId, "flows", flow.Id+".log")
}

func (pk *PodKeeper) saveLogWithFlowId(flowId string) error {

	selector := fmt.Sprintf("deployId=%s", flowId)
	pod_list, _ := pk.client.CoreV1().Pods(pk.namespace).List(meta_v1.ListOptions{LabelSelector: selector})

	if len(pod_list.Items) > 0 {
		for _, pod := range pod_list.Items {

			if pod.Name != "" {
				if err := pk.SavePodLog(pod.Name, "flow", flowId); err != nil {
					base.Error("[PodKeeper.saveLogWithFlowId] Error: ", err)
					return err
				}
			}
		}
	}

	return nil
}

func (pk *PodKeeper) SavePodLog(podId, logDir, logName string) error {
	var r io.Reader
	var log_name string
	var log_dir string

	if podId == "" {
		base.Log("[PodKeeper.SavePodLog] No Pod ID found")
		return nil
	}

	if logName == "" {
		log_name = "pod-" + podId + ".log"
		log_dir = defaultFlowLogDir
	} else {
		log_name = logName
		log_dir = logDir
	}

	body, err := pk.client.CoreV1().Pods(pk.namespace).GetLogs(
		podId, &core_v1.PodLogOptions{}).Timeout(10 * time.Second).Do().Raw()

	if err != nil {
		base.Log("[PodKeeper.SavePodLog] Failed to get pod log: ", podId, err)
		return err
	}

	r = bytes.NewReader(body)
	//TODO: consider appending log instead of overwriting
	obj_path, _, bytes_written, err := pk.logger.SaveObject(log_name, log_dir, r, false)

	base.Log("[PodKeeper.SavePodLog] Log written: ", obj_path, bytes_written, err)

	return nil
}

func (pk *PodKeeper) reconcile() error {
	// read tasks that are in CREATED or RUNNING stage.
	// check if a pod is running against the task
	// if not then start a pod and set status
	return fmt.Errorf("Unimplemented")
}

func (pk *PodKeeper) getPodId(flowId string) (podId string) {

	selector := fmt.Sprintf("deployId=%s", flowId)
	pod_list, _ := pk.client.CoreV1().Pods(pk.namespace).List(meta_v1.ListOptions{LabelSelector: selector})

	if len(pod_list.Items) == 0 {
		fmt.Println("no pod found")
		return
	}

	for _, pod := range pod_list.Items {
		if pod.Name != "" {
			return pod.Name
		}
	}

	return podId
}

// create a new channel and currespoding go rouine to call pod log
// return channel. let the reader read
func (pk *PodKeeper) LogStream(flowId string) (io.ReadCloser, error) {
	//write to channel
	var pod_id string
	pod_id = pk.getPodId(flowId)
	if pod_id == "" {
		return nil, fmt.Errorf("Worker is either finished or yet to start. Please try printing log without stream option.")
	}

	log_stream, err := pk.client.CoreV1().Pods(pk.namespace).GetLogs(pod_id, &core_v1.PodLogOptions{
		Follow: true,
	}).Stream()

	return log_stream, err
}
