package api

import (
	"time"

	"github.com/fatih/structs"
)

func getCurrentTimestamp() string {
	utcTime := time.Now().UTC()
	isoFormat := utcTime.Format("2006-01-02T15:04:05Z")
	return isoFormat
}

func structToMap(obj interface{}, ignore string) map[string]interface{} {
	m := structs.Map(obj)
	delete(m, ignore)

	return m
}

func generateSignature(params map[string]interface{}) string {
	var signature string

	return signature
}
