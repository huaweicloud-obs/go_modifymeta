package logger



import (
	"crypto/md5"
	"errors"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "runtime"
	"strconv"
	"strings"
    "sync"
	"time"
)

type Level int

//const (
//	LEVEL_OFF   Level = 500
//	LEVEL_ERROR Level = 400
//	LEVEL_WARN  Level = 300
//	LEVEL_INFO  Level = 200
//	LEVEL_DEBUG Level = 100
//)

var logLevelMap = map[Level]string{
	LEVEL_OFF:   "[OFF]: ",
	LEVEL_ERROR: "[ERROR]: ",
	LEVEL_WARN:  "[WARN]: ",
	LEVEL_INFO:  "[INFO]: ",
	LEVEL_DEBUG: "[DEBUG]: ",
}

type logConfType struct {
	level        Level
	logToConsole bool
	logFullPath  string
	maxLogSize   int64
	backups      int
}

func getDefaultLogConf() logConfType {
	return logConfType{
		level:        LEVEL_WARN,
		logToConsole: false,
		logFullPath:  "",
		maxLogSize:   1024 * 1024 * 30, //30MB
		backups:      10,
	}
}

var logConf logConfType

type loggerWrapper struct {
	fullPath   string
	fd         *os.File
	ch         chan string
	wg         sync.WaitGroup
	queue      []string
	logger     *log.Logger
	index      int
	cacheCount int
	closed     bool
}

func (lw *loggerWrapper) doInit() {
	fmt.Println("do the init logger")
	lw.queue = make([]string, 0, lw.cacheCount)
	lw.logger = log.New(lw.fd, "", 0)
	lw.ch = make(chan string, lw.cacheCount)
	lw.wg.Add(1)
	go lw.doWrite()
}

