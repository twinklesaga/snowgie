package core

import "github.com/twinklesaga/snowgie"

func NewSnowgieNode(zks []string , zkPath string) snowgie.Snowgie {

	return &SnowgieCore{

	}
}

func NewSnowgie(cfgPath string )snowgie.Snowgie {
	return &SnowgieCore{

	}
}


type SnowgieCore struct {

}

func (s *SnowgieCore) Init(runtime snowgie.SnowgieRuntime) error {
	panic("implement me")
}

func (s *SnowgieCore) Run() error {
	panic("implement me")
}






