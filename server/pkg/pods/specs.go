package pods

import (
	"fmt"
	defaults "github.com/hyper-ml/hyperml/server/pkg/config"
	meta "github.com/hyper-ml/hyperml/server/pkg/types"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	extbetav1 "k8s.io/api/extensions/v1beta1"
	kuberes "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
)

func validatePOD(pod *meta.POD) error {
	if pod == nil {
		return fmt.Errorf("Invalid POD request: empty")
	}

	if pod.ID == 0 {
		return fmt.Errorf("Invalid POD Request")
	}

	if pod.Config == nil {
		return fmt.Errorf("POD config missing in request")
	}

	if pod.Config.Image == nil {
		return fmt.Errorf("Container Image not provided")
	}

	if pod.Config.Command == "" {
		return fmt.Errorf("Container Command not provided")
	}

	if pod.Config.RestartPolicy == "" {
		return fmt.Errorf("Container Restart policy is missing. Can be set to Always")
	}
	return nil
}

func makeNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: corev1.NamespaceSpec{},
	}
}

func makePODTemplateSpec(pod *meta.POD) (corev1.PodTemplateSpec, error) {
	fmt.Println("pod RequestMode:", pod.RequestMode)
	if err := validatePOD(pod); err != nil {
		return corev1.PodTemplateSpec{}, err
	}

	env := []corev1.EnvVar{
		{
			Name:  "PYTHONUNBUFFERED",
			Value: "1",
		}, {
			Name: "WORKER_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.podIP",
				},
			},
		}, {
			Name: "WORKER_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		}, {
			Name: "HOST_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.hostIP",
				},
			},
		},
	}

	// Env Vars passed by user
	if pod.Config.EnvVars != nil {
		for _, v := range pod.Config.EnvVars {
			env = append(env, corev1.EnvVar{
				Name:  v.Name,
				Value: v.Value,
			})
		}
	}

	zeroVal := int64(0)
	cmd := strings.Split(pod.Config.Command, " ")

	resourceSpec := makeResourceSpec(pod.Config.Resources)

	spec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            makeContainerName(pod.PodType, pod.ID, 1),
				Image:           pod.Config.Image.Name,
				ImagePullPolicy: corev1.PullPolicy(pod.Config.ImagePullPolicy),
				Command:         cmd,
				Env:             env,
				Resources:       resourceSpec,
			},
		},
		RestartPolicy:                 corev1.RestartPolicy(pod.Config.RestartPolicy),
		TerminationGracePeriodSeconds: &zeroVal,
	}

	labels := map[string]string{}
	userPodID := fmt.Sprintf("%d", pod.ID)
	labels["userPodID"] = userPodID
	labels["userPodType"] = pod.PodType

	if pod.RequestMode > 0 {
		labels["requestMode"] = fmt.Sprintf("%d", pod.RequestMode)
	}

	fmt.Println("labels:", labels)

	ptSpec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:   makeContainerName(pod.PodType, pod.ID, 1),
			Labels: labels,
		},
		Spec: spec,
	}

	return ptSpec, nil
}

func makeResourceSpec(request *meta.ResourceProfile) corev1.ResourceRequirements {
	memory := request.RAM
	cpu := request.CPU

	if memory == nullString {
		memory = defaults.GetPodMemoryLimit()
	}

	if cpu == nullString {
		cpu = defaults.GetPodCPULimit()
	}

	if cpu == nullString && memory == nullString {
		return corev1.ResourceRequirements{}
	}

	var requests corev1.ResourceList = make(corev1.ResourceList)
	var limits corev1.ResourceList = make(corev1.ResourceList)

	requests[corev1.ResourceCPU] = kuberes.MustParse(cpu)
	requests[corev1.ResourceMemory] = kuberes.MustParse(memory)

	limits[corev1.ResourceCPU] = kuberes.MustParse(cpu)
	limits[corev1.ResourceMemory] = kuberes.MustParse(memory)

	return corev1.ResourceRequirements{
		Requests: requests,
		Limits:   limits,
	}
}

func makeDeploySpec(pod *meta.POD, ptSpec corev1.PodTemplateSpec) (*appsv1.Deployment, error) {
	labels := map[string]string{}
	annotations := map[string]string{}
	userPodID := fmt.Sprintf("%d", pod.ID)

	labels["userPodID"] = userPodID
	labels["userPodType"] = pod.PodType
	if pod.RequestMode > 0 {
		labels["requestMode"] = fmt.Sprintf("%d", pod.RequestMode)
	}
	var replicas int32 = 1

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        userPodID,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"userPodID": userPodID,
				},
			},
			Replicas: &replicas,
			Template: ptSpec,
		},
	}, nil
}

func makeContainerName(podType string, podID uint64, index int) string {
	return strings.ToLower(podType + "-" + fmt.Sprintf("%d-%d", podID, index))
}

func makeServiceName(podType string, podID uint64) string {
	return strings.ToLower(podType + "-" + fmt.Sprintf("%d", podID))
}

func makePortName(podType string) string {
	return strings.ToLower(podType + "-port")
}

