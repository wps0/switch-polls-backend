package polls

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"switch-polls-backend/config"
	"switch-polls-backend/db"
	"switch-polls-backend/utils"
	"time"
)

func PollHandler(w http.ResponseWriter, r *http.Request) {
	if r.Context().Value("recaptcha").(utils.RecaptchaVerifyResponse).Action != "poll_get" {
		log.Printf("PollHandler got invalid recaptcha action from %s", r.RemoteAddr)
		WriteBadRequestResponse(&w)
		return
	}
	_, err := LimitBodySize(w, r, config.Cfg.WebConfig.EndpointsLimits.Polls.PollEndpoint.MaxBodySize)
	if err != nil {
		log.Printf("PollHandler failed to read request body: %v", err)
		return
	}

	args := mux.Vars(r)
	_id, err := strconv.Atoi(args["id"])
	if err != nil {
		log.Printf("PollHandler error when converting id to a string: %v", err)
		WriteBadRequestResponse(&w)
		return
	}

	res, err := db.PollsRepo.GetPoll(db.Poll{Id: _id}, true)
	if err != nil || res == nil || res.Id != _id {
		log.Printf("PollHandler poll with id %d retrieval error: %v", _id, err)
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
	if r.Context().Value("recaptcha").(utils.RecaptchaVerifyResponse).Action != "poll_vote" {
		log.Printf("PollVoteHandler got invalid recaptcha action from %s", r.RemoteAddr)
		WriteBadRequestResponse(&w)
		return
	}
	body, err := LimitBodySize(w, r, config.Cfg.WebConfig.EndpointsLimits.Polls.VotesEndpoint.MaxBodySize)
	if err != nil {
		log.Printf("PollVoteHandler failed to read request body: %v", err)
		WriteBadRequestResponse(&w)
		return
	}

	var reqData VoteRequest
	err = json.Unmarshal(body, &reqData)
	if err != nil {
		log.Println("PollVoteHandler failed to unmarshal body request data", err)
		WriteBadRequestResponse(&w)
		return
	}
	if !utils.ValidateUsername(reqData.UserData.Username) {
		log.Println("PollVoteHandler invalid username format")
		WriteBadRequestResponse(&w)
		return
	}
	email := UsernameToEmail(reqData.UserData.Username)
	if err = utils.ValidateEmail(email); err != nil {
		log.Printf("PollVoteHandler failed to verify the email address '%s'. error: %v", email, err)
		WriteBadRequestResponse(&w)
		return
	}

	option, err := db.PollsRepo.GetPollOption(db.PollOption{Id: reqData.OptionId}, false) //db.GetPollIdByOptionId(reqData.OptionId)
	if err != nil || option.PollId <= 0 {
		log.Println("PollVoteHandler GetPollOption error: ", err)
		WriteBadRequestResponse(&w)
		return
	}

	poll, err := db.PollsRepo.GetPoll(db.Poll{Id: option.PollId}, false)
	if err != nil {
		log.Printf("PollVoteHandler cannot get the poll with id %d, vote request by user %s on option %d, error: %v", option.PollId, email, option.PollId, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if poll.IsReadonly {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Ta ankieta nie przyjmuje już głosów."))
		return
	}

	user, err := db.UsersRepo.GetUser(db.User{Email: email}, true)
	if err != nil {
		log.Printf("PollVoteHandler get user (email: %s) error: %v\n", email, err)
	}

	voted, err := db.CheckIfUserHasAlreadyVotedById(user.Id, option.PollId)
	if voted {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Użytkownik oddał już głos"))
		return
	} else if err != nil {
		log.Println("PollVoteHandler cannot check if user has already voted, error: ", err)
		WriteBadRequestResponse(&w)
		return
	}

	// OK
	vote, err := db.VotesRepo.CreateVote(db.PollVote{
		UserId:   user.Id,
		OptionId: reqData.OptionId,
	})
	if err != nil {
		log.Printf("PollVoteHandler cannot insert the vote of user %s on poll option %d. error: %v", email, reqData.OptionId, err)
		WriteBadRequestResponse(&w)
		return
	}
	token, err := CreateVoteToken(vote.Id)

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
		log.Println("PollVoteHandler cannot send an email to "+email, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func PollConfirmHandler(w http.ResponseWriter, r *http.Request) {
	_, err := LimitBodySize(w, r, config.Cfg.WebConfig.EndpointsLimits.Polls.ConfirmVoteEndpoint.MaxBodySize)
	if err != nil {
		log.Printf("PollConfirmHandler error when reading request body %v", err)
		WriteBadRequestResponse(&w)
		return
	}
	vars := mux.Vars(r)
	token := vars["token"]
	if !utils.IsAlphaWithDash(token) {
		log.Println("PollConfirmHandler invalid token format")
		WriteBadRequestResponse(&w)
		return
	}

	err = VerifyToken(token)
	if err != nil {
		log.Println("PollConfirmHandler invalid token: ", err)
		WriteBadRequestResponse(&w)
		return
	}

	cnf, err := db.GetConfirmationByToken(token)
	if err != nil {
		log.Println("PollConfirmHandler cannot get confirmation by token", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	vote, _ := db.VotesRepo.GetVote(db.PollVote{Id: cnf.VoteId})
	option, _ := db.PollsRepo.GetPollOption(db.PollOption{Id: vote.OptionId}, false)
	poll, err := db.PollsRepo.GetPoll(db.Poll{Id: option.PollId}, false)
	if err != nil {
		log.Printf("PollConfirmHandler cannot get the poll with id %d, vote request by user %d, error: %v", option.PollId, vote.UserId, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if poll.IsReadonly {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Ta ankieta nie przyjmuje już głosów."))
		return
	}

	err = db.ChangeConfirmationStatus(cnf.VoteId, time.Now().Unix())
	if err != nil {
		log.Println("PollConfirmHandler cannot change confirmation status", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, _ := utils.PrepareResponse("Zarejestrowano glos!")
	// TODO: use templates instead of gluing the id to the end
	w.Header().Set("Location", config.Cfg.WebConfig.TokenVerificationRedirectLocation+strconv.Itoa(poll.Id))
	w.WriteHeader(http.StatusSeeOther)
	w.Write(res)
}

func PollResultsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Context().Value("recaptcha").(utils.RecaptchaVerifyResponse).Action != "poll_results_get" {
		log.Printf("PollResultsHandler got invalid recaptcha action from %s", r.RemoteAddr)
		WriteBadRequestResponse(&w)
		return
	}
	_, err := LimitBodySize(w, r, config.Cfg.WebConfig.EndpointsLimits.Polls.ResultsEndpoint.MaxBodySize)
	if err != nil {
		log.Printf("PollResultsHandler error when reading request body %v", err)
		return
	}

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
		log.Println("PollResultsHandler results summary error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp, _ := utils.PrepareResponse(summary)
	w.Write(resp)
}
