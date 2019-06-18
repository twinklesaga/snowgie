package module

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/jlaffaye/ftp"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

var(
	ErrorFtpClientNil= errors.New("ftp client is nil")
	ErrorDirExist = errors.New("directory exists")
)

type FtpSession struct {
	Addr 		string
	User  		string
	Password 	string

	Client 		*ftp.ServerConn
}


func NewFtpSession(addr string , user string , pw string) *FtpSession{

	return &FtpSession{
		Addr:addr,
		User:user,
		Password:pw,
	}

}

func (f *FtpSession)Connect() error {
	var err error
	f.Client, err = ftp.Dial(f.Addr)


	fmt.Println("Dial")
	if err != nil {
		return err
	}

	if err := f.Client.Login(f.User, f.Password); err != nil {
		return err
	}

	return nil
}

func (f *FtpSession)Store(path string , filePath string) error {

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return f.Client.Stor(path , file)
}

func (f *FtpSession) MakeDir(path string) error{
	if f.Client != nil{
		return f.Client.MakeDir(path)
	}
	return ErrorFtpClientNil
}

func (f *FtpSession) MakeDirs(makePath string) error{
	var err error
	sp := strings.Split(makePath  , "/")
	makeP := "/"
	for _,p := range sp {
		if len(p) > 0 {
			makeP = path.Join(makeP, p)
			err = f.Client.ChangeDir(makeP)
			if err != nil {
				log.Println(err)
				err = f.MakeDir(makeP)
				if err != nil {
					return err
				}
			}
		}
	}
	err = f.Client.ChangeDir("/")
	return err
}

func (f *FtpSession) MakeDirs2(makePath string) error {

	if strings.HasSuffix(makePath, "/") {
		makePath = makePath[0 : len(makePath)-1]
	}
	base, dir := path.Split(makePath)

	log.Println(base , dir)
	entries, err := f.Client.List(base)
	log.Println(entries , err)
	if err == nil {

		for _, entry := range entries {
			if dir == entry.Name && entry.Type == ftp.EntryTypeFolder {
				return os.ErrExist
			}
		}
		err = f.Client.MakeDir(makePath)
		if err != nil && strings.HasPrefix( err.Error() , "550") {
			log.Println(err)
			err = os.ErrExist
		}
		return err
	}else if strings.HasSuffix(err.Error(), "No such file or directory") {
		err = f.MakeDirs(base)

		if err == nil {
			for _, entry := range entries {
				if dir == entry.Name && entry.Type == ftp.EntryTypeFolder {
					return os.ErrExist
				}
			}
			err = f.Client.MakeDir(makePath)
			if err != nil && strings.HasPrefix( err.Error() , "550") {
				log.Println(err)
				err = os.ErrExist
			}
			return err
		}
	}else {
		log.Println(err.Error())
	}
	return err
}

func (f *FtpSession) Rename(from string, to string) error{
	return f.Client.Rename(from , to)
}

func (f *FtpSession) Close() {
	if f.Client != nil {
		f.Client.Logout()
		f.Client.Quit()
	}
}

func (f *FtpSession) CheckConnect() error {
	if f.Client == nil {
		return f.Connect()
	}else {
		_ , err := f.Client.CurrentDir()
		if err != nil {
			if !( err == io.EOF || strings.HasPrefix(err.Error(), "421")) {
				log.Println(err)
			}

			f.Close()
			return f.Connect()
		}
		return err
	}
}

func (f *FtpSession) List(dir string) ([]string , error){
	if f.Client != nil{
		return f.Client.NameList(dir)
	}
	return nil, ErrorFtpClientNil
}

func (f *FtpSession)Retr(src string , dest string) error{
	destFile , err :=os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	response , err := f.Client.Retr(src)
	if err != nil {
		return err
	}
	defer response.Close()
	buf, err := ioutil.ReadAll(response)

	w := bufio.NewWriter(destFile)
	_, err = w.Write(buf)
	if err != nil {
		return err
	}
	return w.Flush()
}