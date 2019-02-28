package deploy

import(
  "fmt"
  "path"
  "io/ioutil"
  "encoding/json"
  v1 "k8s.io/api/core/v1"
  rbac_v1 "k8s.io/api/rbac/v1"
  meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  apps_v1 "k8s.io/api/apps/v1beta1"
  "k8s.io/apimachinery/pkg/api/resource"
 

  "hyperflow.in/server/pkg/base"
  hfConfig "hyperflow.in/server/pkg/config"

)
 
 
type StorageTarget string  

const (
  Local StorageTarget = "LOCAL"
  Amazon = "S3"
  Google = "GCS"
  Azure = "AZURE"
  Minio = "MINIO"
  None = "NONE"
)
 

var (
  // namespace 
  product_family = "hf" //hyperflow
  
  app_name = "hflserver"
  Version = "0.1.0"

  server_image = "hflow/server"

  hflow_secret_name = "hf-secret"
  tls_secret_key = "hf-tls-certs"
  tls_volume = "hf-tls-volume" 
  tls_mount_path = "/hf-tls-certs"

  service_acct = "hfuser"

  
  role_name = "hf-role"
  role_bind_name = "hf-role-binding"

  role_policy = []rbac_v1.PolicyRule{{
  	APIGroups:   []string{""},
  	Verbs:	    []string{"get", "list", "watch"},
  	Resources:   []string{"nodes", "pods", "pods/log", "endpoints"},
  }, {
    APIGroups: []string{"apps", "extensions"},
    Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
    Resources: []string{"deployments", "replicasets"},
  }, {
  	APIGroups: []string{""},
  	Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
  	Resources: []string{"replicationcontrollers", "services"},
  }, {
  	APIGroups:     []string{""},
  	Verbs:         []string{"get", "list", "watch", "create", "update", "delete"},
  	Resources:     []string{"secrets"},
  	ResourceNames: []string{hflow_secret_name},
  }}

  default_server_port = int32(8888)
  default_node_port = int32(30001)
  default_num_replicas = int32(1)
  default_worker_image = "hflow/worker"  
  default_storage_target = hfConfig.GCS
  default_mem_limit = "128M"
  default_cpu_limit = "0.25"
  
  default_db_volume_name = "db-volume"
  default_db_volume_claim = "db-storage"
  default_db_storage_size = "200Mi"  // binarySI
  default_db_bucket = "hflow_master" 
  default_file_mount_path = "/var/data"

  default_no_rbac = false
  default_no_auth = false 
  default_db_option = hfConfig.Badger

)
 

type DeployOptions struct {
	
	Namespace string
  PublicInterface string

	ServerIP string
  ServerPort int32 
  ServerNodePort int32

	StorageTarget hfConfig.StorageTarget

	NoRBAC bool
	NoAuth bool

	// incase bolt uses local host volume path
  DbType hfConfig.DbTarget
	DbFileMountPath string
	DbBucket string
  DbUsePVClaim bool

	DockerRegistry string
	ImagePullSecret string
  WorkerImage string

  NodeMemLimit string
	NodeCpuLimit string

	LogLevel string
	Version string
	
	TLSServerCert string
	TLSServerKey string

  APIServerConfig *hfConfig.Config
}

func DefaultDeployOpts() *DeployOptions {
  return &DeployOptions {
    Version: Version,
    Namespace: product_family,
    PublicInterface: ":" + string(default_server_port),
    ServerPort: default_server_port,
    ServerNodePort: default_node_port, 
    StorageTarget: hfConfig.StorageTarget(default_storage_target),
    NoRBAC: default_no_rbac,
    NoAuth: default_no_auth,
    DbType: hfConfig.Badger,
    DbUsePVClaim: true,
    DbFileMountPath: default_file_mount_path,
    DbBucket: default_db_bucket, 
    WorkerImage: default_worker_image,
    NodeMemLimit: default_mem_limit,
    NodeCpuLimit: default_cpu_limit,
  }
}

func prep_labels(tag string) map[string]string {
  return map[string]string{
    "app":            tag,
    "product_family": product_family,
    "release":        Version,
  }
}

