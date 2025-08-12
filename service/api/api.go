package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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
	Data []common.PrinterWithPackage `json:"data"`
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

type UploadInstallInfoResponse struct {
	PublicResponse
	Data string `json:"data"`
}

type UpdateInstallStatusResponse struct {
	PublicResponse
	Data string `json:"data"`
}

type UploadPCInfoRequest struct {
	PublicRequest
	Info common.DetailComputerInfo
}

type GetSeedLabelRequest struct {
	PublicRequest
	KBCode string `json:"kbcode"`
}

type GetSeedLabelResponse struct {
	PublicResponse
	Data []common.SeedLabelInfo `json:"data"`
}

type GetSeedTasksRequest struct {
	PublicRequest
	Seed string `json:"seedlabel"`
}

type GetSeedTasksResponse struct {
	PublicResponse
	Data []common.PackageInfo `json:"data"`
}

type GetAppVersionInfoRequest struct {
	PublicRequest
	Type string `json:"type"`
}

type GetAppVersionInfoResponse struct {
	PackageResponse
	Data common.AppVersionInfo `json:"data"`
}

type CheckSeedLabelRequest struct {
	PublicRequest
	Info common.SeedTimeInfo
}

type CheckSeedLabelResponse struct {
	PublicResponse
	Data bool `json:"data"`
}

type GetCodesByGroupRequest struct {
	PublicRequest
	Group string `json:"group"`
}

type GetCodesByGroupResponse struct {
	PublicResponse
	Data []common.GroupCode `json:"data"`
}

var Client *APIClient
var ACCESS_KEY string = "b3fd07fc731146c7bb5bdc953da719d0"
var ACCESS_SECRET string = "iSkv1/0X/CVk49l+jloSCv7eTGWTFrBZ"

var ACCESS_KEY_RU string = "0efdba78bccf45f496c27e70a7442dd9"
var ACCESS_SECRET_RU string = "jvOIJ/VTSZLPwgpEfOHdZWDBPNiz1xvv"

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

// 修改后的CallAPI方法
func (c *APIClient) CallAPI(
	method string,
	endpoint string,
	body interface{},
	pathParams map[string]interface{},
	queryParams map[string]string,
) ([]byte, int, error) {
	// 替换路径参数
	for key, value := range pathParams {
		placeholder := "{" + key + "}"
		endpoint = strings.Replace(endpoint, placeholder, fmt.Sprintf("%v", value), -1)
	}

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
	common.AppLogger.Info(fmt.Sprintf("get OAServer from ip: %s", ip))

	var request OAServerRequest
	request.IP = ip

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()

	request.PublicReq = public

	m := structToMap(request)
	public.Signature = generateSignature(http.MethodGet, nil, ACCESS_SECRET, m)

	m["signature"] = public.Signature

	data, status, err := Client.CallAPI(http.MethodGet, "/deploy/getOAServer", nil, nil, m)

	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("请求异常: %v", err))
		return ""
	}

	if status != http.StatusOK {
		common.AppLogger.Error(fmt.Sprintf("业务错误: HTTP %d → %s", status, string(data)))
		return ""
	}

	var result OAServerResponse
	if err := json.Unmarshal(data, &result); err != nil {
		common.AppLogger.Error(fmt.Sprintf("JSON解析失败: %v", err))
		return ""
	}

	common.CurrentOA = result.Data
	common.CurrentComputerInfo.OA = result.Data.ServerName
	// common.CurrentOA = "192.168.49.48"
	return common.CurrentComputerInfo.OA
}

func GetPrinterModels() []common.PrinterModel {
	common.AppLogger.Info("get Printer models.")

	var models []common.PrinterModel

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()

	m := structToMap(public)
	public.Signature = generateSignature(http.MethodGet, nil, ACCESS_SECRET, m)

	m["signature"] = public.Signature

	data, status, err := Client.CallAPI(http.MethodGet, "/deploy/getPrinterBrands", nil, nil, m)

	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("请求异常: %v", err))
		return models
	}

	if status != http.StatusOK {
		common.AppLogger.Error(fmt.Sprintf("业务错误: HTTP %d → %s", status, string(data)))
		return models
	}

	var result PrinterModelResponse
	if err := json.Unmarshal(data, &result); err != nil {
		common.AppLogger.Error(fmt.Sprintf("JSON解析失败: %v", err))
		return models
	}

	models = result.Data

	return models
}

