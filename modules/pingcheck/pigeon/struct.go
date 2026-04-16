package pigeon

type GenReport struct {
	Data struct {
		Pigeon struct {
			Alarm []Alarms `json:"alarms"`
			App   struct {
				Source string `json:"source"`
				Key    string `json:"key"`
			} `json:"app"`
		} `json:"pigeon"`
	} `json:"data"`
}



type Alarms struct {
	Fid       int    `json:"fid"`
	AlarmCode string    `json:"alarm_code"`    // none
	Value     string    `json:"value"`
	Subject   string    `json:"subject"`
	Sms       string    `json:"sms"`           // none
	Message   string    `json:"message"`
	Priority  string    `json:"priority"`
	Host      string    `json:"host"`
	HostName  string    `json:"hostname"`      // none
	Domain    string    `json:"domain"`
	Transfer  string    `json:"transfer"`
	AlarmTime string    `json:"alarm_time"`
	ExtArgs   []*ExtArg `json:"ext_args"`
}

//func (p Pigeon) String() {
//	fmt.Printf("")
//}

type ExtArg struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type M3Value struct {
	Body    string `json:"body"`
	Chart   M3Chart `json:"chart"`
	URL string `json:"url"`
}

type M3Chart struct {
	Legend string `json:"legend"`
	Title  string `json:"title"`
}

type M3Body struct {
	Metric       string `json:"metric"`
	DatasourceID int    `json:"datasourceId"`
	From         int64  `json:"from"`
	Step         int    `json:"step"`
	Source       string `json:"source"`
	To           int64  `json:"to"`
	Type         string `json:"type"`
}


type Alarm struct {
	Fid 		int 			`json:"fid"`
	AlarmCode 	string 			`json:"alarm_code"`
	Value 		string 			`json:"value"`
	Subject 	string			`json:"subject"`
	Message     string 			`json:"message"`
	Priority 	string 			`json:"priority"`
	Ip			string 			`json:"host"`
	Hostname 	string 			`json:"hostname"`
	Domain  	string  		`json:"domain"`
	Transfer 	string			`json:"transfer"`
	AlarmTime 	string 			`json:"alarm_time"`
	Metric 		string 			`json:"metric"`
}


type PigeonResopose struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Object  string `json:"object"`
	Message string `json:"message"`
}