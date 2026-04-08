package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/moby/go-archive"
)

type Client struct {
	cli *client.Client
}

func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}
	return &Client{cli: cli}, nil
}

func (c *Client) Close() error {
	return c.cli.Close()
}

func (c *Client) Ping(ctx context.Context) error {
	_, err := c.cli.Ping(ctx)
	return err
}

// BuildImage builds a Docker image from a source directory using a Dockerfile string.
// It writes the Dockerfile to the build context before building.
// logFn receives build output lines in real-time.
func (c *Client) BuildImage(ctx context.Context, contextDir string, dockerfile string, imageTag string, cacheFrom []string, logFn func(string)) error {
	tar, err := archive.TarWithOptions(contextDir, &archive.TarOptions{})
	if err != nil {
		return fmt.Errorf("tar context: %w", err)
	}
	defer tar.Close()

	// Inject Dockerfile into the tar stream
	dockerfileTar, err := archive.Generate("Dockerfile", dockerfile)
	if err != nil {
		return fmt.Errorf("generate dockerfile: %w", err)
	}

	buildContext := io.MultiReader(dockerfileTar, tar)

	resp, err := c.cli.ImageBuild(ctx, buildContext, types.ImageBuildOptions{
		Tags:        []string{imageTag},
		Dockerfile:  "Dockerfile",
		Remove:      true,
		ForceRemove: true,
		NoCache:     false,
		CacheFrom:   cacheFrom,
	})
	if err != nil {
		return fmt.Errorf("image build: %w", err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var msg struct {
			Stream string `json:"stream"`
			Error  string `json:"error"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &msg); err == nil {
			if msg.Error != "" {
				return fmt.Errorf("build error: %s", msg.Error)
			}
			if msg.Stream != "" && logFn != nil {
				logFn(strings.TrimRight(msg.Stream, "\n"))
			}
		}
	}
	return scanner.Err()
}

// RunContainer creates and starts a container from an image.
func (c *Client) RunContainer(ctx context.Context, opts RunContainerOpts) (string, error) {
	// Ensure network exists
	if opts.Network != "" {
		c.ensureNetwork(ctx, opts.Network)
	}

	envList := make([]string, 0, len(opts.Env))
	for k, v := range opts.Env {
		envList = append(envList, k+"="+v)
	}

	exposedPorts := make(map[string]struct{})
	for _, p := range opts.Ports {
		exposedPorts[p+"/tcp"] = struct{}{}
	}

	containerConfig := &container.Config{
		Image:        opts.Image,
		Env:          envList,
		Labels:       opts.Labels,
		ExposedPorts: nil, // Let Docker handle this
	}

	hostConfig := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
		Binds:         opts.Binds,
	}

	if opts.CPULimit > 0 {
		hostConfig.Resources.NanoCPUs = int64(opts.CPULimit * 1e9)
	}
	if opts.MemLimit > 0 {
		hostConfig.Resources.Memory = opts.MemLimit
	}

	if opts.PortBindings != nil {
		hostConfig.PortBindings = opts.PortBindings
	}

	networkConfig := &network.NetworkingConfig{}
	if opts.Network != "" {
		networkConfig.EndpointsConfig = map[string]*network.EndpointSettings{
			opts.Network: {},
		}
	}

	resp, err := c.cli.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, nil, opts.Name)
	if err != nil {
		return "", fmt.Errorf("create container: %w", err)
	}

	if err := c.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		// Cleanup on failure
		c.cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
		return "", fmt.Errorf("start container: %w", err)
	}

	return resp.ID, nil
}

type RunContainerOpts struct {
	Name         string
	Image        string
	Env          map[string]string
	Ports        []string
	PortBindings nat.PortMap
	Labels       map[string]string
	Network      string
	CPULimit     float64 // CPU limit in cores (0 = unlimited)
	MemLimit     int64   // Memory limit in bytes (0 = unlimited)
	Binds        []string // Volume binds (e.g. "vol-name:/path")
}

// StopContainer stops and removes a container.
func (c *Client) StopContainer(ctx context.Context, containerID string) error {
	timeout := 30
	stopOpts := container.StopOptions{Timeout: &timeout}
	if err := c.cli.ContainerStop(ctx, containerID, stopOpts); err != nil {
		// If not found, that's fine
		if !client.IsErrNotFound(err) {
			return err
		}
	}
	return c.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
}

// GetContainerLogs returns container log output as a reader.
func (c *Client) GetContainerLogs(ctx context.Context, containerID string, follow bool, tail string) (io.ReadCloser, error) {
	opts := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Tail:       tail,
		Timestamps: true,
	}
	return c.cli.ContainerLogs(ctx, containerID, opts)
}

// GetContainerLogLines returns log lines as strings using stdcopy demux.
func (c *Client) GetContainerLogLines(ctx context.Context, containerID string, tail string) ([]string, error) {
	reader, err := c.GetContainerLogs(ctx, containerID, false, tail)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var stdout, stderr strings.Builder
	stdcopy.StdCopy(&stdout, &stderr, reader)

	var lines []string
	for _, line := range strings.Split(stdout.String()+stderr.String(), "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, nil
}

// ContainerStatus returns the status of a container.
func (c *Client) ContainerStatus(ctx context.Context, containerID string) (string, error) {
	info, err := c.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		if client.IsErrNotFound(err) {
			return "not_found", nil
		}
		return "", err
	}
	return info.State.Status, nil
}

// FindContainerByName finds a container by name or name prefix.
func (c *Client) FindContainerByName(ctx context.Context, name string) (string, error) {
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return "", err
	}
	for _, ctr := range containers {
		for _, n := range ctr.Names {
			n = strings.TrimPrefix(n, "/")
			if strings.HasPrefix(n, name) {
				return ctr.ID, nil
			}
		}
	}
	return "", nil
}

// RemoveImage removes a Docker image.
func (c *Client) RemoveImage(ctx context.Context, imageTag string) error {
	_, err := c.cli.ImageRemove(ctx, imageTag, image.RemoveOptions{Force: true, PruneChildren: true})
	return err
}

// HealthCheck performs a basic health check by inspecting the container state.
func (c *Client) HealthCheck(ctx context.Context, containerID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		info, err := c.cli.ContainerInspect(ctx, containerID)
		if err != nil {
			return err
		}
		if info.State.Running {
			return nil
		}
		if info.State.Status == "exited" {
			return fmt.Errorf("container exited with code %d", info.State.ExitCode)
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("health check timed out after %s", timeout)
}

func (c *Client) ensureNetwork(ctx context.Context, name string) {
	networks, err := c.cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return
	}
	for _, n := range networks {
		if n.Name == name {
			return
		}
	}
	c.cli.NetworkCreate(ctx, name, network.CreateOptions{Driver: "bridge"})
}

// ListContainers returns all containers (running and stopped).
func (c *Client) ListContainers(ctx context.Context) ([]types.Container, error) {
	return c.cli.ContainerList(ctx, container.ListOptions{All: false})
}

// ContainerStatsOnce returns a one-shot stats response for a container.
func (c *Client) ContainerStatsOnce(ctx context.Context, containerID string) (container.StatsResponseReader, error) {
	return c.cli.ContainerStats(ctx, containerID, false)
}

// PullImage pulls a Docker image from a registry.
func (c *Client) PullImage(ctx context.Context, imageRef string) error {
	reader, err := c.cli.ImagePull(ctx, imageRef, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("pull image: %w", err)
	}
	defer reader.Close()
	// Drain the reader to complete the pull
	io.Copy(io.Discard, reader)
	return nil
}

// TagImage tags a Docker image with a new tag.
func (c *Client) TagImage(ctx context.Context, source, target string) error {
	return c.cli.ImageTag(ctx, source, target)
}

// --- Swarm API ---

// IsSwarmActive returns true if Docker is running in Swarm mode.
func (c *Client) IsSwarmActive(ctx context.Context) bool {
	info, err := c.cli.Info(ctx)
	if err != nil {
		return false
	}
	return info.Swarm.LocalNodeState == swarm.LocalNodeStateActive
}

// SwarmInit initializes Docker Swarm on the current node.
func (c *Client) SwarmInit(ctx context.Context, advertiseAddr string) (string, error) {
	return c.cli.SwarmInit(ctx, swarm.InitRequest{
		ListenAddr:    "0.0.0.0:2377",
		AdvertiseAddr: advertiseAddr,
	})
}

// SwarmJoinToken returns the worker join token.
func (c *Client) SwarmJoinToken(ctx context.Context) (string, error) {
	sw, err := c.cli.SwarmInspect(ctx)
	if err != nil {
		return "", err
	}
	return sw.JoinTokens.Worker, nil
}

// SwarmServiceOpts contains options for creating a Swarm service.
type SwarmServiceOpts struct {
	Name     string
	Image    string
	Env      map[string]string
	Labels   map[string]string
	Network  string
	Replicas uint64
	CPULimit float64
	MemLimit int64
	Mounts   []mount.Mount
}

// CreateSwarmService creates a new Swarm service.
func (c *Client) CreateSwarmService(ctx context.Context, opts SwarmServiceOpts) (string, error) {
	if opts.Network != "" {
		c.ensureOverlayNetwork(ctx, opts.Network)
	}

	envList := make([]string, 0, len(opts.Env))
	for k, v := range opts.Env {
		envList = append(envList, k+"="+v)
	}

	replicas := opts.Replicas
	if replicas == 0 {
		replicas = 1
	}

	spec := swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name:   opts.Name,
			Labels: opts.Labels,
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image: opts.Image,
				Env:   envList,
				Mounts: opts.Mounts,
			},
		},
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: &replicas,
			},
		},
	}

	// Resource limits
	if opts.CPULimit > 0 || opts.MemLimit > 0 {
		spec.TaskTemplate.Resources = &swarm.ResourceRequirements{
			Limits: &swarm.Limit{},
		}
		if opts.CPULimit > 0 {
			spec.TaskTemplate.Resources.Limits.NanoCPUs = int64(opts.CPULimit * 1e9)
		}
		if opts.MemLimit > 0 {
			spec.TaskTemplate.Resources.Limits.MemoryBytes = opts.MemLimit
		}
	}

	// Network
	if opts.Network != "" {
		spec.TaskTemplate.Networks = []swarm.NetworkAttachmentConfig{
			{Target: opts.Network},
		}
	}

	resp, err := c.cli.ServiceCreate(ctx, spec, types.ServiceCreateOptions{})
	if err != nil {
		return "", fmt.Errorf("create service: %w", err)
	}
	return resp.ID, nil
}

// UpdateSwarmService updates a Swarm service (image, replicas, env).
func (c *Client) UpdateSwarmService(ctx context.Context, serviceID string, image string, replicas uint64, env map[string]string) error {
	svc, _, err := c.cli.ServiceInspectWithRaw(ctx, serviceID, types.ServiceInspectOptions{})
	if err != nil {
		return err
	}

	if image != "" {
		svc.Spec.TaskTemplate.ContainerSpec.Image = image
	}
	if replicas > 0 {
		svc.Spec.Mode.Replicated.Replicas = &replicas
	}
	if env != nil {
		envList := make([]string, 0, len(env))
		for k, v := range env {
			envList = append(envList, k+"="+v)
		}
		svc.Spec.TaskTemplate.ContainerSpec.Env = envList
	}

	_, err = c.cli.ServiceUpdate(ctx, serviceID, svc.Version, svc.Spec, types.ServiceUpdateOptions{})
	return err
}

// RemoveSwarmService removes a Swarm service.
func (c *Client) RemoveSwarmService(ctx context.Context, serviceID string) error {
	return c.cli.ServiceRemove(ctx, serviceID)
}

// FindSwarmServiceByName finds a Swarm service by name.
func (c *Client) FindSwarmServiceByName(ctx context.Context, name string) (*swarm.Service, error) {
	services, err := c.cli.ServiceList(ctx, types.ServiceListOptions{})
	if err != nil {
		return nil, err
	}
	for _, svc := range services {
		if svc.Spec.Name == name {
			return &svc, nil
		}
	}
	return nil, nil
}

// SwarmServiceHealthy checks if a Swarm service has running tasks.
func (c *Client) SwarmServiceHealthy(ctx context.Context, serviceID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		tasks, err := c.cli.TaskList(ctx, types.TaskListOptions{
			Filters: filters.NewArgs(filters.Arg("service", serviceID)),
		})
		if err != nil {
			return err
		}
		for _, task := range tasks {
			if task.Status.State == swarm.TaskStateRunning {
				return nil
			}
			if task.Status.State == swarm.TaskStateFailed || task.Status.State == swarm.TaskStateRejected {
				return fmt.Errorf("service task %s: %s", task.Status.State, task.Status.Err)
			}
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("service health check timed out after %s", timeout)
}

// ListSwarmNodes returns all Swarm nodes.
func (c *Client) ListSwarmNodes(ctx context.Context) ([]swarm.Node, error) {
	return c.cli.NodeList(ctx, types.NodeListOptions{})
}

func (c *Client) ensureOverlayNetwork(ctx context.Context, name string) {
	networks, err := c.cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return
	}
	for _, n := range networks {
		if n.Name == name {
			return
		}
	}
	c.cli.NetworkCreate(ctx, name, network.CreateOptions{
		Driver:     "overlay",
		Attachable: true,
	})
}
