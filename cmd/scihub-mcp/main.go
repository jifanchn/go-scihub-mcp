package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jifanchn/go-scihub-mcp/internal/config"
	"github.com/jifanchn/go-scihub-mcp/internal/downloader"
	"github.com/jifanchn/go-scihub-mcp/internal/mcp"
	"github.com/jifanchn/go-scihub-mcp/internal/mcpserver"
	"github.com/jifanchn/go-scihub-mcp/internal/mirror"
	"github.com/jifanchn/go-scihub-mcp/internal/proxy"
)

var (
	version = "1.0.0"
	commit  = "unknown"
	date    = "unknown"
)

// GlobalFlags 全局标志
type GlobalFlags struct {
	ConfigPath     string
	ProxyEnabled   bool
	ProxyHost      string
	ProxyPort      int
	HealthInterval time.Duration
	MCPHost        string
	MCPPort        int
	ShowVersion    bool
	ShowHelp       bool
}

func main() {
	// 解析全局参数
	flags := &GlobalFlags{}
	flag.StringVar(&flags.ConfigPath, "config", "", "Configuration file path")
	flag.BoolVar(&flags.ProxyEnabled, "proxy-enabled", false, "Enable proxy")
	flag.StringVar(&flags.ProxyHost, "proxy-host", "", "Proxy host")
	flag.IntVar(&flags.ProxyPort, "proxy-port", 0, "Proxy port")
	flag.DurationVar(&flags.HealthInterval, "health-interval", 0, "Health check interval")
	flag.StringVar(&flags.MCPHost, "mcp-host", "", "MCP service host")
	flag.IntVar(&flags.MCPPort, "mcp-port", 0, "MCP service port")
	flag.BoolVar(&flags.ShowVersion, "version", false, "Show version information")
	flag.BoolVar(&flags.ShowHelp, "help", false, "Show help information")
	flag.Parse()

	if flags.ShowVersion {
		printVersion()
		return
	}

	if flags.ShowHelp {
		printHelp()
		return
	}

	// 获取子命令
	args := flag.Args()
	if len(args) == 0 {
		// 默认启动服务模式
		runService(flags)
		return
	}

	command := args[0]
	switch command {
	case "fetch":
		runFetch(args[1:], flags)
	case "api":
		runHTTPAPI(args[1:], flags)
	case "mcp":
		runMCPServer(args[1:], flags)
	case "status":
		runStatus(args[1:], flags)
	case "test":
		runTest(args[1:], flags)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}

// runService 运行服务模式（默认模式）
func runService(flags *GlobalFlags) {
	log.Println("Starting SciHub-MCP service...")

	// 加载配置
	cfg, err := loadConfigWithFlags(flags)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 创建组件
	_, mm, _, server, err := createComponents(cfg)
	if err != nil {
		log.Fatalf("Failed to create components: %v", err)
	}

	// 启动镜像管理器
	mm.Start()
	defer mm.Stop()

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动MCP服务器
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start MCP server: %v", err)
		}
	}()

	// 等待信号
	<-sigChan
	log.Println("Received stop signal, shutting down service...")
}

// runFetch 运行文件下载命令
func runFetch(args []string, flags *GlobalFlags) {
	fetchFlags := flag.NewFlagSet("fetch", flag.ExitOnError)
	doi := fetchFlags.String("doi", "", "Paper DOI")
	url := fetchFlags.String("url", "", "Paper URL")
	title := fetchFlags.String("title", "", "Paper title")
	output := fetchFlags.String("output", "", "Output file path")

	fetchFlags.Parse(args)

	if *doi == "" && *url == "" {
		fmt.Println("Must specify either --doi or --url")
		fetchFlags.Usage()
		os.Exit(1)
	}

	// 加载配置
	cfg, err := loadConfigWithFlags(flags)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	_, mm, dl, _, err := createComponents(cfg)
	if err != nil {
		log.Fatalf("Failed to create components: %v", err)
	}

	fmt.Println("Checking mirror availability...")
	// 启动镜像管理器进行快速健康检查
	mm.Start()
	defer mm.Stop()

	// 等待一次健康检查完成
	time.Sleep(3 * time.Second)

	// 显示可用镜像数量
	count := mm.GetMirrorCount()
	available := mm.GetAvailableMirrors()
	fmt.Printf("Found %d online mirrors out of %d total mirrors\n", len(available), count["total"])

	if len(available) == 0 {
		fmt.Println("Warning: No mirrors available, download may fail")
	}

	// 执行下载
	req := &downloader.DownloadRequest{
		DOI:   *doi,
		URL:   *url,
		Title: *title,
	}

	fmt.Printf("Downloading: DOI=%s, URL=%s\n", *doi, *url)
	result, err := dl.Download(req)
	if err != nil {
		log.Fatalf("Download failed: %v", err)
	}

	fmt.Printf("Download successful: %s (size: %d bytes)\n", result.Filename, result.Size)
	if result.Cached {
		fmt.Println("File from cache")
	} else {
		fmt.Printf("Used mirror: %s\n", result.MirrorUsed)
	}

	// 如果指定了输出路径，复制文件
	if *output != "" {
		if err := copyFile(result.FilePath, *output); err != nil {
			log.Fatalf("Failed to copy file: %v", err)
		}
		fmt.Printf("File saved to: %s\n", *output)
	} else {
		fmt.Printf("File path: %s\n", result.FilePath)
	}
}

