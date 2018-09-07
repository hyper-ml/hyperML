package flow

//TODO: add checks to check if pod IP is active. Remove task assignment if not and restart another worker.
// may interfere with kube adm. 

import(
  "io"
  "fmt"
  "time"
  "strings"
  "bytes"
  "hyperview.in/server/base"

  "flag"
  "path/filepath"


  "k8s.io/client-go/kubernetes"
  "k8s.io/client-go/tools/clientcmd"
  "k8s.io/client-go/util/homedir"
  kwatch "k8s.io/apimachinery/pkg/watch"

  //kn_rest "k8s.io/client-go/rest"
  apps_v1 "k8s.io/api/apps/v1"
  core_v1 "k8s.io/api/core/v1"
  meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
 
  db_pkg "hyperview.in/server/core/db"
  //task_pkg "hyperview.in/server/core/tasks"
  "hyperview.in/server/core/storage"

)

const (
  defaultImage = "workhorse:0.2" 
  defaultReplicas = 1
  defaultPullPolicy = "IfNotPresent" // TODO "IfNotPresent"
  SERVER_ADDR = "http://192.168.65.1"
  defaultNameSpace = "hyperflow"
  defaultRestartPolicy = "Always"
)

var (
  podFailureReasons = map[string]bool{
    "CrashLoopBackOff": true,
    "InvalidImageName": true,
    "ErrImagePull":     true,
  }
)


// generate packages / functions for creating/destroying workers
type PodKeeper struct { 
  db *db_pkg.DatabaseContext
  qs *queryServer
  kubeClient *kubernetes.Clientset
  podWatcher kwatch.Interface 
  logStorageServer storage.ObjectAPIServer
}

func NewDefaultPodKeeper(db *db_pkg.DatabaseContext) *PodKeeper {
  c, _ := GetDefaultKubeClient()
  logStorage, _ := storage.NewObjectAPI("logs", 0, storage.GoogleStorage)
  queryServer:= NewQueryServer(db)

  return &PodKeeper {
    kubeClient: c,
    logStorageServer: logStorage,
    qs: queryServer,
  }
}

// Worker/Pod Details to generate kubernetes namespace
type ComputeOptions struct {

  // kubernetes deployment 
  deployId string 
  
  flowAttrs *FlowAttrs 
  currentTaskId string

  labels map[string]string 
  annots map[string]string
  nreplicas int32

  containerImage string
  containerPullPolicy string

  resourceReq *core_v1.ResourceList 
  resLimits *core_v1.ResourceList
  envVars []core_v1.EnvVar
}