func prep_objmeta(objname, tag, ns string, annots map[string]string) meta_v1.ObjectMeta{
  return meta_v1.ObjectMeta{
      Name:        objname,
      Labels:      prep_labels(tag),
      Annotations: annots,
      Namespace:   ns,
    }
}

func namespaceConfig(ns string) v1.Namespace{
  return v1.Namespace{
    TypeMeta: meta_v1.TypeMeta{
      Kind:        "Namespace",
      APIVersion:   "v1",
    }, 
    ObjectMeta: prep_objmeta(ns, app_name, "", nil),
  }
}
 
func serviceAccountConfig(o *DeployOptions) *v1.ServiceAccount {
  return &v1.ServiceAccount {
    TypeMeta: meta_v1.TypeMeta{
      Kind:         "ServiceAccount",
      APIVersion:   "v1",
    },
    ObjectMeta: prep_objmeta(service_acct, app_name, o.Namespace, nil),
  }
}

func roleConfig(o *DeployOptions) (role *rbac_v1.Role, role_binding *rbac_v1.RoleBinding) {
  role = &rbac_v1.Role{
    TypeMeta: meta_v1.TypeMeta{
      Kind:       "Role",
      APIVersion: "rbac.authorization.k8s.io/v1",
    },
    ObjectMeta: prep_objmeta(role_name, app_name, o.Namespace, nil),
    Rules:      role_policy,
  }

  role_binding = &rbac_v1.RoleBinding{
    TypeMeta: meta_v1.TypeMeta{
      Kind:       "RoleBinding",
      APIVersion: "rbac.authorization.k8s.io/v1",
    },
    ObjectMeta: prep_objmeta(role_bind_name, app_name, o.Namespace, nil),
    Subjects: []rbac_v1.Subject{{
      Kind:       "ServiceAccount",
      Name:       service_acct,
      Namespace:  o.Namespace,
    }},
    RoleRef: rbac_v1.RoleRef{
      Kind:       "Role",
      Name:       role_name,
    },
  }
  return 
}


func secretsVolumeConfig() (volume v1.Volume, mount v1.VolumeMount) {
  volume = v1.Volume {
    Name: hflow_secret_name,
    VolumeSource: v1.VolumeSource{
      Secret: &v1.SecretVolumeSource{
        SecretName: hflow_secret_name,
      },
    },
  }

  mount = v1.VolumeMount {
    Name: hflow_secret_name,
    MountPath: "/" + hflow_secret_name,
  }
  return 
}


func getImage(registry, imageName string) string {
  return getVersionedImage(registry, imageName, "")
}

func getVersionedImage(registry, imageName, version string) string {
  i := imageName

  if registry != "" {
    i = path.Join(registry, i)
  }
  if version != "" {
    i = i + ":" + version
  }

  return i
}


func TlsConfig(opts *DeployOptions) (*v1.Secret, error) {
  
  if opts.TLSServerCert == "" && opts.TLSServerKey == "" {
    return nil, nil
  } else if opts.TLSServerCert == ""{
    return nil, fmt.Errorf("Option TLS Server Certificate is not required when TLS Server Key is set")
  } else if opts.TLSServerKey == ""{
    return nil, fmt.Errorf("Option TLS Server Key is not required when TLS Server Certificate is set")
  }

  cert, err := ioutil.ReadFile(opts.TLSServerCert)
  if err != nil {
    return nil, fmt.Errorf("failed to read server cert " + opts.TLSServerCert + ", err:", err)
  }

  key, err := ioutil.ReadFile(opts.TLSServerKey)
  if err != nil {
    return nil, fmt.Errorf("failed to read server cert " + opts.TLSServerKey + ", err:", err)
  }

  data := map[string][]byte{
      "hflow_tls.crt": cert,
      "hflow_tls.key": key,
    }

  return &v1.Secret{
    TypeMeta: meta_v1.TypeMeta{
      Kind:       "Secret",
      APIVersion: "v1",
    },
    ObjectMeta: prep_objmeta(tls_secret_key, tls_secret_key, opts.Namespace, nil),
    Data: data,
  }, nil
}

