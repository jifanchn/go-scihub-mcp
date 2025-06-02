package mcpserver

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jifanchn/go-scihub-mcp/internal/downloader"
	"github.com/jifanchn/go-scihub-mcp/internal/mirror"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// TransportMode 传输模式
type TransportMode string

const (
	TransportSSE TransportMode = "sse"
)

// MCPServer 真正的MCP协议服务器
type MCPServer struct {
	downloader    *downloader.Downloader
	mirrorManager *mirror.MirrorManager
	server        *server.MCPServer
	transport     TransportMode
	host          string
	port          int
	ssePath       string
}

// NewMCPServer 创建新的MCP服务器
func NewMCPServer(d *downloader.Downloader, mm *mirror.MirrorManager, transport TransportMode, host string, port int, ssePath string) *MCPServer {
	// 创建MCP服务器 - 只启用基本工具功能，模仿ScholarAI的配置
	s := server.NewMCPServer(
		"SciHub-MCP",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	mcpServer := &MCPServer{
		downloader:    d,
		mirrorManager: mm,
		server:        s,
		transport:     transport,
		host:          host,
		port:          port,
		ssePath:       ssePath,
	}

	// 注册工具和资源
	mcpServer.registerTools()
	mcpServer.registerResources()

	return mcpServer
}

// registerTools 注册MCP工具
func (m *MCPServer) registerTools() {
	// 下载论文工具
	downloadTool := mcp.NewTool("download_paper",
		mcp.WithDescription("Download scientific paper PDF files"),
		mcp.WithString("doi", mcp.Description("DOI identifier of the paper")),
		mcp.WithString("url", mcp.Description("Original URL of the paper")),
		mcp.WithString("title", mcp.Description("Title of the paper")),
		mcp.WithString("output_path", mcp.Description("Output file path (optional)")),
	)

	m.server.AddTool(downloadTool, m.handleDownloadPaper)

	// 检查镜像状态工具
	statusTool := mcp.NewTool("check_mirror_status",
		mcp.WithDescription("Check availability status of Sci-Hub mirrors"),
	)

	m.server.AddTool(statusTool, m.handleCheckMirrorStatus)

	// 测试特定镜像工具
	testMirrorTool := mcp.NewTool("test_mirror",
		mcp.WithDescription("Test availability of a specific Sci-Hub mirror"),
		mcp.WithString("mirror_url", mcp.Required(), mcp.Description("URL of the mirror to test")),
	)

	m.server.AddTool(testMirrorTool, m.handleTestMirror)

	// 获取可用镜像列表工具
	listMirrorsTool := mcp.NewTool("list_available_mirrors",
		mcp.WithDescription("Get list of currently available Sci-Hub mirrors"),
	)

	m.server.AddTool(listMirrorsTool, m.handleListAvailableMirrors)
}

// registerResources 注册MCP资源
func (m *MCPServer) registerResources() {
	// 缓存目录资源
	cacheResource := mcp.NewResource(
		"scihub://cache",
		"Sci-Hub Cache Directory",
		mcp.WithResourceDescription("List of cached paper files"),
		mcp.WithMIMEType("application/json"),
	)

	m.server.AddResource(cacheResource, m.handleCacheResource)

	// 镜像状态资源
	mirrorResource := mcp.NewResource(
		"scihub://mirrors/status",
		"Mirror Status",
		mcp.WithResourceDescription("Real-time status of all Sci-Hub mirrors"),
		mcp.WithMIMEType("application/json"),
	)

	m.server.AddResource(mirrorResource, m.handleMirrorStatusResource)

	// 动态论文文件资源模板
	paperTemplate := mcp.NewResourceTemplate(
		"scihub://papers/{filename}",
		"Paper Files",
		mcp.WithTemplateDescription("Access cached paper PDF files"),
		mcp.WithTemplateMIMEType("application/pdf"),
	)

	m.server.AddResourceTemplate(paperTemplate, m.handlePaperResource)
}

// handleDownloadPaper 处理下载论文工具
func (m *MCPServer) handleDownloadPaper(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	doi := request.GetString("doi", "")
	url := request.GetString("url", "")
	title := request.GetString("title", "")
	outputPath := request.GetString("output_path", "")

	if doi == "" && url == "" {
		return mcp.NewToolResultError("Must provide either DOI or URL"), nil
	}

	// 创建下载请求
	req := &downloader.DownloadRequest{
		DOI:   doi,
		URL:   url,
		Title: title,
	}

	// 执行下载
	result, err := m.downloader.Download(req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Download failed: %v", err)), nil
	}

	// 如果指定了输出路径，复制文件
	if outputPath != "" {
		if err := copyFile(result.FilePath, outputPath); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to copy file: %v", err)), nil
		}
		result.FilePath = outputPath
	}

	// 准备响应数据
	responseText := fmt.Sprintf(`Download completed!

File information:
- Filename: %s
- File size: %d bytes
- File path: %s
- Mirror used: %s
- From cache: %v

Status: %s
`, result.Filename, result.Size, result.FilePath, result.MirrorUsed, result.Cached, result.Message)

	return mcp.NewToolResultText(responseText), nil
}

