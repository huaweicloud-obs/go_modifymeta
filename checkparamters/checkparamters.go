package checkparamters


func CheckBucketname(BucketName string) (bool,string) {
	valildChars := "abcdefghijklmnopqrstuvwxyz0123456789-."
	bucketDic :=make(map[byte]byte)
	for x:=range valildChars{
		bucketDic[valildChars[x]] =valildChars[x]
	}
	if len(BucketName) <3 || len(BucketName)>64{
		return false,"avalid bucketname must more than 3 char and less than 64 chars"
	} else if (BucketName[0] ==bucketDic['.']||BucketName[0] ==bucketDic['-']||BucketName[len(BucketName)-1] ==bucketDic['.']||BucketName[len(BucketName)-1] ==bucketDic['-']){
		return false,"avalid bucketname must not begin or end with '.' or '-'"
	}else {
		for y:=range BucketName{
			_,ok:=bucketDic[BucketName[y]]
			if !ok{
				return false,"bucketname is Inavalid，must be combine with chars，num，'-','.'"
			}else {
				continue
			}
		}
		return true,"bucketname is valid"
	}
}
func CheckConcurent(con int)(bool){
	if con <1 || con >500{
		return false
	}else {
		return true
	}
}
