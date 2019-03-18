package module

import (
	"errors"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"strings"
	"time"
)

var (
	ErrorZkConnectLost = errors.New("ErrorZkConnectLost")
	ErrorLoadConfigFail = errors.New("ErrorLoadConfigFail")
)

type Zookeeper struct {
	conn *zk.Conn
	event <-chan zk.Event
	done chan error
}


func NewZookeeper() *Zookeeper{

	zk := &Zookeeper{
		conn:nil,
		done : make(chan error),
	}

	return zk
}


func (z *Zookeeper)Connect(zkUrl []string) error {

	var err error

	z.conn , z.event , err = zk.Connect(zkUrl , time.Second * 5)

	z.Run()

	return err
}


func (z *Zookeeper)Run() {
	go func() {
		terminate := false
		for !terminate {
			select {
			case <-z.done:
				terminate = true
				break
			case event := <-z.event:
				if event.Err != nil {
					fmt.Println(event.Err)
				}
				//fmt.Println(event.Type)

			}
		}
		fmt.Println("func (z *Zookeeper)Run() Terminated")
	}()

}


func (z *Zookeeper)Shutdown() error {

	z.done <- nil
	if z.conn != nil {
		z.conn.Close()
		z.conn = nil
	}
	return nil
}

func (z *Zookeeper) CreateForce(path string, data []byte) error {
	if z.conn != nil {

		subPath := strings.Split(path , "/")

		target := ""
		for _, p:=range subPath{

			if len(p) > 0 {
				target += fmt.Sprintf("/%s", p)

				exists, _, err := z.conn.Exists(target)
				if err == nil {
					if !exists {
						flags := int32(0)
						acl := zk.WorldACL(zk.PermAll)
						_, err = z.conn.Create(target, data, flags, acl)
					}
				}
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
	return ErrorZkConnectLost
}

func (z *Zookeeper)Create(path string , data []byte) error  {
	if z.conn != nil {
		flags := int32(0)
		acl := zk.WorldACL(zk.PermAll)

		_ , err := z.conn.Create(path , data , flags , acl)
		return err
	}
	return ErrorLoadConfigFail
}

func (z *Zookeeper)Exists(path string) (bool , error){
	if z.conn != nil {
		exists, _, err := z.conn.Exists(path)

		return exists, err
	}
	return false ,  ErrorZkConnectLost
}

func (z *Zookeeper)Get(path string) ([]byte, error){
	if z.conn != nil {
		data, _, err := z.conn.Get(path)

		return data, err
	}
	return nil , ErrorZkConnectLost
}


func (z *Zookeeper)Set(path string , data []byte) error{

	if z.conn != nil {
		exists , stat , err := z.conn.Exists(path)
		if err == nil {
			if exists {
				_ , err = z.conn.Set(path , data , stat.Version)
			}else {
				flags := int32(0)
				acl := zk.WorldACL(zk.PermAll)
				_ , err = z.conn.Create(path , data , flags , acl)
			}
		}
		return err
	}
	return ErrorZkConnectLost
}



func (z *Zookeeper)Delete(path string)error {
	if z.conn != nil {
		_ , stat , err := z.conn.Get(path)
		if err == nil {
			err = z.conn.Delete(path , stat.Version)
		}
		return err
	}
	return ErrorZkConnectLost
}


func (z *Zookeeper)WatchPath(path string)(chan []byte , chan error){
	pathData := make(chan []byte)
	errors := make(chan error)
	go func() {
		for {
			data, _, events, err := z.conn.GetW(path)
			if err != nil {
				errors <- err
				return
			}
			pathData <- data
			evt := <-events
			if evt.Err != nil {
				errors <- evt.Err
				return
			}
		}
	}()
	return pathData, errors

}

func (z *Zookeeper)WatchChild(path string)(chan []string , chan error){
	snapshots := make(chan []string)
	errors := make(chan error)
	go func() {
		for {
			snapshot, _, events, err := z.conn.ChildrenW(path)
			if err != nil {
				errors <- err
				return
			}
			snapshots <- snapshot
			evt := <-events
			if evt.Err != nil {
				errors <- evt.Err
				return
			}
		}
	}()
	return snapshots, errors

}
