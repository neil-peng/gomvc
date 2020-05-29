package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"math"
	"math/rand"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// in_array
func InIntArray(array []int, target int) bool {
	for _, v := range array {
		if v == target {
			return true
		}
	}
	return false
}

// Base64Encode base64_encode()
func Base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

//sha1
func HashHmac(data, keyStr string) string {
	key := []byte(keyStr)
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(data))
	return string(mac.Sum(nil))
}

// Dechex dechex()
func Dechex(number int64) string {
	return strconv.FormatInt(number, 16)
}

// Round round()
func Round(value float64) float64 {
	return math.Floor(value + 0.5)
}

// Floor floor()
func Floor(value float64) float64 {
	return math.Floor(value)
}

// Hexdec hexdec()
func Hexdec(str string) (int64, error) {
	return strconv.ParseInt(str, 16, 0)
}

// in_array, ingore case
func InStringArray(array []string, target string) bool {
	for _, v := range array {
		if strings.EqualFold(v, target) {
			return true
		}
	}
	return false
}

// IP2long ip2long()
// IPv4
func IP2long(ipAddress string) uint32 {
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return 0
	}
	return binary.BigEndian.Uint32(ip.To4())
}

func IntArrayReverse(s []int) []int {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// Ceil ceil()
func Ceil(value float64) float64 {
	return math.Ceil(value)
}

// StrReplace str_replace()
func StrReplace(search, replace, subject string, count int) string {
	return strings.Replace(subject, search, replace, count)
}

// Substr substr()
func Substr(str string, start int, length int) string {
	if start < 0 || length < -1 {
		return str
	}
	switch {
	case length == -1:
		return str[start:]
	case length == 0:
		return ""
	}
	end := int(start) + length
	if end > len(str) {
		end = len(str)
	}
	return str[start:end]
}

//md5()
func Md5(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	bytes := h.Sum(nil)
	result := hex.EncodeToString(bytes)
	return result
}

// HTTPBuildQuery http_build_query()
func HTTPBuildQuery(queryData url.Values) string {
	return queryData.Encode()
}

//in_array()
func InArray(obj interface{}, target interface{}) bool {
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true
		}
	}

	return false
}

// Basename basename()
func Basename(path string) string {
	return filepath.Base(path)
}

// Strstr strstr()
func Strstr(haystack string, needle string) string {
	if needle == "" {
		return ""
	}
	idx := strings.Index(haystack, needle)
	if idx == -1 {
		return ""
	}
	return haystack[idx:]
}

// Strtr strtr()
//
// If the parameter length is 1, type is: map[string]string
// Strtr("baab", map[string]string{"ab": "01"}) will return "ba01"
// If the parameter length is 2, type is: string, string
// Strtr("baab", "ab", "01") will return "1001", a => 0; b => 1.
func Strtr(haystack string, params ...interface{}) string {
	ac := len(params)
	if ac == 1 {
		pairs := params[0].(map[string]string)
		length := len(pairs)
		if length == 0 {
			return haystack
		}
		oldnew := make([]string, length*2)
		for o, n := range pairs {
			if o == "" {
				return haystack
			}
			oldnew = append(oldnew, o, n)
		}
		return strings.NewReplacer(oldnew...).Replace(haystack)
	} else if ac == 2 {
		from := params[0].(string)
		to := params[1].(string)
		trlen, lt := len(from), len(to)
		if trlen > lt {
			trlen = lt
		}
		if trlen == 0 {
			return haystack
		}

		str := make([]uint8, len(haystack))
		var xlat [256]uint8
		var i int
		var j uint8
		if trlen == 1 {
			for i = 0; i < len(haystack); i++ {
				if haystack[i] == from[0] {
					str[i] = to[0]
				} else {
					str[i] = haystack[i]
				}
			}
			return string(str)
		}
		// trlen != 1
		for {
			xlat[j] = j
			if j++; j == 0 {
				break
			}
		}
		for i = 0; i < trlen; i++ {
			xlat[from[i]] = to[i]
		}
		for i = 0; i < len(haystack); i++ {
			str[i] = xlat[haystack[i]]
		}
		return string(str)
	}

	return haystack
}

// Trim trim()
func Trim(str string, characterMask ...string) string {
	mask := ""
	if len(characterMask) == 0 {
		return strings.TrimSpace(str)
	} else {
		mask = characterMask[0]
	}
	return strings.Trim(str, mask)
}

