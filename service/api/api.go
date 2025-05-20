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
	PublicResponse
	Data []common.PrinterModel `json:"data"`
}

type NetworkPrinterRequest struct {
	PublicRequest
	Keyword string `json:"searchkey"`
}

type NetworkPrinterResponse struct {
	PublicResponse
	Data []common.Printer `json:"data"`
}

type PackageRequest struct {
	PublicRequest
	Pol  string `json:"pol"`
	Seed string `json:"seedlabel"`
}

type PackageResponse struct {
	PublicResponse
	Data []common.PackageInfo `json:"data"`
}

type LocalPrinterDriverRequest struct {
	PublicRequest
	ID string `json:"brandid"`
}

type LocalPrinterDriverResponse struct {
	PublicResponse
	Data []common.PackageInfo `json:"data"`
}

type NetworkPrinterDriverRequest struct {
	PublicRequest
	AppIds string `json:"appids"`
}

type NetworkPrinterDriverResponse struct {
	PublicResponse
	Data []common.PackageInfo `json:"data"`
}

type UploadInstallInfoRequest struct {
	PublicRequest
	Info common.InstallInfo `json:"dto"`
}

type UploadInstallInfoResponse struct {
	PublicResponse
	Data string `json:"data"`
}

type UpdateInstallStatusRequest struct {
	PublicRequest
	App common.AppId `json:"dto"`
}

type UpdateInstallStatusResponse struct {
	PublicResponse
	Data string `json:"data"`
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

	common.CurrentComputerInfo.OA = result.Data.ServerName
	// common.CurrentOA = "192.168.49.48"
	return common.CurrentComputerInfo.OA
}

func GetPrinterModels() []common.PrinterModel {
	log.Printf("get Printer models.")

	var models []common.PrinterModel

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()

	m := structToMap(public)
	public.Signature = generateSignature(m)

	m["signature"] = public.Signature

	data, status, err := Client.CallAPI(http.MethodGet, "/deploy/getPrinterBrands", nil, m)

	if err != nil {
		log.Printf("请求异常: %v", err)
		return models
	}

	if status != http.StatusOK {
		log.Printf("业务错误: HTTP %d → %s", status, string(data))
		return models
	}

	var result PrinterModelResponse
	if err := json.Unmarshal(data, &result); err != nil {
		log.Printf("JSON解析失败: %v", err)
		return models
	}

	models = result.Data

	return models
}

func GetSelectedLocalPrinterDrivers(id string) []common.PackageInfo {
	log.Printf("get driver list of printer from model.")

	var drivers []common.PackageInfo

	var request LocalPrinterDriverRequest
	request.ID = id

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()
	request.PublicRequest = public

	m := structToMap(request)
	request.Signature = generateSignature(m)

	m["signature"] = request.Signature

	data, status, err := Client.CallAPI(http.MethodGet, "/deploy/getAppsByBrand", nil, m)

	if err != nil {
		log.Printf("请求异常: %v", err)
		return drivers
	}

	if status != http.StatusOK {
		log.Printf("业务错误: HTTP %d → %s", status, string(data))
		return drivers
	}

	var result LocalPrinterDriverResponse
	if err := json.Unmarshal(data, &result); err != nil {
		log.Printf("JSON解析失败: %v", err)
		return drivers
	}

	return result.Data
}

func GetNetworkPinterList(keyword string) []common.Printer {
	log.Printf("get network printer list.")

	var printers []common.Printer

	var request NetworkPrinterRequest
	request.Keyword = keyword

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()

	request.PublicRequest = public

	m := structToMap(request)
	request.Signature = generateSignature(m)

	m["signature"] = request.Signature

	data, status, err := Client.CallAPI(http.MethodGet, "/deploy/getNetworkPrinters", nil, m)

	if err != nil {
		log.Printf("请求异常: %v", err)
		return printers
	}

	if status != http.StatusOK {
		log.Printf("业务错误: HTTP %d → %s", status, string(data))
		return printers
	}

	log.Printf("server returns network printer : %s", string(data))

	var result NetworkPrinterResponse
	if err := json.Unmarshal(data, &result); err != nil {
		log.Printf("JSON解析失败: %v", err)
		return printers
	}

	log.Printf("Unmarshal network printer : %v", result)

	return result.Data
}

func GetAllPackages(pol string, seed string) []common.PackageInfo {
	log.Printf("get all application that the computer can install by pol and seed.")

	var apps []common.PackageInfo

	var request PackageRequest
	request.Pol = pol
	request.Seed = seed

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()
	request.PublicRequest = public

	m := structToMap(request)
	request.Signature = generateSignature(m)

	m["signature"] = request.Signature

	data, status, err := Client.CallAPI(http.MethodGet, "/deploy/getInstallableApps", nil, m)

	if err != nil {
		log.Printf("请求异常: %v", err)
		return apps
	}

	if status != http.StatusOK {
		log.Printf("业务错误: HTTP %d → %s", status, string(data))
		return apps
	}

	var result PackageResponse
	if err := json.Unmarshal(data, &result); err != nil {
		log.Printf("JSON解析失败: %v", err)
		return apps
	}

	return result.Data
}