// runHTTPAPI 运行HTTP API服务命令
func runHTTPAPI(args []string, flags *GlobalFlags) {
	apiFlags := flag.NewFlagSet("api", flag.ExitOnError)
	port := apiFlags.Int("port", 0, "Service port")
	host := apiFlags.String("host", "", "Service host")

	apiFlags.Parse(args)

	// 覆盖全局配置
	if *port != 0 {
		flags.MCPPort = *port
	}
	if *host != "" {
		flags.MCPHost = *host
	}

	runService(flags)
}

// runMCPServer 运行真正的MCP协议服务器
func runMCPServer(args []string, flags *GlobalFlags) {
	mcpFlags := flag.NewFlagSet("mcp", flag.ExitOnError)
	mcpFlags.Parse(args)

	log.Println("Starting MCP protocol server...")

	// 加载配置
	cfg, err := loadConfigWithFlags(flags)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 创建组件
	_, mm, dl, _, err := createComponents(cfg)
	if err != nil {
		log.Fatalf("Failed to create components: %v", err)
	}

	// 启动镜像管理器
	mm.Start()
	defer mm.Stop()

	// 创建并启动MCP服务器
	mcpServer := mcpserver.NewMCPServer(dl, mm)
	if err := mcpServer.Start(); err != nil {
		log.Fatalf("Failed to start MCP server: %v", err)
	}
}

// runStatus 运行状态检查命令
func runStatus(args []string, flags *GlobalFlags) {
	cfg, err := loadConfigWithFlags(flags)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	_, mm, _, _, err := createComponents(cfg)
	if err != nil {
		log.Fatalf("Failed to create components: %v", err)
	}

	fmt.Println("Checking mirror status...")
	mm.Start()
	defer mm.Stop()

	// 等待检查完成
	time.Sleep(5 * time.Second)

	// 显示状态
	status := mm.GetMirrorStatus()
	count := mm.GetMirrorCount()

	fmt.Printf("\nMirror Status Report:\n")
	fmt.Printf("Total: %d, Online: %d, Offline: %d, Slow: %d, Unknown: %d\n\n",
		count["total"], count["online"], count["offline"], count["slow"], count["unknown"])

	for url, mirror := range status {
		fmt.Printf("%-30s %s (%v)\n", url, mirror.Status, mirror.ResponseTime)
		if mirror.ErrorMessage != "" {
			fmt.Printf("  Error: %s\n", mirror.ErrorMessage)
		}
	}
}

// runTest 运行镜像测试命令
func runTest(args []string, flags *GlobalFlags) {
	testFlags := flag.NewFlagSet("test", flag.ExitOnError)
	mirrorURL := testFlags.String("mirror", "", "Mirror URL to test")

	testFlags.Parse(args)

	if *mirrorURL == "" {
		fmt.Println("Must specify --mirror")
		testFlags.Usage()
		os.Exit(1)
	}

	cfg, err := loadConfigWithFlags(flags)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	_, mm, _, _, err := createComponents(cfg)
	if err != nil {
		log.Fatalf("Failed to create components: %v", err)
	}

	fmt.Printf("Testing mirror: %s\n", *mirrorURL)
	mirror, err := mm.TestMirror(*mirrorURL)
	if err != nil {
		log.Fatalf("Test failed: %v", err)
	}

	fmt.Printf("Test result: %s\n", mirror.Status)
	fmt.Printf("Response time: %v\n", mirror.ResponseTime)
	if mirror.ErrorMessage != "" {
		fmt.Printf("Error message: %s\n", mirror.ErrorMessage)
	}
}

