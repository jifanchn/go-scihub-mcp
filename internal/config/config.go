package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 主配置结构
type Config struct {
	Mirrors     []string       `yaml:"mirrors" json:"mirrors"`
	Proxy       ProxyConfig    `yaml:"proxy" json:"proxy"`
	HealthCheck HealthConfig   `yaml:"health_check" json:"health_check"`
	MCP         MCPConfig      `yaml:"mcp" json:"mcp"`
	Download    DownloadConfig `yaml:"download" json:"download"`
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	Type     string `yaml:"type" json:"type"` // socks5, http
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

// HealthConfig 健康检查配置
type HealthConfig struct {
	Interval time.Duration `yaml:"interval" json:"interval"`
	Timeout  time.Duration `yaml:"timeout" json:"timeout"`
}

// MCPConfig MCP服务配置
type MCPConfig struct {
	Port      int    `yaml:"port" json:"port"`
	Host      string `yaml:"host" json:"host"`
	Transport string `yaml:"transport" json:"transport"` // sse
	SSEPath   string `yaml:"sse_path" json:"sse_path"`   // SSE端点路径，默认/sse
}

// DownloadConfig 下载配置
type DownloadConfig struct {
	CacheDir   string        `yaml:"cache_dir" json:"cache_dir"`
	MaxRetries int           `yaml:"max_retries" json:"max_retries"`
	Timeout    time.Duration `yaml:"timeout" json:"timeout"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Mirrors: []string{
			"https://sci-hub.ru",
			"https://sci-hub.se",
			"https://sci-hub.st",
			"https://sci-hub.box",
			"https://sci-hub.red",
			"https://sci-hub.al",
			"https://sci-hub.ee",
			"https://sci-hub.lu",
			"https://sci-hub.ren",
			"https://sci-hub.shop",
			"https://sci-hub.vg",
		},
		Proxy: ProxyConfig{
			Enabled: false,
			Type:    "socks5",
			Host:    "127.0.0.1",
			Port:    3080,
		},
		HealthCheck: HealthConfig{
			Interval: 30 * time.Minute,
			Timeout:  10 * time.Second,
		},
		MCP: MCPConfig{
			Port:      8080,
			Host:      "0.0.0.0",
			Transport: "sse",  // 默认使用sse
			SSEPath:   "/sse", // SSE端点路径
		},
		Download: DownloadConfig{
			CacheDir:   "./cache",
			MaxRetries: 3,
			Timeout:    60 * time.Second,
		},
	}
}

// LoadConfig 从文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	// 如果没有指定配置文件，尝试查找默认路径
	if configPath == "" {
		configPath = findConfigFile()
	}

	if configPath != "" {
		if err := loadFromFile(config, configPath); err != nil {
			return nil, fmt.Errorf("加载配置文件失败: %w", err)
		}
	}

	// 确保缓存目录存在
	if err := os.MkdirAll(config.Download.CacheDir, 0755); err != nil {
		return nil, fmt.Errorf("创建缓存目录失败: %w", err)
	}

	return config, nil
}

// findConfigFile 查找配置文件
func findConfigFile() string {
	// 配置文件查找路径
	searchPaths := []string{
		"./config.yaml",
		"./config.yml",
	}

	// 添加用户目录路径
	if homeDir, err := os.UserHomeDir(); err == nil {
		searchPaths = append(searchPaths,
			filepath.Join(homeDir, ".config", "scihub-mcp", "config.yaml"),
			filepath.Join(homeDir, ".config", "scihub-mcp", "config.yml"),
		)
	}

	// 添加系统级路径
	searchPaths = append(searchPaths,
		"/etc/scihub-mcp/config.yaml",
		"/etc/scihub-mcp/config.yml",
	)

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// loadFromFile 从文件加载配置
func loadFromFile(config *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	return nil
}

// SaveConfig 保存配置到文件
func SaveConfig(config *Config, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// GetProxyURL 获取代理URL
func (p *ProxyConfig) GetProxyURL() string {
	if !p.Enabled {
		return ""
	}

	if p.Username != "" && p.Password != "" {
		return fmt.Sprintf("%s://%s:%s@%s:%d", p.Type, p.Username, p.Password, p.Host, p.Port)
	}

	return fmt.Sprintf("%s://%s:%d", p.Type, p.Host, p.Port)
}

// Validate 验证配置
func (c *Config) Validate() error {
	if len(c.Mirrors) == 0 {
		return fmt.Errorf("至少需要配置一个镜像")
	}

	if c.MCP.Port <= 0 || c.MCP.Port > 65535 {
		return fmt.Errorf("MCP端口无效: %d", c.MCP.Port)
	}

	if c.MCP.Transport != "sse" {
		return fmt.Errorf("不支持的传输模式: %s (仅支持: sse)", c.MCP.Transport)
	}

	if c.Download.MaxRetries < 0 {
		return fmt.Errorf("最大重试次数不能为负数")
	}

	if c.HealthCheck.Interval < time.Second {
		return fmt.Errorf("健康检查间隔不能小于1秒")
	}

	if c.HealthCheck.Timeout < time.Second {
		return fmt.Errorf("健康检查超时不能小于1秒")
	}

	return nil
}
