package module



import (
	"fmt"
	"github.com/twinklesaga/snowgie/util"
	"io"
	"log"
	"os"
	"path"
	"time"
)

type RotateFileLogger struct {
	logPath 			string
	id 				string
	fpCreateTime 	time.Time
	fpLog 			*os.File
}

func NewRotateLogger(logPath string , id string) *RotateFileLogger{
	logger := &RotateFileLogger{
		logPath:logPath,
		id:id,
		fpLog:nil,
	}
	logger.createLogFile()

	log.SetOutput(logger)

	return logger
}

func (r *RotateFileLogger) Write(p []byte) (n int, err error) {

	r.rotateLog()

	if r.fpLog != nil {
		return r.fpLog.Write(p)
	}
	return 0, io.ErrClosedPipe
}

func (r *RotateFileLogger) Close() {
	log.SetOutput(os.Stdout)

	if r.fpLog != nil {
		r.fpLog.Close()
		r.fpLog = nil
	}
}


func (r *RotateFileLogger)rotateLog() {
	now := time.Now()

	if r.fpCreateTime.Day() != now.Day() ||
		r.fpCreateTime.Month() != now.Month() ||
		r.fpCreateTime.Year() != now.Year() {

		logfilePath := path.Join(r.logPath, r.id +".log")

		logfileBackupPath := path.Join(r.logPath, fmt.Sprintf("%s_%s.log" ,r.id, r.fpCreateTime.Format("2006-01-02")))
		os.Rename(logfilePath , logfileBackupPath)

		r.createLogFile()
	}
}

func (r *RotateFileLogger)createLogFile(){
	var err error
	typePath := util.PathTypeCheck(r.logPath)
	if typePath != util.PtDirectory {
		os.MkdirAll(r.logPath , os.ModePerm)
	}

	logfilePath := path.Join(r.logPath, r.id +".log")
	if r.fpLog != nil {
		r.fpLog.Close()
		r.fpLog = nil
	}

	r.fpLog, err = os.OpenFile(logfilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		r.fpCreateTime = util.GetFileCreateTime(logfilePath)
	}

	if err != nil {
		r.Close()
	}
}

