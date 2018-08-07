package flow


import(
  "fmt"
  "strings"
  "hyperview.in/server/base"

  "flag"
  "path/filepath"


  "k8s.io/client-go/kubernetes"
  "k8s.io/client-go/tools/clientcmd"
  "k8s.io/client-go/util/homedir"

  //kn_rest "k8s.io/client-go/rest"
  apps_v1 "k8s.io/api/apps/v1"
  core_v1 "k8s.io/api/core/v1"
  meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

)

const (
  defaultImage = "ubuntu:16.04"//"hyperflow/default"
  defaultReplicas = 1
  defaultPullPolicy = "IfNotPresent"
)

// generate packages / functions for creating/destroying workers
type PodKeeper struct { 
  kubeClient *kubernetes.Clientset
}

// Worker/Pod Details to generate kubernetes namespace
type ComputeOptions struct {

  // kubernetes deployment 
  deployId string 
  
  flowAttrs FlowAttrs 
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
  var version = "0"
  if flowVersion != "" {
    version = flowVersion
  } 
  return flowId + "." + version
}


func GetDefaultKubeClient() (*kubernetes.Clientset, error) {
  var kubeconfig *string
  if home := homedir.HomeDir(); home != "" {
    kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
  } else {
    kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
  }
  flag.Parse()

  config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
  if err == nil {
    base.Log("[podKeeper.GetDefaultKubeClient] Using Default config file from home dir:", config)
    return kubernetes.NewForConfig(config)
  }

  base.Log("[podKeeper.GetDefaultKubeClient] Failed to get config file from home: ", err )
  return nil, err
}

func NewDefaultPodKeeper() *PodKeeper {
  c, _ := GetDefaultKubeClient()
  return &PodKeeper {
    kubeClient: c,
  }
}


func (pk PodKeeper) genOptions(taskId string, flowAttrs FlowAttrs) (nsOpts *ComputeOptions, err error) {

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

  deploy_id = getDeployId(flowAttrs.Flow.Id, flowAttrs.Flow.Version)

  //TODO: add version to flows
  labels:= map[string]string{} 
  labels["deployId"] = deploy_id

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

/*func imagePolicy(opts *ComputeOptions) *core_v1.PullPolicy {
  policy := opts.containerPullPolicy

  if policy == "" {
    policy = defaultPullPolicy
  }

  return core_v1.PullPolicy(policy) 
}*/

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
  
  env := []core_v1.EnvVar{{
      Name: "HFLOW_USER_HOME",
      Value: "/workspace",
    }, {
      Name: "HFLOW_ID",
      Value: opts.flowAttrs.Flow.Id,
    }, {
      Name: "HFLOW_TASK_ID",
      Value: opts.currentTaskId,
      }}

  pod_spec = core_v1.PodSpec {
    Containers: []core_v1.Container{
      {
        Name: container_name,
        Image: opts.containerImage,
        Command: []string{"who"},
        Env: env,
      },
    },
    RestartPolicy: "Always",
    TerminationGracePeriodSeconds: &zero_value,
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

// task Id? 
func (pk *PodKeeper) ReleaseWorker(flow Flow ) error {

  selector := fmt.Sprintf("deployId=%s", getDeployId(flow.Id, flow.Version))

  deployer, err := pk.kubeClient.AppsV1().Deployments(core_v1.NamespaceDefault).List(meta_v1.ListOptions{LabelSelector: selector})

  if err != nil {
    base.Log("[PodKeeper.ReleaseWorker] Failed to find deployment record in k8s: ", flow.Id, flow.Version, err)
    return err
  }

  if deployer == nil {
    base.Log("[PodKeeper.ReleaseWorker] No deployment record found. The worker is probably already released.")
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
      base.Log("[PodKeeper.ReleaseWorker] Failed to delete deployer: ", err)
      return err
    }
  }
  return nil
}
 


// launch a new namespace config with K8s 
// 
func (pk *PodKeeper) AssignWorker(taskId string, flowAttrs FlowAttrs) error{
  user_opts, err := pk.genOptions(taskId, flowAttrs)

  if err != nil {
    base.Log("Failed to create flow worker options: ",flowAttrs.Flow.Id, err)
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