// Ltrim ltrim()
func Ltrim(str string, characterMask ...string) string {
	mask := ""
	if len(characterMask) == 0 {
		return strings.TrimLeftFunc(str, unicode.IsSpace)
	} else {
		mask = characterMask[0]
	}
	return strings.TrimLeft(str, mask)
}

// Rtrim rtrim()
func Rtrim(str string, characterMask ...string) string {
	mask := ""
	if len(characterMask) == 0 {
		return strings.TrimRightFunc(str, unicode.IsSpace)
	} else {
		mask = characterMask[0]
	}
	return strings.TrimRight(str, mask)
}

// Ord ord()
func Ord(char string) int {
	r, _ := utf8.DecodeRune([]byte(char))
	return int(r)
}

// JSONDecode json_decode()
func JSONDecode(data []byte, val interface{}) error {
	return json.Unmarshal(data, val)
}

//默认json.Marshal会转移<>&等特殊字符
func JSONEncode(val interface{}) ([]byte, error) {
	var b []byte
	buf := bytes.NewBuffer(b)
	jsEncoder := json.NewEncoder(buf)
	jsEncoder.SetEscapeHTML(false)
	if err := jsEncoder.Encode(val); err != nil {
		return nil, err
	}
	b = buf.Bytes()
	return b[:len(b)-1], nil
}

func InSplitString(src string, target string, split string) bool {
	if len(src) == 0 {
		return false
	}

	srcArray := strings.Split(src, split)
	for _, v := range srcArray {
		vt := Trim(v)
		if strings.EqualFold(vt, target) {
			return true
		}
	}
	return false
}

// Implode implode()
func ImplodeInt(glue string, pieces []int) string {
	var buf bytes.Buffer
	l := len(pieces)
	for _, intP := range pieces {
		buf.WriteString(fmt.Sprintf("%d", intP))
		if l--; l > 0 {
			buf.WriteString(glue)
		}
	}
	return buf.String()
}

func ImplodeInt64(glue string, pieces []int64) string {
	var buf bytes.Buffer
	l := len(pieces)
	for _, intP := range pieces {
		buf.WriteString(fmt.Sprintf("%d", intP))
		if l--; l > 0 {
			buf.WriteString(glue)
		}
	}
	return buf.String()
}

// Addslashes addslashes()
func Addslashes(str string) string {
	var buf bytes.Buffer
	for _, char := range str {
		switch char {
		case '\'', '"', '\\':
			buf.WriteRune('\\')
		}
		buf.WriteRune(char)
	}
	return buf.String()
}

// Stripslashes stripslashes()
func Stripslashes(str string) string {
	var buf bytes.Buffer
	l, skip := len(str), false
	for i, char := range str {
		if skip {
			skip = false
		} else if char == '\\' {
			if i+1 < l && str[i+1] == '\\' {
				skip = true
			}
			continue
		}
		buf.WriteRune(char)
	}
	return buf.String()
}

// URLEncode urlencode()
func URLEncode(str string) string {
	return url.QueryEscape(str)
}

// PregReplace preg_replace()
func PregReplace(pattern string, replacement string, subject string) string {
	re, _ := regexp.Compile(pattern)
	return re.ReplaceAllString(subject, replacement)
}

// Rawurlencode rawurlencode()
func Rawurlencode(str string) string {
	return strings.Replace(url.QueryEscape(str), "+", "%20", -1)
}

// Rawurldecode rawurldecode()
func Rawurldecode(str string) (string, error) {
	return url.QueryUnescape(strings.Replace(str, "%20", "+", -1))
}

// Gethostname gethostname()
func Gethostname() (string, error) {
	return os.Hostname()
}

func GetHostIp() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	var hostIp string
	for _, item := range ifaces {
		addrs, err := item.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			ipStr := fmt.Sprintf("%s", ip)
			if ipStr != "127.0.0.1" {
				hostIp = ipStr
				break
			}
		}
	}
	if len(hostIp) == 0 {
		hostIp = "127.0.0.1"
		return hostIp, nil
	}
	return hostIp, nil
}

// Strtolower strtolower()
func Strtolower(str string) string {
	return strings.ToLower(str)
}

