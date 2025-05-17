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
	Signature   string `json:"signature"`
	Timestamp   string `json:"timestamp"`
	AccessKeyId string `json:"accessKeyId"`
}

type PublicResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type OAServerRequest struct {
	PublicReq PublicRequest
	IP        string `json:"ip"`
}

type OAServerResponse struct {
	PublicResponse
	Data common.OAServer `json:"data"`
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

var Client *APIClient
var ACCESS_KEY string = "b3fd07fc731146c7bb5bdc953da719d0"
var ACCESS_SECRET string = "iSkv1/0X/CVk49l+jloSCv7eTGWTFrBZ"

type APIClient struct {
	BaseURL    string
	Headers    map[string]string
	Timeout    time.Duration
	HTTPClient *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		BaseURL:    baseURL,
		Headers:    make(map[string]string),
		Timeout:    10 * time.Second,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *APIClient) CallAPI(
	method string,
	endpoint string,
	body interface{},
	queryParams map[string]string,
) ([]byte, int, error) {
	// 构建完整URL
	u, err := url.Parse(c.BaseURL + endpoint)
	if err != nil {
		return nil, 0, fmt.Errorf("URL解析失败: %w", err)
	}

	// 添加查询参数
	q := u.Query()
	for k, v := range queryParams {
		q.Add(k, v)
	}
	u.RawQuery = q.Encode()

	// 序列化请求体
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("JSON序列化失败: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	// 创建请求对象
	req, err := http.NewRequest(method, u.String(), reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("请求创建失败: %w", err)
	}

	// 设置请求头
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// 发送请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("网络请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("响应读取失败: %w", err)
	}

	// 检查状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return respBody, resp.StatusCode, fmt.Errorf("异常状态码: %d", resp.StatusCode)
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

	m := structToMap(request)
	public.Signature = generateSignature(m)

	m["signature"] = public.Signature

	data, status, err := Client.CallAPI(http.MethodGet, "/deploy/getOAServer", nil, m)

	if err != nil {
		log.Printf("请求异常: %v", err)
		return ""
	}

	if status != http.StatusOK {
		log.Printf("业务错误: HTTP %d → %s", status, string(data))
		return ""
	}

	var result OAServerResponse
	if err := json.Unmarshal(data, &result); err != nil {
		log.Printf("JSON解析失败: %v", err)
		return ""
	}

	common.CurrentOA = result.Data.ServerName
	// common.CurrentOA = "192.168.49.48"
	return common.CurrentOA
}

func GetPrinterModels() []string {
	// var public PublicRequest
	// public.AccessKeyId = ACCESS_KEY
	// public.Timestamp = getCurrentTimestamp()

	// m := structToMap(public, "Signature")
	// public.Signature = generateSignature(m)

	// //call api
	// params := RequestParams{
	// 	Method: "GET",
	// 	Path:   "/printer/models",
	// 	Body:   public,
	// }

	// body, status, err := Client.SendRequest(params)

	// if err != nil {
	// 	log.Printf("请求异常: %v", err)
	// 	return nil
	// }

	// if status != http.StatusOK {
	// 	log.Printf("业务错误: HTTP %d → %s", status, string(body))
	// 	return nil
	// }

	// var result PrinterModelResponse
	// if err := json.Unmarshal(body, &result); err != nil {
	// 	log.Printf("JSON解析失败: %v", err)
	// 	return nil
	// }

	// return result.Models
	return []string{"HP", "Epson", "Brother"}
}

func GetNetworkPinterList(keyword string) []common.Printer {
	// var request NetworkPrinterRequest
	// request.Keyword = keyword

	// var public PublicRequest
	// public.AccessKeyId = ACCESS_KEY
	// public.Timestamp = getCurrentTimestamp()

	// request.PublicReq = public

	// m := structToMap(public, "Signature")
	// public.Signature = generateSignature(m)

	// //call api
	// params := RequestParams{
	// 	Method: "GET",
	// 	Path:   "/printer/network",
	// 	Body:   request,
	// }

	// body, status, err := Client.SendRequest(params)

	// if err != nil {
	// 	log.Printf("请求异常: %v", err)
	// 	return nil
	// }

	// if status != http.StatusOK {
	// 	log.Printf("业务错误: HTTP %d → %s", status, string(body))
	// 	return nil
	// }

	// var result NetworkPrinterResponse
	// if err := json.Unmarshal(body, &result); err != nil {
	// 	log.Printf("JSON解析失败: %v", err)
	// 	return nil
	// }

	// return result.Printers
	var printer0 common.Printer
	printer0.PolNo = "P87520"
	printer0.IP = "192.168.11.200"

	printers := []common.Printer{printer0}
	return printers
}

func GetAllPackages(pol string) []common.PackageInfo {
	// var request PackageRequest
	// request.Pol = pol

	// var public PublicRequest
	// public.AccessKeyId = ACCESS_KEY
	// public.Timestamp = getCurrentTimestamp()

	// request.PublicReq = public

	// m := structToMap(public, "Signature")
	// public.Signature = generateSignature(m)

	// //call api
	// params := RequestParams{
	// 	Method: "POST",
	// 	Path:   "/packages",
	// 	Body:   request,
	// }

	// body, status, err := Client.SendRequest(params)

	// if err != nil {
	// 	log.Printf("请求异常: %v", err)
	// 	return nil
	// }

	// if status != http.StatusOK {
	// 	log.Printf("业务错误: HTTP %d → %s", status, string(body))
	// 	return nil
	// }

	// var result PackageResponse
	// if err := json.Unmarshal(body, &result); err != nil {
	// 	log.Printf("JSON解析失败: %v", err)
	// 	return nil
	// }

	// return result.Pakcages
	var package0 common.PackageInfo
	package0.AppName = "FireFox08"
	package0.AppType = "ForceApp"
	package0.ID = "0B9111FE-F2F3-05AF-C7CE-35F536E15FDD"
	package0.Status = "Waiting"
	package0.Path = "Package"
	package0.WinFile = "jobs/adobereader.bat"

	var package1 common.PackageInfo
	package1.AppName = "HP - LaserJet 4 Plus"
	package1.AppType = "Printer"
	package1.ID = "A2A50752-E653-5772-CBB4-5D89E4CB8E76"
	package1.Status = "Waiting"

	packages := []common.PackageInfo{package0, package1}
	return packages
}

func GetPrinterDrivers(printers []common.Printer) []common.PackageInfo {
	// var request PrinterDriverRequest
	// for _, value := range printers {
	// 	request.Printers = append(request.Printers, value.ID)
	// }

	// var public PublicRequest
	// public.AccessKeyId = ACCESS_KEY
	// public.Timestamp = getCurrentTimestamp()

	// request.PublicReq = public

	// m := structToMap(public, "Signature")
	// public.Signature = generateSignature(m)

	// //call api
	// params := RequestParams{
	// 	Method: "POST",
	// 	Path:   "/printer/driver",
	// 	Body:   request,
	// }

	// body, status, err := Client.SendRequest(params)

	// if err != nil {
	// 	log.Printf("请求异常: %v", err)
	// 	return nil
	// }

	// if status != http.StatusOK {
	// 	log.Printf("业务错误: HTTP %d → %s", status, string(body))
	// 	return nil
	// }

	// var result PackageResponse
	// if err := json.Unmarshal(body, &result); err != nil {
	// 	log.Printf("JSON解析失败: %v", err)
	// 	return nil
	// }

	// return result.Packages
	var package0 common.PackageInfo
	package0.AppName = "HP - LaserJet 4 Plus"
	package0.AppType = "Printer"
	package0.ID = "A2A50752-E653-5772-CBB4-5D89E4CB8E76"
	package0.Status = "Waiting"

	packages := []common.PackageInfo{package0}
	return packages
}
