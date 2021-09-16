package utils

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"switch-polls-backend/config"
)

type InvalidJson struct{}

func (e *InvalidJson) Error() string {
	return "JSON is not valid"
}

func PrepareResponse(data interface{}) ([]byte, error) {
	dataJs := strings.Builder{}
	js, err := json.Marshal(data)

	if err != nil {
		return nil, err
	}
	if !json.Valid(js) {
		return nil, &InvalidJson{}
	}
	dataJs.Write(js)

	if dataJs.Len() == 0 {
		dataJs.WriteString("{}")
	}

	return []byte(dataJs.String()), nil
}

func BeforeHandling(w *http.ResponseWriter) {
	(*w).Header().Set("Content-Type", "application/json")
	if config.DevMode {
		(*w).Header().Set("Access-Control-Allow-Origin", "*")
	}
	(*w).Header().Set("Access-Control-Allow-Headers", "g-recaptcha-response")
}

func ToLowerCase(ch uint8) uint8 {
	// 32 - the distance between 'A' and 'a' in the ascii table
	if ch >= 'A' && 'Z' >= ch {
		return ch + 32
	}
	return ch
}

var isAlphaRegex = regexp.MustCompile(`^[A-Za-z0-9\-_]+$`)

func IsAlphaWithDashAndUnderscore(s string) bool {
	return isAlphaRegex.MatchString(s)
}

func GetHostname() string {
	return config.Cfg.WebConfig.Domain + ":" + strconv.Itoa(int(config.Cfg.WebConfig.Port))
}
