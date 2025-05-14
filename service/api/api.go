package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"recovery-unit-deploy/service/common"
	"strings"
	"time"
)

type HTTPClient struct {
	Client  *http.Client // 支持自定义超时、重定向策略
	BaseURL string       // 可选，统一服务地址
}

type RequestParams struct {
	Method  string            // 请求方法：GET/POST/PUT等
	Path    string            // 接口路径（如"/api/user"）
	Headers map[string]string // 请求头
	Query   map[string]string // URL查询参数
	Body    interface{}       // 支持JSON/Form/字符串多种类型
	Timeout time.Duration     // 单次请求超时时间
}

type PublicRequest struct {
	AccessKeyId string `json:"accessKeyId"`
	Signature   string `json:"signature"`
	Timestamp   string `json:"timestamp"`
}

type PublicResponse struct {
	RequestId string `json:"requestId"`
	Code      string `json:"code"`
}

type OAServerRequest struct {
	IP        string `json:"ip"`
	PublicReq PublicRequest
}

type OAServerResponse struct {
	Name       string `json:"name"`
	PublicResp PublicResponse
}

type PrinterModelResponse struct {
	Models     []string `json:"models"`
	PublicResp PublicResponse
}

type NetworkPrinterRequest struct {
	Keyword   string `json:"keyword"`
	PublicReq PublicRequest
}

type NetworkPrinterResponse struct {
	Printers   []common.Printer `json:"printers"`
	PublicResp PublicResponse
}

type PackageRequest struct {
	Pol       string `json:"pol"`
	PublicReq PublicRequest
}

type PackageResponse struct {
	Packages   []common.PackageInfo `json:"packages"`
	PublicResp PublicResponse
}

type PrinterDriverRequest struct {
	Printers  []string `json:"printers"`
	PublicReq PublicRequest
}

type PrinterDriverResponse struct {
	Pakcages   []common.PackageInfo `json:"packages"`
	PublicResp PublicResponse
}

var Client HTTPClient
var ACCESS_KEY string = "huDRV7GjHGT"
var ACCESS_SECRET string = "UUJ6JHEDDR90"

func (c *HTTPClient) SendRequest(params RequestParams) ([]byte, int, error) {
	// 1. 构建完整URL
	u, err := url.Parse(c.BaseURL + params.Path)
	if err != nil {
		return nil, 0, fmt.Errorf("URL解析失败: %w", err)
	}

	// 2. 处理查询参数
	q := u.Query()
	for k, v := range params.Query {
		q.Add(k, v)
	}
	u.RawQuery = q.Encode()

	// 3. 构建请求体
	var body io.Reader
	switch v := params.Body.(type) {
	case map[string]interface{}:
		if params.Method == "POST" || params.Method == "PUT" {
			jsonData, _ := json.Marshal(v)
			body = bytes.NewReader(jsonData)
			params.Headers["Content-Type"] = "application/json"
		}
	case url.Values: // Form数据
		body = strings.NewReader(v.Encode())
		params.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	case string:
		body = strings.NewReader(v)
	}

	// 4. 创建请求对象
	req, err := http.NewRequest(params.Method, u.String(), body)
	if err != nil {
		return nil, 0, fmt.Errorf("请求创建失败: %w", err)
	}

	// 5. 设置请求头
	for k, v := range params.Headers {
		req.Header.Set(k, v)
	}

	// 6. 配置超时客户端
	client := c.Client
	if params.Timeout > 0 {
		client = &http.Client{Timeout: params.Timeout}
	}

	// 7. 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("请求发送失败: %w", err)
	}
	defer resp.Body.Close()

	// 8. 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("响应读取失败: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

func GetOAServer(ip string) string {
	log.Printf("get OAServer from ip: %s", ip)

	var request OAServerRequest
	request.IP = ip

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()

	request.PublicReq = public

	m := structToMap(public, "Signature")
	public.Signature = generateSignature(m)

	//call api
	params := RequestParams{
		Method: "POST",
		Path:   "/oa",
		Body:   request,
	}

	body, status, err := Client.SendRequest(params)

	if err != nil {
		log.Printf("请求异常: %v", err)
		return ""
	}

	if status != http.StatusOK {
		log.Printf("业务错误: HTTP %d → %s", status, string(body))
		return ""
	}

	var result OAServerResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("JSON解析失败: %v", err)
		return ""
	}

	// common.CurrentOA = result.Name
	common.CurrentOA = "HPFS3OABAH3"
	return common.CurrentOA
}