func GetNetworkPrinterDrivers(printers []common.Printer) []common.PackageInfo {
	log.Printf("get selected network printer drivers.")
	var apps []common.PackageInfo

	var request NetworkPrinterDriverRequest
	var ids []string
	for _, value := range printers {
		ids = append(ids, value.AppId)
	}

	request.AppIds = strings.Join(ids, ",")

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()
	request.PublicRequest = public

	m := structToMap(request)
	request.Signature = generateSignature(m)

	m["signature"] = request.Signature

	data, status, err := Client.CallAPI(http.MethodGet, "/deploy/getAppsByIds", nil, m)

	if err != nil {
		log.Printf("请求异常: %v", err)
		return apps
	}

	if status != http.StatusOK {
		log.Printf("业务错误: HTTP %d → %s", status, string(data))
		return apps
	}

	var result NetworkPrinterDriverResponse
	if err := json.Unmarshal(data, &result); err != nil {
		log.Printf("JSON解析失败: %v", err)
		return apps
	}

	return result.Data
}

func UploadInstallInfo(info common.InstallInfo) error {
	log.Printf("upload install info.")

	var request UploadInstallInfoRequest
	request.Info = info

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()

	request.PublicRequest = public

	m := structToMap(request)
	request.Signature = generateSignature(m)

	m["signature"] = request.Signature

	data, status, err := Client.CallAPI(http.MethodPost, "/deploy/uploadInstallProgress", nil, m)

	if err != nil {
		log.Printf("请求异常: %v", err)
		return err
	}

	if status != http.StatusOK {
		log.Printf("业务错误: HTTP %d → %s", status, string(data))
		return fmt.Errorf("业务错误: HTTP %d → %s", status, string(data))
	}

	var result UploadInstallInfoResponse
	if err := json.Unmarshal(data, &result); err != nil {
		log.Printf("JSON解析失败: %v", err)
		return err
	}

	log.Printf("Unmarshal UpdateInstallResponse : %v", result)
	return nil
}

func StartInstall(id common.AppId) {
	log.Printf("start install.")

	var request UpdateInstallStatusRequest
	request.App = id

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()

	request.PublicRequest = public

	m := structToMap(request)
	request.Signature = generateSignature(m)

	m["signature"] = request.Signature

	data, status, err := Client.CallAPI(http.MethodPost, "/deploy/startInstall", nil, m)

	if err != nil {
		log.Printf("请求异常: %v", err)
		return
	}

	if status != http.StatusOK {
		log.Printf("业务错误: HTTP %d → %s", status, string(data))
		return
	}

	var result UpdateInstallStatusResponse
	if err := json.Unmarshal(data, &result); err != nil {
		log.Printf("JSON解析失败: %v", err)
		return
	}

	log.Printf("Unmarshal UpdateInstallResponse : %v", result)
}

func InstallationSuccess(id common.AppId) {
	log.Printf("install success.")

	var request UpdateInstallStatusRequest
	request.App = id

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()

	request.PublicRequest = public

	m := structToMap(request)
	request.Signature = generateSignature(m)

	m["signature"] = request.Signature

	data, status, err := Client.CallAPI(http.MethodPost, "/deploy/installationSuccess", nil, m)

	if err != nil {
		log.Printf("请求异常: %v", err)
		return
	}

	if status != http.StatusOK {
		log.Printf("业务错误: HTTP %d → %s", status, string(data))
		return
	}

	var result UpdateInstallStatusResponse
	if err := json.Unmarshal(data, &result); err != nil {
		log.Printf("JSON解析失败: %v", err)
		return
	}

	log.Printf("Unmarshal UpdateInstallResponse : %v", result)
}

func InstallationFailed(id common.AppId) {
	log.Printf("install success.")

	var request UpdateInstallStatusRequest
	request.App = id

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()

	request.PublicRequest = public

	m := structToMap(request)
	request.Signature = generateSignature(m)

	m["signature"] = request.Signature

	data, status, err := Client.CallAPI(http.MethodPost, "/deploy/installationFailed", nil, m)

	if err != nil {
		log.Printf("请求异常: %v", err)
		return
	}

	if status != http.StatusOK {
		log.Printf("业务错误: HTTP %d → %s", status, string(data))
		return
	}

	var result UpdateInstallStatusResponse
	if err := json.Unmarshal(data, &result); err != nil {
		log.Printf("JSON解析失败: %v", err)
		return
	}

	log.Printf("Unmarshal UpdateInstallResponse : %v", result)
}
