package http

type Dto struct {
	Msg 	string			`json:"msg"`
	Data	interface{}		`json:"data"`
}

type MetricResponse struct {
	Name 		string 		`json:"name"`
	Value 		float64		`json:"value"`
	Time 		string 		`json:"time"`
}
