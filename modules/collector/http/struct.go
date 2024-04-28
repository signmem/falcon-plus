package http

type Dto struct {
	Msg 	string			`json:"msg"`
	Data	interface{}		`json:"data"`
}

var (
	Debug bool
)