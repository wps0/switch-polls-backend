package polls

import (
	"log"
	"net/http"
	"switch-polls-backend/utils"
)

func GetBadRequestResponse(w *http.ResponseWriter) []byte {
	msg, err := utils.PrepareResponse("Bad request")
	(*w).WriteHeader(http.StatusBadRequest)
	if err != nil {
		log.Printf("Get invalid argument response > Error when parsing response: %s\n", err)
		return nil
	}
	return msg
}
