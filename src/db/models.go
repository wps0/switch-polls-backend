package db

type OptionExtras struct {
	Type string
	Value string
}

type PollOption struct {
	Id int
	Content string
	Extras []OptionExtras
}

type Poll struct {
	Id int
	Title string
	Description string
	Options []PollOption
	CreateTime int `json:"-"`
}
