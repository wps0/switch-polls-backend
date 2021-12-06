package utils

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"switch-polls-backend/config"

	"github.com/badoux/checkmail"
	"github.com/google/uuid"
)

type InvalidJson struct{}

type EmailTemplateValues struct {
	Receiver    string
	ServiceName string
	VoteOption  string
	PollTitle   string
	PollId      string
	Link        string
}

type RecaptchaVerifyRequest struct {
	Secret   string `json:"secret"`
	Response string `json:"response"`
	RemoteIp string `json:"remoteip"`
}

type RecaptchaVerifyResponse struct {
	Success     bool     `json:"success"`
	ChallengeTs string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
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

var isAlphaDashUnderscoreRegex = regexp.MustCompile(`^[A-Za-z0-9\-_]+$`)
var isAlphaDashRegex = regexp.MustCompile(`^[A-Za-z0-9\-]+$`)
var isAlphaRegex = regexp.MustCompile(`^[A-Za-z0-9]+$`)

func IsAlpha(s string) bool {
	return isAlphaRegex.MatchString(s)
}

func IsAlphaWithDash(s string) bool {
	return isAlphaDashRegex.MatchString(s)
}

func IsAlphaWithDashAndUnderscore(s string) bool {
	return isAlphaDashUnderscoreRegex.MatchString(s)
}

func GetListeningAddress() string {
	return config.Cfg.WebConfig.ListeningAddress + ":" + strconv.Itoa(int(config.Cfg.WebConfig.Port))
}

func ValidateUsername(s string) bool {
	return IsAlpha(s) && len(s) < 64
}

func ValidateEmail(email string) error {
	//if config.Cfg.EmailConfig.ExtendedEmailValidation {
	//	return checkmail.ValidateHostAndUser(config.Cfg.EmailConfig.OrganizationDomain + strconv.Itoa(config.Cfg.EmailConfig.SmtpPort), config.Cfg.EmailConfig.SenderEmail, email)
	//}
	return checkmail.ValidateFormat(email)
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

func VerifyRecaptcha(rq *http.Request) bool {
	origin := rq.Header.Get("Origin")
	token := rq.Header.Get("g-recaptcha-response")
	if len(token) == 0 || len(token) > 1024 || len(origin) > 384 || !IsAlphaWithDashAndUnderscore(token) {
		log.Printf("Required captcha header not found")
		return false
	}
	data := url.Values{}
	data.Set("secret", config.Cfg.WebConfig.RecaptchaSecret)
	data.Set("response", token)
	data.Set("remoteip", rq.RemoteAddr)

	res, err := http.Post(
		config.Cfg.WebConfig.RecaptchaVerifyEndpoint,
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()))
	if err != nil {
		log.Printf("Captcha verification error: %v", err)
		return false
	}
	bodyResp, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Captcha verification error: %v", err)
		return false
	}

	var resp RecaptchaVerifyResponse
	err = json.Unmarshal(bodyResp, &resp)
	if err != nil {
		log.Printf("Captcha verification error: %v", err)
		return false
	}
	if !resp.Success {
		log.Println("Captcha verification error(s): ", resp.ErrorCodes)
	}
	return resp.Success
}
