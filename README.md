# sw6-plugin-analyzer

A dependency analyzer for Shopware 6 plugins that generates visual dependency graphs showing relationships between your local plugins.

## About

sw6-plugin-analyzer scans a directory containing Shopware 6 plugins, analyzes their `composer.json` files, and generates visual dependency graphs. It helps you understand the dependency relationships between your plugins by:

- Generating visual dependency graphs in SVG format using Graphviz
- Creating Mermaid.js compatible diagrams
- Providing a summary of internal and external dependencies
- Optionally showing external dependencies in the visualization
- Highlighting dependency relationships with color-coded nodes

## Installation

### Prerequisites

- Go 1.18 or higher
- Graphviz (for SVG generation)

Install Graphviz:
```bash
# Ubuntu/Debian
sudo apt-get install graphviz

# macOS
brew install graphviz

# Windows
choco install graphviz
```

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/sw6-plugin-analyzer.git
cd sw6-plugin-analyzer

# Build the binary
go build -o sw6-plugin-analyzer
```

Optionally, install it to your Go bin directory:
```bash
go install
```

## Usage

Basic usage:
```bash
sw6-plugin-analyzer -dir /path/to/plugins
```

This will:
1. Scan all plugins in the specified directory
2. Generate dependency graphs (both Mermaid and SVG)
3. Print a dependency summary

### Available Options

```bash
-dir string
    Directory containing plugin folders (required)
    
-format string
    Output format: mermaid, graphviz, or both (default "both")
    
-output string
    Output directory for generated files (default "output")
    
-show-external
    Include external dependencies in the graph (default false)
```

### Examples

Only analyze internal plugin dependencies:
```bash
sw6-plugin-analyzer -dir /path/to/plugins
```

Include external dependencies in the visualization:
```bash
sw6-plugin-analyzer -dir /path/to/plugins -show-external
```

Generate only Graphviz SVG:
```bash
sw6-plugin-analyzer -dir /path/to/plugins -format graphviz
```

Generate only Mermaid diagram:
```bash
sw6-plugin-analyzer -dir /path/to/plugins -format mermaid
```

Custom output directory:
```bash
sw6-plugin-analyzer -dir /path/to/plugins -output ./my-graphs
```

### Output

The tool generates:
1. `dependencies.svg` - Visual graph in SVG format
2. `dependencies.mmd` - Mermaid.js compatible diagram
3. Console output with dependency summary

The SVG graph uses color coding:
- Light gray: Internal plugins
- Light red: External dependencies (when `-show-external` is used)

## License

MIT License

