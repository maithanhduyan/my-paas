package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	dockermount "github.com/docker/docker/api/types/mount"
	core "github.com/my-paas/core"
	"github.com/my-paas/server/docker"
	"github.com/my-paas/server/model"
	"github.com/my-paas/server/store"
)

const (
	buildsDir  = "/data/builds"
	networkName = "mypaas-network"
)

type DeployWorker struct {
	Store  *store.Store
	Docker *docker.Client
}

func (w *DeployWorker) Handle(ctx context.Context, job Job) {
	switch job.Type {
	case JobDeploy:
		w.deploy(ctx, job)
	case JobRollback:
		w.rollback(ctx, job)
	}
}

func (w *DeployWorker) deploy(ctx context.Context, job Job) {
	deployID := job.DeploymentID
	projectID := job.ProjectID

	project, err := w.Store.GetProject(projectID)
	if err != nil || project == nil {
		w.logStep(deployID, "deploy", "error", "project not found: "+projectID)
		w.Store.UpdateDeploymentStatus(deployID, model.DeployFailed)
		return
	}

	// Step 1: Clone
	w.Store.UpdateDeploymentStatus(deployID, model.DeployCloning)
	w.Store.UpdateProjectStatus(projectID, "deploying")
	sourceDir, err := w.cloneRepo(ctx, deployID, project)
	if err != nil {
		w.logStep(deployID, "clone", "error", err.Error())
		w.Store.UpdateDeploymentStatus(deployID, model.DeployFailed)
		w.Store.UpdateProjectStatus(projectID, "failed")
		return
	}

	// Step 2: Detect
	w.Store.UpdateDeploymentStatus(deployID, model.DeployDetecting)
	result := core.Detect(sourceDir)
	if !result.Success {
		w.logStep(deployID, "detect", "error", "detection failed: "+result.Error)
		w.Store.UpdateDeploymentStatus(deployID, model.DeployFailed)
		w.Store.UpdateProjectStatus(projectID, "failed")
		return
	}
	w.logStep(deployID, "detect", "info",
		fmt.Sprintf("detected %s / %s (version: %s)", result.Plan.Provider, result.Plan.Framework, result.Plan.Version))
	w.Store.UpdateProjectDetection(projectID, result.Plan.Provider, result.Plan.Framework)

	// Step 3: Generate Dockerfile & Build
	w.Store.UpdateDeploymentStatus(deployID, model.DeployBuilding)
	imageTag := fmt.Sprintf("mypaas-%s:%s", project.Name, deployID)

	dockerfile, err := core.GenerateDockerfile(sourceDir)
	if err != nil {
		w.logStep(deployID, "build", "error", "dockerfile generation failed: "+err.Error())
		w.Store.UpdateDeploymentStatus(deployID, model.DeployFailed)
		w.Store.UpdateProjectStatus(projectID, "failed")
		return
	}
	w.logStep(deployID, "build", "info", "generated Dockerfile, starting build...")

	cacheFrom := []string{fmt.Sprintf("mypaas-%s:latest", project.Name)}
	err = w.Docker.BuildImage(ctx, sourceDir, dockerfile, imageTag, cacheFrom, func(line string) {
		w.logStep(deployID, "build", "info", line)
	})
	if err != nil {
		w.logStep(deployID, "build", "error", "build failed: "+err.Error())
		w.Store.UpdateDeploymentStatus(deployID, model.DeployFailed)
		w.Store.UpdateProjectStatus(projectID, "failed")
		return
	}
	w.Store.UpdateDeploymentImage(deployID, imageTag)
	w.logStep(deployID, "build", "info", "image built: "+imageTag)

	// Tag as latest for build cache
	latestTag := fmt.Sprintf("mypaas-%s:latest", project.Name)
	w.Docker.TagImage(ctx, imageTag, latestTag)

	// Step 4: Deploy (Swarm service or plain container)
	w.Store.UpdateDeploymentStatus(deployID, model.DeployDeploying)

	// Gather env vars
	envVars, _ := w.Store.GetEnvVars(projectID)
	envMap := make(map[string]string)
	for _, e := range envVars {
		envMap[e.Key] = e.Value
	}
	// Merge plan env vars
	if result.Plan.EnvVars != nil {
		for k, v := range result.Plan.EnvVars {
			if _, exists := envMap[k]; !exists {
				envMap[k] = v
			}
		}
	}

	// Determine port
	port := "8080"
	if len(result.Plan.Ports) > 0 {
		port = result.Plan.Ports[0]
	}

	// Build labels with Traefik routing
	labels := map[string]string{
		"mypaas.project":    projectID,
		"mypaas.deployment": deployID,
		"mypaas.port":       port,
	}
	for k, v := range w.buildTraefikLabels(project, port) {
		labels[k] = v
	}

	// Resource limits
	var cpuLimit float64
	var memLimit int64
	if project.CPULimit > 0 {
		cpuLimit = project.CPULimit
	}
	if project.MemLimit > 0 {
		memLimit = project.MemLimit * 1024 * 1024 // MB to bytes
	}

	// Check if Swarm is active
	if w.Docker.IsSwarmActive(ctx) {
		w.deploySwarmService(ctx, deployID, project, imageTag, envMap, labels, port, cpuLimit, memLimit)
	} else {
		w.deployContainer(ctx, deployID, project, imageTag, envMap, labels, result.Plan.Ports, cpuLimit, memLimit)
	}
}

