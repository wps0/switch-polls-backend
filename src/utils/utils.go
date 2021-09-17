package utils

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/smtp"
	"regexp"
	"strconv"
	"strings"
	"switch-polls-backend/config"

	"github.com/google/uuid"
)

type InvalidJson struct{}

type EmailTemplateValues struct {
	Receiver    string
	ServiceName string
	VoteOption  string
	PollTitle   string
	Link        string
}

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
var isAlphaDashRegex = regexp.MustCompile(`^[A-Za-z0-9\-]+$`)
var isAlphaAtDotRegex = regexp.MustCompile(`^[A-Za-z0-9@.]+$`)
var isAlphaRegex = regexp.MustCompile(`^[A-Za-z0-9]+$`)

func IsAlpha(s string) bool {
	return isAlphaRegex.MatchString(s)
}

func IsAlphaWithAtAndDot(s string) bool {
	return isAlphaAtDotRegex.MatchString(s)
}

func IsAlphaWithDash(s string) bool {
	return isAlphaDashRegex.MatchString(s)
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
	msg = "Subject: " + subject + "\nFrom: " + conf.SenderEmail + "\nMIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" + msg
	receiversList := []string{receiver}

	auth := smtp.PlainAuth("", conf.SenderMailboxUsername, conf.SenderEmailPasswd, conf.SmtpHost)
	err := smtp.SendMail(conf.SmtpHost+":"+strconv.Itoa(conf.SmtpPort), auth, conf.SenderEmail, receiversList, []byte(msg))
	return err
}

func GetNewToken() string {
	return uuid.NewString()
}

func FillEmailTemplate(contents EmailTemplateValues) string {
	var err error

	temp := template.New("email_temp")
	temp, err = temp.Parse(config.Cfg.EmailConfig.EmailTemplate)
	if err != nil {
		log.Printf("template parse error: %s\n", err)
		return ""
	}
	var buf bytes.Buffer
	err = temp.Execute(&buf, &contents)
	if err != nil {
		log.Printf("template execute error: %s\n", err)
		return ""
	}
	return buf.String()
}
