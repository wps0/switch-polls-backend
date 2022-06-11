package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"switch-polls-backend/config"
	"switch-polls-backend/db"
	"switch-polls-backend/polls"
	"switch-polls-backend/utils"
)

func main() {
	log.Println("---- Switch polls backend is starting... ----")
	config.InitConfig()
	if config.Cfg.DebugMode {
		log.Println("Running the application in debug mode.")
	}
	db.ApplyMigrations()
	db.InitDb()
	// routing
	r := mux.NewRouter()
	r.Use(contentTypeJsonMiddleware, loggingMiddleware)

	// subrouters
	apiRouter := r.PathPrefix(config.Cfg.WebConfig.ApiPrefix).Subrouter()
	pollsRoot := apiRouter.PathPrefix("/polls").Subrouter()
	pollsRoot.Use(corsTerminateMiddleware)

	// polls
	pollsRoot.HandleFunc("/confirm_vote/{token:[A-Za-z0-9\\-]+}", polls.PollConfirmHandler).Methods(http.MethodGet)

	pollsRecaptcha := pollsRoot.PathPrefix("").Subrouter()
	pollsRecaptcha.Use(recaptchaMiddleware)
	pollsRecaptcha.HandleFunc("/{id:[0-9]+}", polls.PollHandler).Methods(http.MethodGet, http.MethodOptions)
	pollsRecaptcha.HandleFunc("/{id:[0-9]+}/results", polls.PollResultsHandler).Methods(http.MethodGet, http.MethodOptions)
	pollsRecaptcha.HandleFunc("/vote", polls.PollVoteHandler).Methods(http.MethodPost, http.MethodOptions)

	// start http
	http.Handle("/", r)
	log.Printf("Listening on %s://%s%s\n", config.Cfg.WebConfig.Protocol, utils.GetListeningAddress(), config.Cfg.WebConfig.ApiPrefix)
	log.Fatal(http.ListenAndServe(utils.GetListeningAddress(), nil))
}