func dbVolume(volumeType, claimName string) v1.Volume{
  if (volumeType == "PersistentVolume") {
    return v1.Volume{
      Name:         "db-storage",
      VolumeSource: v1.VolumeSource{
            PersistentVolumeClaim:  &v1.PersistentVolumeClaimVolumeSource{
              ClaimName: claimName,
          },
      },
    }
  }

  return v1.Volume{
      Name:         "db-storage",
      VolumeSource: v1.VolumeSource{
        EmptyDir: &v1.EmptyDirVolumeSource{},
      },
    } 
  
}

func dbVolumeMount(datapath string) v1.VolumeMount {
  return v1.VolumeMount{
      Name: "db-storage",
      MountPath: datapath,
  } 
}

func podConfig(opts *DeployOptions) v1.PodSpec {

  server_image := getVersionedImage(opts.DockerRegistry, server_image, Version)
  worker_image := getImage(opts.DockerRegistry, opts.WorkerImage) 

  mem_resource := resource.MustParse(opts.NodeMemLimit)
  cpu_resource := resource.MustParse(opts.NodeCpuLimit)
  volumes := []v1.Volume{
    {
      Name: "hflow",
    },
  }
  volume_mounts := []v1.VolumeMount{
    {
      Name: "hflow",
      MountPath: "/hflow",
    }, 
  }

  if (opts.DbUsePVClaim) {
    volumes = append(volumes, dbVolume("PersistentVolume", default_db_volume_claim))
    volume_mounts = append(volume_mounts, dbVolumeMount(opts.DbFileMountPath))
  } else {
    volumes = append(volumes, dbVolume("EmptyDir", default_db_volume_claim))
    volume_mounts = append(volume_mounts, dbVolumeMount(opts.DbFileMountPath))    
  }

  if opts.TLSServerCert != "" {
    volumes = append(volumes, v1.Volume{
      Name: tls_volume,
      VolumeSource: v1.VolumeSource{
          Secret: &v1.SecretVolumeSource{
              SecretName: tls_secret_key,
            },
        },
    })

    volume_mounts = append(volume_mounts, v1.VolumeMount{
        Name: tls_volume,
        MountPath: tls_mount_path,
      })
  }
  
  resource_req := v1.ResourceRequirements{
    Requests: v1.ResourceList{
      v1.ResourceCPU: cpu_resource,
      v1.ResourceMemory: mem_resource,
    },
  }   
  return v1.PodSpec{
    ServiceAccountName: service_acct,
    Volumes: volumes,
    /*ImagePullSecrets: []v1.LocalObjectReference{
      {
        Name: opts.ImagePullSecret,
      },  
    }, */ 
    Containers: []v1.Container{
      {
        Name: app_name,
        Image: server_image,
        Env: append([]v1.EnvVar{
          {Name: "HFLOW_ROOT", Value: "/home/hflow"},
          {Name: "DB_BUCKET_PREFIX", Value: opts.DbBucket},
          {Name: "STORAGE_TARGET", Value: string(opts.StorageTarget)},
          {Name: "WORKER_IMAGE", Value: worker_image},
          {Name: "IMAGE_PULL_SECRET", Value: opts.ImagePullSecret},
          {Name: "WORKER_IMAGE_PULL_POLICY", Value: "IfNotPresent"},
          {Name: "VERSION", Value: opts.Version},
          {Name: "LOG_LEVEL", Value: opts.LogLevel}, 
        }, secretVars()),
        
        Ports: []v1.ContainerPort{
          {  
            ContainerPort: opts.ServerPort,
            Protocol:      "TCP",
            Name:          "api-server-port",
          },
        }, 

        VolumeMounts: volume_mounts,
        ImagePullPolicy: "IfNotPresent",
        Resources: resource_req,
        // probes?

 
      },
    },

  }
}
 
func deploymentConfig(opts *DeployOptions) *apps_v1.Deployment {
  if opts.ServerPort == 0 {
    opts.ServerPort = default_server_port
  }
  
  type_meta := meta_v1.TypeMeta{
    Kind:       "Deployment",
    APIVersion: "apps/v1beta1",
  }

  obj_meta := meta_v1.ObjectMeta{
    Name:         app_name,
    Labels:       prep_labels(app_name),
    Annotations:  nil,
    Namespace:    opts.Namespace,
  }

  label_sel := &meta_v1.LabelSelector{
    MatchLabels: prep_labels(app_name),
  }
 
  //todo:  set resource limit
  replicas := default_num_replicas

  return &apps_v1.Deployment {
    TypeMeta: type_meta,
    ObjectMeta: obj_meta,
    Spec: apps_v1.DeploymentSpec{
      Replicas: &replicas,
      Selector: label_sel,
      Template: v1.PodTemplateSpec{
        ObjectMeta: obj_meta, 
        Spec: podConfig(opts),
      },
    },
  }
  
} 

