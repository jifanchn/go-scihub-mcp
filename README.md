# go-scihub-mcp

[中文文档](README_cn.md) | [English](README.md)

A Sci-Hub mirror management and file download tool written in Go, with MCP (Model Context Protocol) compatible API support.

## Features

- 📋 Maintain available Sci-Hub mirror list
- 🔄 Automatic mirror availability detection and updates
- 🌐 SOCKS5 proxy support
- 📁 File download and caching
- 🔗 MCP-compatible HTTP API service
- ⚙️ Flexible command-line configuration with clear priority system

## Installation

### Build from Source

```bash
git clone https://github.com/jifanchn/go-scihub-mcp.git
cd go-scihub-mcp
go mod tidy
go build -o scihub-mcp ./cmd/scihub-mcp
```

### Binary Download

Download the binary for your platform from the [Releases](https://github.com/jifanchn/go-scihub-mcp/releases) page.

## Configuration

### Configuration Priority

The configuration follows a clear priority system:
**Command-line arguments > Configuration file > Default values**

This means global options apply to all commands and can be overridden by command-specific options where applicable.

### Configuration File

The program supports YAML configuration files with the following default search paths:
- `./config.yaml`
- `~/.config/scihub-mcp/config.yaml`
- `/etc/scihub-mcp/config.yaml`

Configuration file example:

```yaml
# Sci-Hub mirror list
mirrors:
  - "https://sci-hub.ru"
  - "https://sci-hub.se"
  - "https://sci-hub.st"
  - "https://sci-hub.box"
  - "https://sci-hub.red"
  - "https://sci-hub.al"
  - "https://sci-hub.ee"
  - "https://sci-hub.lu"
  - "https://sci-hub.ren"
  - "https://sci-hub.shop"
  - "https://sci-hub.vg"

# Proxy configuration
proxy:
  enabled: false
  type: "socks5"
  host: "127.0.0.1"
  port: 3080
  username: ""
  password: ""

# Health check configuration
health_check:
  interval: "30m"  # Check interval
  timeout: "10s"   # Request timeout

# MCP service configuration
mcp:
  port: 8080
  host: "0.0.0.0"
  
# Download configuration
download:
  cache_dir: "./cache"
  max_retries: 3
  timeout: "60s"
```

### Command Line Arguments

Global options apply to all commands and can be specified before the command:

```bash
# Global proxy settings
scihub-mcp --proxy-enabled --proxy-host 127.0.0.1 --proxy-port 3080 [command]

# Global configuration file
scihub-mcp --config /path/to/config.yaml [command]

# Global MCP server settings
scihub-mcp --mcp-host 0.0.0.0 --mcp-port 9090 [command]
```

## Usage

### 1. Basic Running

```bash
# Run with default configuration (HTTP API service)
./scihub-mcp

# Run with global proxy enabled (HTTP API service)
./scihub-mcp --proxy-enabled --proxy-host 127.0.0.1 --proxy-port 3080

# Run with custom configuration file
./scihub-mcp --config /path/to/config.yaml
```

### 2. File Download Command

```bash
# Download file by DOI
./scihub-mcp fetch --doi "10.1038/nature12373"

# Download with global proxy
./scihub-mcp --proxy-enabled fetch --doi "10.1038/nature12373"

# Download with custom output path
./scihub-mcp fetch --doi "10.1038/nature12373" --output "./papers/nature12373.pdf"

# Download by URL
./scihub-mcp fetch --url "https://www.nature.com/articles/nature12373"
```

### 3. HTTP API Service Mode

```bash
# Start HTTP API service (default port 8080)
./scihub-mcp api

# Start with global proxy enabled
./scihub-mcp --proxy-enabled api

# Start on custom port with global settings
./scihub-mcp --mcp-port 9090 api

# Override port for this instance only
./scihub-mcp api --port 9090

# Combine global proxy with custom port
./scihub-mcp --proxy-enabled api --port 9090
```

### 4. MCP Protocol Service Mode

```bash
# Start MCP protocol server (STDIO communication)
./scihub-mcp mcp

# Start with global proxy enabled
./scihub-mcp --proxy-enabled mcp

# MCP protocol communicates through STDIN/STDOUT
# Example tool call:
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}}' | ./scihub-mcp mcp

# Example resource access:
echo '{"jsonrpc": "2.0", "id": 2, "method": "resources/read", "params": {"uri": "scihub://cache"}}' | ./scihub-mcp mcp
```

### 5. Mirror Status Check

```bash
# Check all mirror status
./scihub-mcp status

# Check with proxy enabled
./scihub-mcp --proxy-enabled status

# Test specific mirror
./scihub-mcp test --mirror "https://sci-hub.se"
```

## MCP Server Configuration

To use the MCP protocol server with AI tools like Cursor, Claude Desktop, or other MCP clients, you need to configure the server in the client's settings.

### Cursor AI Configuration

1. **Open Cursor Settings**:
   - Press `Cmd/Ctrl + ,` to open settings
   - Go to "Extensions" -> "MCP Servers" or search for "MCP"

2. **Add SciHub-MCP Server**:
   Create or edit your MCP configuration file (usually `~/.cursor/mcp_servers.json`):
   ```json
   {
     "mcpServers": {
       "scihub-mcp": {
         "command": "/path/to/scihub-mcp",
         "args": ["--proxy-enabled", "--proxy-host", "127.0.0.1", "--proxy-port", "3080", "mcp"],
         "env": {},
         "cwd": "/path/to/working/directory"
       }
     }
   }
   ```

   Or without proxy:
   ```json
   {
     "mcpServers": {
       "scihub-mcp": {
         "command": "/path/to/scihub-mcp",
         "args": ["mcp"],
         "env": {},
         "cwd": "/path/to/working/directory"
       }
     }
   }
   ```

### Claude Desktop Configuration

1. **Locate Configuration File**:
   - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - Windows: `%APPDATA%\Claude\claude_desktop_config.json`
   - Linux: `~/.config/Claude/claude_desktop_config.json`

2. **Add Server Configuration**:
   ```json
   {
     "mcpServers": {
       "scihub-mcp": {
         "command": "/path/to/scihub-mcp",
         "args": ["mcp"]
       }
     }
   }
   ```

   With proxy support:
   ```json
   {
     "mcpServers": {
       "scihub-mcp": {
         "command": "/path/to/scihub-mcp",
         "args": ["--proxy-enabled", "--proxy-host", "127.0.0.1", "--proxy-port", "3080", "mcp"]
       }
     }
   }
   ```

### Generic MCP Client Configuration

For other MCP clients, use these settings:

- **Server Command**: `/path/to/scihub-mcp mcp`
- **Communication**: STDIO (standard input/output)
- **Protocol**: JSON-RPC 2.0 with MCP extensions
- **Environment Variables**: None required
- **Working Directory**: Directory containing the executable

### Testing MCP Connection

You can test the MCP server manually:

```bash
# Test server startup and tool listing
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}}' | ./scihub-mcp mcp

# Expected response should show available tools like:
# {"jsonrpc":"2.0","id":1,"result":{"tools":[{"name":"download_paper",...}]}}
```

### Configuration Examples

For your convenience, we provide example configuration files in the `configs/` directory:

- [`configs/cursor_mcp_config.json`](configs/cursor_mcp_config.json) - Basic Cursor configuration
- [`configs/cursor_mcp_config_with_proxy.json`](configs/cursor_mcp_config_with_proxy.json) - Cursor with proxy
- [`configs/claude_desktop_config.json`](configs/claude_desktop_config.json) - Claude Desktop configuration

**Basic Configuration** (no proxy):
```json
{
  "mcpServers": {
    "scihub-mcp": {
      "command": "/usr/local/bin/scihub-mcp",
      "args": ["mcp"]
    }
  }
}
```

**With Proxy** (for users in restricted networks):
```json
{
  "mcpServers": {
    "scihub-mcp": {
      "command": "/usr/local/bin/scihub-mcp",
      "args": ["--proxy-enabled", "--proxy-host", "127.0.0.1", "--proxy-port", "3080", "mcp"]
    }
  }
}
```

**With Custom Config File**:
```json
{
  "mcpServers": {
    "scihub-mcp": {
      "command": "/usr/local/bin/scihub-mcp",
      "args": ["--config", "/path/to/custom-config.yaml", "mcp"],
      "cwd": "/path/to/cache/directory"
    }
  }
}
```

### Troubleshooting MCP Setup

1. **Server Not Starting**:
   - Check if the binary path is correct
   - Ensure the binary has execution permissions
   - Verify the working directory exists

2. **Proxy Issues**:
   - Test proxy connectivity manually
   - Check proxy server is running on specified host/port
   - Try without proxy first to isolate issues

3. **Tool/Resource Not Found**:
   - Restart the MCP client after configuration changes
   - Check server logs for any startup errors
   - Verify configuration JSON syntax is valid

4. **Permission Errors**:
   - Ensure cache directory is writable
   - Check file permissions for the binary and config files

## Service Types

This tool provides two distinct service modes:

### HTTP API Service (`api`)
- **Type**: REST API over HTTP
- **Communication**: Standard HTTP requests/responses
- **Use Case**: Web applications, curl commands, browser access
- **Endpoints**: `/health`, `/fetch`, `/download/{filename}`, `/mirrors`, `/status`
- **Format**: JSON responses with standard HTTP status codes

**Example HTTP API Usage:**
```bash
# Health check
curl http://localhost:8080/health

# Download paper
curl -X POST http://localhost:8080/fetch \
  -H "Content-Type: application/json" \
  -d '{"doi": "10.1038/nature12373"}'

# Download file directly
curl http://localhost:8080/download/paper.pdf --output paper.pdf
```

### MCP Protocol Service (`mcp`)
- **Type**: Model Context Protocol over STDIO
- **Communication**: JSON-RPC 2.0 messages via STDIN/STDOUT
- **Use Case**: LLM applications, MCP clients, Cursor AI integration
- **Features**: Tools, Resources, and Templates
- **Format**: Standard MCP protocol messages

**Available MCP Tools:**
- `download_paper` - Download scientific papers
- `check_mirror_status` - Check mirror availability
- `test_mirror` - Test specific mirror
- `list_available_mirrors` - List available mirrors

**Available MCP Resources:**
- `scihub://cache` - List cached papers (JSON)
- `scihub://mirrors/status` - Mirror status information (JSON)
- `scihub://papers/{filename}` - Access specific paper files (PDF)

**Example MCP Usage:**
```bash
# List available tools
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}}' | ./scihub-mcp mcp

# Call download tool
echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "download_paper", "arguments": {"doi": "10.1038/nature12373"}}}' | ./scihub-mcp mcp

# List resources
echo '{"jsonrpc": "2.0", "id": 3, "method": "resources/list", "params": {}}' | ./scihub-mcp mcp

# Read cache resource
echo '{"jsonrpc": "2.0", "id": 4, "method": "resources/read", "params": {"uri": "scihub://cache"}}' | ./scihub-mcp mcp
```

## API Endpoints

When running in HTTP API mode (`api`), the following HTTP endpoints are available:

### GET /health
Check service health status

```bash
curl http://localhost:8080/health
```

### POST /fetch
Download paper files with JSON response

```bash
curl -X POST http://localhost:8080/fetch \
  -H "Content-Type: application/json" \
  -d '{"doi": "10.1038/nature12373"}'
```

Request format:
```json
{
  "doi": "10.1038/nature12373",          // DOI (optional)
  "url": "https://example.com/paper",    // Original URL (optional)
  "title": "Paper Title"                 // Paper title (optional)
}
```

Response format:
```json
{
  "success": true,
  "message": "File downloaded successfully",
  "data": {
    "filename": "nature12373.pdf",
    "size": 1024000,
    "mirror_used": "https://sci-hub.se",
    "download_url": "https://sci-hub.se/downloads/...",
    "cached": false,
    "download_link": "/download/nature12373.pdf"
  }
}
```

### GET /download/{filename}
Download file by filename

```bash
curl http://localhost:8080/download/nature12373.pdf --output paper.pdf
```

### GET /mirrors
Get current available mirror status

```bash
curl http://localhost:8080/mirrors
```

### GET /status
Get system status

```bash
curl http://localhost:8080/status
```

## How It Works

1. **Mirror Management**: The program loads the configured mirror list at startup, with background goroutines periodically checking each mirror's availability
2. **Smart Selection**: Automatically selects the fastest available mirror when downloading
3. **Proxy Support**: Supports SOCKS5 proxy for use in network-restricted environments
4. **Caching Mechanism**: Downloaded files are cached locally to avoid duplicate downloads
5. **Error Handling**: Features retry mechanisms and detailed error logging
6. **Flexible Configuration**: Clear priority system ensures predictable behavior

## Development

### Project Structure

```
go-scihub-mcp/
├── cmd/scihub-mcp/     # Main program entry
├── internal/           # Internal packages
│   ├── config/         # Configuration management
│   ├── mirror/         # Mirror management
│   ├── downloader/     # Downloader
│   ├── mcp/           # MCP-compatible HTTP API service
│   └── proxy/         # Proxy management
├── pkg/               # Public packages
├── configs/           # Example configuration files
├── docs/             # Documentation
└── README.md
```

### Running Tests

```bash
go test ./...
```

### Building Release Version

```bash
# Build for current platform
make build

# Cross compile
make build-all

# Create release package
make release
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contributing

Issues and Pull Requests are welcome!

## Disclaimer

This tool is intended for academic research purposes only. Please ensure your usage complies with local laws and regulations and the terms of service of relevant websites.