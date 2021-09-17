package polls

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"switch-polls-backend/config"
	"switch-polls-backend/db"
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
	return username + "@" + config.Cfg.EmailConfig.OrganizationDomain, nil
}

func VerifyRecaptcha(rq *http.Request) bool {
	origin := rq.Header.Get("Origin")
	token := rq.Header.Get("g-recaptcha-response")
	if len(token) == 0 || len(token) > 1024 || len(origin) > 384 || !utils.IsAlphaWithDashAndUnderscore(token) {
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

func CreateVoteToken(voteId int) (string, error) {
	token := utils.GetNewToken()
	return token, db.InsertToken(token, voteId)
}

func VerifyToken(token string) error {
	if !utils.IsAlphaWithDash(token) {
		return errors.New("invalid character in token")
	}

	cnf, err := db.GetConfirmationByToken(token)
	if err != nil {
		return err
	}
	vote, err := db.GetVoteById(cnf.VoteId)
	if err != nil {
		return err
	}
	pollId, err := db.GetPollIdByOptionId(vote.OptionId)
	if err != nil {
		return err
	}
	res, err := db.CheckIfUserHasAlreadyVotedById(vote.UserId, pollId)
	if err != nil {
		return err
	}
	if res {
		return errors.New("user has already voted")
	}
	return nil
}

func GetConfirmationUrl(token string) string {
	port := ""
	if config.Cfg.WebConfig.Port != 80 && config.Cfg.WebConfig.Port != 443 {
		port = ":" + strconv.Itoa(int(config.Cfg.WebConfig.Port))
	}
	return config.Cfg.WebConfig.Protocol + "://" + config.Cfg.WebConfig.Domain + port + config.Cfg.WebConfig.ApiPrefix + "/polls/confirm_vote/" + token
}
