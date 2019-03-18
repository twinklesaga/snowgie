package module

import (
	"errors"
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

var (
	ErrorMqChannelExist = errors.New("ErrorMqChannelExist")

)

type RabbitMq struct {
	conn    *amqp.Connection

	channelMap map[string]*amqp.Channel
}

func NewRabbitMq() *RabbitMq {
	mq :=&RabbitMq{
		conn :nil,
		channelMap:make(map[string]*amqp.Channel),
	}

	return mq
}


func (mq *RabbitMq)Connect(amqpUrl string) error{

	var err error
	mq.conn , err = amqp.Dial(amqpUrl)

	if err == nil {
		log.Printf("mq connect : %s\n",amqpUrl)
		go func() {
			errChan := <-mq.conn.NotifyClose(make(chan *amqp.Error))

			if errChan != nil {
				log.Println(errChan)
			}
		}()
	}
	return err

}


func (mq *RabbitMq)Consume(id string , queue string) (<-chan amqp.Delivery ,error){

	var err error
	if mq.conn != nil {

		_,ok := mq.channelMap[id]
		if ok {
			return nil , ErrorMqChannelExist
		}

		chnl , err := mq.conn.Channel()
		if err == nil {
			chnl.Qos(1 , 0 , false)

			deliveries , err := chnl.Consume(
				queue,
				"",
				false,
				false,
				false,
				false,
				nil,
			)

			if err == nil {
				mq.channelMap[id] = chnl
				return deliveries , nil
			}
		}

	}
	return nil , err
}

func (mq *RabbitMq)Publish(exchange string , exchangeType string)(chan<- amqp.Publishing , error){
	chnl , err := mq.conn.Channel()
	if err == nil {
		err = chnl.ExchangeDeclare(
			exchange,
			exchangeType,
			true,
			false,
			false,
			false,
			nil,
		)
		if err == nil {
			err = chnl.Confirm( false )
			if err == nil {


				confirms := chnl.NotifyPublish(make(chan amqp.Confirmation, 1))

				pubChan := make(chan amqp.Publishing)
				go func(){

					for {
						select {
						case pub :=<-pubChan:
							err := chnl.Publish(
								exchange,
								"",
								false,
								false,
								pub)
							if err != nil {
								log.Println(err)
							}
							confirmOne(confirms)
						}
					}
				}()

				return pubChan, nil
			}
		}
	}
	return nil , err
}

func confirmOne(confirms <-chan amqp.Confirmation) {
	log.Printf("waiting for confirmation of one publishing")
	confirmed := <-confirms
	if !confirmed.Ack{
		log.Printf("failed delivery of delivery tag: %d", confirmed.DeliveryTag)
	}
}

func (mq *RabbitMq)Shutdown() error{
	if mq.conn == nil {
		log.Println("mq.conn nil")
	}
	for k, v := range mq.channelMap {
		if err :=v.Cancel("", true); err != nil {

			return fmt.Errorf("Consumer_cancel_failed: %s %v\n", k,err)
		}
		if err :=v.Close(); err != nil {

			return fmt.Errorf("Close(): %s %v\n", k,err)
		}

	}
	if mq.conn != nil {
		if err := mq.conn.Close(); err != nil {
			return fmt.Errorf("AMQP connection close error: %s", err)
		}
	}
	return nil
}

func (mq *RabbitMq)StopConsume(id string) error {

	if mq.conn == nil {
		log.Println("StopConsume mq.conn nil")
	}

	chnl ,ok := mq.channelMap[id]
	if ok {
		if err :=chnl.Cancel("", true); err != nil {
			return fmt.Errorf("Consumer_cancel_failed: %s %v\n", chnl,err)
		}

		if err :=chnl.Close(); err != nil {

			return fmt.Errorf("Close(): %s %v\n", chnl,err)
		}
		mq.channelMap[id] = nil
		delete(mq.channelMap,id)
	}
	return nil
}