// loadConfigWithFlags 使用全局标志加载配置
func loadConfigWithFlags(flags *GlobalFlags) (*config.Config, error) {
	cfg, err := config.LoadConfig(flags.ConfigPath)
	if err != nil {
		return nil, err
	}

	// 优先级：命令行参数 > 配置文件 > 默认值
	if flags.ProxyEnabled {
		cfg.Proxy.Enabled = true
	}
	if flags.ProxyHost != "" {
		cfg.Proxy.Host = flags.ProxyHost
	}
	if flags.ProxyPort != 0 {
		cfg.Proxy.Port = flags.ProxyPort
	}
	if flags.HealthInterval != 0 {
		cfg.HealthCheck.Interval = flags.HealthInterval
	}
	if flags.MCPHost != "" {
		cfg.MCP.Host = flags.MCPHost
	}
	if flags.MCPPort != 0 {
		cfg.MCP.Port = flags.MCPPort
	}

	return cfg, cfg.Validate()
}

// createComponents 创建所有组件
func createComponents(cfg *config.Config) (*proxy.ProxyManager, *mirror.MirrorManager, *downloader.Downloader, *mcp.Server, error) {
	// 创建代理管理器
	pm, err := proxy.NewProxyManager(cfg.Proxy.Enabled, cfg.Proxy.GetProxyURL())
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("Failed to create proxy manager: %w", err)
	}

	// 创建镜像管理器
	mm := mirror.NewMirrorManager(cfg.Mirrors, pm, cfg.HealthCheck.Interval, cfg.HealthCheck.Timeout)

	// 创建下载器
	dl := downloader.NewDownloader(mm, pm, cfg.Download.CacheDir, cfg.Download.MaxRetries, cfg.Download.Timeout)

	// 创建HTTP API服务器（兼容MCP接口格式）
	server := mcp.NewServer(dl, mm, cfg.MCP.Host, cfg.MCP.Port)

	return pm, mm, dl, server, nil
}

// copyFile 复制文件
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

// printVersion 打印版本信息
func printVersion() {
	fmt.Printf("SciHub-MCP %s\n", version)
	fmt.Printf("Commit: %s\n", commit)
	fmt.Printf("Built: %s\n", date)
}

// printHelp 打印帮助信息
func printHelp() {
	fmt.Printf(`SciHub-MCP %s - Sci-Hub 镜像管理和文件下载工具

使用方法:
  scihub-mcp [全局选项] [命令] [命令选项]

命令:
  fetch       下载论文文件
  api         启动HTTP API服务 (兼容MCP格式的REST API)
  mcp         启动真正的MCP协议服务器 (标准STDIO通信)
  status      检查镜像状态
  test        测试特定镜像

全局选项 (适用于所有命令):
  --config string              配置文件路径
  --proxy-enabled              启用代理
  --proxy-host string          代理主机 (默认: 127.0.0.1)
  --proxy-port int             代理端口 (默认: 3080)
  --health-interval duration   健康检查间隔 (默认: 30m)
  --mcp-host string            HTTP API服务主机 (默认: 0.0.0.0)
  --mcp-port int               HTTP API服务端口 (默认: 8080)
  --version                    显示版本信息
  --help                       显示此帮助信息

fetch 命令选项:
  --doi string                 论文DOI
  --url string                 论文URL
  --title string               论文标题
  --output string              输出文件路径

api 命令选项:
  --port int                   HTTP API端口 (覆盖全局 --mcp-port)
  --host string                HTTP API主机 (覆盖全局 --mcp-host)

mcp 命令选项:
  无选项，启动标准MCP协议服务器，通过STDIO通信

test 命令选项:
  --mirror string              要测试的镜像URL

配置优先级:
  命令行参数 > 配置文件 > 默认值

服务说明:
  api: 启动HTTP REST API服务，可通过curl或浏览器访问
           支持 /fetch, /download/, /mirrors, /status 等端点
           
  mcp:     启动标准MCP协议服务器，通过STDIO与客户端通信
           提供工具: download_paper, check_mirror_status, test_mirror
           提供资源: scihub://cache, scihub://mirrors/status, scihub://papers/{filename}

示例:
  # 启动HTTP API服务（默认模式）
  scihub-mcp
  scihub-mcp api

  # 启动MCP协议服务器
  scihub-mcp mcp

  # 全局启用代理
  scihub-mcp --proxy-enabled --proxy-host 127.0.0.1 --proxy-port 3080

  # 下载论文
  scihub-mcp fetch --doi "10.1038/nature12373"
  scihub-mcp --proxy-enabled fetch --doi "10.1038/nature12373"

  # 启动HTTP API在指定端口
  scihub-mcp api --port 9090
  scihub-mcp --proxy-enabled api --port 9090

  # 检查镜像状态
  scihub-mcp status
  scihub-mcp --proxy-enabled status

更多信息请访问: https://github.com/jifanchn/go-scihub-mcp
`, version)
}
