package release

import (
	"fmt"
	"log"
	"os"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/servehub/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("release.ingress", ReleaseIngress{})
}

type ReleaseIngress struct{}

func (p ReleaseIngress) Run(data manifest.Manifest) error {
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("HOME")+"/.kube/config")
	if err != nil {
		return err
	}

	kube, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	if err := runService(kube, data); err != nil {
		return err
	}

	return runIngress(kube, data)
}

func runIngress(kube *kubernetes.Clientset, data manifest.Manifest) error {
	appName := data.GetString("name")

	rules := make([]v1beta1.IngressRule, 0)
	for _, route := range data.GetArray("routes") {
		rules = append(rules, v1beta1.IngressRule{
			Host: route.GetString("host"),
			IngressRuleValue: v1beta1.IngressRuleValue{
				HTTP: &v1beta1.HTTPIngressRuleValue{
					Paths: []v1beta1.HTTPIngressPath{
						{
							Path: route.GetStringOr("location", "/"),
							Backend: v1beta1.IngressBackend{
								ServiceName: appName,
								ServicePort: intstr.FromInt(route.GetIntOr("port", 80)),
							},
						},
					},
				},
			},
		})
	}

	ingressSpec := &v1beta1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind: "Ingress",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: appName,
		},
		Spec: v1beta1.IngressSpec{
			Rules: rules,
		},
	}

	ingresses := kube.Extensions().Ingresses(data.GetString("namespace"))
	_, err := ingresses.Update(ingressSpec)

	switch {
	case err == nil:
		log.Println("Ingress updated")

	case !errors.IsNotFound(err):
		return fmt.Errorf("Could not update ingress: %s", err)

	default:
		_, err = ingresses.Create(ingressSpec)
		if err != nil {
			return fmt.Errorf("Could not create ingress: %s", err)
		}
		log.Println("Ingress created")
	}

	return nil
}
