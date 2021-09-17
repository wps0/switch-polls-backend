package polls

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"switch-polls-backend/config"
	"switch-polls-backend/utils"
)

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

func WriteBadRequestResponse(w *http.ResponseWriter) {
	msg, err := utils.PrepareResponse("Bad request")
	(*w).WriteHeader(http.StatusBadRequest)
	if err != nil {
		log.Printf("Get invalid argument response > Error when parsing response: %s\n", err)
		return
	}
	(*w).Write(msg)
}

func UsernameToEmail(username string) (string, error) {
	if !utils.IsAlpha(username) {
		return "", errors.New("username can only contain alphanumeric characters")
	}
	return username + config.Cfg.EmailConfig.OrganizationDomain, nil
}

func VerifyRecaptcha(rq *http.Request) bool {
	origin := rq.Header.Get("Origin")
	token := rq.Header.Get("g-recaptcha-response")
	if len(token) == 0 || len(token) > 1024 || len(origin) > 384 || !utils.IsAlphaWithDashAndUnderscore(token) {
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
