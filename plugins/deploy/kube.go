package deploy

import (
	"fmt"
	"log"
	"os"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/servehub/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("deploy.kube", DeployKube{})
}

type DeployKube struct{}

func (p DeployKube) Run(data manifest.Manifest) error {
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("HOME")+"/.kube/config")
	if err != nil {
		return err
	}

	kube, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	return runDeployment(kube, data)
}

func runDeployment(kube *kubernetes.Clientset, data manifest.Manifest) error {
	appName := data.GetString("name")

	ports := make([]v1.ContainerPort, 0)
	for _, port := range data.GetArray("ports") {
		ports = append(ports, v1.ContainerPort{
			ContainerPort: int32(port.GetInt("containerPort")),
			Protocol:      v1.ProtocolTCP,
			Name:          port.GetStringOr("name", ""),
		})
	}

	envs := make([]v1.EnvVar, 0)
	for k, v := range data.GetMap("environment") {
		envs = append(envs, v1.EnvVar{Name: k, Value: fmt.Sprintf("%s", v.Unwrap())})
	}

	resources := v1.ResourceList{}
	if data.GetStringOr("cpu", "") != "" {
		resources[v1.ResourceCPU] = resource.MustParse(data.GetString("cpu"))
	}

	if data.GetStringOr("mem", "") != "" {
		resources[v1.ResourceMemory] = resource.MustParse(data.GetString("mem"))
	}

	containers := []v1.Container{
		{
			Name:            appName,
			Image:           data.GetString("image"),
			Ports:           ports,
			Resources:       v1.ResourceRequirements{Limits: resources, Requests: resources},
			ImagePullPolicy: v1.PullIfNotPresent,
			Env:             envs,
		},
	}

	deploySpec := &v1beta1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: appName,
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: int32p(int32(data.GetInt("replicas"))),
			Strategy: v1beta1.DeploymentStrategy{
				Type: v1beta1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &v1beta1.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: int32(0),
					},
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: int32(1),
					},
				},
			},
			RevisionHistoryLimit: int32p(3),
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   appName,
					Labels: map[string]string{"app": appName},
				},
				Spec: v1.PodSpec{
					Containers:    containers,
					RestartPolicy: v1.RestartPolicyAlways,
					DNSPolicy:     v1.DNSClusterFirst,
				},
			},
		},
	}

	deploy := kube.Extensions().Deployments(data.GetString("namespace"))
	_, err := deploy.Update(deploySpec)

	switch {
	case err == nil:
		log.Println("Deployment controller updated")

	case !errors.IsNotFound(err):
		return fmt.Errorf("Could not update deployment controller: %s", err)

	default:
		_, err = deploy.Create(deploySpec)
		if err != nil {
			return fmt.Errorf("Could not create deployment controller: %s", err)
		}
		log.Println("Deployment controller created")
	}

	return nil
}

func int32p(i int32) *int32 {
	return &i
}
