package python

import (
	"regexp"
	"strings"

	"github.com/my-paas/core/app"
	"github.com/my-paas/core/plan"
)

type PythonProvider struct{}

func (p *PythonProvider) Name() string { return "python" }

func (p *PythonProvider) Detect(a *app.App) (bool, error) {
	return a.HasFile("requirements.txt") ||
		a.HasFile("pyproject.toml") ||
		a.HasFile("Pipfile") ||
		a.HasFile("main.py") ||
		a.HasFile("app.py"), nil
}

func (p *PythonProvider) Plan(a *app.App) (*plan.BuildPlan, error) {
	pyVersion := "3.12"
	pkgMgr := detectPythonPkgManager(a)
	framework := detectPythonFramework(a)

	if a.HasFile("pyproject.toml") {
		if content, err := a.ReadFile("pyproject.toml"); err == nil {
			if v := parsePythonVersion(content); v != "" {
				pyVersion = v
			}
		}
	}

	bp := &plan.BuildPlan{
		Provider:  "python",
		Language:  "Python",
		Version:   pyVersion,
		Framework: framework,
		BaseImage: "python:" + pyVersion + "-slim",
		Ports:     []string{"8000"},
		EnvVars: map[string]string{
			"PYTHONUNBUFFERED": "1",
			"PYTHONDONTWRITEBYTECODE": "1",
		},
	}

	switch pkgMgr {
	case "uv":
		bp.InstallCmd = "pip install uv && uv sync --frozen"
	case "poetry":
		bp.InstallCmd = "pip install poetry && poetry install --no-dev --no-interaction"
	case "pipenv":
		bp.InstallCmd = "pip install pipenv && pipenv install --deploy --system"
	default:
		bp.InstallCmd = "pip install --no-cache-dir -r requirements.txt"
	}

	bp.StartCmd = getStartCommand(a, framework)

	return bp, nil
}

func detectPythonPkgManager(a *app.App) string {
	if a.HasFile("uv.lock") {
		return "uv"
	}
	if a.HasFile("poetry.lock") {
		return "poetry"
	}
	if a.HasFile("Pipfile.lock") || a.HasFile("Pipfile") {
		return "pipenv"
	}
	return "pip"
}

func detectPythonFramework(a *app.App) string {
	files := []string{"requirements.txt", "pyproject.toml", "Pipfile"}
	for _, f := range files {
		content, err := a.ReadFile(f)
		if err != nil {
			continue
		}
		lower := strings.ToLower(content)
		if strings.Contains(lower, "django") {
			return "django"
		}
		if strings.Contains(lower, "fastapi") {
			return "fastapi"
		}
		if strings.Contains(lower, "flask") {
			return "flask"
		}
		if strings.Contains(lower, "streamlit") {
			return "streamlit"
		}
	}
	return ""
}

func getStartCommand(a *app.App, framework string) string {
	// Check for Procfile
	if a.HasFile("Procfile") {
		if content, err := a.ReadFile("Procfile"); err == nil {
			for _, line := range strings.Split(content, "\n") {
				if strings.HasPrefix(line, "web:") {
					return strings.TrimSpace(strings.TrimPrefix(line, "web:"))
				}
			}
		}
	}

	switch framework {
	case "django":
		return "python manage.py runserver 0.0.0.0:8000"
	case "fastapi":
		return "uvicorn main:app --host 0.0.0.0 --port 8000"
	case "flask":
		return "gunicorn --bind 0.0.0.0:8000 app:app"
	case "streamlit":
		return "streamlit run app.py --server.port 8000 --server.address 0.0.0.0"
	}

	if a.HasFile("main.py") {
		return "python main.py"
	}
	if a.HasFile("app.py") {
		return "python app.py"
	}
	return "python main.py"
}

func parsePythonVersion(pyproject string) string {
	re := regexp.MustCompile(`python_requires\s*=\s*">=?(\d+\.\d+)"`)
	matches := re.FindStringSubmatch(pyproject)
	if len(matches) > 1 {
		return matches[1]
	}

	re2 := regexp.MustCompile(`requires-python\s*=\s*">=?(\d+\.\d+)"`)
	matches2 := re2.FindStringSubmatch(pyproject)
	if len(matches2) > 1 {
		return matches2[1]
	}

	return ""
}
