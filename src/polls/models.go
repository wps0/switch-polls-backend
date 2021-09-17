package polls

type VoteRequest struct {
	OptionId int      `json:"optionId"`
	UserData UserData `json:"userData"`
}

type UserData struct {
	UserAgent string `json:"userAgent"`
	Username  string `json:"username"`
}
