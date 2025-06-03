package ics_node

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type ICS struct {
	Client *client.Client
	nodes  []*DeviceConfig
}

func NewICS() *ICS {
	Client, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
		client.WithVersion("1.49"),
	)
	if err != nil {
		log.Fatal("Error creating ICS docker ics.Client", err)
	}
	return &ICS{
		Client: Client,
		nodes:  make([]*DeviceConfig, 0),
	}
}

// BuildAndRunContainer builds an image from a Dockerfile and runs the container.
func (ics ICS) BuildAndRunContainer(ctx context.Context, config DeviceConfig) error {
	log.Println("Building ics container...")
	log.Printf("device %v\n", config)
	// 1. Create tar of Dockerfile directory
	tarBuf := new(bytes.Buffer)
	tw := tar.NewWriter(tarBuf)

	dockerfileDir := filepath.Dir(config.DockerfilePath)
	err := filepath.Walk(dockerfileDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(dockerfileDir, file)
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		header, err := tar.FileInfoHeader(fi, relPath)
		if err != nil {
			return err
		}
		header.Name = relPath
		err = tw.WriteHeader(header)
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		io.Copy(tw, f)
		return nil
	})
	tw.Close()
	if err != nil {
		return fmt.Errorf("Error creating build context: %w", err)
	}

	// 2. Build the image
	buildResp, err := ics.Client.ImageBuild(ctx, bytes.NewReader(tarBuf.Bytes()), build.ImageBuildOptions{
		Tags:       []string{config.ImageName},
		Dockerfile: filepath.Base(config.DockerfilePath),
		Remove:     true,
	})
	if err != nil {
		return fmt.Errorf("Build failed: %w", err)
	}
	defer buildResp.Body.Close()
	scanner := bufio.NewScanner(buildResp.Body)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	// 3. Create container
	// 1. Define exposed ports (for internal container networking)
	exposedPorts := nat.PortSet{
		"502/tcp": struct{}{},
	}

	// 2. Define host port bindings (optional — only if you need external access)

	//portBindings := nat.PortMap{
	//	"502/tcp": []nat.PortBinding{
	//		{
	//			HostIP:   "0.0.0.0",
	//			HostPort: "502",
	//		},
	//	},
	//}
	//
	// 3. Create HostConfig
	hostConfig := &container.HostConfig{
		//PortBindings: portBindings,
		NetworkMode: "honeynet",
		AutoRemove:  true,
	}

	// 4. Optionally define NetworkingConfig if you want to set endpoint config (e.g., static IPs)
	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"honeynet": {
				IPAddress: "172.18.0.15", // Static internal IP
			},
		},
	}

	// 5. Container config
	containerConfig := &container.Config{
		Image:        config.ImageName,
		ExposedPorts: exposedPorts,
	}

	// 6. Create the container
	_, err = ics.Client.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		networkingConfig,
		nil,
		config.ContainerName,
	)
	if err != nil {
		return fmt.Errorf("Container create failed: %w", err)
	}

	// 4. Start container
	err = ics.Client.ContainerStart(ctx, config.ContainerName, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("Container start failed: %w", err)
	}

	fmt.Printf("Container %s started successfully.\n", config.ContainerName)
	return nil
}
