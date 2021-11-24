package polls

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"switch-polls-backend/config"
	"switch-polls-backend/db"
	"switch-polls-backend/utils"
)

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
	option, err := db.PollsRepo.GetPollOption(db.PollOption{Id: vote.OptionId}, false)
	if err != nil {
		return err
	}
	res, err := db.CheckIfUserHasAlreadyVotedById(vote.UserId, option.PollId)
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
