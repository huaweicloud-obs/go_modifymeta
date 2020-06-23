package main

import (
	. "checkparamters"
	. "process"
	"flag"
	"fmt"
	. "logger"
	"obs"
	"os"
	"runtime"
	"sync"
	"time"
)

const BUFFER  =1024
var Config = flag.String("config", "", "configure argument paramter ,ex:/home/source/config.dat")
var Concurrent = flag.Int("job",20,"set the Concurrent numbers for calll source file process")
//var timework = flag.Bool("ti",true,"if true,the man always lazy work all time range;if false,the man will hard work in 00:00~6:00 ")

var obsClient *obs.ObsClient


func main(){


	flag.Parse()
	//==========================输入部分语法检查=========================================
	_,errin :=os.Stat(*Config)
	if errin !=nil{
		fmt.Println("invaild config file or wrong path",errin)
		return
	}
	config,errrConfig:=ReadConfigFile(*Config)
	if errrConfig !=nil{
		fmt.Println("invaild config data or less power",errin)
		return
	}


	_,errbucket :=CheckBucketname(config[BUCKET])
	if errbucket !="bucketname is valid" {
		fmt.Println("invaild destination bucket name or ",errbucket)
		LoggerNormal(LEVEL_ERROR,"invaild destination bucket name or ",errbucket)
		return
	}
	errconcurrent :=CheckConcurent(*Concurrent)
	if errconcurrent !=true {
		fmt.Println("invaild concurrent number,the value must be  1 =< x <= 500 .")
		LoggerNormal(LEVEL_ERROR,"invaild concurrent number,the value must be  1 =< x <= 500 .")
		return
	}


	fmt.Println("StaretTime:",time.Now().Format(time.RFC3339Nano))


	StartTime :=time.Now();
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("Total Work logic CPU :",runtime.NumCPU())
	inputstr := make(chan KeyInfo,BUFFER)
	obsClient:=GetobsClient(config[REGION],config[AK], config[SK],config[BUCKET],config[DEFINEDOMAIN])

	var wsyn sync.WaitGroup
	go ListObjects(obsClient,config[BUCKET],config[PREFIX],inputstr,&wsyn)

	go DisplayCurrent()
	SetObjectsMeta(obsClient,config[BUCKET],inputstr,&wsyn,*Concurrent)
	wsyn.Wait()

	CloseLog()
	CloseLogcall()
	CloseLogcall_fail()
	time.Sleep(3*time.Second)
	fmt.Println("EndTime:",time.Now().Format(time.RFC3339Nano))
	fmt.Println("Cost Time:",time.Since(StartTime))

}