func GetPrinterModels() []string {
	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()

	m := structToMap(public, "Signature")
	public.Signature = generateSignature(m)

	//call api
	params := RequestParams{
		Method: "GET",
		Path:   "/printer/models",
		Body:   public,
	}

	body, status, err := Client.SendRequest(params)

	if err != nil {
		log.Printf("请求异常: %v", err)
		return nil
	}

	if status != http.StatusOK {
		log.Printf("业务错误: HTTP %d → %s", status, string(body))
		return nil
	}

	var result PrinterModelResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("JSON解析失败: %v", err)
		return nil
	}

	// return result.Models
	return []string{"HP", "Epson", "Brother"}
}

func GetNetworkPinterList(keyword string) []common.Printer {
	var request NetworkPrinterRequest
	request.Keyword = keyword

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()

	request.PublicReq = public

	m := structToMap(public, "Signature")
	public.Signature = generateSignature(m)

	//call api
	params := RequestParams{
		Method: "GET",
		Path:   "/printer/network",
		Body:   request,
	}

	body, status, err := Client.SendRequest(params)

	if err != nil {
		log.Printf("请求异常: %v", err)
		return nil
	}

	if status != http.StatusOK {
		log.Printf("业务错误: HTTP %d → %s", status, string(body))
		return nil
	}

	var result NetworkPrinterResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("JSON解析失败: %v", err)
		return nil
	}

	// return result.Printers
	var printer0 common.Printer
	printer0.PolNo = "P87520"
	printer0.IP = "192.168.11.200"

	printers := []common.Printer{printer0}
	return printers
}

func GetAllPackages(pol string) []common.PackageInfo {
	var request PackageRequest
	request.Pol = pol

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()

	request.PublicReq = public

	m := structToMap(public, "Signature")
	public.Signature = generateSignature(m)

	//call api
	params := RequestParams{
		Method: "POST",
		Path:   "/packages",
		Body:   request,
	}

	body, status, err := Client.SendRequest(params)

	if err != nil {
		log.Printf("请求异常: %v", err)
		return nil
	}

	if status != http.StatusOK {
		log.Printf("业务错误: HTTP %d → %s", status, string(body))
		return nil
	}

	var result PackageResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("JSON解析失败: %v", err)
		return nil
	}

	// return result.Pakcages
	var package0 common.PackageInfo
	package0.AppName = "FireFox08"
	package0.AppType = "ForceApp"
	package0.ID = "0B9111FE-F2F3-05AF-C7CE-35F536E15FDD"
	package0.Status = "Waiting"

	var package1 common.PackageInfo
	package1.AppName = "HP - LaserJet 4 Plus"
	package1.AppType = "Printer"
	package1.ID = "A2A50752-E653-5772-CBB4-5D89E4CB8E76"
	package1.Status = "Waiting"

	packages := []common.PackageInfo{package0, package1}
	return packages
}

func GetPrinterDrivers(printers []common.Printer) []common.PackageInfo {
	var request PrinterDriverRequest
	for _, value := range printers {
		request.Printers = append(request.Printers, value.ID)
	}

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()

	request.PublicReq = public

	m := structToMap(public, "Signature")
	public.Signature = generateSignature(m)

	//call api
	params := RequestParams{
		Method: "POST",
		Path:   "/printer/driver",
		Body:   request,
	}

	body, status, err := Client.SendRequest(params)

	if err != nil {
		log.Printf("请求异常: %v", err)
		return nil
	}

	if status != http.StatusOK {
		log.Printf("业务错误: HTTP %d → %s", status, string(body))
		return nil
	}

	var result PackageResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("JSON解析失败: %v", err)
		return nil
	}

	// return result.Packages
	var package0 common.PackageInfo
	package0.AppName = "HP - LaserJet 4 Plus"
	package0.AppType = "Printer"
	package0.ID = "A2A50752-E653-5772-CBB4-5D89E4CB8E76"
	package0.Status = "Waiting"

	packages := []common.PackageInfo{package0}
	return packages
}
