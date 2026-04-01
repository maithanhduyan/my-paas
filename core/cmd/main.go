package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/my-paas/core"
)

const usage = `mypaas-core - Auto-detect & generate Dockerfiles

Usage:
  mypaas-core detect <path>       Detect language/framework, output JSON plan
  mypaas-core dockerfile <path>   Generate Dockerfile for the source directory
  mypaas-core help                Show this help message
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "detect":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: mypaas-core detect <path>")
			os.Exit(1)
		}
		result := core.Detect(os.Args[2])
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
		if !result.Success {
			os.Exit(1)
		}

	case "dockerfile":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: mypaas-core dockerfile <path>")
			os.Exit(1)
		}
		dockerfile, err := core.GenerateDockerfile(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		fmt.Print(dockerfile)

	case "help", "--help", "-h":
		fmt.Print(usage)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		fmt.Print(usage)
		os.Exit(1)
	}
}
