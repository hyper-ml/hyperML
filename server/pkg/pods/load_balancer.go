package pods

import (
	"encoding/json"
	"fmt"
	meta "github.com/hyper-ml/hyperml/server/pkg/types"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kns "k8s.io/client-go/kubernetes"

	extbetav1 "k8s.io/api/extensions/v1beta1"
)

// LoadBalancer :
type LoadBalancer struct {
	ingress   *extbetav1.Ingress
	namespace string
	client    *kns.Clientset
	domain    string
}

func newLoadBalancer(ns string, client *kns.Clientset, domain string) (*LoadBalancer, error) {
	lb := &LoadBalancer{
		client:    client,
		namespace: ns,
		domain:    domain,
	}

	return lb, nil
}

// Init :
func (lb *LoadBalancer) Init() error {

	// look for existing ingress. ?
	result, err := lb.client.ExtensionsV1beta1().Ingresses(lb.namespace).Get(DefaultIngressName, metav1.GetOptions{})

	if err != nil {
		errstr := err.Error()
		if !(strings.HasPrefix(errstr, "ingress") &&
			strings.HasSuffix(errstr, "not found")) {
			return err
		}
	}

	if result == nil || result.ObjectMeta.Name == "" {

		// generate ingress config
		spec, err := makeIngressSpec(lb.domain)
		if err != nil {
			return err
		}
		jsonSpec, _ := json.Marshal(&spec)
		fmt.Println("Ingress Spec:", string(jsonSpec))
		// launch with keeper
		result, err = lb.client.ExtensionsV1beta1().Ingresses(lb.namespace).Create(&spec)
		if err != nil {
			return fmt.Errorf("Failed to Initialize Ingress Controller: " + err.Error())
		}
	} else {
		fmt.Println("choosing existing ingress : ", DefaultIngressName)
	}

	lb.ingress = result
	return nil
}

// AddPod :
func (lb *LoadBalancer) AddPod(pod *meta.POD) error {
	ingress, err := lb.client.ExtensionsV1beta1().Ingresses(lb.namespace).Get(DefaultIngressName, metav1.GetOptions{})

	if err != nil {
		errstr := err.Error()
		if strings.HasPrefix(errstr, "ingress") &&
			strings.HasSuffix(errstr, "not found") {
			_ = lb.Init()
		}
		return err
	}

	ingress, err = addPODtoIngressSpec(pod, ingress, lb.domain)
	if err != nil {
		return err
	}

	ingress, err = lb.client.ExtensionsV1beta1().Ingresses(lb.namespace).Update(ingress)
	fmt.Println("err:", err)
	if err != nil {
		return err
	}
	lb.ingress = ingress
	return nil

}

// RemovePod :
func (lb *LoadBalancer) RemovePod(pod *meta.POD) error {
	ingress, err := lb.client.ExtensionsV1beta1().Ingresses(lb.namespace).Get(DefaultIngressName, metav1.GetOptions{})

	if err != nil {
		return err
	}
	ingress, err = removePODfromIngressSpec(pod, ingress, lb.domain)
	if err != nil {
		return err
	}
	ingress, err = lb.client.ExtensionsV1beta1().Ingresses(lb.namespace).Update(ingress)

	lb.ingress = ingress
	return nil
}
