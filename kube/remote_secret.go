package kube

import (
	"context"
	"encoding/base64"
	"flag"
	"log"
	"path/filepath"

	"github.com/frauniki/istio-eks-sonar/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
)

func getK8sConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	return clientcmd.BuildConfigFromFlags("", *kubeconfig)
}

func CreateRemoteSecrets(ctx context.Context, config *config.GenerateConfig, remoteKubeconfigs map[string]*api.Config) error {
	cfg, err := getK8sConfig()
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}

	for _, namespace := range config.TargetNamespaces {
		for clusterName, kubeconfig := range remoteKubeconfigs {
			kubeconfigBytes, err := clientcmd.Write(*kubeconfig)
			if err != nil {
				return err
			}
			b64EncodedKubeconfig := make([]byte, base64.StdEncoding.EncodedLen(len(kubeconfigBytes)))
			base64.StdEncoding.Encode(b64EncodedKubeconfig, kubeconfigBytes)

			resp, err := clientset.CoreV1().Secrets(namespace).Create(ctx, &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: config.SecretNamePrefix + clusterName,
				},
				Type: v1.SecretTypeOpaque,
				Data: map[string][]byte{
					clusterName: b64EncodedKubeconfig,
				},
			}, metav1.CreateOptions{})
			if err != nil {
				return err
			}

			log.Printf("Created remote secret %s in namespace %s", resp.Name, resp.Namespace)
		}
	}

	return nil
}
