package polls

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"switch-polls-backend/db"
	"switch-polls-backend/utils"
)

func PollHandler(w http.ResponseWriter, r *http.Request) {
	utils.BeforeHandling(&w)

	args := mux.Vars(r)
	_id, err := strconv.Atoi(args["id"])
	if err != nil {
		log.Printf("PollHandler > Error when converting id to a string")
		w.Write(GetBadRequestResponse(&w))
		return
	}

	res, err := db.GetPollById(_id)
	if err != nil || res.Id != _id {
		w.WriteHeader(http.StatusNotFound)
		resp, _ := utils.PrepareResponse("the poll with the given id was not found")
		w.Write(resp)
		return
	}

	resp, _ := utils.PrepareResponse(res)
	w.Write(resp)
}

//func PollVoteHandler(w http.ResponseWriter, r *http.Request) {
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
//	body, err := ioutil.ReadAll(r.Body)
//	if err != nil {
//		w.WriteHeader(http.StatusBadRequest)
//		w.Write(utils.PrepareResponseIgnoreErrors("bad request"))
//		return
//	}
//	var vote JsonPollVote
//	err = json.Unmarshal(body, &vote)
//	if err != nil {
//		w.WriteHeader(http.StatusBadRequest)
//		w.Write(utils.PrepareResponseIgnoreErrors("bad request"))
//		return
//	}
//
//	err = PollsMgr.AddVoteUsingSortID(user, poll, vote.SortId)
//	if err != nil {
//		log.Printf("Failed to add the vote of user with id %d on %d in poll %d. Error: %s", user.ID, vote.SortId, poll.ID, err)
//		w.WriteHeader(http.StatusBadRequest)
//		w.Write(utils.PrepareResponseIgnoreErrors(err.Error()))
//		return
//	}
//	w.WriteHeader(204)
//}
//
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