// ArrayChunk array_chunk()
func ArrayChunk(s []int64, size int) [][]int64 {
	if size < 1 {
		panic("size: cannot be less than 1")
	}
	length := len(s)
	chunks := int(math.Ceil(float64(length) / float64(size)))
	var n [][]int64
	for i, end := 0, 0; chunks > 0; chunks-- {
		end = (i + 1) * size
		if end > length {
			end = length
		}
		n = append(n, s[i*size:end])
		i++
	}
	return n
}

// array_chunk map
func ArrayChunkMap(s []map[string]interface{}, size int) [][]map[string]interface{} {
	if size < 1 {
		panic("size: cannot be less than 1")
	}
	length := len(s)
	chunks := int(math.Ceil(float64(length) / float64(size)))
	var n [][]map[string]interface{}
	for i, end := 0, 0; chunks > 0; chunks-- {
		end = (i + 1) * size
		if end > length {
			end = length
		}
		n = append(n, s[i*size:end])
		i++
	}
	return n
}

// Empty empty()
func Empty(val interface{}) bool {
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.String, reflect.Array:
		return v.Len() == 0
	case reflect.Map, reflect.Slice:
		return v.Len() == 0 || v.IsNil()
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}

	return reflect.DeepEqual(val, reflect.Zero(v.Type()).Interface())
}

// Rand rand()
func Rand(min, max int) int {
	if min > max {
		min, max = max, min
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := r.Intn(math.MaxInt32)
	return n/((math.MaxInt32+1)/(max-min+1)) + min
}

// Explode explode()
func Explode(delimiter, str string) []string {
	return strings.Split(str, delimiter)
}

//dirname()
func Dirname(path string) string {
	rpos := strings.LastIndex(path, "/")
	return Substr(path, 0, rpos)
}

//mt_rand()
func Mt_rand(min, max int64) int64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Int63n(max-min+1) + min
}

// Crc32 crc32()
func Crc32(str string) uint32 {
	return crc32.ChecksumIEEE([]byte(str))
}

func IsNumeric(b byte) bool {
	if _, err := strconv.Atoi(string(b)); err == nil {
		return true
	}
	return false
}

func PostfixType(fileName string, fileSchemes map[string][]string) string {
	fileType := "unknown"
	fileName = Trim(fileName)
	if index := strings.LastIndex(fileName, "."); index > 0 && index < len(fileName) {
		postFix := fileName[index+1:]
		for typeName, typeList := range fileSchemes {
			if InStringArray(typeList, postFix) {
				fileType = typeName
				break
			}
		}
	}
	return fileType
}

func InterfaceToString(val interface{}) string {
	switch val.(type) {
	case string:
		return val.(string)
	case int:
		return strconv.FormatInt(int64(val.(int)), 10)
	case int64:
		return strconv.FormatInt(val.(int64), 10)
	case uint64:
		return strconv.FormatUint(val.(uint64), 10)
	case float32:
		return strconv.FormatFloat(float64(val.(float32)), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(val.(float64), 'f', -1, 64)
	default:
		bytes, _ := json.Marshal(val)
		return string(bytes)
	}
}

func Mstos(ms int) (s int) {
	if ms%1000 == 0 {
		s = ms / 1000
	} else {
		s = ms/1000 + 1
	}

	return s
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func SplitHexStringToArry(sign string) (res [4]int64, err error) {
	if len(sign) != 32 {
		err = errors.New("invalid sign:" + sign)
		return
	}
	res[0], err = strconv.ParseInt(sign[:8], 16, 64)
	if err != nil {
		return
	}
	res[1], err = strconv.ParseInt(sign[8:16], 16, 64)
	if err != nil {
		return
	}
	res[2], err = strconv.ParseInt(sign[16:24], 16, 64)
	if err != nil {
		return
	}
	res[3], err = strconv.ParseInt(sign[24:], 16, 64)
	return
}

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandId() string {
	date := time.Now()
	base := time.Date(date.Year(), 1, 1, 0, 0, 0, 0, date.Location())
	offset := time.Since(base).Nanoseconds() / (int64(time.Millisecond) / int64(time.Nanosecond))
	rand_num := r.Int63n(2<<27 - 1)
	return fmt.Sprintf("%d", uint64(offset*2<<28+rand_num))
}