func (w *DeployWorker) deployContainer(ctx context.Context, deployID string, project *model.Project, imageTag string, envMap, labels map[string]string, ports []string, cpuLimit float64, memLimit int64) {
	projectID := project.ID
	containerName := fmt.Sprintf("mypaas-%s-%s", project.Name, deployID)

	// Stop previous container if exists
	oldContainerID, _ := w.Docker.FindContainerByName(ctx, fmt.Sprintf("mypaas-%s-", project.Name))
	if oldContainerID == "" {
		oldContainerID = w.findProjectContainer(ctx, project.Name)
	}

	// Gather volume mounts
	var binds []string
	volumes, _ := w.Store.ListVolumes(projectID)
	for _, v := range volumes {
		binds = append(binds, fmt.Sprintf("mypaas-%s-%s:%s", project.Name, v.Name, v.MountPath))
	}

	containerID, err := w.Docker.RunContainer(ctx, docker.RunContainerOpts{
		Name:     containerName,
		Image:    imageTag,
		Env:      envMap,
		Ports:    ports,
		Network:  networkName,
		Labels:   labels,
		CPULimit: cpuLimit,
		MemLimit: memLimit,
		Binds:    binds,
	})
	if err != nil {
		w.logStep(deployID, "deploy", "error", "failed to start container: "+err.Error())
		w.Store.UpdateDeploymentStatus(deployID, model.DeployFailed)
		w.Store.UpdateProjectStatus(projectID, "failed")
		return
	}
	w.logStep(deployID, "deploy", "info", "container started: "+containerName)

	// Health check
	err = w.Docker.HealthCheck(ctx, containerID, 60*time.Second)
	if err != nil {
		w.logStep(deployID, "healthcheck", "error", "health check failed: "+err.Error())
		w.Docker.StopContainer(ctx, containerID)
		w.Store.UpdateDeploymentStatus(deployID, model.DeployFailed)
		w.Store.UpdateProjectStatus(projectID, "failed")
		return
	}

	// Success: stop old container
	if oldContainerID != "" {
		w.logStep(deployID, "deploy", "info", "stopping old container")
		w.Docker.StopContainer(ctx, oldContainerID)
	}

	w.logStep(deployID, "healthcheck", "info", "container is healthy")
	w.Store.UpdateDeploymentStatus(deployID, model.DeployHealthy)
	w.Store.UpdateProjectStatus(projectID, "healthy")
	log.Printf("[deploy] project=%s deployment=%s completed successfully", projectID, deployID)
}

func (w *DeployWorker) deploySwarmService(ctx context.Context, deployID string, project *model.Project, imageTag string, envMap, labels map[string]string, port string, cpuLimit float64, memLimit int64) {
	projectID := project.ID
	serviceName := fmt.Sprintf("mypaas-%s", project.Name)

	// Gather volume mounts
	var mounts []dockermount.Mount
	volumes, _ := w.Store.ListVolumes(projectID)
	for _, v := range volumes {
		mounts = append(mounts, dockermount.Mount{
			Type:   dockermount.TypeVolume,
			Source: fmt.Sprintf("mypaas-%s-%s", project.Name, v.Name),
			Target: v.MountPath,
		})
	}

	// Check if service already exists (update vs create)
	existing, _ := w.Docker.FindSwarmServiceByName(ctx, serviceName)
	if existing != nil {
		w.logStep(deployID, "deploy", "info", "updating existing Swarm service: "+serviceName)
		err := w.Docker.UpdateSwarmService(ctx, existing.ID, imageTag, project.Replicas, envMap)
		if err != nil {
			w.logStep(deployID, "deploy", "error", "failed to update service: "+err.Error())
			w.Store.UpdateDeploymentStatus(deployID, model.DeployFailed)
			w.Store.UpdateProjectStatus(projectID, "failed")
			return
		}
		// Update labels on existing service
		w.Docker.UpdateSwarmServiceLabels(ctx, existing.ID, labels)
	} else {
		w.logStep(deployID, "deploy", "info", "creating Swarm service: "+serviceName)
		replicas := project.Replicas
		if replicas == 0 {
			replicas = 1
		}
		_, err := w.Docker.CreateSwarmService(ctx, docker.SwarmServiceOpts{
			Name:     serviceName,
			Image:    imageTag,
			Env:      envMap,
			Labels:   labels,
			Network:  networkName,
			Replicas: replicas,
			CPULimit: cpuLimit,
			MemLimit: memLimit,
			Mounts:   mounts,
		})
		if err != nil {
			w.logStep(deployID, "deploy", "error", "failed to create service: "+err.Error())
			w.Store.UpdateDeploymentStatus(deployID, model.DeployFailed)
			w.Store.UpdateProjectStatus(projectID, "failed")
			return
		}
	}

	// Health check via Swarm task state
	svc, _ := w.Docker.FindSwarmServiceByName(ctx, serviceName)
	if svc != nil {
		err := w.Docker.SwarmServiceHealthy(ctx, svc.ID, 90*time.Second)
		if err != nil {
			w.logStep(deployID, "healthcheck", "error", "service health check failed: "+err.Error())
			w.Store.UpdateDeploymentStatus(deployID, model.DeployFailed)
			w.Store.UpdateProjectStatus(projectID, "failed")
			return
		}
	}

	w.logStep(deployID, "healthcheck", "info", "service is healthy (Swarm mode)")
	w.Store.UpdateDeploymentStatus(deployID, model.DeployHealthy)
	w.Store.UpdateProjectStatus(projectID, "healthy")
	log.Printf("[deploy-swarm] project=%s deployment=%s completed successfully", projectID, deployID)
}

