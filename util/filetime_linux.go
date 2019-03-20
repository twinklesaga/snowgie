package util

import (
	"os"
	"syscall"
	"time"
)

func GetFileCreateTime(path string) time.Time  {

	fi , err := os.Stat(path)
	if err == nil {
		stat := fi.Sys().(*syscall.Stat_t)
		return time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))
	}
	return time.Now()
}

