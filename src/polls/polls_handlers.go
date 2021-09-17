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
	utils.BeforeHandling(&w)
	if !VerifyRecaptcha(r) {
		WriteBadRequestResponse(&w)
		return
	}

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

// weryfikacja przez maila (6h):
//  - wysyłanie powiadomień
//  - weryfikacja legitności gdy user kliknie
//  - redirect do strony wyjściowej
// votes endpoint zwracający w json ilość głosów per opcja (2h)
// wyświetlanie tych głosów na frontendzie (6h)

// db cleanup raz na x h - usuwa, gdy zachodzi jakis warunek (np. uplynal czas od stworzenia / user juz potwierdzil)

// nie mozna potwierdzic, gdy user juz glosowal
//  /confirm_vote?token=UUID

func PollVoteHandler(w http.ResponseWriter, r *http.Request) {
	utils.BeforeHandling(&w)
	if !VerifyRecaptcha(r) {
		WriteBadRequestResponse(&w)
		return
	}

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
		WriteBadRequestResponse(&w)
		return
	}

	pollId, err := db.GetPollIdByOptionId(reqData.OptionId)
	if pollId <= 0 || err != nil {
		log.Println("option id was not found or other error has occurred. error: ", err)
		WriteBadRequestResponse(&w)
		return
	}

	if voted, err := db.CheckIfUserHasAlreadyVoted(email, pollId); voted {
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
		Link:        token,
	})
	err = utils.SendEmail(&config.Cfg.EmailConfig, config.Cfg.EmailConfig.EmailSubject, template, email)
	if err != nil {
		log.Println("cannot send an email to "+email, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

//func PollVotesHandler(w http.ResponseWriter, r *http.Request) {
//	utils.BeforeHandling(&w)
//	user := auth.AuthMgr.GetLoggedInUser(&w, r)
//	if user == nil {
//		return
//	}
//	args := mux.Vars(r)
//	id, _ := strconv.Atoi(args["id"])
//	poll := PollsMgr.GetPollById(uint(id))
//	if poll == nil {
//		w.WriteHeader(http.StatusNotFound)
//		w.Write(utils.PrepareResponseIgnoreErrors("poll not found"))
//		return
//	}
//	w.Write(utils.PrepareOKResponseIgnoreErrors(poll.PrepareVotesSummary()))
//}
