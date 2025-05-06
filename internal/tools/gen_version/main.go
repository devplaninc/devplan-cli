package main

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/pb/cli"
	"github.com/devplaninc/devplan-cli/internal/utils/git"
	"google.golang.org/protobuf/encoding/protojson"
	"os"
	"path/filepath"
)

func main() {
	version := os.Getenv("DEVPLAN_RELEASE_VERSION")
	if version == "" {
		panic("DEVPLAN_RELEASE_VERSION not set")
	}
	ver := cli.Version_builder{
		ProductionVersion: version,
	}.Build()
	root, err := git.GetRoot()
	if err != nil {
		panic(err)
	}
	m := protojson.MarshalOptions{Indent: "  "}

	// Convert the map to JSON
	jsonData, err := m.Marshal(ver)
	if err != nil {
		panic(fmt.Errorf("failed to marshal version data to JSON [%+v]: %w", ver, err))
	}

	// Write the JSON to version.json at the repository root
	outputPath := filepath.Join(root, "version.json")
	err = os.WriteFile(outputPath, jsonData, 0644)
	if err != nil {
		panic(fmt.Errorf("failed to write version.json: %w", err))
	}

	fmt.Printf("Successfully wrote version information to %s\n", outputPath)
}
