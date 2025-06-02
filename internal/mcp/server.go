package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jifanchn/go-scihub-mcp/internal/downloader"
	"github.com/jifanchn/go-scihub-mcp/internal/mirror"
)

// Server MCP服务器
type Server struct {
	downloader    *downloader.Downloader
	mirrorManager *mirror.MirrorManager
	host          string
	port          int
}

// APIResponse 通用API响应
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewServer 创建MCP服务器
func NewServer(d *downloader.Downloader, mm *mirror.MirrorManager, host string, port int) *Server {
	return &Server{
		downloader:    d,
		mirrorManager: mm,
		host:          host,
		port:          port,
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// 设置路由
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/fetch", s.handleFetch)
	mux.HandleFunc("/download/", s.handleDownload)
	mux.HandleFunc("/mirrors", s.handleMirrors)
	mux.HandleFunc("/status", s.handleStatus)

	// 添加中间件
	handler := s.corsMiddleware(s.loggingMiddleware(mux))

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	log.Printf("MCP server started at %s", addr)

	return http.ListenAndServe(addr, handler)
}

// handleHealth 健康检查接口
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeErrorResponse(w, http.StatusMethodNotAllowed, "Only GET method is supported")
		return
	}

	count := s.mirrorManager.GetMirrorCount()
	data := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"mirrors":   count,
		"uptime":    time.Since(time.Now()).String(), // 这里应该记录启动时间
	}

	s.writeSuccessResponse(w, "Service is running normally", data)
}

// handleFetch 文件下载接口
func (s *Server) handleFetch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeErrorResponse(w, http.StatusMethodNotAllowed, "Only POST method is supported")
		return
	}

	var req downloader.DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	// 验证请求
	if req.DOI == "" && req.URL == "" {
		s.writeErrorResponse(w, http.StatusBadRequest, "Must provide either DOI or URL")
		return
	}

	result, err := s.downloader.Download(&req)
	if err != nil {
		log.Printf("Download failed: %v", err)
		s.writeErrorResponse(w, http.StatusInternalServerError, result.Message)
		return
	}

	// 检查是否请求返回文件内容
	returnFile := r.URL.Query().Get("return_file") == "true"

	if returnFile {
		// 直接返回文件内容
		s.serveFile(w, result.FilePath, result.Filename)
	} else {
		// 返回下载信息和下载链接
		responseData := map[string]interface{}{
			"success":       result.Success,
			"message":       result.Message,
			"filename":      result.Filename,
			"size":          result.Size,
			"mirror_used":   result.MirrorUsed,
			"download_url":  result.DownloadURL,
			"cached":        result.Cached,
			"file_path":     result.FilePath,
			"download_link": fmt.Sprintf("/download/%s", result.Filename),
		}

		s.writeSuccessResponse(w, result.Message, responseData)
	}
}

// handleDownload 文件下载接口
func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeErrorResponse(w, http.StatusMethodNotAllowed, "Only GET method is supported")
		return
	}

	// 从URL路径提取文件名
	filename := r.URL.Path[len("/download/"):]
	if filename == "" {
		s.writeErrorResponse(w, http.StatusBadRequest, "Filename cannot be empty")
		return
	}

	// 构建文件路径
	filePath := fmt.Sprintf("cache/%s", filename)

	s.serveFile(w, filePath, filename)
}

// serveFile 提供文件下载
func (s *Server) serveFile(w http.ResponseWriter, filePath, filename string) {
	// 检查文件是否存在
	file, err := os.Open(filePath)
	if err != nil {
		s.writeErrorResponse(w, http.StatusNotFound, "File not found")
		return
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get file information")
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// 复制文件内容到响应
	_, err = io.Copy(w, file)
	if err != nil {
		log.Printf("File transfer failed: %v", err)
	}
}

// handleMirrors 镜像状态接口
func (s *Server) handleMirrors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeErrorResponse(w, http.StatusMethodNotAllowed, "Only GET method is supported")
		return
	}

	mirrors := s.mirrorManager.GetMirrorStatus()
	count := s.mirrorManager.GetMirrorCount()

	data := map[string]interface{}{
		"mirrors": mirrors,
		"summary": count,
	}

	s.writeSuccessResponse(w, "Mirror status retrieved successfully", data)
}

// handleStatus 系统状态接口
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeErrorResponse(w, http.StatusMethodNotAllowed, "Only GET method is supported")
		return
	}

	available := s.mirrorManager.GetAvailableMirrors()
	count := s.mirrorManager.GetMirrorCount()
	best := s.mirrorManager.GetBestMirror()

	data := map[string]interface{}{
		"mirror_count":      count,
		"available_mirrors": len(available),
		"best_mirror":       best,
		"timestamp":         time.Now().Unix(),
	}

	s.writeSuccessResponse(w, "System status retrieved successfully", data)
}

// writeSuccessResponse 写入成功响应
func (s *Server) writeSuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	response := APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("编码响应失败: %v", err)
	}
}

// writeErrorResponse 写入错误响应
func (s *Server) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := APIResponse{
		Success: false,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("编码错误响应失败: %v", err)
	}
}

// corsMiddleware CORS中间件
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware 日志中间件
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 创建响应记录器来捕获状态码
		recorder := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(recorder, r)

		duration := time.Since(start)
		log.Printf("%s %s %d %v %s",
			r.Method,
			r.URL.Path,
			recorder.statusCode,
			duration,
			r.RemoteAddr,
		)
	})
}

// responseRecorder 响应记录器
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

// GetAddr 获取服务器地址
func (s *Server) GetAddr() string {
	return fmt.Sprintf("%s:%d", s.host, s.port)
}

// GetPort 获取端口
func (s *Server) GetPort() int {
	return s.port
}

// GetHost 获取主机
func (s *Server) GetHost() string {
	return s.host
}

// Shutdown 关闭服务器 (这里是简化版本，实际应该使用context进行优雅关闭)
func (s *Server) Shutdown() error {
	log.Println("MCP服务器正在关闭...")
	return nil
}
