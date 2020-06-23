package logger

import (
	"fmt"
	"obs"
	"strings"
	"time"
)


var Readfilenum =make(chan int64,1024)
var CallProcesnum =make(chan int64,1024)

//var ProgressDisplay Progress
const (
	TIMESTD="2006-01-02_15:04:05"
	TIMESTD1="2006-01-02_15-0405"
	TIMETODAY="2006-01-02"
	LEVEL_OFF   Level = 500
	LEVEL_ERROR Level = 400
	LEVEL_WARN  Level = 300
	LEVEL_INFO  Level = 200
	LEVEL_DEBUG Level = 100
    MP4MIME="video/mp4"
	AK = "ak"
	SK = "sk"
	REGION ="region"
	BUCKET ="bucketname"
	DEFINEDOMAIN="definedomain"

	PREFIX="prefix"
	FIXIDLE=3000
	FIXBUSY=50
	MAXCONNECTIONS=1000


	OBS= "obs."
	HUAWEICLOUD=".myhuaweicloud.com"
)

type KeyInfo struct {
	KeyName string
	KeyModifyTime string
}
func init(){

	var logFullPath string = "./logs/OBS-SDK.log"
	// 设置每个日志文件的大小，单位：字节
	var maxLogSize int64 = 1024 * 1024 * 100
	// 设置保留日志文件的个数
	var backups int = 10
	var backupA int = 6000
	// 设置日志的级别
	var level = obs.LEVEL_DEBUG
	// 设置是否打印日志到控制台
	var logToConsole bool = false
	var logPrefix []string
	TimeNow:=time.Now().Format(TIMESTD1)
	logPrefix = append(logPrefix,"./logs/CallData_")
	logPrefix = append(logPrefix,TimeNow)

	filename:=strings.Join(logPrefix,"_")

	// 开启日志
	obs.InitLog(logFullPath, maxLogSize, backups, level, logToConsole)

	error1:=InitLogNormal(filename+"SystemsRun", maxLogSize, backupA, LEVEL_DEBUG, false)
	error2:=InitLogSucces(filename+"ReportSuccess", maxLogSize, backupA, LEVEL_DEBUG, false)
	error3:=InitLogFail(filename+"ReportError", maxLogSize, backupA, LEVEL_DEBUG, false)
	if error1 !=nil ||	error2 !=nil || error3 !=nil{
		fmt.Println("the Logger Init Failed,Error is",error1,error2,error3)
		return
	}

}

func DisplayCurrent( ){
	var x int64 =0
	var y int64 =0
	var tick <-chan time.Time
	tick = time.Tick(500 * time.Millisecond)
loop:
	for  {
		select {
		case _,ok:=<-Readfilenum:
			if ok{
				x++
			}

		case _,ok:=<-CallProcesnum:
			if !ok{
				break loop
			}
			y++

		case <-tick:
			if y>0{
				fmt.Printf("\rNow do the object call Procesnum/Readnum: %d/%d",y,x)
				//time.Sleep(1 * time.Second)
			}else {
				fmt.Printf("\rNow Read the file lines: %d",x)
			}
		}
	}
	fmt.Printf("\nFinished do the object call Procesnum/Readnum: %d/%d\n",y,x)
}
