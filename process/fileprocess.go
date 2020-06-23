package process

import (
	"bufio"
	"fmt"
	"io"
	. "logger"
	"os"
	"strings"
)

func ReadConfigFile (filename string)( config map[string]string, err error){
	fi, err := os.Open(filename)
	Config:=map[string]string{}
	if err != nil {
		//close(srcStr)
		fmt.Println("the Read file error:",err)
		LoggerNormal(LEVEL_ERROR,"the Read file error:",err)
		return nil,err

	}
	defer fi.Close()
	reader := bufio.NewReader(fi);
	for {
		line, err := reader.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				fmt.Println("\nthe Read Config file finished,config are:",Config)
				//Logger.Printf("the Read Config file finished,config are::%s\n", Config)
				return Config,nil

			}
			fmt.Println("the Read Config file Failed,error are:", err)
			return Config,err
		}
		if strings.Contains(line,"=")&&!strings.HasPrefix(line,"#"){
			keyValue:=strings.SplitN(line,"=",2)
			if keyValue !=nil{
				Config[strings.ToLower(strings.TrimSpace(keyValue[0]))]=strings.TrimSpace(keyValue[1])
			}
		}

	}

}

