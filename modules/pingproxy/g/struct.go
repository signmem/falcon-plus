package g

type GlobalConfig struct {
	Debug                   bool			`json:"debug"`
	LogFile                 string          `json:"logfile"`
	LogMaxAge				int				`json:"logmaxage"`
	LogRotateAge			int				`json:"logrotateage"`
	Http	 				*HttpConfig		`json:"http"`
}

type HttpConfig struct {
	Enabled			bool			`json:"enabled"`
	Listen			string			`json:"listen"`
	Port 			string 			`json:"port"`
}

type HttpPingRequest struct {
	Ipaddr 			string 			`json:"ipaddr"`
}

type HttpPingResponse struct {
	PingStatus 		bool			`json:"pingcheck"`
	Ipaddr 			string 			`json:"ipaddr"`
}