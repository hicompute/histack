package k8s

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewClient() (client.Client, error) {
	cfg, err := createConfig()
	if err != nil {
		return nil, err
	}

	return client.New(cfg, client.Options{
		Scheme: Scheme,
	})
}

func createConfig() (*rest.Config, error) {
	configFile := filepath.Join(homedir.HomeDir(), ".kube", "config")
	_, err := os.Stat(configFile)
	if err != nil {
		return rest.InClusterConfig()
	}
	return clientcmd.BuildConfigFromFlags("", configFile)
}