func GetSelectedLocalPrinterDrivers(id string) []common.PackageInfo {
	common.AppLogger.Info("get driver list of printer from model.")

	var drivers []common.PackageInfo

	var request LocalPrinterDriverRequest
	request.ID = id

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()
	request.PublicRequest = public

	m := structToMap(request)
	request.Signature = generateSignature(http.MethodGet, nil, ACCESS_SECRET, m)

	m["signature"] = request.Signature

	data, status, err := Client.CallAPI(http.MethodGet, "/deploy/getAppsByBrand", nil, nil, m)

	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("请求异常: %v", err))
		return drivers
	}

	if status != http.StatusOK {
		common.AppLogger.Error(fmt.Sprintf("业务错误: HTTP %d → %s", status, string(data)))
		return drivers
	}

	var result LocalPrinterDriverResponse
	if err := json.Unmarshal(data, &result); err != nil {
		common.AppLogger.Error(fmt.Sprintf("JSON解析失败: %v", err))
		return drivers
	}

	return result.Data
}

func GetNetworkPinterList(keyword string) []common.PrinterWithPackage {
	common.AppLogger.Info("get network printer list.")

	var printers []common.PrinterWithPackage

	var request NetworkPrinterRequest
	request.Keyword = keyword

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()

	request.PublicRequest = public

	m := structToMap(request)
	request.Signature = generateSignature(http.MethodGet, nil, ACCESS_SECRET, m)

	m["signature"] = request.Signature

	data, status, err := Client.CallAPI(http.MethodGet, "/deploy/getNetworkPrinters", nil, nil, m)

	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("请求异常: %v", err))
		return printers
	}

	if status != http.StatusOK {
		common.AppLogger.Error(fmt.Sprintf("业务错误: HTTP %d → %s", status, string(data)))
		return printers
	}

	common.AppLogger.Info(fmt.Sprintf("server returns network printer : %s", string(data)))

	var result NetworkPrinterResponse
	if err := json.Unmarshal(data, &result); err != nil {
		common.AppLogger.Error(fmt.Sprintf("JSON解析失败: %v", err))
		return printers
	}

	common.AppLogger.Info(fmt.Sprintf("Unmarshal network printer : %v", result))

	return result.Data
}

func GetAllPackages(pol string, seed string) []common.PackageInfo {
	common.AppLogger.Info("get all application that the computer can install by pol and seed.")

	var apps []common.PackageInfo

	var request PackageRequest
	request.Pol = pol
	request.Seed = seed

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()
	request.PublicRequest = public

	m := structToMap(request)
	request.Signature = generateSignature(http.MethodGet, nil, ACCESS_SECRET, m)

	m["signature"] = request.Signature

	data, status, err := Client.CallAPI(http.MethodGet, "/deploy/getInstallableApps", nil, nil, m)

	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("请求异常: %v", err))
		return apps
	}

	if status != http.StatusOK {
		common.AppLogger.Error(fmt.Sprintf("业务错误: HTTP %d → %s", status, string(data)))
		return apps
	}

	var result PackageResponse
	if err := json.Unmarshal(data, &result); err != nil {
		common.AppLogger.Error(fmt.Sprintf("JSON解析失败: %v", err))
		return apps
	}

	return result.Data
}

func GetNetworkPrinterDrivers(printers []common.Printer) []common.PackageInfo {
	common.AppLogger.Info("get selected network printer drivers.")
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
	request.Signature = generateSignature(http.MethodGet, nil, ACCESS_SECRET, m)

	m["signature"] = request.Signature

	data, status, err := Client.CallAPI(http.MethodGet, "/deploy/getAppsByIds", nil, nil, m)

	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("请求异常: %v", err))
		return apps
	}

	if status != http.StatusOK {
		common.AppLogger.Error(fmt.Sprintf("业务错误: HTTP %d → %s", status, string(data)))
		return apps
	}

	var result NetworkPrinterDriverResponse
	if err := json.Unmarshal(data, &result); err != nil {
		common.AppLogger.Error(fmt.Sprintf("JSON解析失败: %v", err))
		return apps
	}

	return result.Data
}

