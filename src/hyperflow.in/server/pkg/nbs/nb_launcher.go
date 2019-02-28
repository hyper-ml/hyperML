package nbs

import(
  "k8s.io/client-go/kubernetes"
  "k8s.io/client-go/tools/clientcmd" 
  apps_v1 "k8s.io/api/apps/v1"
  core_v1 "k8s.io/api/core/v1"
  meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

  "hyperflow.in/server/pkg/base"

)

const (
  defaultImage = "jupyterhub/k8s-singleuser-sample"
  defaultReplicas = 1
  defaultNotebookPort = "8888"
  defaultImagePullPolicy = "IfNotPresent"
)

var (
  podImageFailures = map[string]bool{
    //"CrashLoopBackOff": true,
    "InvalidImageName": true,
    "ErrImagePull":     true,
  }
)

type ComputeOptions struct {
  deployId string 
  labels map[string]string 
  annots map[string]string

  image string
  imagePullPolicy string

  nreplicas int32

  volumes          []core_v1.Volume       
  volumeMounts     []core_v1.VolumeMount 

  resourceReq *core_v1.ResourceList 
  resLimits *core_v1.ResourceList
  envVars []core_v1.EnvVar

  notebookPort string

}

type NbLauncher struct {
  k8client *kubernetes.Clientset
}

func NewNbLauncher(){
  c, _ := GetDefaultKubeClient()
  return &NbLauncher {
    k8client: c,
  }
}


func GetDefaultKubeClient() (*kubernetes.Clientset, error) {
  config_path := hflow_config.GetK8sConfigPath()
  base.Info("[GetDefaultKubeClient] Kube Config: ", config_path)

  config, err := clientcmd.BuildConfigFromFlags("", config_path)
  if err != nil {
    return nil, err
  }
  return kubernetes.NewForConfig(config)
}

func (l *NbLauncher) genOptions(nbId string) (*ComputeOptions, error) {

  var image string = defaultImage

  labels:= map[string]string{} 
  labels["nbId"] = nbId
  labels["type"] = "notebook" 

  return &ComputeOptions {
    deployId: nbId,
    notebookPort: defaultNotebookPort,
    nreplicas: defaultReplicas, 
    image: image,
    imagePullPolicy: defaultImagePullPolicy,
  }, nil
}

func (pk *PodKeeper) genDeploySpec(opts *ComputeOptions) (*apps_v1.Deployment) {
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


func (l *NbLauncher) genWorkerServiceSpec(opts *ComputeOptions) (*core_v1.Service, error) {
  
  var ws_spec *core_v1.Service 
  ws_spec = &core_v1.Service {
    TypeMeta: meta_v1.TypeMeta {
      Kind:       "Service",
      APIVersion: "v1",
    },
    ObjectMeta: meta_v1.ObjectMeta {
      Name:        "notebook-client-" + opts.deployId,
      Labels:      opts.labels, 
    },
    Spec: core_v1.ServiceSpec {
      Selector: opts.labels,
      Ports: []core_v1.ServicePort {
        { Port: opts.notebookPort,
          Name: "notebook-port",
        },
      },
    },
  }

  return ws_spec, nil
}

func (l *NbLauncher) genPodTemplate(opts *ComputeOptions) (core_v1.PodTemplateSpec, error) {

  pod_spec := core_v1.PodSpec{}
  container_name := "notebook: " + opts.deployId
  zero_value := int64(0)

  pod_spec = core_v1.PodSpec {
    Containers: []core_v1.Container {
      {
        Name: container_name,
        Image: opts.image,
        Command: []string{"jupyterhub-singleuser"},
        ImagePullPolicy: core_v1.PullPolicy(opts.imagePolicy),
//        VolumeMounts: opts.volumeMounts,
      },
    },
    RestartPolicy: defaultRestartPolicy,
    TerminationGracePeriodSeconds: &zero_value,
//    Volumes: opts.volumes,

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


func (l *NbLauncher) regWorkerService(ns string, spec *core_v1.Service) error {
  if _, err:= l.k8client.CoreV1().Services(ns).Create(spec);  scrubError(err) != nil {
    return err
  }
  return nil
}

func (l *NbLauncher) NewNotebook(nbId string) error {
  user_opts, _ := l.genOptions(nbId)

  d_spec := l.genDeploySpec(user_opts)
  pod_template, _ := l.genPodTemplate(user_opts)
  d_spec.Spec.Template = pod_template 
  ws_spec, _ := l.genWorkerServiceSpec(user_opts)  
  result, err := l.kubeClient.AppsV1().Deployments(core_v1.NamespaceDefault).Create(rc_spec)
  if err != nil {
    return err
  }

  if err := l.regWorkerService(core_v1.NamespaceDefault, ws_spec); scrubError(err) != nil {
    base.Error("[NbLauncher.NewNotebook] Worker Service Creation failed: ", err)
    return err 
  }

  return nil  


}
