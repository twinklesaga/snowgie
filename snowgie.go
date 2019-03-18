package snowgie


type Snowgie interface {
	Init(runtime SnowgieRuntime) error
	Run() error
}

type SnowgieController interface {
	GetExchange(id string)
}

type SnowgieRuntime interface {
	Init(controller SnowgieController)
	ProcessConsume(id string, body []byte)(error , bool)
}