func UploadInstallInfo(info common.InstallInfo) (string, error) {
	common.AppLogger.Info("upload install info.")

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()

	m := structToMap(public)

	public.Signature = generateSignature(http.MethodPost, info, ACCESS_SECRET, m)
	m["signature"] = public.Signature

	delete(m, "body")

	data, status, err := Client.CallAPI(http.MethodPost, "/deploy/uploadInstallProgress", info, nil, m)

	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("请求异常: %v", err))
		return "", err
	}

	if status != http.StatusOK {
		common.AppLogger.Error(fmt.Sprintf("业务错误: HTTP %d → %s", status, string(data)))
		return "", fmt.Errorf("业务错误: HTTP %d → %s", status, string(data))
	}

	var result UploadInstallInfoResponse
	if err := json.Unmarshal(data, &result); err != nil {
		common.AppLogger.Error(fmt.Sprintf("JSON解析失败: %v", err))
		return "", err
	}

	common.AppLogger.Info(fmt.Sprintf("Unmarshal UpdateInstallResponse : %v", result))
	return result.Data, nil
}

func StartInstall(id common.AppStatus) {
	common.AppLogger.Info("start install.")

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()

	m := structToMap(public)

	public.Signature = generateSignature(http.MethodPost, id, ACCESS_SECRET, m)
	m["signature"] = public.Signature

	delete(m, "body")

	data, status, err := Client.CallAPI(http.MethodPost, "/deploy/startInstall", id, nil, m)

	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("请求异常: %v", err))
		return
	}

	if status != http.StatusOK {
		common.AppLogger.Error(fmt.Sprintf("业务错误: HTTP %d → %s", status, string(data)))
		return
	}

	var result UpdateInstallStatusResponse
	if err := json.Unmarshal(data, &result); err != nil {
		common.AppLogger.Error(fmt.Sprintf("JSON解析失败: %v", err))
		return
	}

	common.AppLogger.Info(fmt.Sprintf("Unmarshal UpdateInstallStatusResponse : %v", result))
}

func InstallationSuccess(id common.AppStatus) {
	common.AppLogger.Info("install success.")

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()

	m := structToMap(public)

	public.Signature = generateSignature(http.MethodPost, id, ACCESS_SECRET, m)
	m["signature"] = public.Signature

	delete(m, "body")

	data, status, err := Client.CallAPI(http.MethodPost, "/deploy/installationSuccess", id, nil, m)

	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("请求异常: %v", err))
		return
	}

	if status != http.StatusOK {
		common.AppLogger.Error(fmt.Sprintf("业务错误: HTTP %d → %s", status, string(data)))
		return
	}

	var result UpdateInstallStatusResponse
	if err := json.Unmarshal(data, &result); err != nil {
		common.AppLogger.Error(fmt.Sprintf("JSON解析失败: %v", err))
		return
	}

	common.AppLogger.Info(fmt.Sprintf("Unmarshal UpdateInstallStatusResponse : %v", result))
}

func InstallationFailed(id common.FailedAppStatus) {
	common.AppLogger.Info("install failed.")

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()

	m := structToMap(public)

	public.Signature = generateSignature(http.MethodPost, id, ACCESS_SECRET, m)
	m["signature"] = public.Signature

	delete(m, "body")

	data, status, err := Client.CallAPI(http.MethodPost, "/deploy/installationFailed", id, nil, m)

	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("请求异常: %v", err))
		return
	}

	if status != http.StatusOK {
		common.AppLogger.Error(fmt.Sprintf("业务错误: HTTP %d → %s", status, string(data)))
		return
	}

	var result UpdateInstallStatusResponse
	if err := json.Unmarshal(data, &result); err != nil {
		common.AppLogger.Error(fmt.Sprintf("JSON解析失败: %v", err))
		return
	}

	common.AppLogger.Info(fmt.Sprintf("Unmarshal UpdateInstallStatusResponse : %v", result))
}