func makeServiceSpec(pod *meta.POD) (corev1.Service, error) {
	srvName := makeServiceName(pod.PodType, pod.ID)

	labels := map[string]string{}
	labels["userPodID"] = fmt.Sprintf("%d", pod.ID)
	labels["userPodType"] = pod.PodType
	if pod.RequestMode > 0 {
		labels["requestMode"] = fmt.Sprintf("%d", pod.RequestMode)
	}

	srvPortName := makePortName(pod.PodType)

	return corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   srvName,
			Labels: labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Port: pod.ServicePort,
					Name: srvPortName,
				},
			},
		},
	}, nil
}

func makeIngressSpec(domain string) (extbetav1.Ingress, error) {

	labels := map[string]string{}
	labels["namespace"] = "hyperml"
	labels["domain"] = domain

	defaultRule := makeIngressRule(domain, "default", "/", 80)
	iRules := []extbetav1.IngressRule{defaultRule}

	return extbetav1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   DefaultIngressName,
			Labels: labels,
		},
		Spec: extbetav1.IngressSpec{
			Rules: iRules,
		},
	}, nil
}

func makeIngressRule(domain, srvName, path string, srvPort int) extbetav1.IngressRule {
	ir := extbetav1.IngressRule{
		Host: domain,
	}

	ir.IngressRuleValue = extbetav1.IngressRuleValue{
		HTTP: &extbetav1.HTTPIngressRuleValue{
			Paths: []extbetav1.HTTPIngressPath{
				{
					Path: path,
					Backend: extbetav1.IngressBackend{
						ServiceName: srvName,
						ServicePort: intstr.FromInt(srvPort),
					},
				},
			},
		},
	}

	return ir
}

func makeNotebookPath(userKey string) string {
	return "/" + userKey
}

func addPODtoIngressSpec(pod *meta.POD, ingress *extbetav1.Ingress, domain string) (*extbetav1.Ingress, error) {

	if err := validatePOD(pod); err != nil {
		return nil, err
	}

	if len(ingress.Spec.Rules) == 0 {
		newRule := makeIngressRule(domain, makeServiceName(pod.PodType, pod.ID), makeNotebookPath(pod.UserKey), int(pod.ServicePort))
		ingress.Spec.Rules = []extbetav1.IngressRule{newRule}
		return ingress, nil
	}

	// TODO: support multiple domains
	for i, r := range ingress.Spec.Rules {

		// assume domain was added at the init
		if r.Host == domain {
			path := extbetav1.HTTPIngressPath{
				Path: makeNotebookPath(pod.UserKey),
				Backend: extbetav1.IngressBackend{
					ServiceName: makeServiceName(pod.PodType, pod.ID),
					ServicePort: intstr.FromInt(int(pod.ServicePort)),
				},
			}

			ingress.Spec.Rules[i].HTTP.Paths = append(ingress.Spec.Rules[i].HTTP.Paths, path)
		}
	}

	return ingress, nil
}

func removePODfromIngressSpec(pod *meta.POD, ingress *extbetav1.Ingress, domain string) (*extbetav1.Ingress, error) {

	if err := validatePOD(pod); err != nil {
		return nil, err
	}

	if len(ingress.Spec.Rules) == 0 {
		return ingress, nil
	}

	for ri, r := range ingress.Spec.Rules {
		if r.Host == domain {
			paths := ingress.Spec.Rules[ri].HTTP.Paths

			changed := false
			for i, p := range paths {
				if p.Path == makeNotebookPath(pod.UserKey) {
					paths[i] = paths[len(paths)-1]
					paths[len(paths)-1] = extbetav1.HTTPIngressPath{}
					paths = paths[:len(paths)-1]
					changed = true
				}
			}

			if changed {
				ingress.Spec.Rules[ri].HTTP.Paths = paths
			}
		}
	}

	return ingress, nil
}

func makeDeletePolicySpec() *metav1.DeleteOptions {
	propogationPolicy := metav1.DeletePropagationForeground

	return &metav1.DeleteOptions{
		///OrphanDependents: &false_value,
		PropagationPolicy: &propogationPolicy,
	}
}

func makeJobSpec(pod *meta.POD, jobConfig *defaults.JobConfig) (*batchv1.Job, error) {

	ptSpec, err := makePODTemplateSpec(pod)
	if err != nil {
		return nil, fmt.Errorf("Failed to make pod template: %v", err)
	}

	intval := int32(1)
	deadline := int64(jobConfig.DeadlineSeconds)
	backoff := int32(jobConfig.BackoffLimit)
	//ttl := int32(jobDefaults.TTL)

	annotations := map[string]string{}
	podIDstr := fmt.Sprintf("%d", pod.ID)

	labels := map[string]string{}
	labels["userPodID"] = podIDstr
	labels["userPodType"] = pod.PodType
	if pod.RequestMode > 0 {
		labels["requestMode"] = fmt.Sprintf("%d", pod.RequestMode)
	}
	tru := true
	spec := batchv1.JobSpec{
		Parallelism:           &intval,
		Completions:           &intval,
		ActiveDeadlineSeconds: &deadline,
		BackoffLimit:          &backoff,
		//TTLSecondsAfterFinished: &ttl,
		ManualSelector: &tru,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"userPodID": podIDstr,
			},
		},
		Template: ptSpec,
	}
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        podIDstr,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: spec,
	}, nil
}