func errInvalidFlowAttrs() error{
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


func GetDefaultKubeClient() (*kubernetes.Clientset, error) {
  var kubeconfig *string
  if home := homedir.HomeDir(); home != "" {
    kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
  } else {
    kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
  }
  flag.Parse()
  base.Log("[GetDefaultKubeClient] Kube Config: ", *kubeconfig)

  config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
  if err == nil {
    base.Log("[podKeeper.GetDefaultKubeClient] Using Default config file from home dir:", config)
    return kubernetes.NewForConfig(config)
  }

  base.Log("[podKeeper.GetDefaultKubeClient] Failed to get config file from home: ", err )
  return nil, err
}
  
func (pk PodKeeper) genOptions(taskId string, flowAttrs *FlowAttrs) (nsOpts *ComputeOptions, err error) {

  defer func() {
      if err != nil {
        base.Debug("[podKeeper.genOptions] Namespace Options :", nsOpts)
      }      
    }()

  var image string
  var deploy_id string

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
  base.Debug("[PodKeeper.genOptions] Image: ", image)

  deploy_id = getDeployId(flowAttrs.Flow.Id, flowAttrs.Flow.Version)

  //TODO: add version to flows
  labels:= map[string]string{} 
  labels["deployId"] = deploy_id
  labels["type"] = "worker"

  return  &ComputeOptions{
    deployId: deploy_id,
    flowAttrs: flowAttrs,
    currentTaskId: taskId, 
    containerImage:  image,
    nreplicas: defaultReplicas,
    labels: labels, 
  }, nil
}

// skip names space if exists
func (pk *PodKeeper) genRcSpec(opts *ComputeOptions) (*apps_v1.Deployment) {
  
  base.Log("[podKeeper.genRcSpec] Deploy Id: ", opts.deployId)

  return &apps_v1.Deployment{
    TypeMeta: meta_v1.TypeMeta{
      Kind: "Deployment",
      APIVersion: "v1",
    },
    ObjectMeta: meta_v1.ObjectMeta{
      Name: opts.deployId,
      Labels: opts.labels,
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

func resourceBounds(opts *ComputeOptions) core_v1.ResourceRequirements{
  res_req := core_v1.ResourceRequirements{}

  if opts.resourceReq != nil {
    res_req.Requests = *opts.resourceReq
  }

  if opts.resLimits != nil {
    res_req.Limits = *opts.resLimits
  }

  return res_req
}


func (pk *PodKeeper) genPodTemplate(opts *ComputeOptions) (core_v1.PodTemplateSpec, error) {

  pod_spec := core_v1.PodSpec{}
  container_name := "task-" + opts.currentTaskId
  zero_value := int64(0)

  policy := imagePolicy(opts)

  env := []core_v1.EnvVar{{
      Name: "API_SERVER_IP",
      Value: "192.168.64.1", 
      // IP of localhost is defaulted from topology of minikube ip and 1 at the end
      /*ValueFrom: &core_v1.EnvVarSource{
        FieldRef: &core_v1.ObjectFieldSelector{
          APIVersion: "v1",
          FieldPath:  "status.hostIP", },
        },*/
    }, {
      Name: "API_SERVER_PORT",
      Value: "8888",
    },{
      Name: "API_SERVER_PROTOCOL",
      Value: "http://",
    },{
      Name: "FLOW_ID",
      Value: opts.flowAttrs.Flow.Id,
    }, {
      Name: "TASK_ID",
      Value: opts.currentTaskId,
    }, {
      Name: "WORKSPACE_DIR",
      Value: "/home",
    },  {
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
 
  pod_spec = core_v1.PodSpec {
    Containers: []core_v1.Container{
      {
        Name: container_name,
        Image: opts.containerImage,
        Command: []string{"workhorse"},
        Env: env,
        ImagePullPolicy: core_v1.PullPolicy(policy),
      },
    },
    RestartPolicy: defaultRestartPolicy,
    TerminationGracePeriodSeconds: &zero_value,
//    HostNetwork: true,
//    DNSPolicy: "ClusterFirstWithHostNet",
    //TODO: add service account name from providers
  }

  pod_spec.Containers[0].Resources = resourceBounds(opts)  
    
  return core_v1.PodTemplateSpec{
    ObjectMeta: meta_v1.ObjectMeta{
      Name: opts.deployId,
      Labels: opts.labels,
      Annotations: opts.annots,
    },
    Spec: pod_spec,
  }, nil
}

func (pk *PodKeeper) deleteDeployment(deployId string) error {

  selector := fmt.Sprintf("deployId=%s", deployId)
  deployer, err := pk.kubeClient.AppsV1().Deployments(core_v1.NamespaceDefault).List(meta_v1.ListOptions{LabelSelector: selector})

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
    if err := pk.kubeClient.AppsV1().Deployments(core_v1.NamespaceDefault).Delete(dn.Name, options); err != nil {
      base.Log("[PodKeeper.deleteDeployment] Failed to delete deployer: ", err)
      return err
    }
  }
  return nil 
}

// task Id? 
func (pk *PodKeeper) ReleaseWorker(flow Flow ) error {
  deploy_id:= getDeployId(flow.Id, flow.Version)
  return pk.deleteDeployment(deploy_id) 
}
 


// launch a new namespace config with K8s 
// 
func (pk *PodKeeper) AssignWorker(taskId string, flowAttrs *FlowAttrs) error{
  user_opts, err := pk.genOptions(taskId, flowAttrs)

  if err != nil {
    base.Log("[PodKeeper.AssignWorker] Failed to create flow worker options: ",flowAttrs.Flow.Id, err)
    return err
  }

  rc_spec := pk.genRcSpec(user_opts)
  pod_template, err := pk.genPodTemplate(user_opts)
  rc_spec.Spec.Template = pod_template
  //ns := "hyperflow_ns"

  result, err := pk.kubeClient.AppsV1().Deployments(core_v1.NamespaceDefault).Create(rc_spec)
  
  if err != nil {
    if !strings.Contains(err.Error(), "already exists") {
      base.Log("[PodKeeper.AssignWorker] Failed to create deployment for namespace: FlowId, TaskId", flowAttrs.Flow.Id, taskId)
      base.Log("[PodKeeper.AssignWorker] Error: ", err)
      return err
    }
  }

  base.Log("[PodKeeper.AssignWorker] Created deployment .\n", result.GetObjectMeta().GetName())

  return nil

}


func (pk *PodKeeper)  WorkerExists(flowId, taskId string) bool {
  return false
}

func(pk *PodKeeper)  Watch(eventCh chan WorkerEvent) {

  var pod_events <-chan kwatch.Event
  pod_watcher, err := pk.kubeClient.CoreV1().Pods(core_v1.NamespaceDefault).Watch(
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

  // start watching here 
  // return events as they are received in workerEvent channel 
  for {
    select {
      case pod_event := <- pod_events:
        wevt := WorkerEvent {}

        // TODO: check if kwatch.error is watch error
        //       pod_event.Type == kwatch.Error

        if  pod_event.Type == "" {
          continue
        }

        pod_evt_str := string(pod_event.Type)
        base.Log("[podKeeper.Watch] Pod event:", pod_evt_str)

        pod, ok := pod_event.Object.(*core_v1.Pod)
        if !ok  {
          continue
        }

        base.Log("[podKeeper.Watch] PodId: ", pod.Name, string(pod.Status.Phase))
        
        deploy_id := pod.ObjectMeta.Labels["deployId"]
        wevt.Flow = Flow{Id: deploy_id}
        wevt.Worker = Worker{PodId: pod.Name, PodPhase: string(pod.Status.Phase)}
        
        //flow_id := deploy_id  
        //task_attrs, _ := pk.qs.GetTaskByFlowId(flow_id, "")

        /*if task_attrs.Status >= task_pkg.TASK_COMPLETED ||
           task_attrs.Status == task_pkg.TASK_FAILED {
          // skip events as task is likely completed and  
          // deployment is being deleted
          base.Log("[PodKeeper.Watch] Skipping events as task is already completed")
          
          pk.deleteDeployment(deploy_id)
          continue
        } */

        if pod_evt_str == "MODIFIED" && len(pod.Status.ContainerStatuses) == 1 {
          if pod.Status.ContainerStatuses[0].State.Running != nil {
            // Do not escalate. Just capture log 
            pk.SaveWorkerLog(wevt.Worker, wevt.Flow)
            continue
          }
        }
        
        // if pod succeeded or failed then escalate 
        if pod.Status.Phase == core_v1.PodFailed ||
           pod.Status.Phase == core_v1.PodSucceeded || 
           pod.Status.Phase == core_v1.PodUnknown {
          base.Log("[PodKeeper.Watch] pod completed: ", pod.Status.Phase, pod.Status.Message)
          wevt.Type = WorkerEventType(WorkerSucceeded)
          eventCh <- wevt
          continue
        }


        // escalate any success or failures at container level
        for _, status := range pod.Status.ContainerStatuses {

          if status.State.Terminated != nil {
            if status.State.Terminated.ExitCode == 0 {
              base.Log("[PodKeeper.Watch] Contaner terminated with 0 code", status.State.Terminated)
              // TODO:  delete deployment here 
              // dont wait

              wevt.Type = WorkerEventType(WorkerSucceeded)
              eventCh <- wevt
              pk.SaveWorkerLog(wevt.Worker, wevt.Flow)              
              pk.deleteDeployment(deploy_id)
          
              continue
              }
            }

          if status.State.Running != nil { 
              wevt.Type = WorkerEventType(WorkerSucceeded)
              eventCh <- wevt
              pk.SaveWorkerLog(wevt.Worker, wevt.Flow)
              continue
            }

          if status.State.Waiting != nil && podFailureReasons[status.State.Waiting.Reason] {
              wevt.Type = WorkerEventType(WorkerError)
              eventCh <- wevt
              pk.SaveWorkerLog(wevt.Worker, wevt.Flow)
              pk.deleteDeployment(deploy_id)
          
              continue
            }
        }
    }
  }
  return 
}

func (pk *PodKeeper) CloseWatch() {
  pk.podWatcher.Stop()
}

func (pk *PodKeeper) SaveWorkerLog(worker Worker, flow Flow) (error) {
  var r io.Reader

  if worker.PodId == "" {
    base.Log("[PodKeeper.SaveWorkerLog] No Pod ID found")
    return nil
  }

  body, err := pk.kubeClient.CoreV1().Pods(core_v1.NamespaceDefault).GetLogs(
          worker.PodId, &core_v1.PodLogOptions{
            Container: "task-" + flow.Id,
          }).Timeout(10 * time.Second).Do().Raw()
  if err != nil {
    base.Log("[PodKeeper.SaveWorkerLog] Failed to get pod log: ", worker.PodId)
    return err
  }

  r = bytes.NewReader(body)
  task_logdir := "flow"  
  //TODO: consider appending log instead of overwriting
  obj_path, _, bytes_written, err := pk.logStorageServer.SaveObject(flow.Id + ".log", task_logdir, r, false)

  base.Log("[PodKeeper.SaveWorkerLog] Log written: ", obj_path, bytes_written, err)
  return nil
}