func CancelInstallation(id common.AppStatus) {
	common.AppLogger.Info("cancel install.")

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()

	m := structToMap(public)

	public.Signature = generateSignature(http.MethodPost, id, ACCESS_SECRET, m)
	m["signature"] = public.Signature

	delete(m, "body")

	data, status, err := Client.CallAPI(http.MethodPost, "/deploy/cancelInstall", id, nil, m)

	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("请求异常: %v", err))
		return
	}

	if status != http.StatusOK {
		common.AppLogger.Error(fmt.Sprintf("业务错误: HTTP %d → %s", status, string(data)))
		return
	}

	var result UpdateInstallStatusResponse
	if err := json.Unmarshal(data, &result); err != nil {
		common.AppLogger.Error(fmt.Sprintf("JSON解析失败: %v", err))
		return
	}

	common.AppLogger.Info(fmt.Sprintf("Unmarshal CancelInstallation : %v", result))
}

func UploadPCInfo(info common.DetailComputerInfo) {
	common.AppLogger.Info("upload pc info.")

	var request UploadPCInfoRequest
	request.Info = info

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()
	request.PublicRequest = public

	m := structToMap(request)
	request.Signature = generateSignature(http.MethodPost, info, ACCESS_SECRET, m)

	m["signature"] = request.Signature

	data, status, err := Client.CallAPI(http.MethodPost, "/deploy/uploadPcInfo", info, nil, m)

	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("请求异常: %v", err))
		return
	}

	if status != http.StatusOK {
		common.AppLogger.Error(fmt.Sprintf("业务错误: HTTP %d → %s", status, string(data)))
		return
	}

	var result PublicResponse
	if err := json.Unmarshal(data, &result); err != nil {
		common.AppLogger.Error(fmt.Sprintf("JSON解析失败: %v", err))
		return
	}

	if result.Code == 0 {
		common.AppLogger.Info("Upload PC Info Successfully.")
	} else {
		common.AppLogger.Error(fmt.Sprintf("Upload PC Info Failed. code: %d, msg: %s", result.Code, result.Message))
	}
}

func GetSeedLabel(kbcode string) common.SeedLabelInfo {
	common.AppLogger.Info("get seedlabel from kbcode")

	var seedlabel common.SeedLabelInfo

	var request GetSeedLabelRequest
	request.KBCode = kbcode

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()
	request.PublicRequest = public

	m := structToMap(request)
	request.Signature = generateSignature(http.MethodGet, nil, ACCESS_SECRET, m)

	m["signature"] = request.Signature

	data, status, err := Client.CallAPI(http.MethodGet, "/deploy/getSeedInfoByKb", nil, nil, m)

	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("请求异常: %v", err))
		return seedlabel
	}

	if status != http.StatusOK {
		common.AppLogger.Error(fmt.Sprintf("业务错误: HTTP %d → %s", status, string(data)))
		return seedlabel
	}

	var result GetSeedLabelResponse
	if err := json.Unmarshal(data, &result); err != nil {
		common.AppLogger.Error(fmt.Sprintf("JSON解析失败: %v", err))
		return seedlabel
	}

	for _, sl := range result.Data {
		filename := fmt.Sprintf("C:\\%s.seedlabel.txt", sl.SeedLabel)
		_, err := os.Stat(filename)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}

		seedlabel = sl
		common.CurrentComputerInfo.Seed = sl.SeedLabel
		common.CurrentSeed = sl
	}

	return seedlabel
}

