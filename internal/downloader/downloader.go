package downloader

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jifanchn/go-scihub-mcp/internal/mirror"
	"github.com/jifanchn/go-scihub-mcp/internal/proxy"
)

// DownloadRequest 下载请求
type DownloadRequest struct {
	DOI   string `json:"doi"`
	URL   string `json:"url"`
	Title string `json:"title"`
}

// DownloadResult 下载结果
type DownloadResult struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	MirrorUsed  string `json:"mirror_used"`
	DownloadURL string `json:"download_url"`
	Cached      bool   `json:"cached"`
	FilePath    string `json:"file_path"`
}

// Downloader 下载器
type Downloader struct {
	mirrorManager *mirror.MirrorManager
	proxyManager  *proxy.ProxyManager
	cacheDir      string
	maxRetries    int
	timeout       time.Duration
}

// NewDownloader 创建下载器
func NewDownloader(mm *mirror.MirrorManager, pm *proxy.ProxyManager, cacheDir string, maxRetries int, timeout time.Duration) *Downloader {
	return &Downloader{
		mirrorManager: mm,
		proxyManager:  pm,
		cacheDir:      cacheDir,
		maxRetries:    maxRetries,
		timeout:       timeout,
	}
}

// Download 下载文件
func (d *Downloader) Download(req *DownloadRequest) (*DownloadResult, error) {
	// 验证请求
	if req.DOI == "" && req.URL == "" {
		return &DownloadResult{
			Success: false,
			Message: "Must provide DOI or URL",
		}, fmt.Errorf("Invalid download request")
	}

	// 生成缓存文件名
	cacheFilename := d.generateCacheFilename(req)
	cachePath := filepath.Join(d.cacheDir, cacheFilename)

	// 检查缓存
	if info, err := os.Stat(cachePath); err == nil && info.Size() > 0 {
		return &DownloadResult{
			Success:  true,
			Message:  "File found in cache",
			Filename: cacheFilename,
			Size:     info.Size(),
			Cached:   true,
			FilePath: cachePath,
		}, nil
	}

	// 尝试从各个镜像下载
	return d.downloadFromMirrors(req, cachePath, cacheFilename)
}

// downloadFromMirrors 从镜像下载
func (d *Downloader) downloadFromMirrors(req *DownloadRequest, cachePath, filename string) (*DownloadResult, error) {
	available := d.mirrorManager.GetAvailableMirrors()
	if len(available) == 0 {
		return &DownloadResult{
			Success: false,
			Message: "No available mirrors",
		}, fmt.Errorf("No available mirrors")
	}

	var lastError error

	// 按响应时间排序尝试每个镜像
	for _, mirror := range available {
		result, err := d.downloadFromMirror(req, mirror.URL, cachePath, filename)
		if err == nil {
			result.MirrorUsed = mirror.URL
			return result, nil
		}
		lastError = err
	}

	return &DownloadResult{
		Success: false,
		Message: fmt.Sprintf("Download failed: %v", lastError),
	}, lastError
}

