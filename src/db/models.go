package db

type OptionExtras struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type PollOption struct {
	Id      int            `json:"id"`
	Content string         `json:"content"`
	Extras  []OptionExtras `json:"extras"`
}

type Poll struct {
	Id          int          `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Options     []PollOption `json:"options"`
	CreateTime  int          `json:"-"`
}