func serviceConfig(opts *DeployOptions) *v1.Service{
  
  obj_meta := prep_objmeta(app_name, app_name, opts.Namespace, nil)

  return &v1.Service{
    TypeMeta: meta_v1.TypeMeta{
      Kind:       "Service",
      APIVersion: "v1",
    },
    ObjectMeta: obj_meta,
    Spec: v1.ServiceSpec{
      Type: v1.ServiceTypeNodePort,
      Selector: map[string]string{
        "app": app_name,
      },
      Ports: []v1.ServicePort{
        {
          Port: opts.ServerPort,
          Name: "api-service-port",
          NodePort: opts.ServerNodePort,
        }, 
      },
    },
  }
}

func secretVars() v1.EnvVar{
  t := true
  return v1.EnvVar{
    Name: hfConfig.ConfigVarsEnvName,  // HFLOW_CONFIG
    ValueFrom: &v1.EnvVarSource{
      SecretKeyRef: &v1.SecretKeySelector{
        LocalObjectReference: v1.LocalObjectReference{
          Name: hflow_secret_name,
        },
        Key:      "config",
        Optional: &t,
      },
    },
  }
}

func genAPIServerConfig(opts *DeployOptions) *hfConfig.Config {
  port_string := fmt.Sprint(opts.ServerPort)
  
  c, err := hfConfig.NewConfig(opts.ServerIP, port_string, "")
  if err != nil {
    panic(err)
  }

  c.DB = &hfConfig.DBConfig{
      Driver: opts.DbType,
      Name: opts.DbBucket,
      DataDirPath: opts.DbFileMountPath,
    }

  return c
}

func configSecret(ns string, c *hfConfig.Config) *v1.Secret{

  conf_data, _ := json.Marshal(c) 
  return &v1.Secret{
    TypeMeta: meta_v1.TypeMeta{
      Kind:       "Secret",
      APIVersion: "v1",
    },
    ObjectMeta: prep_objmeta(hflow_secret_name, hflow_secret_name, ns, nil),
    Data: map[string][]byte{
      "config": conf_data,
    },
  }
}

// database can be stored on local disk or cloud disk 
// target for DB can be different than object storage target   
func DatabasePV(namespace string, name string, target StorageTarget, cloudDiskName string, NodePath string, volumeSize string) (v1.PersistentVolume, error) {
 
 pv :=  v1.PersistentVolume{
  TypeMeta: meta_v1.TypeMeta{
    Kind:       "PersistentVolume",
    APIVersion: "v1",
  },
  ObjectMeta: prep_objmeta(name, app_name, namespace, nil),
  Spec: v1.PersistentVolumeSpec{
    Capacity: map[v1.ResourceName]resource.Quantity{
      "storage":  resource.MustParse(volumeSize),
    },
    AccessModes:  []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
    PersistentVolumeReclaimPolicy: v1.PersistentVolumeReclaimRetain,
    StorageClassName: "standard",
  }, 
 } 

 switch target {
 case Local:
    pv.Spec.PersistentVolumeSource = v1.PersistentVolumeSource{
      HostPath: &v1.HostPathVolumeSource{
        Path: NodePath,
      },
    }
  case Google:
    pv.Spec.PersistentVolumeSource = v1.PersistentVolumeSource{
      GCEPersistentDisk: &v1.GCEPersistentDiskVolumeSource{
        PDName: cloudDiskName,
        //todo: FSType
      },
    }
 }  

 return pv, nil
}

func DatabasePVClaim(namespace string, volumeName string, claimName string, claimSize string) v1.PersistentVolumeClaim {
  return v1.PersistentVolumeClaim{
      TypeMeta: meta_v1.TypeMeta{
      Kind:       "PersistentVolumeClaim",
      APIVersion: "v1",
    },
    ObjectMeta: prep_objmeta(claimName, app_name, namespace, nil),
    Spec: v1.PersistentVolumeClaimSpec{
      Resources:  v1.ResourceRequirements{
        Requests: map[v1.ResourceName]resource.Quantity{
          "storage": resource.MustParse(claimSize),
        },
      },
      AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
      VolumeName:  volumeName,
    },
  }
}

