package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ComposerJSON struct {
	Name    string            `json:"name"`
	Require map[string]string `json:"require"`
}

type Plugin struct {
	Name         string
	FolderName   string
	Dependencies []string
	IsExternal   bool
}

type PluginAnalyzer struct {
	PluginsDir        string
	Plugins           map[string]*Plugin
	ShowExternalDeps  bool
	ExternalDepsCount map[string]int
}

func NewPluginAnalyzer(dir string, showExternal bool) *PluginAnalyzer {
	return &PluginAnalyzer{
		PluginsDir:        dir,
		Plugins:           make(map[string]*Plugin),
		ShowExternalDeps:  showExternal,
		ExternalDepsCount: make(map[string]int),
	}
}

func (pa *PluginAnalyzer) ScanPlugins() error {
	entries, err := os.ReadDir(pa.PluginsDir)
	if err != nil {
		return fmt.Errorf("failed to read plugins directory: %w", err)
	}

	// First pass: collect all internal plugins
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		composerPath := filepath.Join(pa.PluginsDir, entry.Name(), "composer.json")
		if _, err := os.Stat(composerPath); os.IsNotExist(err) {
			log.Printf("Warning: No composer.json found in %s", entry.Name())
			continue
		}

		composerData, err := ioutil.ReadFile(composerPath)
		if err != nil {
			log.Printf("Error reading composer.json in %s: %v", entry.Name(), err)
			continue
		}

		var composer ComposerJSON
		if err := json.Unmarshal(composerData, &composer); err != nil {
			log.Printf("Error parsing composer.json in %s: %v", entry.Name(), err)
			continue
		}

		pa.Plugins[composer.Name] = &Plugin{
			Name:       composer.Name,
			FolderName: entry.Name(),
			IsExternal: false,
		}
	}

	// Second pass: collect dependencies
	for _, plugin := range pa.Plugins {
		composerPath := filepath.Join(pa.PluginsDir, plugin.FolderName, "composer.json")
		composerData, _ := ioutil.ReadFile(composerPath)
		var composer ComposerJSON
		json.Unmarshal(composerData, &composer)

		for dep := range composer.Require {
			if strings.Contains(dep, "/") {
				if _, isInternal := pa.Plugins[dep]; isInternal {
					plugin.Dependencies = append(plugin.Dependencies, dep)
				} else if pa.ShowExternalDeps {
					plugin.Dependencies = append(plugin.Dependencies, dep)
					// Create external plugin node if it doesn't exist
					if _, exists := pa.Plugins[dep]; !exists {
						pa.Plugins[dep] = &Plugin{
							Name:       dep,
							FolderName: dep,
							IsExternal: true,
						}
					}
					pa.ExternalDepsCount[dep]++
				} else {
					pa.ExternalDepsCount[dep]++
				}
			}
		}
	}

	return nil
}

func (pa *PluginAnalyzer) GenerateMermaid() string {
	var sb strings.Builder
	sb.WriteString("graph TD\n")

	for _, plugin := range pa.Plugins {
		if plugin.IsExternal && !pa.ShowExternalDeps {
			continue
		}

		for _, dep := range plugin.Dependencies {
			depPlugin := pa.Plugins[dep]
			if depPlugin.IsExternal && !pa.ShowExternalDeps {
				continue
			}
			sb.WriteString(fmt.Sprintf("    \"%s\" --> \"%s\"\n", plugin.FolderName, depPlugin.FolderName))
		}
	}

	return sb.String()
}

