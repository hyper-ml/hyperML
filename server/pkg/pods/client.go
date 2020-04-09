package pods

import (
	"fmt"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	"github.com/hyper-ml/hyperml/server/pkg/config"
	osUtils "github.com/hyper-ml/hyperml/server/pkg/utils/osutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cube "k8s.io/client-go/kubernetes"
	crest "k8s.io/client-go/rest"
	"strings"

	cubecmd "k8s.io/client-go/tools/clientcmd"
)

func getClient(c *config.KubeConfig) (*cube.Clientset, error) {
	var clusterConf *crest.Config
	var err error
	if c.InCluster {
		base.Info("Using K8s Incluster Config")
		clusterConf, err = crest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to read config incluster: %v", err)
		}
		return cube.NewForConfig(clusterConf)
	}

	if c.Path != "" {
		if osUtils.PathExists(c.Path) {

			clusterConf, err = cubecmd.BuildConfigFromFlags("", c.Path)
			if err != nil {
				return nil, fmt.Errorf("Failed to create client given path(%v) : %v", c.Path, err)
			}
			base.Info("Using K8s Config from :", c.Path)
		} else {
			return nil, fmt.Errorf("Failed to load kuberenetes config from path %v: %v", c.Path, err)
		}
	} else {
		envPath, err := osUtils.K8ConfigValidPath()
		if err != nil {
			return nil, fmt.Errorf("Failed to read env for k8s config path: %v", err)
		}

		clusterConf, err = cubecmd.BuildConfigFromFlags("", envPath)
		if err != nil {
			return nil, fmt.Errorf("Failed to generate client from envpath: %v", err)
		}

		base.Info("Using K8s Config from :", envPath)
	}
	return cube.NewForConfig(clusterConf)
}

func checkClient(client *cube.Clientset) (fnerr error) {
	fmt.Println("Checking Kubernetes Connection....")
	defer func() {

		if r := recover(); r != nil {
			fnerr = fmt.Errorf("Failed to connect to kubernetes: %v", fnerr)
		}

	}()

	ns, err := client.CoreV1().Namespaces().Get(DefaultNamespace, metav1.GetOptions{})
	fmt.Println("namespace:", ns.ObjectMeta.Name)

	if err != nil {
		errstr := err.Error()
		if !(strings.HasPrefix(errstr, "namespaces") &&
			strings.HasSuffix(errstr, "not found")) {
			return fmt.Errorf("Unable to connect to k8s cluster: %v", err.Error())
		}
	}

	if ns == nil || ns.ObjectMeta.Name == "" {
		fmt.Println("namespace hyperml not found. Creating new ... ")
		ns, err = client.CoreV1().Namespaces().Create(makeNamespace(DefaultNamespace))
		if err != nil {
			return fmt.Errorf("Failed to create namespace in k8s hyperml: %v", err)
		}
	} else {
		fmt.Println("choosing existing namespace hyperml")
	}
	return nil
}
