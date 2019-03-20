package snowgie

type Snowgie interface {
	Init(runtime SnowgieRuntime) error
	Run() error
}

type SnowgieContext interface {
	ProcessPublish(id string , body []byte , priority uint8) error
	GetConfig() SnowgieConfig
}

type SnowgieRuntime interface {
	Init(ctx SnowgieContext) error
	ProcessConsume(id string, body []byte)(error , bool)

}