// downloadFromMirror 从指定镜像下载
func (d *Downloader) downloadFromMirror(req *DownloadRequest, mirrorURL, cachePath, filename string) (*DownloadResult, error) {
	for attempt := 0; attempt < d.maxRetries; attempt++ {
		result, err := d.attemptDownload(req, mirrorURL, cachePath, filename)
		if err == nil {
			return result, nil
		}

		if attempt < d.maxRetries-1 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	return nil, fmt.Errorf("Download failed, retried %d times", d.maxRetries)
}

// attemptDownload 尝试下载
func (d *Downloader) attemptDownload(req *DownloadRequest, mirrorURL, cachePath, filename string) (*DownloadResult, error) {
	// 构建下载URL
	downloadURL, err := d.buildDownloadURL(mirrorURL, req)
	if err != nil {
		return nil, fmt.Errorf("Failed to build download URL: %w", err)
	}

	// 首先获取论文页面，解析真实的PDF链接
	pdfURL, err := d.getPDFURL(downloadURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to get PDF link: %w", err)
	}

	// 下载PDF文件
	err = d.downloadFile(pdfURL, cachePath)
	if err != nil {
		return nil, fmt.Errorf("Download file failed: %w", err)
	}

	// 获取文件信息
	info, err := os.Stat(cachePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to get file info: %w", err)
	}

	return &DownloadResult{
		Success:     true,
		Message:     "Download succeeded",
		Filename:    filename,
		Size:        info.Size(),
		DownloadURL: pdfURL,
		Cached:      false,
		FilePath:    cachePath,
	}, nil
}

// buildDownloadURL 构建下载URL
func (d *Downloader) buildDownloadURL(mirrorURL string, req *DownloadRequest) (string, error) {
	baseURL := strings.TrimSuffix(mirrorURL, "/")

	if req.DOI != "" {
		// 清理DOI
		doi := strings.TrimSpace(req.DOI)
		doi = strings.TrimPrefix(doi, "doi:")
		doi = strings.TrimPrefix(doi, "DOI:")
		return fmt.Sprintf("%s/%s", baseURL, url.QueryEscape(doi)), nil
	}

	if req.URL != "" {
		return fmt.Sprintf("%s/%s", baseURL, url.QueryEscape(req.URL)), nil
	}

	return "", fmt.Errorf("Cannot build download URL")
}

// getPDFURL 从Sci-Hub页面获取PDF链接
func (d *Downloader) getPDFURL(pageURL string) (string, error) {
	client := d.proxyManager.GetHTTPClient()
	client.Timeout = d.timeout

	resp, err := client.Get(pageURL)
	if err != nil {
		return "", fmt.Errorf("Failed to request page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Page returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read page content: %w", err)
	}

	content := string(body)

	// 尝试多种PDF链接模式
	patterns := []string{
		`<embed[^>]+src="([^"]*\.pdf[^"]*)"`,
		`<iframe[^>]+src="([^"]*\.pdf[^"]*)"`,
		`<a[^>]+href="([^"]*\.pdf[^"]*)"`,
		`location\.href\s*=\s*["']([^"']*\.pdf[^"']*)["']`,
		`window\.location\s*=\s*["']([^"']*\.pdf[^"']*)["']`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			pdfURL := matches[1]

			// 如果是相对路径，转换为绝对路径
			if strings.HasPrefix(pdfURL, "/") {
				u, err := url.Parse(pageURL)
				if err == nil {
					pdfURL = fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, pdfURL)
				}
			} else if strings.HasPrefix(pdfURL, "//") {
				u, err := url.Parse(pageURL)
				if err == nil {
					pdfURL = fmt.Sprintf("%s:%s", u.Scheme, pdfURL)
				}
			}

			return pdfURL, nil
		}
	}

	return "", fmt.Errorf("PDF link not found")
}

// downloadFile 下载文件
func (d *Downloader) downloadFile(url, filepath string) error {
	client := d.proxyManager.GetHTTPClient()
	client.Timeout = d.timeout

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("Download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Download returned status code: %d", resp.StatusCode)
	}

	// 确保目录存在
	dir := filepath[:strings.LastIndex(filepath, "/")]
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("Failed to create directory: %w", err)
	}

	// 创建文件
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("Failed to create file: %w", err)
	}
	defer file.Close()

	// 复制内容
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		os.Remove(filepath) // 下载失败时删除不完整的文件
		return fmt.Errorf("Failed to write file: %w", err)
	}

	return nil
}

// generateCacheFilename 生成缓存文件名
func (d *Downloader) generateCacheFilename(req *DownloadRequest) string {
	var identifier string

	if req.DOI != "" {
		identifier = req.DOI
	} else if req.URL != "" {
		identifier = req.URL
	} else if req.Title != "" {
		identifier = req.Title
	} else {
		identifier = fmt.Sprintf("unknown_%d", time.Now().Unix())
	}

	// 生成MD5哈希作为文件名
	hasher := md5.New()
	hasher.Write([]byte(identifier))
	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	return hash + ".pdf"
}

// GetCachedFile 获取缓存文件
func (d *Downloader) GetCachedFile(req *DownloadRequest) (string, bool) {
	filename := d.generateCacheFilename(req)
	cachePath := filepath.Join(d.cacheDir, filename)

	if info, err := os.Stat(cachePath); err == nil && info.Size() > 0 {
		return cachePath, true
	}

	return "", false
}

// ClearCache 清理缓存
func (d *Downloader) ClearCache() error {
	entries, err := os.ReadDir(d.cacheDir)
	if err != nil {
		return fmt.Errorf("Failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".pdf") {
			path := filepath.Join(d.cacheDir, entry.Name())
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("Failed to delete cache file %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}
