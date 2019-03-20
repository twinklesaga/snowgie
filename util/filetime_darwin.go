package util

import (
	"syscall"
	"time"
)

func GetFileCreateTime(path string) time.Time  {

	var st syscall.Stat_t
	syscall.Stat(path , &st)

	return time.Unix(int64(st.Ctimespec.Sec), int64(st.Ctimespec.Nsec))
}


