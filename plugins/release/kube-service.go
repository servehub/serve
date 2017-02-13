package release

import (
	"fmt"
	"log"
	"os"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/servehub/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("release.kube-service", ReleaseKubeService{})
}

type ReleaseKubeService struct{}

func (p ReleaseKubeService) Run(data manifest.Manifest) error {
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("HOME")+"/.kube/config")
	if err != nil {
		return err
	}

	kube, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	return runService(kube, data)
}

func runService(kube *kubernetes.Clientset, data manifest.Manifest) error {
	appName := data.GetString("name")

	ports := make([]v1.ServicePort, 0)
	for _, port := range data.GetArray("ports") {
		ports = append(ports, v1.ServicePort{
			Port:     int32(port.GetInt("port")),
			Protocol: v1.ProtocolTCP,
			Name:     port.GetStringOr("name", ""),
		})
	}

	serviceSpec := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: appName,
		},
		Spec: v1.ServiceSpec{
			Ports:    ports,
			Selector: map[string]string{"app": appName},
		},
	}

	services := kube.Services(data.GetString("namespace"))

	if serv, err := services.Get(appName, metav1.GetOptions{}); err == nil {
		log.Printf("Service `%s` already exists!\n", serv.Name)
		return nil
	}

	_, err := services.Update(serviceSpec)

	switch {
	case err == nil:
		log.Println("Service updated")

	case !errors.IsNotFound(err):
		return fmt.Errorf("Could not update service: %s", err)

	default:
		_, err = services.Create(serviceSpec)
		if err != nil {
			return fmt.Errorf("Could not create service: %s", err)
		}
		log.Println("Service created")
	}

	return nil
}
