package codeutil

import (
	"bytes"
	"encoding/json"
)

// GetPrettyJsonStr json format
func GetPrettyJsonStr(obj interface{}) string {
	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	jsonEncoder.SetIndent("", "	")
	_ = jsonEncoder.Encode(obj)
	return bf.String()
}

// GetUglyJsonStr json format
func GetUglyJsonStr(obj interface{}) string {
	ans, _ := json.Marshal(obj)
	return string(ans)
}
