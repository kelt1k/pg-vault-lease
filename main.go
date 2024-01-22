package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type SGCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// ... other fields according to SGCluster CRD
}

func main() {

	var config *rest.Config
	var err error
	var namespace string

	envType := os.Getenv("ENV_TYPE")
	targetNamespace := os.Getenv("TARGET_NAMESPACE")
	if envType == "local" {
		// Use the kubeconfig file from the user's home directory
		var kubeconfig string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		} else {
			kubeconfig = os.Getenv("KUBECONFIG")
		}

		// Use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}
	} else {
		// Use in-cluster configuration
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	if envType == "local" {
		namespace = targetNamespace
	} else {
		// Check Namespace where App's pod is deployed
		discovered_namespace, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
		namespace = string(discovered_namespace)
		if err != nil {
			panic(err.Error())
		}
	}
	// Define the API Group and Version for SGCluster
	sgClusterAPIGroup := "stackgres.io"
	sgClusterAPIVersion := "v1"

	// Get SGClusters in Current Namespace

	sgClusters, err := clientset.RESTClient().
		Get().
		AbsPath("/apis", sgClusterAPIGroup, sgClusterAPIVersion, "sgclusters").
		Namespace(namespace).
		Do(context.TODO()).
		Get()

	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("SGClusters found: %+v, in %v namespace\n", sgClusters, namespace)

	secrets, err := clientset.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Found secrets: %v\n", secrets)

}
