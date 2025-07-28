package api

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"recovery-unit-deploy/service/common"
	"reflect"
	"sort"
	"strings"
	"time"
)

func getCurrentTimestamp() string {
	utcTime := time.Now().UTC()
	isoFormat := utcTime.Format("2006-01-02T15:04:05Z")
	return isoFormat
}

func flatten(prefix string, v reflect.Value, result map[string]string) {
	// 解引用指针和接口
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		// 处理 Map 类型
		for _, key := range v.MapKeys() {
			childKey := key.String()
			newPrefix := prefix
			if newPrefix != "" {
				newPrefix += "."
			}
			newPrefix += childKey
			flatten(newPrefix, v.MapIndex(key), result)
		}

	case reflect.Struct:
		// 处理结构体类型
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			// 跳过未导出字段
			if !field.IsExported() {
				continue
			}
			newPrefix := prefix
			if newPrefix != "" {
				newPrefix += "."
			}
			// 使用 JSON 标签名（如果有）
			tag := field.Tag.Get("json")
			if tag != "" && tag != "-" {
				newPrefix += tag
			}
			flatten(newPrefix, v.Field(i), result)
		}

	default:
		// 基本类型直接存储
		if prefix != "" && v.IsValid() && v.CanInterface() {
			result[prefix] = fmt.Sprintf("%v", v.Interface())
		}
	}
}

func structToMap(obj interface{}) map[string]string {
	m := make(map[string]string)
	flatten("", reflect.ValueOf(obj), m)
	return m
}

// 特殊字符替换逻辑（符合RFC3986）
func percentEncode(s string) string {
	// 先进行标准URL编码
	encoded := url.QueryEscape(s)
	// 替换特定字符
	encoded = strings.Replace(encoded, "+", "%20", -1)
	encoded = strings.Replace(encoded, "*", "%2A", -1)
	encoded = strings.Replace(encoded, "%7E", "~", -1)
	return encoded
}

func canonicalizeParams(params map[string]string) string {
	// 提取并排序参数名
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys) // 字典序排序[1](@ref)

	// 构建键值对
	var buf strings.Builder
	for _, k := range keys {
		// 跳过签名参数
		if k == "signature" {
			continue
		}
		// 双重编码：参数名和值均需编码[8](@ref)
		encodedKey := percentEncode(k)
		encodedVal := percentEncode(params[k])
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(encodedKey)
		buf.WriteByte('=')
		buf.WriteString(encodedVal)
	}
	return buf.String()
}

func buildStringToSign(method, canonicalQuery string) string {
	// 编码固定路径"/"
	encodedPath := percentEncode("/")
	// 编码规范化后的查询字符串
	encodedQuery := percentEncode(canonicalQuery)
	// 拼接签名字符串
	return strings.Join([]string{
		strings.ToUpper(method),
		encodedPath,
		encodedQuery,
	}, "&")
}

func computeSignature(secret, stringToSign string) string {
	// 转换密钥为字节
	key := []byte(secret)
	// 创建HMAC-SHA1哈希器[4](@ref)
	hasher := hmac.New(sha1.New, key)
	hasher.Write([]byte(stringToSign))
	// Base64编码结果
	return base64.StdEncoding.EncodeToString(hasher.Sum(nil))
}

func generateSignature(method string, body interface{}, key string, params map[string]string) string {
	if body != nil {
		// 序列化为JSON字节数组
		jsonData, err := json.Marshal(body)
		if err != nil {
			common.AppLogger.Error(fmt.Sprintf("json.Marshal: %v", err))
			return ""
		}
		// 转换为字符串
		jsonStr := string(jsonData)
		params["body"] = jsonStr

	}
	canonicalQuery := canonicalizeParams(params)

	stringToSign := buildStringToSign(method, canonicalQuery)

	// 3. 计算签名
	accessSecret := key
	signature := computeSignature(accessSecret+"&", stringToSign)
	return signature
}
