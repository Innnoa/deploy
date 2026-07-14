package common

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// NginxDownloader Nginx文件下载器
type NginxDownloader struct {
	client     *http.Client
	concurrent int // 并发下载数量
	username   string
	password   string
}

// NewNginxDownloader 创建新的Nginx下载器实例
func NewNginxDownloader(concurrent int, username, password string) *NginxDownloader {
	if concurrent <= 0 {
		concurrent = 5 // 默认并发数
	}
	return &NginxDownloader{
		client: &http.Client{
			Timeout: 300 * time.Second, // 5分钟超时
		},
		concurrent: concurrent,
		username:   username,
		password:   password,
	}
}

// DownloadFromNginx 从Nginx下载文件或文件列表
func (nd *NginxDownloader) DownloadFromNginx(url, outputDir string) error {
	AppLogger.Info(fmt.Sprintf("开始检查URL类型: %s", url))

	contentType, err := nd.getContentType(url)
	if err != nil {
		return fmt.Errorf("获取URL内容类型失败: %w", err)
	}
	AppLogger.Debug(fmt.Sprintf("URL内容类型: %s", contentType))

	// 判断是否为文件列表页面（Nginx默认的目录列表页面通常是text/html）
	if strings.Contains(contentType, "text/html") {
		AppLogger.Info("检测到URL是文件列表，开始批量下载")
		return nd.downloadFileList(url, outputDir)
	} else {
		AppLogger.Info("检测到URL是单个文件，开始下载")
		// 从URL中提取文件名
		fileName := filepath.Base(url)
		outputPath := filepath.Join(outputDir, fileName)
		return nd.downloadSingleFile(url, outputPath)
	}
}

func (nd *NginxDownloader) getContentType(url string) (string, error) {
	// 创建HEAD请求
	req, err := nd.buildBasicAuthHeader(url)
	if err != nil {
		return "", err
	}

	// 发送HEAD请求
	resp, err := nd.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check for authentication errors
	if resp.StatusCode == http.StatusUnauthorized {
		return "", errors.New("authentication failed: invalid username or password")
	}

	return resp.Header.Get("Content-Type"), nil
}

// downloadSingleFile 下载单个文件
func (nd *NginxDownloader) downloadSingleFile(url, outputPath string) error {
	AppLogger.Info(fmt.Sprintf("下载文件: %s -> %s", url, outputPath))

	// 创建目标文件
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer outFile.Close()

	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 添加Basic Auth认证头
	if nd.username != "" || nd.password != "" {
		req.SetBasicAuth(nd.username, nd.password)
	}

	// 下载文件
	resp, err := nd.client.Do(req)
	if err != nil {
		return fmt.Errorf("下载文件失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	// 写入文件
	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	AppLogger.Info(fmt.Sprintf("文件下载成功: %s", outputPath))
	return nil
}

func (nd *NginxDownloader) buildBasicAuthHeader(url string) (*http.Request, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}

	// Add Basic Auth header
	if nd.username != "" && nd.password != "" {
		req.SetBasicAuth(nd.username, nd.password)
	}
	return req, nil
}

// recursiveDownload 递归下载目录结构
func (nd *NginxDownloader) recursiveDownload(listURL, outputDir string, wg *sync.WaitGroup, semaphore chan struct{}, errors *[]string, errorsMutex *sync.Mutex) {
	defer wg.Done()

	AppLogger.Info(fmt.Sprintf("递归解析目录: %s -> %s", listURL, outputDir))

	// 创建请求获取文件列表页面
	req, err := http.NewRequest("GET", listURL, nil)
	if err != nil {
		errorsMutex.Lock()
		*errors = append(*errors, fmt.Sprintf("创建请求失败: %v", err))
		errorsMutex.Unlock()
		return
	}

	// 添加Basic Auth认证头
	if nd.username != "" || nd.password != "" {
		req.SetBasicAuth(nd.username, nd.password)
	}

	// 发送请求
	resp, err := nd.client.Do(req)
	if err != nil {
		errorsMutex.Lock()
		*errors = append(*errors, fmt.Sprintf("获取目录列表失败: %v", err))
		errorsMutex.Unlock()
		return
	}
	if resp.StatusCode == http.StatusNotFound {
		errorsMutex.Lock()
		*errors = append(*errors, fmt.Sprintf("获取目录列表失败: %d", 404))
		errorsMutex.Unlock()
		return
	}
	defer resp.Body.Close()

	// 解析Nginx默认目录列表页面，提取文件和目录
	entries, err := nd.parseNginxFileList(resp.Body, listURL)
	if err != nil {
		errorsMutex.Lock()
		*errors = append(*errors, fmt.Sprintf("解析列表失败: %v", err))
		errorsMutex.Unlock()
		return
	}

	// 处理所有条目
	for _, entry := range entries {
		if entry.IsDir {
			// 如果是目录，创建对应的本地目录并递归下载
			subDirName := strings.TrimSuffix(entry.Name, "/")
			// 确保子目录名称是安全的文件名
			subDirName = filepath.Base(subDirName)
			subDirPath := filepath.Join(outputDir, subDirName)

			// 创建本地子目录
			if err := os.MkdirAll(subDirPath, 0755); err != nil {
				errorsMutex.Lock()
				*errors = append(*errors, fmt.Sprintf("创建目录失败 %s: %v", subDirPath, err))
				errorsMutex.Unlock()
				continue
			}

			// 确保目录URL以斜杠结尾
			dirURL := entry.URL
			if !strings.HasSuffix(dirURL, "/") {
				dirURL = dirURL + "/"
			}

			// 递归下载子目录
			wg.Add(1)
			go nd.recursiveDownload(dirURL, subDirPath, wg, semaphore, errors, errorsMutex)
		} else {
			// 如果是文件，下载文件
			wg.Add(1)
			semaphore <- struct{}{}

			go func(entryURL, fileName, targetDir string) {
				defer wg.Done()
				defer func() { <-semaphore }()

				// 确保文件名安全，避免路径遍历攻击
				safeFileName := filepath.Base(fileName)
				outputPath := filepath.Join(targetDir, safeFileName)

				if err := nd.downloadSingleFile(entryURL, outputPath); err != nil {
					errorsMutex.Lock()
					*errors = append(*errors, fmt.Sprintf("文件 %s 下载失败: %v", safeFileName, err))
					errorsMutex.Unlock()
				}
			}(entry.URL, entry.Name, outputDir)
		}
	}
}

// downloadFileList 从Nginx文件列表页面下载所有文件（支持目录结构）
func (nd *NginxDownloader) downloadFileList(listURL, outputDir string) error {
	AppLogger.Info(fmt.Sprintf("开始下载目录: %s -> %s", listURL, outputDir))

	// 确保输出目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 使用并发和递归下载
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, nd.concurrent)
	errors := make([]string, 0)
	errorsMutex := sync.Mutex{}

	// 启动递归下载
	wg.Add(1)
	nd.recursiveDownload(listURL, outputDir, &wg, semaphore, &errors, &errorsMutex)

	// 等待所有下载任务完成
	wg.Wait()

	// 报告错误
	if len(errors) > 0 {
		return fmt.Errorf("部分内容下载失败:\n%s", strings.Join(errors, "\n"))
	}

	AppLogger.Info(fmt.Sprintf("目录下载完成: %s", outputDir))
	return nil
}

