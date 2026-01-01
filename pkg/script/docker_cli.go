package script

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

type DockerCli struct {
	cli *client.Client
}

func (cli *DockerCli) RunImage(tag string, cmd []string, options map[string]any) (string, error) {
	ctx := context.Background()

	// Check if output capture is requested
	captureOutput := false
	if options != nil {
		if capture, ok := options["captureOutput"].(bool); ok {
			captureOutput = capture
		}
	}

	// Build container config
	config := &container.Config{
		Image:        tag,
		Cmd:          cmd,
		AttachStdout: captureOutput,
		AttachStderr: captureOutput,
	}

	// Build host config
	hostConfig := &container.HostConfig{}

	// Parse options
	if options != nil {
		// User
		if user, ok := options["user"].(string); ok {
			config.User = user
		}

		// Volumes/Binds
		if volumes, ok := options["volumes"].([]string); ok {
			hostConfig.Binds = volumes
		}

		// AutoRemove (--rm flag)
		if autoRemove, ok := options["autoRemove"].(bool); ok {
			hostConfig.AutoRemove = autoRemove
		}

		// Environment variables
		if env, ok := options["env"].([]string); ok {
			config.Env = env
		}

		// Working directory
		if workDir, ok := options["workDir"].(string); ok {
			config.WorkingDir = workDir
		}
	}

	// Create container
	resp, err := cli.cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return "", err
	}

	// Attach to container if output capture is requested
	var attachResp types.HijackedResponse
	if captureOutput {
		attachResp, err = cli.cli.ContainerAttach(ctx, resp.ID, container.AttachOptions{
			Stream: true,
			Stdout: true,
			Stderr: true,
		})
		if err != nil {
			return "", err
		}
		defer attachResp.Close()
	}

	// Start container
	err = cli.cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return "", err
	}

	// Wait for container to finish (if autoRemove or captureOutput)
	if hostConfig.AutoRemove || captureOutput {
		statusCh, errCh := cli.cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
		select {
		case err := <-errCh:
			if err != nil {
				return "", err
			}
		case <-statusCh:
		}
	}

	// Return output if captured, otherwise return container ID
	if captureOutput {
		output, err := io.ReadAll(attachResp.Reader)
		if err != nil {
			return "", err
		}
		return string(output), nil
	}

	return resp.ID, nil
}

func (cli *DockerCli) ContainerExec(containerID string, cmd []string) (string, error) {
	return cli.ContainerExecAsUser(containerID, "", cmd)
}

func (cli *DockerCli) ContainerExecAsUser(containerID string, user string, cmd []string) (string, error) {
	ctx := context.Background()

	// Create exec instance
	execConfig := container.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
		User:         user,
	}

	execID, err := cli.cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return "", err
	}

	// Start exec and attach to get output
	resp, err := cli.cli.ContainerExecAttach(ctx, execID.ID, container.ExecAttachOptions{})
	if err != nil {
		return "", err
	}
	defer resp.Close()

	// Read output
	output, err := io.ReadAll(resp.Reader)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func (cli *DockerCli) ContainerExecToFile(containerID, user string, cmd []string, outputFile string) error {
	ctx := context.Background()

	// Create exec instance
	execConfig := container.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
		User:         user,
	}

	execID, err := cli.cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return err
	}

	// Start exec and attach to get output
	resp, err := cli.cli.ContainerExecAttach(ctx, execID.ID, container.ExecAttachOptions{})
	if err != nil {
		return err
	}
	defer resp.Close()

	// Create output file
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Stream output directly to file
	_, err = io.Copy(file, resp.Reader)
	return err
}

func (cli *DockerCli) ListContainers(options map[string]any) ([]container.Summary, error) {
	option := container.ListOptions{}
	if options != nil {
		if all, ok := options["all"]; ok {
			option.All = all.(bool)
		}
	}

	ctx := context.Background()
	return cli.cli.ContainerList(ctx, option)
}

func (cli *DockerCli) InspectContainer(id string) (container.InspectResponse, error) {
	ctx := context.Background()
	return cli.cli.ContainerInspect(ctx, id)
}

func (cli *DockerCli) StartContainer(id string) error {
	return cli.cli.ContainerStart(context.Background(), id, container.StartOptions{})
}

func (cli *DockerCli) StopContainer(id string) error {
	return cli.cli.ContainerStop(context.Background(), id, container.StopOptions{})
}

func (cli *DockerCli) RemoveContainer(id string) error {
	return cli.cli.ContainerRemove(context.Background(), id, container.RemoveOptions{})
}

func (cli *DockerCli) PauseContainer(id string) error {
	return cli.cli.ContainerPause(context.Background(), id)
}

func (cli *DockerCli) UnpauseContainer(id string) error {
	return cli.cli.ContainerUnpause(context.Background(), id)
}

func (cli *DockerCli) ListVolumes() (volume.ListResponse, error) {
	ctx := context.Background()
	return cli.cli.VolumeList(ctx, volume.ListOptions{})
}

func (cli *DockerCli) InspectVolume(name string) (volume.Volume, error) {
	ctx := context.Background()
	return cli.cli.VolumeInspect(ctx, name)
}

func (cli *DockerCli) RemoveVolume(name string) error {
	ctx := context.Background()
	return cli.cli.VolumeRemove(ctx, name, true)
}

func (cli *DockerCli) RemoveImage(name string) ([]image.DeleteResponse, error) {
	ctx := context.Background()
	return cli.cli.ImageRemove(ctx, name, image.RemoveOptions{})
}

func (cli *DockerCli) ListImages() ([]image.Summary, error) {
	ctx := context.Background()
	return cli.cli.ImageList(ctx, image.ListOptions{})
}

func (cli *DockerCli) InspectImage(name string) (image.InspectResponse, error) {
	ctx := context.Background()
	return cli.cli.ImageInspect(ctx, name)
}

func (cli *DockerCli) PullImage(name string) error {
	ctx := context.Background()
	out, err := cli.cli.ImagePull(ctx, name, image.PullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()
	// Read the output to ensure the pull completes
	_, err = io.ReadAll(out)
	return err
}

func NewDockerCli(host *string) *DockerCli {
	var cli *client.Client
	var err error
	if host != nil && len(*host) > 0 {
		cli, err = client.NewClientWithOpts(
			client.WithHost(*host),
			client.WithAPIVersionNegotiation(),
		)
	} else {
		cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	}

	if err != nil {
		log.Fatal(err)
	}
	return &DockerCli{
		cli: cli,
	}
}
