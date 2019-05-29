package snowgie

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/streadway/amqp"
	"github.com/twinklesaga/snowgie/module"
	"github.com/twinklesaga/snowgie/util"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

func NewSnowgieNode(zks []string , zkPath string , id string)  (Snowgie , error) {

	zk := module.NewZookeeper()
	if err := zk.Connect(zks) ; err != nil {
		return nil , err
	}

	cfg , err := zk.Get(zkPath)
	if err != nil {
		return nil , err
	}

	snowgiePath := path.Join(zkPath , id)

	var config SnowgieConfig
	if err := json.Unmarshal(cfg , &config) ; err != nil {
		return nil ,err
	}

	return &SnowgieCore{
		id:id,
		zk:zk,
		zkPath:zkPath,
		snowgiePath:snowgiePath,
		config:config,
		mq:module.NewRabbitMq(),
		wg:new(sync.WaitGroup),
		doneList: make([]chan bool , 0),
		exchangeMap:make(map[string]chan<- amqp.Publishing),
	},nil
}

func NewSnowgie(cfgPath string , id string) (Snowgie , error){

	cfg , err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return nil , err
	}

	var config SnowgieConfig
	if err := json.Unmarshal(cfg , &config) ; err != nil {
		return nil ,err
	}
	log.Println(cfgPath, id ,string(cfg) , config)

	return &SnowgieCore{
		id:id,
		zk:nil,
		zkPath:"",
		snowgiePath:"",
		config:config,
		mq:module.NewRabbitMq(),
		wg:new(sync.WaitGroup),
		doneList: make([]chan bool , 0),
		exchangeMap:make(map[string]chan<- amqp.Publishing),
	},nil
}


type SnowgieCore struct {
	id 			string
	zk			*module.Zookeeper
	zkPath 		string
	snowgiePath string
	config		SnowgieConfig
	mq 			*module.RabbitMq

	runtime 	SnowgieRuntime

	wg 			*sync.WaitGroup
	doneList 	[]chan bool

	exchangeMap map[string]chan<- amqp.Publishing

	logger 		*module.RotateFileLogger
}

func (s *SnowgieCore) Init(runtime SnowgieRuntime) error {

	log.Println(s.config)

	logPath := path.Join(s.config.LogPath, s.config.NodeType)

	s.logger = module.NewRotateLogger(logPath, s.id)
	log.SetOutput(s.logger)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	s.runtime = runtime

	var err error
	err = s.lockPidFile()
	if err != nil {
		return err
	}
	err = runtime.Init(s)
	if err != nil {
		return err
	}


	if err := s.mq.Connect(s.config.AmqpUrl); err != nil {
		return err
	}

	//input queue
	for _, q := range s.config.Input {
		consume, err := s.mq.Consume(q.QueueId, q.Queue)
		if err != nil {
			return err
		}
		done := make(chan bool)
		s.runConsume(q.QueueId, consume, done)

		s.doneList = append(s.doneList, done)
	}

	//output exchange
	for _, e := range s.config.Output {
		done := make(chan bool)
		errorExchange, err := s.mq.Publish(e.Exchange, e.ExchangeType)
		if err != nil {
			return err
		}
		s.exchangeMap[e.ExchangeId] = s.runPublish(e.ExchangeId, errorExchange, done)
		s.doneList = append(s.doneList, done)
	}

	if err != nil {
		s.shutdown()
	}

	return err
}

func (s *SnowgieCore) Run() error {

	defer s.shutdown()

	done := make(chan bool)

	// zookeeper
	if s.zk != nil {
		child , event := s.zk.WatchChild(s.snowgiePath)
		go func(){
			for {
				select {
				case cmd := <-child:
					if s.commandProcess(cmd) {
						done <- true
					}
				case err := <-event:
					log.Println(err)
				}
			}
		}()
	} else {
		sig := make(chan os.Signal)
		signal.Notify(sig , syscall.SIGINT)
		go func() {
			<-sig
			log.Println("SIGNAL")
			done <- true
		}()
	}

	<-done

	for _ , c := range s.doneList {
		c <- true
	}

	s.wg.Wait()
	return nil
}

func (s *SnowgieCore)GetConfig() SnowgieConfig{
	return s.config
}

func (s *SnowgieCore)GetResource(resId SGResourceId) (SnowgieResource , error){
	panic("//")
}

