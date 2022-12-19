package utils

import (
	"os"

	docker "github.com/fsouza/go-dockerclient"
)

type DockerClient struct {
	Client *docker.Client
	Auth   *docker.AuthConfiguration
}

func NewDockerClient() (*DockerClient, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}
	return &DockerClient{
		Client: client,
		Auth: &docker.AuthConfiguration{
			Username: "",
			Password: "",
		},
	}, nil
}

func (cl *DockerClient) BuildImage(path string, name string, isRemoveTmpContainer bool) error {
	opt := docker.BuildImageOptions{
		Name:           cl.Auth.Username + "/" + name,
		Dockerfile:     "Dockerfile",
		OutputStream:   os.Stdout,
		ContextDir:     path,
		RmTmpContainer: isRemoveTmpContainer,
	}
	return cl.Client.BuildImage(opt)
}

func (cl *DockerClient) RemoveImage(name string) error {
	return cl.Client.RemoveImage(name)
}

func (cl *DockerClient) PushImage(name string, version string) error {
	opt := docker.PushImageOptions{
		Name: cl.Auth.Username + "/" + name + ":" + version,
		Tag:  version,
	}
	return cl.Client.PushImage(opt, *cl.Auth)
}
