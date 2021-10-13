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

	res, err := db.GetPollById(_id)
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

	pollId, err := db.GetPollIdByOptionId(reqData.OptionId)
	if pollId <= 0 || err != nil {
		log.Println("option id was not found or other error has occurred. error: ", err)
		WriteBadRequestResponse(&w)
		return
	}

	voted, err := db.CheckIfUserHasAlreadyVoted(email, pollId)
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
	voteId, err := db.InsertVote(email, reqData.OptionId)
	if err != nil {
		log.Printf("cannot insert the vote of user %s on poll option %d. error: %v", email, reqData.OptionId, err)
		WriteBadRequestResponse(&w)
		return
	}
	token, err := CreateVoteToken(voteId)

	template := utils.FillEmailTemplate(utils.EmailTemplateValues{
		Receiver:    email,
		ServiceName: "SWITCH POLLS",
		VoteOption:  strconv.Itoa(reqData.OptionId),
		PollTitle:   strconv.Itoa(pollId),
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

	vote, err := db.GetVoteById(cnf.VoteId)
	pollId := 0
	if err != nil {
		log.Println("vote by id error", err)
	} else {
		pollId, err = db.GetPollIdByOptionId(vote.OptionId)
		if err != nil {
			log.Println("cannot get pollId by option id when confirming user's vote", err)
		}
	}

	res, _ := utils.PrepareResponse("Zarejestrowano glos!")
	w.Header().Set("Location", config.Cfg.WebConfig.TokenVerificationRedirectLocation+strconv.Itoa(pollId))
	w.WriteHeader(http.StatusSeeOther)
	w.Write(res)
}

func PollResultsHandler(w http.ResponseWriter, r *http.Request) {
	args := mux.Vars(r)
	id, _ := strconv.Atoi(args["id"])
	poll, err := db.GetPollById(id)
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
