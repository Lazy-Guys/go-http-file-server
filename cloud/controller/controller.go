package controller

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/Lazy-Guys/utils"
	v1 "k8s.io/api/core/v1"
)

type Node struct {
	Name        string
	NodeVersion string
}

type Pod struct {
	Name   string
	Status string
}

type Controller struct {
	Client    *utils.K8sClient
	Ofc       *utils.OpenFaasFunc
	Docker    *utils.DockerClient
	NameSpace string
}

const (
	podName   = "test"
	nameSpace = "default"
	yamlName  = "./" + utils.PodYamlPrefix + podName + ".yaml"
)

func NewController(isInCluster bool, nameSpace string) (*Controller, error) {
	cl, err := utils.NewK8sClient(isInCluster)
	if err != nil {
		fmt.Printf("Error: k8s client created failed: %s\n", err.Error())
		return nil, err
	}
	dk, err := utils.NewDockerClient()
	if err != nil {
		fmt.Println("Error: docker client created failed:", err.Error())
		return nil, err
	}
	return &Controller{
		Client: cl,
		Ofc: &utils.OpenFaasFunc{
			Name:             "",
			FuncOpt:          "",
			IsCreateFromFile: false,
			YmlFileAddress:   "",
		},
		Docker:    dk,
		NameSpace: nameSpace,
	}, nil
}

func (ct *Controller) GetNameSpace() string {
	return ct.NameSpace
}

func (ct *Controller) GetPodList() ([]Pod, error) {
	ret, err := ct.Client.GetPodList(ct.NameSpace)
	if err != nil {
		fmt.Printf("Get Pod List failed: %s\n", err.Error())
		return nil, err
	}
	pod := make([]Pod, 0, 1)
	for _, x := range ret.Items {
		pod = append(pod, TransPod(&x))
	}
	return pod, nil
}

func (ct *Controller) GenerateYamlFile(nameSpace string, podName string, isHostNetwork bool) error {
	_, err := utils.GeneratePodYaml(nameSpace, podName, isHostNetwork)
	return err
}

func (ct *Controller) ApplyPod(namespace string, filename string) error {
	err := ct.Client.ApplyYaml(namespace, filename)
	if err != nil {
		fmt.Println("Error: apply pod failed:", err.Error())
	}
	return err
}
func (ct *Controller) DeletePod(nameSpace string, podName string) error {

	err := ct.Client.DeletePod(nameSpace, podName)
	if err != nil {
		fmt.Println("Error: delete pod failed:", err.Error())
	}
	return err
}

func (ct *Controller) GenerateYamlAddress(podName string) string {
	var fileName string
	fileName = "./" + utils.PodYamlPrefix + podName + ".yaml"
	return fileName
}

func (ct *Controller) ExecOpenFaasCmd() error {
	return ct.Ofc.ExecOpenFaasCmd()
}

func (ct *Controller) BuildImage(path string, name string) error {
	return ct.Docker.BuildImage(path, name, true)
}

func (ct *Controller) RemoveImage(name string) error {
	return ct.Docker.RemoveImage(name)
}

func (ct *Controller) CreatePath(path string) error {
	return utils.CreatePath(path)
}

func (ct *Controller) PushImage(name string, version string) error {
	return ct.Docker.PushImage(name, version)
}

func (ct *Controller) CopyTmpFile(srcpath string, dstpath string) error {
	var cmd *exec.Cmd
	cmd = exec.Command("cp", "-rf", srcpath+"/template", dstpath)
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func TransPod(pod *v1.Pod) Pod {
	if pod == nil {
		fmt.Printf("Error: cannot read pod!\n")
	}
	name := pod.Name
	sta := pod.Status.ContainerStatuses[0].Ready
	var newPod Pod
	if sta {
		newPod = Pod{
			Name:   name,
			Status: "Ready",
		}
	} else {
		newPod = Pod{
			Name:   name,
			Status: "Error",
		}
	}

	return newPod
}
