package snowgie

type Snowgie interface {
	Init(runtime SnowgieRuntime) error
	Run() error
}

type SnowgieController interface {
	ProcessPublish(id string , body []byte , priority uint8) error
	GetConfig() SnowgieConfig
}

type SnowgieRuntime interface {
	Init(controller SnowgieController) error
	ProcessConsume(id string, body []byte)(error , bool)

}
