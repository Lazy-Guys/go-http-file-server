package utils

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type K8sClient struct {
	Client *kubernetes.Clientset
	conf   *rest.Config
}

func NewK8sClient(isInCluster bool) (*K8sClient, error) {
	var config *rest.Config
	var confErr error
	if isInCluster {
		config, confErr = rest.InClusterConfig()
		if confErr != nil {
			panic(confErr.Error())
		}
	} else {
		config, confErr = OutClusterConf()
		if confErr != nil {
			panic(confErr.Error())
		}
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &K8sClient{
		Client: clientset,
		conf:   config,
	}, nil
}

func (cl *K8sClient) GetPodList(nameSpace string) (*v1.PodList, error) {
	return cl.Client.CoreV1().Pods(nameSpace).List(context.Background(), metav1.ListOptions{})
}

func (cl *K8sClient) GetPod(nameSpace string, podName string) (pod *v1.Pod, err error) {
	pod, err = cl.Client.CoreV1().Pods(nameSpace).Get(context.Background(), podName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		fmt.Printf("Pod %s in namespace %s not found\n", pod, nameSpace)
		return nil, err
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		fmt.Printf("Error getting pod %s in namespace %s: %v\n",
			pod, nameSpace, statusError.ErrStatus.Message)
		return nil, err
	} else if err != nil {
		panic(err.Error())
	} else {
		return pod, nil
	}
}

func (cl *K8sClient) DeletePod(nameSpace string, podName string) (err error) {
	deletePolicy := metav1.DeletePropagationForeground
	err = cl.Client.AppsV1().Deployments(nameSpace).Delete(context.Background(), podName, metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
	return err
}

func (cl *K8sClient) GetNodeList() (*v1.NodeList, error) {
	return cl.Client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
}

func GetNodeStatus(node v1.Node) string {
	for _, x := range node.Status.Conditions {
		if x.Status == "True" {
			return string(x.Type)
		}
	}
	return "NotReady"
}
