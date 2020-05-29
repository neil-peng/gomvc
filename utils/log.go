//come from someone, change later.
package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// RFC5424
const (
	LEVEL_EMERGENCY = iota
	LEVEL_ALERT
	LEVEL_CRITICAL // used for always
	LEVEL_WARNING
	LEVEL_NOTICE  //5
	LEVEL_INFO    //6
	LEVEL_DEBUG   //7
	LEVEL_VERBOSE //8
)

var (
	levelStrings = []string{
		"[EMERGENCY]",
		"[ALERT]",
		"[CRITICAL]",
		"[WARNING]",
		"[NOTICE]",
		"[INFO]",
		"[DEBUG]",
		"[VERB]",
	}
)

// A Logger represents an active logging object that generates lines of
// output to an io.Writer.  Each logging operation makes a single call to
// the Writer's Write method.  A Logger can be used simultaneously from
// multiple goroutines; it guarantees to serialize access to the Writer.
type LoggerBase struct {
	level        int32
	mu           sync.Mutex // ensures atomic writes; protects the following fields
	out          *os.File   // destination for output
	out_wf       *os.File   // destination for output
	path         string     // log file path
	buf          []byte     // for accumulating text to write
	backupCount  int
	microseconds bool
	shortfile    bool
	printCatal   bool
}

type Logger struct {
	reqinfo    map[string]string
	noticeinfo map[string]string
	buf        string
	mu         sync.Mutex
}

/*
 * global static var
 */
var _log = &LoggerBase{
	out:          os.Stderr,
	out_wf:       os.Stderr,
	level:        LEVEL_NOTICE,
	backupCount:  0,
	microseconds: true,
	shortfile:    true,
	printCatal:   false,
}

func NewLogger() *Logger {
	return &Logger{
		reqinfo:    make(map[string]string),
		noticeinfo: make(map[string]string),
		buf:        "",
	}
}

func SetLogLevel(level int32) {
	Critical("set log level to %v", level)
	atomic.StoreInt32(&_log.level, level)
}

func GetLogLevel() int32 {
	v := atomic.LoadInt32(&_log.level)
	return v
}

func SetLogbackupCount(cnt int) {
	_log.backupCount = cnt
}

func SetPrintCatal() {
	_log.printCatal = true
}

func SetLogFile(path string) {
	//Critical("set log file to %v", path)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		Critical("error on SetLogFile: err: %s", err)
	}

	_log.out = f
	_log.path = path

	f_wf, err := os.OpenFile(path+".wf", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		Critical("error on SetLogFile: err: %s", err)
	}
	_log.out_wf = f_wf
}

func ReOpen(path string) {
	if _log.path == "" {
		return
	}
	_log.mu.Lock()
	defer _log.mu.Unlock()

	_log.out.Close()
	_log.out_wf.Close()
	SetLogFile(_log.path)
}