func (lw *loggerWrapper) rotate() {
	stat, err := lw.fd.Stat()
	if err != nil {
		lw.fd.Close()
		panic(err)
	}
	if stat.Size() >= logConf.maxLogSize {
		lw.fd.Sync()
		lw.fd.Close()
		if lw.index > logConf.backups {
			lw.index = 1
		}
		os.Rename(lw.fullPath, lw.fullPath+"."+IntToString(lw.index))
		lw.index += 1

		fd, err := os.OpenFile(lw.fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		lw.fd = fd
		lw.logger.SetOutput(lw.fd)
	}
}

func (lw *loggerWrapper) doFlush() {
	lw.rotate()
	for _, m := range lw.queue {
		lw.logger.Println(m)
	}
	_ = lw.fd.Sync()
}

func (lw *loggerWrapper) doClose() {
	lw.closed = true
	close(lw.ch)
	lw.wg.Wait()
}

func (lw *loggerWrapper) doWrite() {
	defer lw.wg.Done()
	for {
		msg, ok := <-lw.ch
		if !ok {
			lw.doFlush()
			_ = lw.fd.Close()
			break
		}
		if len(lw.queue) >= lw.cacheCount {
			lw.doFlush()
			lw.queue = make([]string, 0, lw.cacheCount)
		}
		lw.queue = append(lw.queue, msg)
	}

}

func (lw *loggerWrapper) Printf(format string, v ...interface{}) {
	if !lw.closed {
		msg := fmt.Sprintf(format, v...)
		lw.ch <- msg
	}
}

var consoleLogger *log.Logger
var fileLoggerNormal *loggerWrapper
var fileLoggerFail *loggerWrapper
var fileLoggerSucess *loggerWrapper
var lock *sync.RWMutex = new(sync.RWMutex)

func isDebugLogEnabled() bool {
	return logConf.level <= LEVEL_DEBUG
}

func isErrorLogEnabled() bool {
	return logConf.level <= LEVEL_ERROR
}

func isWarnLogEnabled() bool {
	return logConf.level <= LEVEL_WARN
}

func isInfoLogEnabled() bool {
	return logConf.level <= LEVEL_INFO
}

func reset() {
	if fileLoggerNormal != nil {
		fileLoggerNormal.doClose()
		fileLoggerNormal = nil
	}
	consoleLogger = nil
	logConf = getDefaultLogConf()
}

func resetFail() {
	if fileLoggerFail != nil {
		fileLoggerFail.doClose()
		fileLoggerFail = nil
	}
	consoleLogger = nil
	logConf = getDefaultLogConf()
}
func resetSuccess() {
	if fileLoggerSucess != nil {
		fileLoggerSucess.doClose()
		fileLoggerSucess = nil
	}
	consoleLogger = nil
	logConf = getDefaultLogConf()
}

func InitLogSucces(logFullPath string, maxLogSize int64, backups int, level Level, logToConsole bool) error {
	return InitLogWithCacheCntSucces(logFullPath, maxLogSize, backups, level, logToConsole, 50)
}
func InitLogNormal(logFullPath string, maxLogSize int64, backups int, level Level, logToConsole bool) error {
	return InitLogWithCacheCntNormal(logFullPath, maxLogSize, backups, level, logToConsole, 50)
}
func InitLogFail(logFullPath string, maxLogSize int64, backups int, level Level, logToConsole bool) error {
	return InitLogWithCacheCntFail(logFullPath, maxLogSize, backups, level, logToConsole, 50)
}
func InitLogWithCacheCntSucces(logFullPath string, maxLogSize int64, backups int, level Level, logToConsole bool, cacheCnt int) error {
	lock.Lock()
	defer lock.Unlock()
	if cacheCnt <= 0 {
		cacheCnt = 50
	}
	//resetSuccess()
	if fullPath := strings.TrimSpace(logFullPath); fullPath != "" {
		_fullPath, err := filepath.Abs(fullPath)
		if err != nil {
			return err
		}

		if !strings.HasSuffix(_fullPath, ".log") {
			_fullPath += ".log"
		}

		stat, err := os.Stat(_fullPath)
		if err == nil && stat.IsDir() {
			return errors.New(fmt.Sprintf("logFullPath:[%s] is a directory", _fullPath))
		} else if err := os.MkdirAll(filepath.Dir(_fullPath), os.ModePerm); err != nil {
			return err
		}

		fd, err := os.OpenFile(_fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}

		if stat == nil {
			stat, err = os.Stat(_fullPath)
			if err != nil {
				fd.Close()
				return err
			}
		}

		prefix := stat.Name() + "."
		index := 1
		walkFunc := func(path string, info os.FileInfo, err error) error {
			if err == nil {
				if name := info.Name(); strings.HasPrefix(name, prefix) {
					if i := StringToInt(name[len(prefix):], 0); i >= index {
						index = i + 1
					}
				}
			}
			return err
		}

		if err = filepath.Walk(filepath.Dir(_fullPath), walkFunc); err != nil {
			fd.Close()
			return err
		}

		fileLoggerSucess = &loggerWrapper{fullPath: _fullPath, fd: fd, index: index, cacheCount: cacheCnt, closed: false}
		fileLoggerSucess.doInit()
	}
	if maxLogSize > 0 {
		logConf.maxLogSize = maxLogSize
	}
	if backups > 0 {
		logConf.backups = backups
	}
	logConf.level = level
	if logToConsole {
		consoleLogger = log.New(os.Stdout, "", log.LstdFlags)
	}
	return nil
}
func InitLogWithCacheCntFail(logFullPath string, maxLogSize int64, backups int, level Level, logToConsole bool, cacheCnt int) error {
	lock.Lock()
	defer lock.Unlock()
	if cacheCnt <= 0 {
		cacheCnt = 50
	}
	//resetFail()
	if fullPath := strings.TrimSpace(logFullPath); fullPath != "" {
		_fullPath, err := filepath.Abs(fullPath)
		if err != nil {
			return err
		}

		if !strings.HasSuffix(_fullPath, ".log") {
			_fullPath += ".log"
		}

		stat, err := os.Stat(_fullPath)
		if err == nil && stat.IsDir() {
			return errors.New(fmt.Sprintf("logFullPath:[%s] is a directory", _fullPath))
		} else if err := os.MkdirAll(filepath.Dir(_fullPath), os.ModePerm); err != nil {
			return err
		}

		fd, err := os.OpenFile(_fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}

		if stat == nil {
			stat, err = os.Stat(_fullPath)
			if err != nil {
				fd.Close()
				return err
			}
		}

		prefix := stat.Name() + "."
		index := 1
		walkFunc := func(path string, info os.FileInfo, err error) error {
			if err == nil {
				if name := info.Name(); strings.HasPrefix(name, prefix) {
					if i := StringToInt(name[len(prefix):], 0); i >= index {
						index = i + 1
					}
				}
			}
			return err
		}

		if err = filepath.Walk(filepath.Dir(_fullPath), walkFunc); err != nil {
			fd.Close()
			return err
		}

		fileLoggerFail = &loggerWrapper{fullPath: _fullPath, fd: fd, index: index, cacheCount: cacheCnt, closed: false}
		fileLoggerFail.doInit()
	}
	if maxLogSize > 0 {
		logConf.maxLogSize = maxLogSize
	}
	if backups > 0 {
		logConf.backups = backups
	}
	logConf.level = level
	if logToConsole {
		consoleLogger = log.New(os.Stdout, "", log.LstdFlags)
	}
	return nil
}
func InitLogWithCacheCntNormal(logFullPath string, maxLogSize int64, backups int, level Level, logToConsole bool, cacheCnt int) error {
	lock.Lock()
	defer lock.Unlock()
	if cacheCnt <= 0 {
		cacheCnt = 50
	}
	//reset()
	if fullPath := strings.TrimSpace(logFullPath); fullPath != "" {
		_fullPath, err := filepath.Abs(fullPath)
		if err != nil {
			return err
		}

		if !strings.HasSuffix(_fullPath, ".log") {
			_fullPath += ".log"
		}

		stat, err := os.Stat(_fullPath)
		if err == nil && stat.IsDir() {
			return errors.New(fmt.Sprintf("logFullPath:[%s] is a directory", _fullPath))
		} else if err := os.MkdirAll(filepath.Dir(_fullPath), os.ModePerm); err != nil {
			return err
		}

		fd, err := os.OpenFile(_fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}

		if stat == nil {
			stat, err = os.Stat(_fullPath)
			if err != nil {
				fd.Close()
				return err
			}
		}

		prefix := stat.Name() + "."
		index := 1
		walkFunc := func(path string, info os.FileInfo, err error) error {
			if err == nil {
				if name := info.Name(); strings.HasPrefix(name, prefix) {
					if i := StringToInt(name[len(prefix):], 0); i >= index {
						index = i + 1
					}
				}
			}
			return err
		}

		if err = filepath.Walk(filepath.Dir(_fullPath), walkFunc); err != nil {
			fd.Close()
			return err
		}

		fileLoggerNormal = &loggerWrapper{fullPath: _fullPath, fd: fd, index: index, cacheCount: cacheCnt, closed: false}
		fileLoggerNormal.doInit()
	}
	if maxLogSize > 0 {
		logConf.maxLogSize = maxLogSize
	}
	if backups > 0 {
		logConf.backups = backups
	}
	logConf.level = level
	if logToConsole {
		consoleLogger = log.New(os.Stdout, "", log.LstdFlags)
	}
	return nil
}

func CloseLog() {
	if logEnabled() {
		lock.Lock()
		defer lock.Unlock()
		reset()
	}
}
func CloseLogcall_fail() {
	if logEnabled() {
		lock.Lock()
		defer lock.Unlock()
		resetFail()
		//reset()
	}
}
func CloseLogcall() {
	if logEnabled() {
		lock.Lock()
		defer lock.Unlock()
		resetSuccess()
		//reset()
	}
}

func SyncLog() {
}

func logEnabled() bool {
	return consoleLogger != nil || fileLoggerNormal != nil || fileLoggerFail!= nil || fileLoggerSucess != nil
}

func DoLog(level Level, format string, v ...interface{}) {
	LoggerNormal(level, format, v)
}

func LoggerNormal(level Level, format string, v ...interface{}) {
	if  logConf.level <= level {
		msg := fmt.Sprintf(format, v...)
		if _, file, line, ok := runtime.Caller(1); ok {
			index := strings.LastIndex(file, "/")
			if index >= 0 {
				file = file[index+1:]
			}
			msg = fmt.Sprintf("%s:%d|%s", file, line, msg)
		}
		prefix := logLevelMap[level]
		if consoleLogger != nil {
			consoleLogger.Printf("%s%s", prefix, msg)
		}
		if fileLoggerNormal != nil {
			nowDate := FormatUtcNow("2006-01-02T15:04:05Z")
			fileLoggerNormal.Printf("%s %s%s", nowDate, prefix, msg)
		}
	}
}

func Logcall_fail(level Level, format string, v ...interface{}) {
	if  logConf.level <= level {
		msg := fmt.Sprintf(format, v...)
		if _, file, line, ok := runtime.Caller(1); ok {
			index := strings.LastIndex(file, "/")
			if index >= 0 {
				file = file[index+1:]
			}
			msg = fmt.Sprintf("%s:%d|%s", file, line, msg)
		}
		prefix := logLevelMap[level]
		if consoleLogger != nil {
			consoleLogger.Printf("%s%s", prefix, msg)
		}
		if fileLoggerFail != nil {
			nowDate := FormatUtcNow("2006-01-02T15:04:05Z")
			fileLoggerFail.Printf("%s %s%s", nowDate, prefix, msg)
		}
	}
}


func Logcall(level Level, format string, v ...interface{}) {
	if  logConf.level <= level {
		msg := fmt.Sprintf(format, v...)
		if _, file, line, ok := runtime.Caller(1); ok {
			index := strings.LastIndex(file, "/")
			if index >= 0 {
				file = file[index+1:]
			}
			msg = fmt.Sprintf("%s:%d|%s", file, line, msg)
		}
		prefix := logLevelMap[level]
		if consoleLogger != nil {
			consoleLogger.Printf("%s%s", prefix, msg)
		}
		if fileLoggerSucess != nil {
			nowDate := FormatUtcNow("2006-01-02T15:04:05Z")
			fileLoggerSucess.Printf("%s %s,%s", nowDate, prefix, msg)
		}
	}
}

func StringToInt(value string, def int) int {
	ret, err := strconv.Atoi(value)
	if err != nil {
		ret = def
	}
	return ret
}

func StringToInt64(value string, def int64) int64 {
	ret, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		ret = def
	}
	return ret
}

func IntToString(value int) string {
	return strconv.Itoa(value)
}

func Int64ToString(value int64) string {
	return strconv.FormatInt(value, 10)
}

func GetCurrentTimestamp() int64 {
	return time.Now().UnixNano() / 1000000
}

func FormatUtcNow(format string) string {
	return time.Now().UTC().Format(format)
}

func FormatUtcToRfc1123(t time.Time) string {
	ret := t.UTC().Format(time.RFC1123)
	return ret[:strings.LastIndex(ret, "UTC")] + "GMT"
}

func Md5(value []byte) []byte {
	m := md5.New()
	m.Write(value)
	return m.Sum(nil)
}