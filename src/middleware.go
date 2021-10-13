package main

import (
	"log"
	"net/http"
	"switch-polls-backend/config"
	"switch-polls-backend/utils"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request to %s from %s (method: %s).", r.URL, r.RemoteAddr, r.Method)
		next.ServeHTTP(w, r)
	})
}

func corsTerminateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", config.Cfg.WebConfig.CORS.AccessControlAllowOrigin)
		w.Header().Set("Access-Control-Allow-Headers", config.Cfg.WebConfig.CORS.AccessControlAllowHeaders)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func contentTypeJsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func recaptchaMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if utils.VerifyRecaptcha(r) {
			next.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid recaptcha token"))
		}
	})
}
