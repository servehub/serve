package deploy

import (
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/tools/clientcmd"

	"fmt"
	"github.com/servehub/serve/manifest"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"log"
	"k8s.io/apimachinery/pkg/api/errors"
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
	allPorts := make([]v1.ContainerPort, 0)

	containers := make([]v1.Container, 0)
	for _, cont := range data.GetArray("containers") {

		ports := make([]v1.ContainerPort, 0)
		for _, port := range cont.GetArray("ports") {
			ports = append(ports, v1.ContainerPort{
				ContainerPort: port.GetString("containerPort"),
				Protocol:      v1.ProtocolTCP,
				Name:          port.GetStringOr("name", nil),
			})
		}

		allPorts = append(allPorts, ports...)

		envs := make([]v1.EnvVar, 0)
		for k, v := range data.GetMap("environment") {
			envs = append(envs, v1.EnvVar{Name: k, Value: fmt.Sprintf("%s", v.Unwrap())})
		}

		containers = append(containers, v1.Container{
			Name:  appName,
			Image: cont.GetString("image"),
			Ports: ports,
			Resources: v1.ResourceRequirements{
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(cont.GetString("cpu")),
					v1.ResourceMemory: resource.MustParse(cont.GetString("mem")),
				},
			},
			ImagePullPolicy: v1.PullAlways,
			Env:             envs,
		})
	}

	deploySpec := &v1beta1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: appName,
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: int32p(data.GetInt("replicas")),
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
				ObjectMeta: v1.ObjectMeta{
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
