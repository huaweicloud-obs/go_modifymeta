package process

import (
	"limiter"
	. "logger"
	"net/http"
	"obs"
	"strings"
	"sync"
)



func GetobsClient(RegionID string,ak string,sk string,bucket string,DefineDomain string) *obs.ObsClient {
	if len(DefineDomain)>=5{
		var err error
		var obsclient *obs.ObsClient
		obsclient, err = obs.New(ak, sk, DefineDomain,
			obs.WithCustomDomainName(true),
			obs.WithMaxConnections(MAXCONNECTIONS),
			)
		if err != nil {
			panic(err)
		}
		_,errbucket:=obsclient.HeadBucket(bucket)
		if errbucket !=nil{
			panic(errbucket)
		}
		return obsclient
	}else {
		endpoint := OBS+RegionID+HUAWEICLOUD

		var err error
		var obsclient *obs.ObsClient
		obsclient, err = obs.New(ak, sk, endpoint,)
		if err != nil {
			panic(err)
		}
		_,errbucket:=obsclient.HeadBucket(bucket)
		if errbucket !=nil{
			panic(errbucket)
		}
		return obsclient
	}

}


func SetObjectsMeta(client *obs.ObsClient,BucketName string,inputstr chan KeyInfo,wsyn *sync.WaitGroup,concurrent int) {
	ConcurLimiterS := limiter.NewConcurrencyLimiter(concurrent)
    for objectkey :=range inputstr{
    	objectname:=objectkey
		ConcurLimiterS.ExecuteWithParams(func(para ...interface{}) {

			result_headobject:=getobjectMetadata(client,BucketName,objectname.KeyName)
			// object does not exist in bucket
			if result_headobject {
				Setobjectmeta(client,BucketName,objectname)
			}
			CallProcesnum <-1
			wsyn.Done()
		},nil)

	}
	//close(CallProcesnum)
}

func ListObjects(client *obs.ObsClient,BucketName string,prefix string,inputstr chan KeyInfo,wsyn *sync.WaitGroup)  {
	is_Turncated:=true
	inputlist:= &obs.ListObjectsInput{}
	inputlist.Bucket=BucketName
	if len(prefix)==0{
		inputlist.Marker=""
	}else {
		inputlist.Marker=prefix
	}

	inputlist.MaxKeys=1000
	for   {
		if is_Turncated==false{
			break
		}else {
			resplist,errlist:=client.ListObjects(inputlist)
			if errlist==nil{
				if resplist.StatusCode == 200 {
					inputlist.Marker=resplist.NextMarker
					is_Turncated=resplist.IsTruncated
					for _,value := range resplist.Contents{
						if strings.HasSuffix(value.Key,".mp4")==true{
							var keyinfo KeyInfo
							keyinfo.KeyName=value.Key
							keyinfo.KeyModifyTime=value.LastModified.Format(http.TimeFormat)
							wsyn.Add(1)
							inputstr<-keyinfo
							Readfilenum<-1
						}
					}
				}
			}
		}
	}
	close(inputstr)
	close(Readfilenum)
}

func getobjectMetadata(client *obs.ObsClient,bucket_name string,object_name string) bool {
	input := &obs.GetObjectMetadataInput{}
	input.Bucket = bucket_name
	input.Key = object_name
	output, err := client.GetObjectMetadata(input)
	if err == nil {
		if output.StatusCode ==200&&output.ContentType!=MP4MIME{
			Logcall(LEVEL_INFO,"The objecet type of origin|bucket|%s|key|%s|contenttype|%s",bucket_name,object_name,output.ContentType)
			return true
		}else {
			return false
		}

	} else {
		if obsError, ok := err.(obs.ObsError); ok {
			//fmt.Printf("StatusCode:%d\n", obsError.StatusCode)
			if obsError.StatusCode == 404{
				Logcall_fail(LEVEL_WARN,"bucket name:%s,object key is:%s,StatusCode:%d,error message is:%s",bucket_name,object_name, obsError.StatusCode,obsError)
				return false
			}else {
				Logcall_fail(LEVEL_ERROR,"bucket name:%s,object key is:%s,StatusCode:%d,error message is:%s",bucket_name,object_name, obsError.StatusCode,obsError)
			}
		} else {
			//fmt.Println(err)
			LoggerNormal(LEVEL_ERROR,"bucket name:%s,object key is:%s,error:%v.",bucket_name,object_name, err)
		}
		return false
	}
}



func Setobjectmeta(client *obs.ObsClient,bucketname string,objectinfo KeyInfo) {
	input := &obs.SetObjectMetadataInput{}
	input.Bucket = bucketname
	input.Key = objectinfo.KeyName
	input.ContentType=MP4MIME
	output, err := client.SetObjectMetadata(input)
	if err == nil {
		if output.StatusCode==200{
			//_ = output.Body.Close()
			Logcall(LEVEL_INFO,"set objecet MetaData is success|bucket|%s||key|%s|Meta|%s",input.Bucket,input.Key,input.ContentType)
		}else {
			Logcall_fail(LEVEL_ERROR,"Set object MetaData is Failed|bucket|%s||key|%s|Meta|%s",input.Bucket,input.Key,input.ContentType)
		}
	} else {
		if obsError, ok := err.(obs.ObsError); ok {

			Logcall_fail(LEVEL_ERROR,"Set object meta error,bucket name:%s,object:%s,errorcode:%d,errortitile:%s,errormessage:%s," +
					"erroall:%v\n",input.Bucket,input.Key,obsError.StatusCode,obsError.Code,obsError.Message,obsError)

		} else {
			//fmt.Println(err)
			LoggerNormal(LEVEL_ERROR,"Set object meta error:",input.Bucket,input.Key,err.Error())

		}
	}
}