func (pa *PluginAnalyzer) GenerateGraphviz(outputPath string) error {
	dotContent := new(strings.Builder)
	dotContent.WriteString("digraph PluginDependencies {\n")
	dotContent.WriteString("    rankdir=TB;\n")
	dotContent.WriteString("    node [shape=box, style=rounded];\n")
	dotContent.WriteString("    edge [color=\"#666666\"];\n")

	// Add nodes
	for _, plugin := range pa.Plugins {
		if plugin.IsExternal && !pa.ShowExternalDeps {
			continue
		}

		style := "rounded,filled"
		fillColor := "#f0f0f0"
		if plugin.IsExternal {
			fillColor = "#ffe0e0"  // Light red for external deps
		}
		
		dotContent.WriteString(fmt.Sprintf("    \"%s\" [label=\"%s\", fillcolor=\"%s\", style=\"%s\"];\n",
			plugin.Name, plugin.FolderName, fillColor, style))
	}

	// Add edges
	for _, plugin := range pa.Plugins {
		if plugin.IsExternal && !pa.ShowExternalDeps {
			continue
		}

		for _, dep := range plugin.Dependencies {
			depPlugin := pa.Plugins[dep]
			if depPlugin.IsExternal && !pa.ShowExternalDeps {
				continue
			}
			dotContent.WriteString(fmt.Sprintf("    \"%s\" -> \"%s\";\n", plugin.Name, dep))
		}
	}

	dotContent.WriteString("}\n")

	// Write to temporary file
	tmpFile, err := os.CreateTemp("", "deps*.dot")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(dotContent.String()); err != nil {
		return fmt.Errorf("failed to write DOT content: %w", err)
	}
	tmpFile.Close()

	// Run dot command to generate SVG
	cmd := exec.Command("dot", "-Tsvg", "-o", outputPath, tmpFile.Name())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run dot command: %w", err)
	}

	return nil
}

func checkGraphvizInstalled() bool {
	_, err := exec.LookPath("dot")
	return err == nil
}

func main() {
	pluginsDir := flag.String("dir", "", "Directory containing plugin folders")
	outputFormat := flag.String("format", "both", "Output format: mermaid, graphviz, or both")
	outputDir := flag.String("output", "output", "Output directory for generated files")
	showExternal := flag.Bool("show-external", false, "Include external dependencies in the graph")
	flag.Parse()

	if *pluginsDir == "" {
		log.Fatal("Please specify plugins directory with -dir flag")
	}

	if !checkGraphvizInstalled() {
		log.Fatal("Graphviz is not installed. Please install it first.")
	}

	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	analyzer := NewPluginAnalyzer(*pluginsDir, *showExternal)
	if err := analyzer.ScanPlugins(); err != nil {
		log.Fatalf("Failed to scan plugins: %v", err)
	}

	if *outputFormat == "mermaid" || *outputFormat == "both" {
		mermaid := analyzer.GenerateMermaid()
		mermaidPath := filepath.Join(*outputDir, "dependencies.mmd")
		if err := ioutil.WriteFile(mermaidPath, []byte(mermaid), 0644); err != nil {
			log.Printf("Failed to write Mermaid file: %v", err)
		}
		fmt.Printf("Mermaid graph saved to %s\n", mermaidPath)
	}

	if *outputFormat == "graphviz" || *outputFormat == "both" {
		svgPath := filepath.Join(*outputDir, "dependencies.svg")
		if err := analyzer.GenerateGraphviz(svgPath); err != nil {
			log.Printf("Failed to generate SVG: %v", err)
		} else {
			fmt.Printf("SVG graph saved to %s\n", svgPath)
		}
	}

	// Print summary
	fmt.Println("\nInternal Dependencies Summary:")
	for _, plugin := range analyzer.Plugins {
		if plugin.IsExternal {
			continue
		}
		if len(plugin.Dependencies) > 0 {
			fmt.Printf("\n%s:\n", plugin.FolderName)
			for _, dep := range plugin.Dependencies {
				depPlugin := analyzer.Plugins[dep]
				if depPlugin.IsExternal {
					fmt.Printf("  ├─ %s (external)\n", dep)
				} else {
					fmt.Printf("  ├─ %s\n", depPlugin.FolderName)
				}
			}
		}
	}

	// Print external dependencies summary
	if len(analyzer.ExternalDepsCount) > 0 {
		fmt.Println("\nExternal Dependencies Summary:")
		for dep, count := range analyzer.ExternalDepsCount {
			fmt.Printf("  %s: used by %d plugin(s)\n", dep, count)
		}
	}
}

