package mirror

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/jifanchn/go-scihub-mcp/internal/proxy"
)

// MirrorStatus 镜像状态
type MirrorStatus int

const (
	StatusUnknown MirrorStatus = iota
	StatusOnline
	StatusOffline
	StatusSlow
)

func (s MirrorStatus) String() string {
	switch s {
	case StatusOnline:
		return "在线"
	case StatusOffline:
		return "离线"
	case StatusSlow:
		return "缓慢"
	default:
		return "未知"
	}
}

// Mirror 镜像信息
type Mirror struct {
	URL          string        `json:"url"`
	Status       MirrorStatus  `json:"status"`
	ResponseTime time.Duration `json:"response_time"`
	LastChecked  time.Time     `json:"last_checked"`
	ErrorCount   int           `json:"error_count"`
	ErrorMessage string        `json:"error_message"`
}

// MirrorManager 镜像管理器
type MirrorManager struct {
	mirrors       map[string]*Mirror
	proxyManager  *proxy.ProxyManager
	checkInterval time.Duration
	checkTimeout  time.Duration
	silent        bool
	mu            sync.RWMutex
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// NewMirrorManager 创建新的镜像管理器
func NewMirrorManager(mirrorURLs []string, proxyManager *proxy.ProxyManager, checkInterval, checkTimeout time.Duration, silent bool) *MirrorManager {
	mm := &MirrorManager{
		mirrors:       make(map[string]*Mirror),
		proxyManager:  proxyManager,
		checkInterval: checkInterval,
		checkTimeout:  checkTimeout,
		silent:        silent,
		stopChan:      make(chan struct{}),
	}

	// 初始化镜像
	for _, url := range mirrorURLs {
		mm.mirrors[url] = &Mirror{
			URL:    url,
			Status: StatusUnknown,
		}
	}

	return mm
}

// Start 启动镜像管理器
func (mm *MirrorManager) Start() {
	mm.wg.Add(1)
	go mm.healthCheckLoop()
}

// Stop 停止镜像管理器
func (mm *MirrorManager) Stop() {
	close(mm.stopChan)
	mm.wg.Wait()
}

// healthCheckLoop 健康检查循环
func (mm *MirrorManager) healthCheckLoop() {
	defer mm.wg.Done()

	// 立即执行一次检查
	mm.checkAllMirrors()

	ticker := time.NewTicker(mm.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mm.checkAllMirrors()
		case <-mm.stopChan:
			return
		}
	}
}

// checkAllMirrors 检查所有镜像
func (mm *MirrorManager) checkAllMirrors() {
	mm.mu.RLock()
	urls := make([]string, 0, len(mm.mirrors))
	for url := range mm.mirrors {
		urls = append(urls, url)
	}
	mm.mu.RUnlock()

	// 并发检查所有镜像
	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			mm.checkMirror(url)
		}(url)
	}

	wg.Wait()
	if !mm.silent {
		log.Printf("Mirror health check completed, checked %d mirrors", len(urls))
	}
}

// checkMirror 检查单个镜像
func (mm *MirrorManager) checkMirror(url string) {
	start := time.Now()
	status := StatusOffline
	errorMsg := ""

	defer func() {
		mm.mu.Lock()
		mirror := mm.mirrors[url]
		mirror.Status = status
		mirror.ResponseTime = time.Since(start)
		mirror.LastChecked = time.Now()
		if status == StatusOffline {
			mirror.ErrorCount++
			mirror.ErrorMessage = errorMsg
		} else {
			mirror.ErrorCount = 0
			mirror.ErrorMessage = ""
		}
		mm.mu.Unlock()
	}()

	// 创建HTTP请求
	ctx, cancel := context.WithTimeout(context.Background(), mm.checkTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		errorMsg = fmt.Sprintf("创建请求失败: %v", err)
		return
	}

	req.Header.Set("User-Agent", "SciHub-MCP/1.0 Health Check")

	// 发送请求
	resp, err := mm.proxyManager.GetHTTPClient().Do(req)
	if err != nil {
		errorMsg = fmt.Sprintf("请求失败: %v", err)
		return
	}
	defer resp.Body.Close()

	responseTime := time.Since(start)

	// 判断状态
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		if responseTime > 5*time.Second {
			status = StatusSlow
		} else {
			status = StatusOnline
		}
	} else {
		errorMsg = fmt.Sprintf("HTTP状态码: %d", resp.StatusCode)
	}
}

// GetAvailableMirrors 获取可用镜像
func (mm *MirrorManager) GetAvailableMirrors() []*Mirror {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	var available []*Mirror
	for _, mirror := range mm.mirrors {
		if mirror.Status == StatusOnline || mirror.Status == StatusSlow {
			// 创建副本以避免并发访问问题
			mirrorCopy := *mirror
			available = append(available, &mirrorCopy)
		}
	}

	return available
}

// GetBestMirror 获取最佳镜像
func (mm *MirrorManager) GetBestMirror() *Mirror {
	available := mm.GetAvailableMirrors()
	if len(available) == 0 {
		return nil
	}

	// 按响应时间排序，选择最快的在线镜像
	var best *Mirror
	for _, mirror := range available {
		if mirror.Status == StatusOnline {
			if best == nil || mirror.ResponseTime < best.ResponseTime {
				best = mirror
			}
		}
	}

	// 如果没有在线的，选择最快的慢镜像
	if best == nil {
		for _, mirror := range available {
			if mirror.Status == StatusSlow {
				if best == nil || mirror.ResponseTime < best.ResponseTime {
					best = mirror
				}
			}
		}
	}

	return best
}

// GetMirrorStatus 获取所有镜像状态
func (mm *MirrorManager) GetMirrorStatus() map[string]*Mirror {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	result := make(map[string]*Mirror)
	for url, mirror := range mm.mirrors {
		// 创建副本
		mirrorCopy := *mirror
		result[url] = &mirrorCopy
	}

	return result
}

// TestMirror 测试特定镜像
func (mm *MirrorManager) TestMirror(url string) (*Mirror, error) {
	mm.mu.RLock()
	_, exists := mm.mirrors[url]
	mm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("镜像 %s 不存在", url)
	}

	// 执行检查
	mm.checkMirror(url)

	// 返回更新后的状态
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	mirrorCopy := *mm.mirrors[url]
	return &mirrorCopy, nil
}

// AddMirror 添加镜像
func (mm *MirrorManager) AddMirror(url string) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if _, exists := mm.mirrors[url]; !exists {
		mm.mirrors[url] = &Mirror{
			URL:    url,
			Status: StatusUnknown,
		}
	}
}

// RemoveMirror 移除镜像
func (mm *MirrorManager) RemoveMirror(url string) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	delete(mm.mirrors, url)
}

// GetMirrorCount 获取镜像数量统计
func (mm *MirrorManager) GetMirrorCount() map[string]int {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	count := map[string]int{
		"total":   0,
		"online":  0,
		"offline": 0,
		"slow":    0,
		"unknown": 0,
	}

	for _, mirror := range mm.mirrors {
		count["total"]++
		switch mirror.Status {
		case StatusOnline:
			count["online"]++
		case StatusOffline:
			count["offline"]++
		case StatusSlow:
			count["slow"]++
		case StatusUnknown:
			count["unknown"]++
		}
	}

	return count
}
