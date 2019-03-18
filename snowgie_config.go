package snowgie


type SnowgieConfig struct {
	NodeType 	string			`json:"nodeType"`
	AmqpUrl 	string			`json:"amqpUrl"`
	BinPath		string			`json:"binPath"`
	LogPath		string 			`json:"logPath"`

	Input 		[]QueueInfo		`json:"input"`
	Output		[]ExchangeInfo	`json:"output"`
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