func GetSeedTasks(seed string) []common.PackageInfo {
	common.AppLogger.Info("get tasks from seedlabel")

	var tasks []common.PackageInfo

	var request GetSeedTasksRequest
	request.Seed = seed

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()
	request.PublicRequest = public

	m := structToMap(request)

	request.Signature = generateSignature(http.MethodGet, nil, ACCESS_SECRET, m)

	m["signature"] = request.Signature

	data, status, err := Client.CallAPI(http.MethodGet, "/deploy/getTaskApps", nil, nil, m)

	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("请求异常: %v", err))
		return tasks
	}

	if status != http.StatusOK {
		common.AppLogger.Error(fmt.Sprintf("业务错误: HTTP %d → %s", status, string(data)))
		return tasks
	}

	var result GetSeedTasksResponse
	if err := json.Unmarshal(data, &result); err != nil {
		common.AppLogger.Error(fmt.Sprintf("JSON解析失败: %v", err))
		return tasks
	}

	tasks = append(tasks, result.Data...)

	return tasks
}

func GetAppVersionInfo(t string) common.AppVersionInfo {
	common.AppLogger.Info("get newest app verion info")

	var appInfo common.AppVersionInfo
	var request GetAppVersionInfoRequest
	request.Type = t

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY_RU
	public.Timestamp = getCurrentTimestamp()
	request.PublicRequest = public

	m := structToMap(request)
	request.Signature = generateSignature(http.MethodPost, t, ACCESS_SECRET_RU, m)

	m["signature"] = request.Signature

	data, status, err := Client.CallAPI(http.MethodPost, "/ru/version/info", t, nil, m)

	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("请求异常: %v", err))
		return appInfo
	}

	if status != http.StatusOK {
		common.AppLogger.Error(fmt.Sprintf("业务错误: HTTP %d → %s", status, string(data)))
		return appInfo
	}

	var result GetAppVersionInfoResponse
	if err := json.Unmarshal(data, &result); err != nil {
		common.AppLogger.Error(fmt.Sprintf("JSON解析失败: %v", err))
		return appInfo
	}

	appInfo = result.Data
	return appInfo
}

func CheckSeedLabel(seed string, modTime string, createTime string) bool {
	common.AppLogger.Info("check seedlabel by created time and modified time")

	var request CheckSeedLabelRequest
	var seedTimeInfo common.SeedTimeInfo
	seedTimeInfo.SeedLabel = seed
	seedTimeInfo.CreateTime = createTime
	seedTimeInfo.UpdateTime = modTime
	request.Info = seedTimeInfo

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()
	request.PublicRequest = public

	m := structToMap(request)
	request.Signature = generateSignature(http.MethodPost, seedTimeInfo, ACCESS_SECRET, m)

	m["signature"] = request.Signature

	data, status, err := Client.CallAPI(http.MethodPost, "/deploy/checkSeedFile", seedTimeInfo, nil, m)

	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("请求异常: %v", err))
		return false
	}

	if status != http.StatusOK {
		common.AppLogger.Error(fmt.Sprintf("业务错误: HTTP %d → %s", status, string(data)))
		return false
	}

	var result CheckSeedLabelResponse
	if err := json.Unmarshal(data, &result); err != nil {
		common.AppLogger.Error(fmt.Sprintf("JSON解析失败: %v", err))
		return false
	}

	return result.Data
}

func GetCodesByGroup(group string) []common.GroupCode {
	common.AppLogger.Info("get code from group")

	var codes []common.GroupCode

	var request GetCodesByGroupRequest
	request.Group = group

	var public PublicRequest
	public.AccessKeyId = ACCESS_KEY
	public.Timestamp = getCurrentTimestamp()
	request.PublicRequest = public

	m := structToMap(request)

	request.Signature = generateSignature(http.MethodGet, nil, ACCESS_SECRET, m)

	m["signature"] = request.Signature

	data, status, err := Client.CallAPI(http.MethodGet, "/deploy/getCodesByGroup", nil, nil, m)

	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("请求异常: %v", err))
		return codes
	}

	if status != http.StatusOK {
		common.AppLogger.Error(fmt.Sprintf("业务错误: HTTP %d → %s", status, string(data)))
		return codes
	}

	var result GetCodesByGroupResponse
	if err := json.Unmarshal(data, &result); err != nil {
		common.AppLogger.Error(fmt.Sprintf("JSON解析失败: %v", err))
		return codes
	}

	return result.Data
}
