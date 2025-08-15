package main

import (
	"fmt"
	"os"
)

func printCreateConfigUsage() {
	fmt.Println(`Usage:
  gitbackup create-config [path]
  gitbackup validate-config [path]

  [path] is optional, defaults to ./gitbackup.yml
`)
}

func handleCreateConfig(args []string) {
	path := "gitbackup.yml"
	if len(args) > 0 {
		path = args[0]
	}
	err := WriteSampleYAMLConfig(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Sample config written to %s\n", path)
}

func handleValidateConfig(args []string) {
	path := "gitbackup.yml"
	if len(args) > 0 {
		path = args[0]
	}
	err := validateYAMLConfig(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config validation failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Config is valid.")
}