func DatabaseConfig(manifest Configurator, opts *DeployOptions, dbStorage StorageTarget, cloudDiskName string, hostpath string, allocSize string) (Configurator, error){
  var err error 
  volume_name := default_db_volume_name
  volume_claim_name := default_db_volume_claim

  pv, err := DatabasePV(opts.Namespace, volume_name, dbStorage, cloudDiskName, hostpath, allocSize)
  if err != nil {
    return manifest, err
  }

  if err = manifest.AddEntry(pv); err != nil {
    return manifest, err
  }   

  if err = manifest.AddEntry(DatabasePVClaim(opts.Namespace, volume_name, volume_claim_name, allocSize)); err != nil {
    return manifest, err
  }

  return manifest, nil
}

func BaseConfig(manifest Configurator, opts *DeployOptions) (Configurator, error) {
  if manifest == nil {
    manifest = NewConfigurator()
  }

  if opts == nil {
    opts = DefaultDeployOpts()
  }

  if err:= manifest.AddEntry(namespaceConfig(opts.Namespace)); err != nil {
    base.Error("failed to write namespace onfig, err: ", err)
    return nil, err
  }

  if err := manifest.AddEntry(serviceAccountConfig(opts)); err != nil{
    return nil, err
  }

  if !opts.NoRBAC{
    role, role_binding := roleConfig(opts)

    if err := manifest.AddEntry(role); err != nil {
      base.Error("failed to write role config, err: ", err)
    }
    if err := manifest.AddEntry(role_binding); err != nil {
      base.Error("failed to write role binding config, err: ", err)  
    }
  }

  if err := manifest.AddEntry(serviceConfig(opts)); err != nil {
    base.Error("failed to write service config, err: ", err)
  }

  if err := manifest.AddEntry(deploymentConfig(opts)); err != nil {
    base.Error("failed to write deploy config, err: ", err)    
  }
 
  return manifest, nil
}

// req params: 
// creds - google credentials file content
// objectStorageBucket - s3 or GCS bucket name
// -- DbHostPath when dbStorage: LOCAL
// -- DbCloudDiskName when dbStorage: Google, AWS, etc

func GoogleConfig(manifest Configurator, opts *DeployOptions, creds string, objectStorageBucket string, dbStorage StorageTarget, dbCloudDiskName string, dbHostPath string, dbMountPath string, dbDiskSize string) (Configurator, error) {
  
  if opts == nil {
    opts = DefaultDeployOpts()
  }

  if dbMountPath == "" {
    opts.DbFileMountPath = dbMountPath
  }

  if (dbStorage == None || dbStorage == "") {
    opts.DbUsePVClaim = false
  }

  cfg := genAPIServerConfig(opts)
  
  if opts.ServerPort != opts.ServerNodePort {
    cfg.MasterExternalPort = opts.ServerNodePort
  }

  if cfg.ObjStorage == nil {
    cfg.ObjStorage = &hfConfig.ObjStorageConfig{}
  }

  cfg.ObjStorage.StorageTarget = hfConfig.GCS

  cfg.ObjStorage.Gcs = &hfConfig.GcsConfig{
    Bucket: objectStorageBucket,
    Creds: []byte(creds), 
  }
  
  if cfg.K8 == nil {
    cfg.K8 = &hfConfig.KubeConfig{}
  }

  cfg.K8 = &hfConfig.KubeConfig{
    InCluster: true,
    Namespace: product_family,
  }

  opts.APIServerConfig = cfg

  manifest, err:= BaseConfig(manifest, opts)
  if err != nil {
    return nil, err
  }
  
  if manifest.AddEntry(configSecret(opts.Namespace, cfg)); err != nil {
    return nil, err
  }
   
  if (opts.DbUsePVClaim) { 
    if manifest, err := DatabaseConfig(manifest, opts, dbStorage, dbCloudDiskName, dbHostPath, dbDiskSize); err != nil {
      return manifest, err
    }
  }

  return manifest, nil
}




