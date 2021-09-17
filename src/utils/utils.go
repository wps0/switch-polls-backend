package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"net/smtp"
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

var isAlphaDashUnderscoreRegex = regexp.MustCompile(`^[A-Za-z0-9\-_]+$`)
var isAlphaAtDotRegex = regexp.MustCompile(`^[A-Za-z0-9@.]+$`)
var isAlphaRegex = regexp.MustCompile(`^[A-Za-z0-9]+$`)

func IsAlpha(s string) bool {
	return isAlphaRegex.MatchString(s)
}

func IsAlphaWithAtAndDot(s string) bool {
	return isAlphaAtDotRegex.MatchString(s)
}

func IsAlphaWithDashAndUnderscore(s string) bool {
	return isAlphaDashUnderscoreRegex.MatchString(s)
}

func GetHostname() string {
	return config.Cfg.WebConfig.Domain + ":" + strconv.Itoa(int(config.Cfg.WebConfig.Port))
}

func VerifyUsername(s string) bool {
	return IsAlpha(s)
}

func SendEmail(conf *config.EmailConfiguration, subject string, msg string, receiver string) error {
	msg = "Subject: " + subject + "\nMIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" + msg
	receiversList := []string{receiver}

	auth := smtp.PlainAuth("", conf.SenderEmail, conf.SenderEmailPasswd, conf.SmtpHost)
	err := smtp.SendMail(conf.SmtpHost+":"+strconv.Itoa(conf.SmtpPort), auth, conf.SenderEmail, receiversList, []byte(msg))
	if err != nil {
		log.Println(err)
	}
	return err
}
