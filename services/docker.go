package services

import (
	"context"
	"fmt"
	"strconv"

	"github.com/docker/distribution/uuid"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func CreateDocker(port int, url string) (containerId string, err error) {
	// 创建Docker客户端
	cli, err := client.NewClientWithOpts(client.WithHost(url), client.WithAPIVersionNegotiation())
	if err != nil {
		return
	}
	// 容器端口映射规则
	portBindings := nat.PortMap{
		"8000/tcp": []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",          // 绑定到所有网络接口
				HostPort: strconv.Itoa(port), // 映射到主机的端口
			},
		},
	}

	// 创建一个容器配置
	containerConfig := &container.Config{Image: "keepwork/jupyterhub:v1.1", // 选择要运行的Docker镜像
		ExposedPorts: map[nat.Port]struct{}{
			"8000/tcp": {}, // 暴露80端口
		},
	}

	// 创建容器
	dockerName := uuid.Generate().String()
	resp, err := cli.ContainerCreate(context.Background(), containerConfig, &container.HostConfig{
		PortBindings: portBindings, // 应用端口映射规则
	}, nil, nil, dockerName)
	if err != nil {
		return
	}

	fmt.Printf("create success %s\n", resp.ID)

	// 启动容器
	if err := cli.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}
	fmt.Println("启动成功啦")
	return resp.ID, nil
}

func DestroyDocker(containerId string, url string) (err error) {
	// 创建Docker客户端
	cli, err := client.NewClientWithOpts(client.WithHost(url), client.WithAPIVersionNegotiation())
	if err != nil {
		return
	}
	// 删除容器
	if err := cli.ContainerRemove(context.Background(), containerId, types.ContainerRemoveOptions{Force: true}); err != nil {
		return err
	}
	return
}
