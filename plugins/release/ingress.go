package release

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

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
			Kind:       "Ingress",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: appName,
			Labels: map[string]string{
				"service": data.GetString("service"),
				"stage":   data.GetString("stage"),
				"version": data.GetString("version"),
			},
		},
		Spec: v1beta1.IngressSpec{
			Rules: rules,
		},
	}

	ingresses := kube.Extensions().Ingresses(data.GetString("namespace"))

	if exists, err := ingresses.List(metav1.ListOptions{
		LabelSelector: "service=" + data.GetString("service") + ",stage=" + data.GetString("stage") + ",version!=" + data.GetString("version"),
	}); err != nil {
		return err
	} else if len(exists.Items) > 0 {
		items, _ := json.MarshalIndent(exists.Items, "", "  ")
		log.Println("Exists ingress: ", string(items))
		log.Println("Mark as outdated!")

		for _, ing := range exists.Items {
			ing.Labels["stage"] = "outdated"

			if ing.Annotations == nil {
				ing.Annotations = make(map[string]string, 0)
			}
			ing.Annotations["endOfLife"] = time.Now().Add(15 * time.Minute).String()

			if _, err := ingresses.Update(&ing); err != nil {
				return err
			}
		}
	}

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
