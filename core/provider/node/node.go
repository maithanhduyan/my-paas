package node

import (
	"encoding/json"
	"strings"

	"github.com/my-paas/core/app"
	"github.com/my-paas/core/plan"
)

type packageJSON struct {
	Name         string            `json:"name"`
	Scripts      map[string]string `json:"scripts"`
	Dependencies map[string]string `json:"dependencies"`
	DevDeps      map[string]string `json:"devDependencies"`
	Engines      struct {
		Node string `json:"node"`
	} `json:"engines"`
	PackageManager string `json:"packageManager"`
}

type NodeProvider struct{}

func (p *NodeProvider) Name() string { return "node" }

func (p *NodeProvider) Detect(a *app.App) (bool, error) {
	return a.HasFile("package.json"), nil
}

func (p *NodeProvider) Plan(a *app.App) (*plan.BuildPlan, error) {
	var pkg packageJSON
	if err := a.ReadJSON("package.json", &pkg); err != nil {
		return nil, err
	}

	nodeVersion := "22"
	if pkg.Engines.Node != "" {
		nodeVersion = cleanVersion(pkg.Engines.Node)
	}

	pm := detectPackageManager(a, &pkg)
	framework := detectFramework(&pkg)

	bp := &plan.BuildPlan{
		Provider:  "node",
		Language:  "Node.js",
		Version:   nodeVersion,
		Framework: framework,
		BaseImage: "node:" + nodeVersion + "-alpine",
		Ports:     []string{"3000"},
		EnvVars:   map[string]string{"NODE_ENV": "production"},
	}

	switch pm {
	case "bun":
		bp.InstallCmd = "bun install --frozen-lockfile"
	case "pnpm":
		bp.InstallCmd = "corepack enable && pnpm install --frozen-lockfile"
	case "yarn":
		bp.InstallCmd = "yarn install --frozen-lockfile"
	default:
		bp.InstallCmd = "npm ci"
	}

	if _, ok := pkg.Scripts["build"]; ok {
		switch pm {
		case "bun":
			bp.BuildCmd = "bun run build"
		case "pnpm":
			bp.BuildCmd = "pnpm run build"
		case "yarn":
			bp.BuildCmd = "yarn build"
		default:
			bp.BuildCmd = "npm run build"
		}
	}

	bp.StartCmd = getStartCommand(&pkg, pm, framework)

	if framework == "nextjs" {
		bp.Ports = []string{"3000"}
		bp.EnvVars["HOSTNAME"] = "0.0.0.0"
	} else if framework == "nuxt" || framework == "remix" {
		bp.Ports = []string{"3000"}
	} else if framework == "vite" || framework == "react" || framework == "vue" {
		bp.Static = true
		bp.Ports = []string{"80"}
	}

	return bp, nil
}

func detectPackageManager(a *app.App, pkg *packageJSON) string {
	if pkg.PackageManager != "" {
		if strings.HasPrefix(pkg.PackageManager, "pnpm") {
			return "pnpm"
		}
		if strings.HasPrefix(pkg.PackageManager, "yarn") {
			return "yarn"
		}
		if strings.HasPrefix(pkg.PackageManager, "bun") {
			return "bun"
		}
	}
	if a.HasFile("bun.lockb") || a.HasFile("bun.lock") {
		return "bun"
	}
	if a.HasFile("pnpm-lock.yaml") {
		return "pnpm"
	}
	if a.HasFile("yarn.lock") {
		return "yarn"
	}
	return "npm"
}

func detectFramework(pkg *packageJSON) string {
	deps := mergeMaps(pkg.Dependencies, pkg.DevDeps)

	if _, ok := deps["next"]; ok {
		return "nextjs"
	}
	if _, ok := deps["nuxt"]; ok {
		return "nuxt"
	}
	if _, ok := deps["@remix-run/node"]; ok {
		return "remix"
	}
	if _, ok := deps["@nestjs/core"]; ok {
		return "nestjs"
	}
	if _, ok := deps["express"]; ok {
		return "express"
	}
	if _, ok := deps["fastify"]; ok {
		return "fastify"
	}
	if _, ok := deps["vite"]; ok {
		if _, ok2 := deps["react"]; ok2 {
			return "react"
		}
		if _, ok2 := deps["vue"]; ok2 {
			return "vue"
		}
		return "vite"
	}
	if _, ok := deps["react-scripts"]; ok {
		return "react"
	}

	return ""
}

func getStartCommand(pkg *packageJSON, pm string, framework string) string {
	if framework == "nextjs" {
		return "node server.js"
	}

	if _, ok := pkg.Scripts["start"]; ok {
		switch pm {
		case "bun":
			return "bun run start"
		case "pnpm":
			return "pnpm start"
		case "yarn":
			return "yarn start"
		default:
			return "npm start"
		}
	}

	return "node index.js"
}

func cleanVersion(v string) string {
	v = strings.TrimLeft(v, "^~>=<")
	parts := strings.Split(v, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return v
}

func mergeMaps(a, b map[string]string) map[string]string {
	merged := make(map[string]string)
	for k, v := range a {
		merged[k] = v
	}
	for k, v := range b {
		merged[k] = v
	}
	return merged
}

// for future use: read raw package.json
func readPackageJSON(a *app.App) (*packageJSON, error) {
	data, err := a.ReadFile("package.json")
	if err != nil {
		return nil, err
	}
	var pkg packageJSON
	if err := json.Unmarshal([]byte(data), &pkg); err != nil {
		return nil, err
	}
	return &pkg, nil
}
