package java

import (
	"strings"

	"github.com/my-paas/core/app"
	"github.com/my-paas/core/plan"
)

type JavaProvider struct{}

func (p *JavaProvider) Name() string { return "java" }

func (p *JavaProvider) Detect(a *app.App) (bool, error) {
	return a.HasFile("pom.xml") || a.HasFile("build.gradle") || a.HasFile("build.gradle.kts"), nil
}

func (p *JavaProvider) Plan(a *app.App) (*plan.BuildPlan, error) {
	javaVersion := "21"
	buildTool := detectBuildTool(a)
	framework := detectJavaFramework(a)

	bp := &plan.BuildPlan{
		Provider:  "java",
		Language:  "Java",
		Version:   javaVersion,
		Framework: framework,
		BaseImage: "eclipse-temurin:" + javaVersion + "-jre-alpine",
		Ports:     []string{"8080"},
		EnvVars:   map[string]string{},
	}

	switch buildTool {
	case "gradle":
		bp.InstallCmd = ""
		bp.BuildCmd = "./gradlew build -x test"
		bp.StartCmd = "java -jar build/libs/*.jar"
	default:
		bp.InstallCmd = ""
		bp.BuildCmd = "./mvnw package -DskipTests"
		bp.StartCmd = "java -jar target/*.jar"
	}

	return bp, nil
}

func detectBuildTool(a *app.App) string {
	if a.HasFile("build.gradle") || a.HasFile("build.gradle.kts") {
		return "gradle"
	}
	return "maven"
}

func detectJavaFramework(a *app.App) string {
	files := []string{"pom.xml", "build.gradle", "build.gradle.kts"}
	for _, f := range files {
		content, err := a.ReadFile(f)
		if err != nil {
			continue
		}
		lower := strings.ToLower(content)
		if strings.Contains(lower, "spring-boot") {
			return "spring-boot"
		}
		if strings.Contains(lower, "quarkus") {
			return "quarkus"
		}
		if strings.Contains(lower, "micronaut") {
			return "micronaut"
		}
	}
	return ""
}
