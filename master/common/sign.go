package common

import (
	"crypto/md5"
	"encoding/hex"
	"net/url"
	"sort"
	"strings"
)

var (
	appInfo = map[string]string{
		"8500001": "6afbb38a5ebb1c95", // 云手机
		"8500002": "c1f10e979ea4a754", // 模拟器
		"8500003": "d1e71900688f7c1a", // 破解api
	}
)

// ValidSign 校验sign
func ValidSign(values url.Values, appkey, s string) bool {
	secKey := appInfo[appkey]
	cks := GenerateSign(values, secKey)
	return cks == s
}

// GenerateSign 生成url签名sign
func GenerateSign(values url.Values, secKey string) string {
	params := make([]string, 0)
	for k, v := range values {
		if len(k) > 0 && k != "s" && len(v) > 0 {
			if len(v[0]) == 0 {
				params = append(params, k+"-")
			} else {
				params = append(params, k+v[0])
			}
		}
	}
	sort.Sort(sort.StringSlice(params))
	params = append(params, secKey)
	stringToSign := strings.Join(params, "")
	s := md5.New()
	s.Write([]byte(stringToSign))
	return hex.EncodeToString(s.Sum(nil))
}
