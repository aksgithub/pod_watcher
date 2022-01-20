package k8s

import (
	"context"
	"go.uber.org/zap"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

type ClientInterface interface {
	GetPods(namespace string, options metav1.ListOptions) (*apiv1.PodList, error)
	DeletePod(pod *apiv1.Pod) error
	NewSharedInformerFactory(ns string) (informers.SharedInformerFactory, error)
}

type Client struct {
	logger    *zap.Logger
	clientSet *kubernetes.Clientset
}

func (c *Client) GetPods(namespace string, options metav1.ListOptions) (*apiv1.PodList, error) {
	ctx := context.Background()
	return c.clientSet.CoreV1().Pods(namespace).List(ctx, options)
}

func (c *Client) DeletePod(pod *apiv1.Pod) error {
	ctx := context.Background()
	return c.clientSet.CoreV1().Pods(pod.ObjectMeta.Namespace).Delete(ctx, pod.ObjectMeta.Name, metav1.DeleteOptions{})
}

func (c *Client) NewSharedInformerFactory(ns string) (informers.SharedInformerFactory, error) {
	factory := informers.NewSharedInformerFactoryWithOptions(c.clientSet, 0, informers.WithNamespace(ns))
	return factory, nil
}

func newClientSet() (*kubernetes.Clientset, error) {
	var err error
	var config *restclient.Config
	if os.Getenv("KUBERNETES_SERVICE_HOST") == "" {
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		// Reads config when in cluster
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func NewClient(logger *zap.Logger) (*Client, error) {
	clientSet, err := newClientSet()
	if err != nil {
		return nil, err
	}

	return &Client{clientSet: clientSet, logger: logger}, err
}