func (w *DeployWorker) rollback(ctx context.Context, job Job) {
	deployID := job.DeploymentID
	projectID := job.ProjectID

	// Find the target deployment to rollback to
	target, err := w.Store.GetDeployment(deployID)
	if err != nil || target == nil {
		w.logStep(deployID, "rollback", "error", "deployment not found")
		return
	}

	if target.ImageTag == "" {
		w.logStep(deployID, "rollback", "error", "no image tag for rollback target")
		return
	}

	project, err := w.Store.GetProject(projectID)
	if err != nil || project == nil {
		w.logStep(deployID, "rollback", "error", "project not found")
		return
	}

	// Create a new deployment record for the rollback
	rollbackDeploy, err := w.Store.CreateDeployment(projectID, model.TriggerRollback)
	if err != nil {
		return
	}

	w.Store.UpdateDeploymentStatus(rollbackDeploy.ID, model.DeployDeploying)

	envVars, _ := w.Store.GetEnvVars(projectID)
	envMap := make(map[string]string)
	for _, e := range envVars {
		envMap[e.Key] = e.Value
	}

	// Build labels with Traefik routing
	port := "8080"
	labels := map[string]string{
		"mypaas.project":    projectID,
		"mypaas.deployment": rollbackDeploy.ID,
		"mypaas.port":       port,
	}
	for k, v := range w.buildTraefikLabels(project, port) {
		labels[k] = v
	}

	// Resource limits
	var cpuLimit float64
	var memLimit int64
	if project.CPULimit > 0 {
		cpuLimit = project.CPULimit
	}
	if project.MemLimit > 0 {
		memLimit = project.MemLimit * 1024 * 1024
	}

	// Swarm mode: update existing service with the old image
	if w.Docker.IsSwarmActive(ctx) {
		serviceName := fmt.Sprintf("mypaas-%s", project.Name)
		existing, _ := w.Docker.FindSwarmServiceByName(ctx, serviceName)
		if existing != nil {
			w.logStep(rollbackDeploy.ID, "rollback", "info", "rolling back Swarm service to "+target.ImageTag)
			err := w.Docker.UpdateSwarmService(ctx, existing.ID, target.ImageTag, project.Replicas, envMap)
			if err != nil {
				w.logStep(rollbackDeploy.ID, "rollback", "error", "failed to update service: "+err.Error())
				w.Store.UpdateDeploymentStatus(rollbackDeploy.ID, model.DeployFailed)
				return
			}
			// Wait for service to converge
			err = w.Docker.SwarmServiceHealthy(ctx, existing.ID, 90*time.Second)
			if err != nil {
				w.logStep(rollbackDeploy.ID, "rollback", "error", "health check failed: "+err.Error())
				w.Store.UpdateDeploymentStatus(rollbackDeploy.ID, model.DeployFailed)
				return
			}
		} else {
			w.logStep(rollbackDeploy.ID, "rollback", "error", "Swarm service not found: "+serviceName)
			w.Store.UpdateDeploymentStatus(rollbackDeploy.ID, model.DeployFailed)
			return
		}
	} else {
		// Container mode: stop old, start new
		containerName := fmt.Sprintf("mypaas-%s-%s", project.Name, rollbackDeploy.ID)
		oldContainerID := w.findProjectContainer(ctx, project.Name)

		containerID, err := w.Docker.RunContainer(ctx, docker.RunContainerOpts{
			Name:     containerName,
			Image:    target.ImageTag,
			Env:      envMap,
			Network:  networkName,
			Labels:   labels,
			CPULimit: cpuLimit,
			MemLimit: memLimit,
		})
		if err != nil {
			w.logStep(rollbackDeploy.ID, "rollback", "error", err.Error())
			w.Store.UpdateDeploymentStatus(rollbackDeploy.ID, model.DeployFailed)
			return
		}

		err = w.Docker.HealthCheck(ctx, containerID, 30*time.Second)
		if err != nil {
			w.Docker.StopContainer(ctx, containerID)
			w.Store.UpdateDeploymentStatus(rollbackDeploy.ID, model.DeployFailed)
			return
		}

		if oldContainerID != "" {
			w.Docker.StopContainer(ctx, oldContainerID)
		}
	}

	w.Store.UpdateDeploymentImage(rollbackDeploy.ID, target.ImageTag)
	w.Store.UpdateDeploymentStatus(rollbackDeploy.ID, model.DeployHealthy)
	w.Store.UpdateProjectStatus(projectID, "healthy")
	w.logStep(rollbackDeploy.ID, "rollback", "info", "rolled back to deployment "+deployID)
}

