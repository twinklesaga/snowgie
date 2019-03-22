package snowgie

type Snowgie interface {
	Init(runtime SnowgieRuntime) error
	Run() error
}

type SGResourceId int

const SGResFtp	SGResourceId = iota


type SnowgieResource interface {
	GetId() SGResourceId
	GetData() ([]byte ,error)
	Reg() error
	Close() error
}

type SnowgieContext interface {
	ProcessPublish(id string , body []byte , priority uint8) error
	GetConfig() SnowgieConfig
	GetResource(resId SGResourceId) (SnowgieResource , error)
}

type SnowgieRuntime interface {
	Init(ctx SnowgieContext) error
	ProcessConsume(id string, body []byte)(error , bool)
	Shutdown()
}
