package utils

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"switch-polls-backend/config"
)

type InvalidJson struct{}

func (e *InvalidJson) Error() string {
	return "JSON is not valid"
}

func PrepareResponse(data ...interface{}) ([]byte, error) {
	dataJs := strings.Builder{}
	dataJs.WriteString("{[")
	for item := range data {
		js, err := json.Marshal(item)

		if err != nil {
			return nil, err
		}
		if !json.Valid(js) {
			return nil, &InvalidJson{}
		}
	}
	dataJs.WriteString("]}")

	return []byte(dataJs.String()), nil
}

func BeforeHandling(w *http.ResponseWriter) {
	(*w).Header().Set("Content-Type", "application/json")
}

func ToLowerCase(ch uint8) uint8 {
	// 32 - the distance between 'A' and 'a' in the ascii table
	if ch >= 'A' && 'Z' >= ch {
		return ch + 32
	}
	return ch
}

func GetHostname() string {
	return config.Cfg.WebConfig.Domain + ":" + strconv.Itoa(int(config.Cfg.WebConfig.Port))
}
