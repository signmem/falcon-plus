package cron

type MailDto struct {
	Priority int    `json:"priority"`
	Metric   string `json:"metric"`
	Subject  string `json:"subject"`
	Content  string `json:"content"`
	Email    string `json:"email"`
	Status   string `json:"status"`
}

type SmsDto struct {
	Priority int    `json:"priority"`
	Metric   string `json:"metric"`
	Content  string `json:"content"`
	Phone    string `json:"phone"`
	Status   string `json:"status"`
}

type ImDto struct {
	Priority int    `json:"priority"`
	Metric   string `json:"metric"`
	Content  string `json:"content"`
	IM       string `json:"im"`
	Status   string `json:"status"`
}

//add by vincent.zhang for pigeon
type PigeonDto struct {
	Priority   int    `json:"priority"`
	Status     string `json:"status"` // OK or PROBLEM
	Endpoint   string `json:"endpoint"`
	Note       string `json;"note"`
	Metric     string `json:"metric"`
	Tags       string `json:"tags"`
	LeftValue  string `json:"leftValue"`
	Func       string `json:"func"`
	Operator   string `json:"operator"`
	RightValue string `json:"rightValue"`
	EventTime  string `json:"eventTime"`
	IP         string `json:"ip"`
	Domain     string `json:"Domain"`
}

type PigeonFid struct {
	Fid int64 `json:"fid"`
}

type PigeonFidResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Object  *PigeonFid `json:"object"`
	Success bool       `json:"success"`
}

type PigeonResponse struct {
	Success bool   `json:"success"`
	Object string `json:"object"`
	Message string `json:"message"`
}