func timestr(period time.Duration) string {
	t := time.Now().Add(time.Second * -10)

	if period == time.Minute {
		return fmt.Sprintf("%04d%02d%02d%02d%02d",
			t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
	}
	if period == time.Hour {
		return fmt.Sprintf("%04d%02d%02d%02d",
			t.Year(), t.Month(), t.Day(), t.Hour())
	}
	if period == time.Hour*24 {
		return fmt.Sprintf("%04d%02d%02d",
			t.Year(), t.Month(), t.Day())
	}

	return fmt.Sprintf("%04d%02d%02d%02d%02d%02d",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}

func getFilesToDelete(path string, fileFilter *regexp.Regexp, backupCount int) []string {
	var result []string
	if backupCount <= 0 {
		return result
	}

	dirName := filepath.Dir(path)
	baseName := filepath.Base(path)
	fileInfos, err := ioutil.ReadDir(dirName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FileLogWriter(%s): %s\n", path, err)
		return result
	}

	prefix := baseName + "."
	plen := len(prefix)
	for _, fileInfo := range fileInfos {
		fileName := fileInfo.Name()
		if len(fileName) >= plen {
			if fileName[:plen] == prefix {
				suffix := fileName[plen:]
				if fileFilter.MatchString(suffix) {
					result = append(result, filepath.Join(dirName, fileName))
				}
			}
		}
	}

	sort.Sort(sort.StringSlice(result))

	if len(result) < backupCount {
		result = result[0:0]
	} else {
		result = result[:len(result)-backupCount]
	}
	return result
}

/*
 * enable rotate whit peirod
 * peirod can be: time.Minute, time.Hour, 24 * time.Hour
 */
func SetLogRotate(period time.Duration) {
	if period != time.Minute && period != time.Hour && period != time.Hour*24 {
		Critical("bad rotate peirod: %s", period)
		return
	}
	var fileFilter *regexp.Regexp
	if period == time.Minute { //Min
		fileFilter = regexp.MustCompile(`^\d{4}\d{2}\d{2}\d{2}\d{2}$`)
	} else if period == time.Hour { //Hour
		fileFilter = regexp.MustCompile(`^\d{4}\d{2}\d{2}\d{2}$`)
	} else { //Day
		fileFilter = regexp.MustCompile(`^\d{4}\d{2}\d{2}$`)
	}
	ch := make(chan bool)

	go func() {
		for {
			now := time.Now()
			nextHour := now.Truncate(period).Add(period).Add(time.Second)
			timer := time.NewTimer(nextHour.Sub(now))
			<-timer.C
			ch <- true
		}
	}()

	go func() {
		for {
			<-ch
			t := timestr(period)
			//log
			filename := fmt.Sprintf("%s.%s", _log.path, t)
			os.Rename(_log.path, filename)
			for _, fileName := range getFilesToDelete(_log.path, fileFilter, _log.backupCount) {
				os.Remove(fileName)
			}
			//log.wf
			filename = fmt.Sprintf("%s.wf.%s", _log.path, t)
			os.Rename(_log.path+".wf", filename)
			for _, fileName := range getFilesToDelete(_log.path+".wf", fileFilter, _log.backupCount) {
				os.Remove(fileName)
			}
			ReOpen(_log.path)
		}
	}()
}

func (L *Logger) formatBaseInfo() {
	L.mu.Lock()
	defer L.mu.Unlock()

	if L.buf == "" && len(L.reqinfo) > 0 {
		for k, v := range L.reqinfo {
			L.buf += "[" + k + ":" + v + "] "
		}
	}
}

func (L *Logger) SetBaseInfo(key string, value interface{}) {
	L.mu.Lock()
	defer L.mu.Unlock()

	L.reqinfo[key] = fmt.Sprintf("%v", value)
}

func (L *Logger) PushNotice(key string, value interface{}) {
	L.mu.Lock()
	defer L.mu.Unlock()

	L.noticeinfo[key] = fmt.Sprintf("%v", value)
}

func Critical(format string, v ...interface{}) {
	_log.output(LEVEL_CRITICAL, "", format, v...)
}

func (L *Logger) Critical(format string, v ...interface{}) {
	L.formatBaseInfo()
	_log.output(LEVEL_CRITICAL, L.buf, format, v...)
}

func Warn(format string, v ...interface{}) {
	_log.output(LEVEL_WARNING, "", format, v...)
}

func (L *Logger) Warn(format string, v ...interface{}) {
	L.formatBaseInfo()
	_log.output(LEVEL_WARNING, L.buf, format, v...)
}

func Notice(format string, v ...interface{}) {
	_log.output(LEVEL_NOTICE, "", format, v...)
}

func (L *Logger) Notice(format string, v ...interface{}) {
	L.formatBaseInfo()

	L.mu.Lock()
	defer L.mu.Unlock()

	for k, v := range L.noticeinfo {
		L.buf += "[" + k + ":" + v + "] "
	}

	_log.output(LEVEL_NOTICE, L.buf, format, v...)
}

func Info(format string, v ...interface{}) {
	_log.output(LEVEL_INFO, "", format, v...)
}

func (L *Logger) Info(format string, v ...interface{}) {
	L.formatBaseInfo()
	_log.output(LEVEL_INFO, L.buf, format, v...)
}

func Debug(format string, v ...interface{}) {
	_log.output(LEVEL_DEBUG, "", format, v...)
}

func (L *Logger) Debug(format string, v ...interface{}) {
	L.formatBaseInfo()
	_log.output(LEVEL_DEBUG, L.buf, format, v...)
}

func Verbose(format string, v ...interface{}) {
	_log.output(LEVEL_VERBOSE, "", format, v...)
}

func (L *Logger) Verbose(format string, v ...interface{}) {
	L.formatBaseInfo()
	_log.output(LEVEL_VERBOSE, L.buf, format, v...)
}

func Stacktrace(level int32, format string, v ...interface{}) {
	if level > GetLogLevel() {
		return
	}
	_log.output(level, "", format+" --- stack: \n%s", v, debug.Stack())
}

func (L *Logger) Stacktrace(level int32, format string, v ...interface{}) {
	if level > GetLogLevel() {
		return
	}
	L.formatBaseInfo()
	_log.output(level, L.buf, format+" --- stack: \n%s", v, debug.Stack())
}

/*
 * variadic is slow because we create temp slices
 * so we add some help functions
 */
func Debug1(format string, a interface{}) {
	if LEVEL_DEBUG > GetLogLevel() {
		return
	}

	_log.output(LEVEL_DEBUG, "", format, a)
}

func (L *Logger) Debug1(format string, a interface{}) {
	if LEVEL_DEBUG > GetLogLevel() {
		return
	}
	L.formatBaseInfo()
	_log.output(LEVEL_DEBUG, L.buf, format, a)
}

func Debug2(format string, a interface{}, b interface{}) {
	if LEVEL_DEBUG > GetLogLevel() {
		return
	}

	_log.output(LEVEL_DEBUG, "", format, a, b)
}

func (L *Logger) Debug2(format string, a interface{}, b interface{}) {
	if LEVEL_DEBUG > GetLogLevel() {
		return
	}
	L.formatBaseInfo()
	_log.output(LEVEL_DEBUG, L.buf, format, a, b)
}

func Debug3(format string, a interface{}, b interface{}, c interface{}) {
	if LEVEL_DEBUG > GetLogLevel() {
		return
	}

	_log.output(LEVEL_DEBUG, "", format, a, b, c)
}

func (L *Logger) Debug3(format string, a interface{}, b interface{}, c interface{}) {
	if LEVEL_DEBUG > GetLogLevel() {
		return
	}
	L.formatBaseInfo()
	_log.output(LEVEL_DEBUG, L.buf, format, a, b, c)
}

func Debug4(format string, a interface{}, b interface{}, c interface{}, d interface{}) {
	if LEVEL_DEBUG > GetLogLevel() {
		return
	}

	_log.output(LEVEL_DEBUG, "", format, a, b, c, d)
}

func (L *Logger) Debug4(format string, a interface{}, b interface{}, c interface{}, d interface{}) {
	if LEVEL_DEBUG > GetLogLevel() {
		return
	}
	L.formatBaseInfo()
	_log.output(LEVEL_DEBUG, L.buf, format, a, b, c, d)
}

func Info1(format string, a interface{}) {
	if LEVEL_INFO > GetLogLevel() {
		return
	}

	_log.output(LEVEL_INFO, "", format, a)
}

func (L *Logger) Info1(format string, a interface{}) {
	if LEVEL_INFO > GetLogLevel() {
		return
	}
	L.formatBaseInfo()
	_log.output(LEVEL_INFO, L.buf, format, a)
}

func Info2(format string, a interface{}, b interface{}) {
	if LEVEL_INFO > GetLogLevel() {
		return
	}

	_log.output(LEVEL_INFO, "", format, a, b)
}

func (L *Logger) Info2(format string, a interface{}, b interface{}) {
	if LEVEL_INFO > GetLogLevel() {
		return
	}
	L.formatBaseInfo()
	_log.output(LEVEL_INFO, L.buf, format, a, b)
}

func Info3(format string, a interface{}, b interface{}, c interface{}) {
	if LEVEL_INFO > GetLogLevel() {
		return
	}

	_log.output(LEVEL_INFO, "", format, a, b, c)
}

func (L *Logger) Info3(format string, a interface{}, b interface{}, c interface{}) {
	if LEVEL_INFO > GetLogLevel() {
		return
	}
	L.formatBaseInfo()
	_log.output(LEVEL_INFO, L.buf, format, a, b, c)
}

func Info4(format string, a interface{}, b interface{}, c interface{}, d interface{}) {
	if LEVEL_INFO > GetLogLevel() {
		return
	}

	_log.output(LEVEL_INFO, "", format, a, b, c, d)
}

func (L *Logger) Info4(format string, a interface{}, b interface{}, c interface{}, d interface{}) {
	if LEVEL_INFO > GetLogLevel() {
		return
	}
	L.formatBaseInfo()
	_log.output(LEVEL_INFO, L.buf, format, a, b, c, d)
}

// Cheap integer to fixed-width decimal ASCII.
// Give a negative width to avoid zero-padding.
// Knows the buffer has capacity.
func itoa(buf *[]byte, i int, wid int) {
	var u = uint(i)
	if u == 0 && wid <= 1 {
		*buf = append(*buf, '0')
		return
	}

	// Assemble decimal in reverse order.
	var b [32]byte
	bp := len(b)
	for ; u > 0 || wid > 0; u /= 10 {
		bp--
		wid--
		b[bp] = byte(u%10) + '0'
	}
	*buf = append(*buf, b[bp:]...)
}

func (l *LoggerBase) formatHeader(buf *[]byte, t time.Time,
	level int32, file string, line int) {

	//2015-05-14
	year, month, day := t.Date()
	itoa(buf, year, 4)
	*buf = append(*buf, '-')
	itoa(buf, int(month), 2)
	*buf = append(*buf, '-')
	itoa(buf, day, 2)
	*buf = append(*buf, ' ')

	//09:56:00.023132
	hour, min, sec := t.Clock()
	itoa(buf, hour, 2)
	*buf = append(*buf, ':')
	itoa(buf, min, 2)
	*buf = append(*buf, ':')
	itoa(buf, sec, 2)
	if l.microseconds {
		*buf = append(*buf, '.')
		itoa(buf, t.Nanosecond()/1e3, 6)
	}
	*buf = append(*buf, ' ')

	*buf = append(*buf, levelStrings[level]...)
	*buf = append(*buf, ' ')

	short := file
	if l.printCatal {
		index := strings.LastIndex(file, "src")
		file = file[index:]
	} else {
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
	}

	*buf = append(*buf, file...)
	*buf = append(*buf, ':')
	itoa(buf, line, -1)
	*buf = append(*buf, ": "...)
}

func (l *LoggerBase) formatHeaderBasic(buf *[]byte, t time.Time, file string, line int) {

	//2015-05-14
	year, month, day := t.Date()
	itoa(buf, year, 4)
	*buf = append(*buf, '-')
	itoa(buf, int(month), 2)
	*buf = append(*buf, '-')
	itoa(buf, day, 2)
	*buf = append(*buf, ' ')

	//09:56:00.023132
	hour, min, sec := t.Clock()
	itoa(buf, hour, 2)
	*buf = append(*buf, ':')
	itoa(buf, min, 2)
	*buf = append(*buf, ':')
	itoa(buf, sec, 2)
	if l.microseconds {
		*buf = append(*buf, '.')
		itoa(buf, t.Nanosecond()/1e3, 6)
	}
	*buf = append(*buf, ' ')

	// xxx.go (filename)
	short := file
	if l.printCatal {
		index := strings.LastIndex(file, "src")
		file = file[index:]
	} else {
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
	}

	*buf = append(*buf, file...)
	*buf = append(*buf, ':')
	itoa(buf, line, -1)
	*buf = append(*buf, ": "...)
}

func (l *LoggerBase) output(level int32, baseinfo, format string, v ...interface{}) error {
	if level > GetLogLevel() {
		return nil
	}

	var s string
	if len(v) == 0 {
		s = format
	} else {
		s = fmt.Sprintf(format, v...)
	}

	now := time.Now() // get this early.
	var file string
	var line int
	l.mu.Lock()
	defer l.mu.Unlock()

	// release lock while getting caller info - it's expensive.
	l.mu.Unlock()
	var ok bool
	_, file, line, ok = runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}
	l.mu.Lock()

	l.buf = l.buf[:0]
	l.formatHeader(&l.buf, now, level, file, line)
	l.buf = append(l.buf, baseinfo...)
	l.buf = append(l.buf, s...)
	if len(s) > 0 && s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}

	var err error
	if level >= LEVEL_NOTICE {
		_, err = l.out.Write(l.buf)
	} else if level <= LEVEL_WARNING {
		_, err = l.out_wf.Write(l.buf)
	}

	return err
}

func GetDefaultLogger() *LoggerBase {
	return _log
}

func (l *LoggerBase) Output(calldepth int, s string) error {
	now := time.Now() // get this early.
	var file string
	var line int
	l.mu.Lock()
	defer l.mu.Unlock()

	// release lock while getting caller info - it's expensive.
	l.mu.Unlock()
	var ok bool
	_, file, line, ok = runtime.Caller(calldepth)
	if !ok {
		file = "???"
		line = 0
	}
	l.mu.Lock()

	l.buf = l.buf[:0]
	l.formatHeaderBasic(&l.buf, now, file, line)
	l.buf = append(l.buf, s...)
	if len(s) > 0 && s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}

	_, err := l.out.Write(l.buf)

	return err
}
