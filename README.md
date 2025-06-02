# go-scihub-mcp

[中文文档](README_cn.md) | [English](README.md)

A Sci-Hub mirror management and file download tool written in Go, with MCP (Model Context Protocol) compatible API support.

## Features

- 📋 Maintain available Sci-Hub mirror list
- 🔄 Automatic mirror availability detection and updates
- 🌐 SOCKS5 proxy support
- 📁 File download and caching
- 🔗 MCP-compatible SSE API service
- ⚙️ Flexible command-line configuration with clear priority system
- 🐳 Docker support with docker-compose

## Installation

### Build from Source

```bash
git clone https://github.com/jifanchn/go-scihub-mcp.git
cd go-scihub-mcp
go mod tidy
go build -o scihub-mcp ./cmd/scihub-mcp
```

### Docker

```bash
# Using docker-compose (recommended)
docker-compose up -d

# Or build and run manually
docker build -t scihub-mcp .
docker run -d -p 8088:8088 --name scihub-mcp scihub-mcp
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
  transport: "sse"  # SSE mode only
  sse_path: "/sse"
  
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

### 4. MCP Protocol Service Mode (SSE)

```bash
# Start MCP protocol server with SSE transport
./scihub-mcp mcp

# Start with custom port
./scihub-mcp --mcp-port 9090 mcp

# Start with proxy enabled
./scihub-mcp --proxy-enabled --proxy-host 127.0.0.1 --proxy-port 3080 mcp

# Use custom configuration file
./scihub-mcp --config configs/config.yaml mcp
```

**SSE Mode Endpoints:**
- SSE Stream: `http://localhost:8080/sse`
- Message Endpoint: `http://localhost:8080/message`
- Health Check: `http://localhost:8080/health`

**SSE Transport Characteristics:**
- Real-time bidirectional communication over HTTP
- WebSocket-like functionality using standard HTTP
- Better for web applications and remote clients
- Supports concurrent connections
- Built-in reconnection and error handling
- Compatible with HTTP/2 and modern web infrastructure

### 5. Mirror Status Check

```bash
# Check all mirror status
./scihub-mcp status

# Check with proxy enabled
./scihub-mcp --proxy-enabled status
```

### 6. Docker Deployment

```bash
# Using docker-compose (recommended)
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f

# Stop service
docker-compose down
```

The Docker service will:
- Run on port 8088 (mapped to host)
- Use SOCKS5 proxy at `127.0.0.1:3080` on the host
- Persist cache data in `./cache` directory
- Automatically restart unless stopped

## MCP Server Configuration

To use the MCP protocol server with AI tools like Cursor, Claude Desktop, or other MCP clients, configure the server using SSE transport mode.

### Cursor AI Configuration

Create or edit your MCP configuration file (usually `~/.cursor/mcp_servers.json`):

```json
{
  "mcpServers": {
    "scihub-mcp": {
      "description": "SciHub MCP Server - SSE Mode",
      "url": "http://localhost:8080/sse"
    }
  }
}
```

For Docker deployment:
```json
{
  "mcpServers": {
    "scihub-mcp": {
      "description": "SciHub MCP Server - Docker",
      "url": "http://localhost:8088/sse"
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
      "url": "http://localhost:8080/sse"
    }
  }
}
```

For Docker deployment:
```json
{
  "mcpServers": {
    "scihub-mcp": {
      "url": "http://localhost:8088/sse"
    }
  }
}
```

## MCP Tools and Resources

The MCP server provides the following tools and resources:

### Available Tools

1. **download_paper**: Download scientific paper PDF files
   - Parameters: `doi`, `url`, `title`, `output_path`
   
2. **check_mirror_status**: Check availability status of Sci-Hub mirrors
   
3. **test_mirror**: Test availability of a specific Sci-Hub mirror
   - Parameters: `mirror_url`
   
4. **list_available_mirrors**: Get list of currently available Sci-Hub mirrors

### Available Resources

1. **scihub://cache**: List of cached paper files
2. **scihub://mirrors/status**: Real-time status of all Sci-Hub mirrors  
3. **scihub://papers/{filename}**: Access cached paper PDF files

### Testing MCP Connection

You can test the SSE server:

```bash
# Check health endpoint
curl http://localhost:8080/health

# For Docker deployment
curl http://localhost:8088/health
```

## API Endpoints

### HTTP API Endpoints

- `GET /health` - Health check
- `GET /status` - Mirror status
- `GET /mirrors` - Available mirrors list
- `POST /download` - Download paper
- `GET /cache` - List cached files

### MCP SSE Endpoints

- `GET /sse` - SSE stream for MCP communication
- `POST /message` - Send messages to MCP server
- `GET /health` - Health check

## Development

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Build with Docker
make docker
```

### Testing

```bash
# Run unit tests
make test

# Run with coverage
make coverage
```

### Development Mode

```bash
# Start HTTP API server
make dev

# Start MCP server
make dev-mcp
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Sci-Hub](https://sci-hub.se/) for providing access to scientific literature
- [MCP](https://github.com/modelcontextprotocol) for the Model Context Protocol specification
- All contributors and users of this project

## Disclaimer

This tool is for educational and research purposes only. Users are responsible for ensuring compliance with applicable laws and regulations in their jurisdiction.