// FileEntry 表示文件或目录条目
type FileEntry struct {
	URL    string
	Name   string
	IsDir  bool
	Parent string
}

// parseNginxFileList 解析Nginx默认目录列表页面，提取文件和目录
func (nd *NginxDownloader) parseNginxFileList(htmlContent io.Reader, baseURL string) ([]FileEntry, error) {
	// 规范化基础URL
	if !strings.HasSuffix(baseURL, "/") {
		baseURL = baseURL + "/"
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(htmlContent)
	if err != nil {
		return nil, err
	}
	// Find all links in the directory listing
	links := make([]string, 0)
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			links = append(links, href)
		}
	})

	entries := make([]FileEntry, 0)
	processedEntries := make(map[string]bool)

	for _, link := range links {
		if len(link) < 3 {
			continue
		}

		// 跳过上级目录链接和当前目录
		if link == ".." || link == "../" || link == "." || link == "./" || strings.TrimSpace(link) == "Parent Directory" {
			continue
		}

		// 判断是否为目录
		isDir := strings.HasSuffix(link, "/")

		// 移除目录后缀用于解码
		linkName := link
		if isDir {
			linkName = strings.TrimSuffix(link, "/")
		}

		// URL解码文件名，处理中文等非ASCII字符
		decodedName, err := url.QueryUnescape(linkName)
		if err != nil {
			// 如果解码失败，使用原始链接名
			decodedName = linkName
		}

		// 构建完整的URL
		var fullURL string
		if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
			fullURL = link
		} else {
			fullURL = baseURL + link
		}

		// 去重
		if !processedEntries[fullURL] {
			processedEntries[fullURL] = true
			entries = append(entries, FileEntry{
				URL:    fullURL,
				Name:   decodedName,
				IsDir:  isDir,
				Parent: baseURL,
			})
		}
	}

	AppLogger.Info(fmt.Sprintf("解析到 %d 个条目", len(entries)))
	return entries, nil
}

// DownloadConfig 下载配置
type DownloadConfig struct {
	URL        string            `json:"url"`
	OutputDir  string            `json:"output_dir"`
	Headers    map[string]string `json:"headers,omitempty"`
	Concurrent int               `json:"concurrent,omitempty"`
}

// DownloadWithConfig 使用配置下载文件或文件列表
func (nd *NginxDownloader) DownloadWithConfig(config *DownloadConfig) error {
	// 如果配置了并发数，则更新
	if config.Concurrent > 0 {
		nd.concurrent = config.Concurrent
	}

	// 确保输出目录路径格式统一，处理不同操作系统的路径分隔符问题
	normalizedOutputDir := filepath.Clean(config.OutputDir)

	// 确保输出目录存在，创建目录时保持目录结构一致性
	if err := os.MkdirAll(normalizedOutputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	return nd.DownloadFromNginx(config.URL, normalizedOutputDir)
}