func (s *SnowgieCore)GetNodConfig(nodConfig interface{}) error{
	if len(s.config.NodeConfigPath) > 0 {
		log.Println("load node config")
		data , err := ioutil.ReadFile(s.config.NodeConfigPath)

		if err != nil{
			return err
		}
		return json.Unmarshal(data, nodConfig)
	}else {
		return errors.New("cannot find nodeConfigPath property")
	}

	return nil
}
func (s *SnowgieCore)shutdown() {
	log.Println("Shutdown start")

	s.runtime.Shutdown()

	err := s.mq.Shutdown()
	if err != nil {
		log.Println(err)
	}

	if s.zk != nil {
		time.Sleep(time.Second)
		err = s.zk.Delete(s.snowgiePath)
		if err != nil {
			log.Println(err)
		}

		s.zk.Shutdown()
	}
	s.unlockPidFile()
	log.Println("Shutdown End")
	s.logger.Close()

}


func (s *SnowgieCore)ProcessPublish(id string , body []byte , priority uint8) error {

	pubChan , ok := s.exchangeMap[id]
	if ok {
		pub := amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "application/json",
			ContentEncoding: "",
			Body:            body,
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        priority,
		}
		pubChan <- pub

	}

	return nil
}


func (s *SnowgieCore) runConsume(id string , consume <-chan amqp.Delivery , done <-chan bool){

	s.wg.Add(1)
	go func(){
		defer s.wg.Done()
		terminated := false
		for !terminated {
			select {
			case <-done:
				terminated = true
				break
			case d , ok :=<-consume:
				if ok {
					err , ack := s.runtime.ProcessConsume(id , d.Body)
					if ack {
						d.Ack(false)
					}
					if err != nil {
						log.Println(err)
					}
				}else {
					terminated = true
				}
			}
		}

	}()
}

func (s *SnowgieCore)runPublish(id string , publishes chan<- amqp.Publishing , done <-chan bool) chan<- amqp.Publishing {
	pubChan := make(chan amqp.Publishing)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		terminated := false
		for !terminated {
			select {
			case <-done:
				terminated = true
				break
			case msg := <-pubChan:
				publishes<-msg
			}
		}
	}()
	return pubChan
}



func (s *SnowgieCore)commandProcess(cmd []string) bool{

	terminate := false
	for _,c := range cmd {

		cmdPath := path.Join(s.snowgiePath , c)
		if strings.EqualFold("COMMAND" , c) {
			data , err :=s.zk.Get(cmdPath)
			if err == nil {
				cmd := string(data)
				log.Printf("COMMAND : %s\n" , cmd)
				err := s.zk.Delete(cmdPath)
				if err != nil{
					log.Println(err)
				}
				if strings.EqualFold(cmd , "QUIT"){
					terminate = true
				}
			}
		}else {
			log.Printf("COMMAND : Unknown %s\n" , c )
		}

	}
	return terminate
}


func (s *SnowgieCore)lockPidFile() error {

	if len(s.config.PidPath) > 0 {
		pidFilePath := path.Join(s.config.PidPath ,s.config.NodeType, fmt.Sprintf("%s.pid" , s.id))

		return MakePidFile(pidFilePath , s.id)
	}
	return nil
}

func MakePidFile(pidFile string , id string )error {
	if util.Exists(pidFile) {
		var err error
		var pidData []byte
		var pid int
		pidData ,err = ioutil.ReadFile(pidFile)

		if err == nil {
			pid , err = strconv.Atoi(string(pidData))
			if err == nil {
				_ , err := os.FindProcess(pid)
				if err == nil {
					return errors.New(fmt.Sprintf("%s is already run : %d",id , pid))
				}
			}
		}
		return err
	}else {
		return ioutil.WriteFile(pidFile ,  []byte(fmt.Sprintf("%d", os.Getpid())), 0664)
	}
}

func (s *SnowgieCore)unlockPidFile() error {

	if len(s.config.PidPath) > 0 {
		pidFilePath := path.Join(s.config.PidPath, s.config.NodeType, fmt.Sprintf("%s.pid", s.id))
		RemovePidFile(pidFilePath)
	}
	return nil
}

func RemovePidFile(pidFile string) error{
	if util.Exists(pidFile) {
		return os.Remove(pidFile)
	}
	return errors.New("cannot find pid file")
}