func (w *DeployWorker) cloneRepo(ctx context.Context, deployID string, project *model.Project) (string, error) {
	if project.GitURL == "" {
		return "", fmt.Errorf("no git URL configured")
	}

	destDir := filepath.Join(buildsDir, project.ID, deployID)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir: %w", err)
	}

	w.logStep(deployID, "clone", "info", fmt.Sprintf("cloning %s (branch: %s)", project.GitURL, project.Branch))

	cmd := exec.CommandContext(ctx, "git", "clone",
		"--depth", "1",
		"--branch", project.Branch,
		project.GitURL,
		destDir,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git clone failed: %s\n%s", err, string(output))
	}

	// Get commit info
	hashCmd := exec.CommandContext(ctx, "git", "-C", destDir, "rev-parse", "--short", "HEAD")
	hashOut, _ := hashCmd.Output()
	msgCmd := exec.CommandContext(ctx, "git", "-C", destDir, "log", "-1", "--pretty=%s")
	msgOut, _ := msgCmd.Output()

	commitHash := strings.TrimSpace(string(hashOut))
	commitMsg := strings.TrimSpace(string(msgOut))
	w.Store.UpdateDeploymentCommit(deployID, commitHash, commitMsg)
	w.logStep(deployID, "clone", "info", fmt.Sprintf("cloned commit %s: %s", commitHash, commitMsg))

	return destDir, nil
}

func (w *DeployWorker) findProjectContainer(ctx context.Context, projectName string) string {
	// Search for any container with the mypaas project label
	prefix := "mypaas-" + projectName + "-"
	containers, err := w.Docker.FindContainerByName(ctx, prefix)
	if err == nil && containers != "" {
		return containers
	}
	return ""
}

func (w *DeployWorker) logStep(deploymentID, step, level, message string) {
	w.Store.AddDeploymentLog(deploymentID, step, level, message)
	log.Printf("[%s] [%s] %s: %s", deploymentID, step, level, message)
}

func (w *DeployWorker) buildTraefikLabels(project *model.Project, port string) map[string]string {
	labels := make(map[string]string)
	baseDomain := os.Getenv("MYPAAS_DOMAIN")
	if baseDomain == "" {
		return labels
	}

	routerName := "mypaas-" + project.Name
	subdomain := project.Name + "." + baseDomain

	labels["traefik.enable"] = "true"
	labels["traefik.http.services."+routerName+".loadbalancer.server.port"] = port
	labels["traefik.http.routers."+routerName+".rule"] = "Host(`" + subdomain + "`)"
	labels["traefik.http.routers."+routerName+".entrypoints"] = "web"
	labels["traefik.http.routers."+routerName+".service"] = routerName

	// Custom domain routers
	domains, _ := w.Store.GetDomains(project.ID)
	for i, d := range domains {
		dr := fmt.Sprintf("%s-d%d", routerName, i)
		labels["traefik.http.routers."+dr+".rule"] = "Host(`" + d.Domain + "`)"
		labels["traefik.http.routers."+dr+".service"] = routerName
		if d.SSLAuto {
			labels["traefik.http.routers."+dr+".entrypoints"] = "websecure"
			labels["traefik.http.routers."+dr+".tls.certresolver"] = "letsencrypt"
		} else {
			labels["traefik.http.routers."+dr+".entrypoints"] = "web"
		}
	}

	return labels
}