// handleCheckMirrorStatus 处理检查镜像状态工具
func (m *MCPServer) handleCheckMirrorStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	status := m.mirrorManager.GetMirrorStatus()
	count := m.mirrorManager.GetMirrorCount()

	responseText := fmt.Sprintf(`Mirror Status Report:

Total: %d
Online: %d
Offline: %d
Slow: %d
Unknown: %d

Detailed status:
`, count["total"], count["online"], count["offline"], count["slow"], count["unknown"])

	for url, mirror := range status {
		responseText += fmt.Sprintf("- %s: %s (%v)\n", url, mirror.Status, mirror.ResponseTime)
		if mirror.ErrorMessage != "" {
			responseText += fmt.Sprintf("  Error: %s\n", mirror.ErrorMessage)
		}
	}

	return mcp.NewToolResultText(responseText), nil
}

// handleTestMirror 处理测试镜像工具
func (m *MCPServer) handleTestMirror(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	mirrorURL, err := request.RequireString("mirror_url")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mirror, err := m.mirrorManager.TestMirror(mirrorURL)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Test failed: %v", err)), nil
	}

	responseText := fmt.Sprintf(`Mirror Test Result:

URL: %s
Status: %s
Response time: %v
`, mirrorURL, mirror.Status, mirror.ResponseTime)

	if mirror.ErrorMessage != "" {
		responseText += fmt.Sprintf("Error message: %s\n", mirror.ErrorMessage)
	}

	return mcp.NewToolResultText(responseText), nil
}

// handleListAvailableMirrors 处理获取可用镜像列表工具
func (m *MCPServer) handleListAvailableMirrors(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	available := m.mirrorManager.GetAvailableMirrors()

	responseText := fmt.Sprintf("Currently available mirrors (%d):\n\n", len(available))
	for i, mirror := range available {
		responseText += fmt.Sprintf("%d. %s\n", i+1, mirror)
	}

	if len(available) == 0 {
		responseText = "No mirrors currently available. Please check network connection or wait for health check to complete."
	}

	return mcp.NewToolResultText(responseText), nil
}

// handleCacheResource 处理缓存资源
func (m *MCPServer) handleCacheResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	cacheDir := "cache" // 应该从配置获取

	files, err := os.ReadDir(cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      "scihub://cache",
					MIMEType: "application/json",
					Text:     `{"files": [], "message": "Cache directory does not exist"}`,
				},
			}, nil
		}
		return nil, err
	}

	var fileList []map[string]interface{}
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".pdf" {
			info, _ := file.Info()
			fileList = append(fileList, map[string]interface{}{
				"name":         file.Name(),
				"size":         info.Size(),
				"modified":     info.ModTime(),
				"resource_uri": fmt.Sprintf("scihub://papers/%s", file.Name()),
			})
		}
	}

	responseJSON := fmt.Sprintf(`{
  "files": %s,
  "count": %d,
  "cache_directory": "%s"
}`, formatFileList(fileList), len(fileList), cacheDir)

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "scihub://cache",
			MIMEType: "application/json",
			Text:     responseJSON,
		},
	}, nil
}

// handleMirrorStatusResource 处理镜像状态资源
func (m *MCPServer) handleMirrorStatusResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	count := m.mirrorManager.GetMirrorCount()

	// 简化的JSON格式化
	responseJSON := `{
  "mirrors": {},
  "summary": ` + fmt.Sprintf(`{
    "total": %d,
    "online": %d,
    "offline": %d,
    "slow": %d,
    "unknown": %d
  }`, count["total"], count["online"], count["offline"], count["slow"], count["unknown"]) + `,
  "timestamp": {
    "unix": "1234567890",
    "format": "2024-01-01T00:00:00Z"
  }
}`

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "scihub://mirrors/status",
			MIMEType: "application/json",
			Text:     responseJSON,
		},
	}, nil
}

// handlePaperResource 处理论文文件资源
func (m *MCPServer) handlePaperResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	// 从URI提取文件名
	uri := request.Params.URI
	filename := filepath.Base(uri[len("scihub://papers/"):])

	filePath := filepath.Join("cache", filename)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", filename)
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	return []mcp.ResourceContents{
		mcp.BlobResourceContents{
			URI:      uri,
			MIMEType: "application/pdf",
			Blob:     string(content),
		},
	}, nil
}

// Start 启动MCP服务器
func (m *MCPServer) Start() error {
	switch m.transport {
	case TransportSSE:
		log.Printf("Starting MCP protocol server with SSE transport on %s:%d%s...", m.host, m.port, m.ssePath)
		return m.startSSEServer()
	default:
		return fmt.Errorf("unsupported transport mode: %s", m.transport)
	}
}

// startSSEServer 启动SSE服务器
func (m *MCPServer) startSSEServer() error {
	// 创建SSE服务器
	sseServer := server.NewSSEServer(m.server,
		server.WithBaseURL(fmt.Sprintf("http://%s:%d", m.host, m.port)),
		server.WithSSEEndpoint(m.ssePath),
		server.WithMessageEndpoint("/message"),
	)

	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	log.Printf("SSE server listening on %s", addr)
	log.Printf("SSE endpoint: http://%s%s", addr, m.ssePath)
	log.Printf("Message endpoint: http://%s/message", addr)
	log.Printf("Health check: http://%s/health", addr)

	// 启动SSE服务器
	return sseServer.Start(addr)
}

// 辅助函数
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = dstFile.ReadFrom(srcFile)
	return err
}

func formatFileList(files []map[string]interface{}) string {
	if len(files) == 0 {
		return "[]"
	}

	result := "[\n"
	for i, file := range files {
		result += fmt.Sprintf(`    {
      "name": "%s",
      "size": %v,
      "resource_uri": "%s"
    }`, file["name"], file["size"], file["resource_uri"])

		if i < len(files)-1 {
			result += ","
		}
		result += "\n"
	}
	result += "  ]"
	return result
}
