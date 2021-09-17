package main

import (
	"database/sql"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"switch-polls-backend/config"
	"switch-polls-backend/db"
	"switch-polls-backend/polls"
	"switch-polls-backend/utils"
)

var Db *sql.DB

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Printf("Request to %s from %s (method: %s).", r.URL, r.RemoteAddr, r.Method)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func HandleCORSOptionsRequest(w http.ResponseWriter, r *http.Request) {
	utils.BeforeHandling(&w)
	w.WriteHeader(http.StatusOK)
}

func main() {
	config.InitConfig()
	if config.DevMode {
		log.Println("Running the application in development mode.")
	}
	db.InitDb()
	// routing
	r := mux.NewRouter()
	r.Use(loggingMiddleware)

	// subrouters
	apiRouter := r.PathPrefix(config.Cfg.WebConfig.ApiPrefix).Subrouter()
	rPolls := apiRouter.PathPrefix("/polls").Subrouter()

	// polls
	rPolls.HandleFunc("/{id:[0-9]+}", polls.PollHandler).Methods("GET")
	rPolls.HandleFunc("/{id:[0-9]+}", HandleCORSOptionsRequest).Methods("OPTIONS")
	rPolls.HandleFunc("/vote", polls.PollVoteHandler).Methods("POST")
	rPolls.HandleFunc("/vote", HandleCORSOptionsRequest).Methods("OPTIONS")
	// chyba ok ten regex?
	rPolls.HandleFunc("/confirm_vote/{token:[A-Za-z0-9\\-]+}", polls.PollConfirmHandler).Methods("GET")

	// start http
	http.Handle("/", r)
	log.Printf("Listening at %s://%s%s\n", config.Cfg.WebConfig.Protocol, utils.GetHostname(), config.Cfg.WebConfig.ApiPrefix)
	log.Fatal(http.ListenAndServe(utils.GetHostname(), nil))
}
