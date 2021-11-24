package polls

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"strconv"
	"switch-polls-backend/config"
	"switch-polls-backend/db"
	"switch-polls-backend/utils"
)

func PollHandler(w http.ResponseWriter, r *http.Request) {
	args := mux.Vars(r)
	_id, err := strconv.Atoi(args["id"])
	if err != nil {
		log.Printf("PollHandler > Error when converting id to a string")
		WriteBadRequestResponse(&w)
		return
	}

	res, err := db.PollsRepo.GetPoll(db.Poll{Id: _id}, true)
	if err != nil || res == nil || res.Id != _id {
		w.WriteHeader(http.StatusNotFound)
		resp, _ := utils.PrepareResponse("the poll with the given id was not found")
		w.Write(resp)
		return
	}

	resp, _ := utils.PrepareResponse(res)
	w.Write(resp)
}

// TODO: db cleanup raz na x h - usuwa, gdy zachodzi jakis warunek (np. uplynal czas od stworzenia / user juz potwierdzil)

func PollVoteHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Failed to read request body: ", err)
		WriteBadRequestResponse(&w)
		return
	}

	var reqData VoteRequest
	err = json.Unmarshal(body, &reqData)
	if err != nil {
		log.Println("failed to unmarshal body request data", err)
		WriteBadRequestResponse(&w)
		return
	}
	email, err := UsernameToEmail(reqData.UserData.Username)
	if err != nil || !utils.VerifyUsername(reqData.UserData.Username) {
		log.Println("cannot convert username to an email")
		WriteBadRequestResponse(&w)
		return
	}

	option, err := db.PollsRepo.GetPollOption(db.PollOption{Id: reqData.OptionId}, false) //db.GetPollIdByOptionId(reqData.OptionId)
	if err != nil || option.PollId <= 0 {
		log.Println("option id was not found or other error has occurred. error: ", err)
		WriteBadRequestResponse(&w)
		return
	}

	voted, err := db.CheckIfUserHasAlreadyVoted(email, option.PollId)
	if voted {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("user has already voted"))
		return
	} else if err != nil {
		log.Println("check if user has voted error", err)
		WriteBadRequestResponse(&w)
		return
	}

	// OK
	user, err := db.UsersRepo.GetUser(db.User{Email: email}, false)
	if err != nil {
		log.Printf("PollVoteHandler get user (email: %s) error: %v\n", email, err)
	}
	vote, err := db.VotesRepo.CreateVote(db.PollVote{
		UserId:   user.Id,
		OptionId: reqData.OptionId,
	})
	if err != nil {
		log.Printf("cannot insert the vote of user %s on poll option %d. error: %v", email, reqData.OptionId, err)
		WriteBadRequestResponse(&w)
		return
	}
	token, err := CreateVoteToken(vote.Id)

	poll, err := db.PollsRepo.GetPoll(db.Poll{Id: option.PollId}, false)
	if err != nil {
		log.Printf("PollVoteHandler cannog get the poll with id %v - vote request by user %s on option option %d. error: %v", option.PollId, email, reqData.OptionId, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	template := utils.FillEmailTemplate(utils.EmailTemplateValues{
		Receiver:    email,
		ServiceName: "SWITCH POLLS",
		VoteOption:  option.Content, // TODO: limit the length to n chars and append '...' to the end if the threshold is reached
		PollTitle:   poll.Title,
		PollId:      strconv.Itoa(poll.Id),
		Link:        GetConfirmationUrl(token),
	})
	err = utils.SendEmail(&config.Cfg.EmailConfig, config.Cfg.EmailConfig.EmailSubject, template, email)
	if err != nil {
		log.Println("cannot send an email to "+email, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func PollConfirmHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]
	if !utils.IsAlphaWithDash(token) {
		log.Println("invalid token format")
		WriteBadRequestResponse(&w)
		return
	}

	err := VerifyToken(token)
	if err != nil {
		log.Println("invalid token: ", err)
		WriteBadRequestResponse(&w)
		return
	}

	cnf, err := db.GetConfirmationByToken(token)
	if err != nil {
		log.Println("cannot get confirmation by token", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = db.ChangeConfirmationStatus(cnf.VoteId, true)
	if err != nil {
		log.Println("cannot change confirmation status", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	vote, err := db.VotesRepo.GetVote(db.PollVote{Id: cnf.VoteId})
	var option db.PollOption
	if err != nil {
		log.Println("vote by id error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		option, err = db.PollsRepo.GetPollOption(db.PollOption{Id: vote.OptionId}, false) //db.GetPollIdByOptionId(vote.OptionId)
		if err != nil {
			log.Println("cannot get pollId by option id when confirming user's vote", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	res, _ := utils.PrepareResponse("Zarejestrowano glos!")
	// TODO: use templates instead of gluing the id to the end
	w.Header().Set("Location", config.Cfg.WebConfig.TokenVerificationRedirectLocation+strconv.Itoa(option.PollId))
	w.WriteHeader(http.StatusSeeOther)
	w.Write(res)
}

func PollResultsHandler(w http.ResponseWriter, r *http.Request) {
	args := mux.Vars(r)
	id, _ := strconv.Atoi(args["id"])
	poll, err := db.PollsRepo.GetPoll(db.Poll{Id: id}, false)
	if err != nil {
		WriteBadRequestResponse(&w)
		return
	}
	if poll == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	summary, err := db.PrepareResultsSummary(poll.Id)
	if err != nil {
		log.Println("results summary error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp, _ := utils.PrepareResponse(summary)
	w.Write(resp)
}
