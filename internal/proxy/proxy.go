package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"
)

// ProxyManager 代理管理器
type ProxyManager struct {
	enabled  bool
	proxyURL string
	client   *http.Client
}

// NewProxyManager 创建代理管理器
func NewProxyManager(enabled bool, proxyURL string) (*ProxyManager, error) {
	pm := &ProxyManager{
		enabled:  enabled,
		proxyURL: proxyURL,
	}

	client, err := pm.createHTTPClient()
	if err != nil {
		return nil, fmt.Errorf("创建HTTP客户端失败: %w", err)
	}

	pm.client = client
	return pm, nil
}

// createHTTPClient 创建HTTP客户端
func (pm *ProxyManager) createHTTPClient() (*http.Client, error) {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
	}

	if pm.enabled && pm.proxyURL != "" {
		if err := pm.configureProxy(transport); err != nil {
			return nil, err
		}
	}

	return &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}, nil
}

// configureProxy 配置代理
func (pm *ProxyManager) configureProxy(transport *http.Transport) error {
	proxyURL, err := url.Parse(pm.proxyURL)
	if err != nil {
		return fmt.Errorf("解析代理URL失败: %w", err)
	}

	switch proxyURL.Scheme {
	case "socks5":
		return pm.configureSocks5Proxy(transport, proxyURL)
	case "http", "https":
		transport.Proxy = http.ProxyURL(proxyURL)
		return nil
	default:
		return fmt.Errorf("不支持的代理类型: %s", proxyURL.Scheme)
	}
}

// configureSocks5Proxy 配置SOCKS5代理
func (pm *ProxyManager) configureSocks5Proxy(transport *http.Transport, proxyURL *url.URL) error {
	// 创建SOCKS5拨号器
	var auth *proxy.Auth
	if proxyURL.User != nil {
		password, _ := proxyURL.User.Password()
		auth = &proxy.Auth{
			User:     proxyURL.User.Username(),
			Password: password,
		}
	}

	dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, auth, proxy.Direct)
	if err != nil {
		return fmt.Errorf("创建SOCKS5拨号器失败: %w", err)
	}

	// 设置自定义拨号函数
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.Dial(network, addr)
	}

	return nil
}

// GetHTTPClient 获取HTTP客户端
func (pm *ProxyManager) GetHTTPClient() *http.Client {
	return pm.client
}

// IsEnabled 检查代理是否启用
func (pm *ProxyManager) IsEnabled() bool {
	return pm.enabled
}

// GetProxyURL 获取代理URL
func (pm *ProxyManager) GetProxyURL() string {
	return pm.proxyURL
}

// TestConnection 测试代理连接
func (pm *ProxyManager) TestConnection(targetURL string) error {
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return fmt.Errorf("创建测试请求失败: %w", err)
	}

	req.Header.Set("User-Agent", "SciHub-MCP/1.0")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req = req.WithContext(ctx)

	resp, err := pm.client.Do(req)
	if err != nil {
		return fmt.Errorf("代理连接测试失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("代理连接测试返回错误状态码: %d", resp.StatusCode)
	}

	return nil
}

// SetTimeout 设置超时时间
func (pm *ProxyManager) SetTimeout(timeout time.Duration) {
	pm.client.Timeout = timeout
}

// Clone 克隆代理管理器
func (pm *ProxyManager) Clone() (*ProxyManager, error) {
	return NewProxyManager(pm.enabled, pm.proxyURL)
}
