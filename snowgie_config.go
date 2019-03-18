package snowgie


type SnowgieConfig struct {
	nodeType 	string			`json:"nodeType"`
	amqpUrl 	string			`json:"amqpUrl"`
	binPath		string			`json:"binPath"`
	logPath		string 			`json:"logPath"`

	input 		[]QueueInfo		`json:"input"`
	output		[]ExchangeInfo	`json:"output"`
}


type QueueInfo struct {
	QueueId string				`json:"id"`
	Queue string				`json:"queue"`
}

type ExchangeInfo struct {
	ExchangeId 	 string			`json:"id"`
	Exchange	 string			`json:"exchange"`
	ExchangeType string			`json:"exchangeType"`
}