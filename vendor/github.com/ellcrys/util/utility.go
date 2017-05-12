package util

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	r "math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	pkgErrors "github.com/pkg/errors"

	"encoding/hex"

	"github.com/cbroglie/mustache"
	"github.com/franela/goreq"
	"github.com/hokaccha/go-prettyjson"
	spin "github.com/ncodes/go-spin"
	"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2/bson"
)

func init() {
	r.Seed(time.Now().UnixNano())
	goreq.SetConnectTimeout(5 * time.Second)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// UUID4 returns a UUID v4 string
func UUID4() string {
	return uuid.NewV4().String()
}

// Sha1 returns a sha1 hash
func Sha1(str string) string {
	h := sha1.New()
	h.Write([]byte(str))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Sha256 returns a sha256 hash
func Sha256(str string) string {
	h := sha256.New()
	h.Write([]byte(str))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// DecodeJSONToMap decode json string to map
func DecodeJSONToMap(str string) (map[string]interface{}, error) {
	var dat map[string]interface{}
	if err := json.Unmarshal([]byte(str), &dat); err != nil {
		return map[string]interface{}{}, err
	}
	return dat, nil
}

// RandString gets random string of fixed length
func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[r.Intn(len(letterBytes))]
	}
	return string(b)
}

// ReadReader reads n bytes from a pipe and pass bytes read to a callback. If an error occurs
// error is passed to the callback. The callback signature is:
// Func(err error, bs []byte, done bool).
func ReadReader(reader io.Reader, nBytes int, cb func(err error, bs []byte, done bool) bool) {
	r := bufio.NewReader(reader)
	buf := make([]byte, 0, nBytes)
	for {
		n, err := r.Read(buf[:cap(buf)])
		buf = buf[:n]
		if n == 0 {
			if err == nil {
				continue
			}
			if err == io.EOF {
				break
			}
			if cb(err, buf, false) == false {
				return
			}
			break
		}

		// process buf
		if err != nil && err != io.EOF {
			if cb(err, buf, false) == false {
				return
			}
		}

		if cb(err, buf, false) == false {
			return
		}
	}

	cb(nil, buf, true)
}

// Map casts an interface to Map type
func Map(m interface{}) map[string]interface{} {
	return m.(map[string]interface{})
}

// JSONToSliceString converts stringified JSON array to slice of string.
// JSON array must contain only string values
func JSONToSliceString(jsonStr string) ([]string, error) {
	var data []string
	d := json.NewDecoder(strings.NewReader(jsonStr))
	d.UseNumber()
	if err := d.Decode(&data); err != nil {
		return data, errors.New("unable to parse json string")
	}
	return data, nil
}

// JSONToSliceOfMap converts stringified JSON array to slice of maps.
// JSON array must contain only maps
func JSONToSliceOfMap(jsonStr string) ([]map[string]interface{}, error) {
	var data []map[string]interface{}
	d := json.NewDecoder(strings.NewReader(jsonStr))
	d.UseNumber()
	if err := d.Decode(&data); err != nil {
		return data, errors.New("unable to parse json string")
	}
	return data, nil
}

// ByteArrToString converts a byte array to string
func ByteArrToString(byteArr []byte) string {
	return fmt.Sprintf("%s", byteArr)
}

// ReadFromFixtures reads files from /tests/fixtures/ directory
func ReadFromFixtures(path string) string {
	absPath, _ := filepath.Abs(path)
	dat, err := ioutil.ReadFile(absPath)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s", dat)
}

// RandNum generates random numbers between a range
func RandNum(min, max int) int {
	s1 := r.NewSource(time.Now().UnixNano())
	r1 := r.New(s1)
	return r1.Intn(max-min) + min
}

// NewID generates an id for use as an id.
func NewID() string {
	curTime := int(time.Now().Unix())
	id := fmt.Sprintf("%s:%d", UUID4(), curTime)
	return Sha1(id)
}

// InStringSlice takes an interface that it expects to be
// a slice of string or a slice of interface and attempts to find
// a string value in the slice. The the list is a slice of interface
// it cast the values of the list to string and compares with the value
// being searched for. Returns true if the value is found or false.
func InStringSlice(list interface{}, val string) bool {
	switch v := list.(type) {
	case []interface{}:
		for _, s := range v {
			if s.(string) == val {
				return true
			}
		}
		break
	case []string:
		for _, s := range v {
			if s == val {
				return true
			}
		}
	default:
		panic("unsupported type")
	}
	return false
}

// InStringSliceRx check whether a regex pattern matches an item in
// a string only slice
func InStringSliceRx(strs []string, pattern string) bool {
	for _, str := range strs {
		if match, _ := regexp.MatchString(pattern, str); match {
			return true
		}
	}
	return false
}

// HasKey checks if a key exists in a map
func HasKey(m map[string]interface{}, key string) bool {
	for k := range m {
		if k == key {
			return true
		}
	}
	return false
}

// IsStringValue checks that a value type is string
func IsStringValue(any interface{}) bool {
	switch any.(type) {
	case string:
		return true
	default:
		return false
	}
}

// GetMapKeys gets all the keys of a map
func GetMapKeys(m map[string]interface{}) []string {
	mk := make([]string, len(m))
	i := 0
	for key := range m {
		mk[i] = key
		i++
	}
	return mk
}

// IsMapOfAny checks that a variable value type is a map of any value
func IsMapOfAny(any interface{}) bool {
	switch any.(type) {
	case map[string]interface{}:
		return true
		break
	default:
		return false
		break
	}
	return false
}

// IsSlice checks that a variable value type is a slice
func IsSlice(any interface{}) bool {
	switch any.(type) {
	case []interface{}:
		return true
		break
	default:
		return false
		break
	}
	return false
}

// ContainsOnlyMapType checks that a slice contains map[string]interface{} type
func ContainsOnlyMapType(s []interface{}) bool {
	for _, v := range s {
		switch v.(type) {
		case map[string]interface{}:
			continue
			break
		default:
			return false
		}
	}
	return true
}

// IsSliceOfStrings Checks if a slice contains only string values
func IsSliceOfStrings(s []interface{}) bool {
	for _, v := range s {
		switch v.(type) {
		case string:
			continue
			break
		default:
			return false
		}
	}
	return true
}

// UnixToTime converts a unix time to time object
func UnixToTime(i int64) time.Time {
	return time.Unix(i, 0)
}

// IsNumberValue checks whether the value passed is int, float64, float32 or int64
func IsNumberValue(val interface{}) bool {
	switch val.(type) {
	case int, int64, float32, float64:
		return true
	default:
		return false
	}
}

// IsInt checks whether the value passed is an integer
func IsInt(val interface{}) bool {
	switch val.(type) {
	case int, int64:
		return true
	default:
		return false
	}
}

// IsJSONNumber checks whether the value passed is a json.Number type
func IsJSONNumber(val interface{}) bool {
	switch val.(type) {
	case json.Number:
		return true
	default:
		return false
	}
}

// IntToFloat64 cast int value to float64
func IntToFloat64(num interface{}) float64 {
	switch v := num.(type) {
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		panic("failed to cast unsupported type to float64")
	}
}

// ToInt64 converts int, float32 and float64 to int64
func ToInt64(num interface{}) int64 {
	switch v := num.(type) {
	case int:
		return int64(v)
		break
	case int64:
		return v
		break
	case float64:
		return int64(v)
		break
	case string:
		val, _ := strconv.ParseInt(v, 10, 64)
		return val
		break
	default:
		panic("type is unsupported")
	}
	return 0
}

// Env gets environment variable or return a default value when no set
func Env(key, defStr string) string {
	val := os.Getenv(key)
	if val == "" && defStr != "" {
		return defStr
	}
	return val
}

// IsMapEmpty check if a map is empty
func IsMapEmpty(m map[string]interface{}) bool {
	return len(GetMapKeys(m)) == 0
}

// IntToString converts int to string
func IntToString(v int64) string {
	return fmt.Sprintf("%d", v)
}

// MapToJSON returns a json string representation of a map
func MapToJSON(m map[string]interface{}) (string, error) {
	bs, err := json.Marshal(&m)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

// JSONToMap decodes a json string to a map
func JSONToMap(jsonStr string) (map[string]interface{}, error) {
	var data map[string]interface{}
	d := json.NewDecoder(strings.NewReader(jsonStr))
	d.UseNumber()
	if err := d.Decode(&data); err != nil {
		return make(map[string]interface{}), errors.New("unable to parse json string")
	}
	return data, nil
}

// GetJWSPayload gets the encoded payload from a JWS token
func GetJWSPayload(token string) (string, error) {
	var parts = strings.Split(token, ".")
	if len(parts) != 3 {
		return "", errors.New("parameter is not a valid token")
	}
	return parts[1], nil
}

// ReadJSONFile reads and decode json file
func ReadJSONFile(filePath string) (map[string]interface{}, error) {

	var key map[string]interface{}

	// load file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return key, errors.New("failed to load file: " + filePath)
	}

	// parse file to json
	jsonData, err := DecodeJSONToMap(string(data))
	if err != nil {
		return key, errors.New("failed to decode file: " + filePath)
	}

	return jsonData, nil
}

// GetIPs returns all the ip chains of the request.
// The first ip is the remote address and the
// rest are extracted from the x-forwarded-for header
func GetIPs(req *http.Request) []string {

	var ips []string
	var remoteAddr = req.RemoteAddr
	if remoteAddr != "" {
		ipParts := strings.Split(remoteAddr, ":")
		ips = append(ips, ipParts[0])
	}

	// fetch ips in x-forwarded-for header
	var xForwardedFor = strings.TrimSpace(req.Header.Get("x-forwarded-for"))
	if xForwardedFor != "" {
		xForwardedForParts := strings.Split(xForwardedFor, ", ")
		for _, ip := range xForwardedForParts {
			if !InStringSlice(ips, ip) {
				ips = append(ips, ip)
			}
		}
	}

	return ips
}

// NewGetRequest creates a http GET request
func NewGetRequest(url string, headers map[string]string) (*http.Response, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, strings.NewReader(""))
	for key, val := range headers {
		req.Header.Set(key, val)
	}
	return client.Do(req)
}

// NewPostRequest creates a http POST request
func NewPostRequest(url, body string, headers map[string]string) (*http.Response, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, strings.NewReader(body))
	for key, val := range headers {
		req.Header.Set(key, val)
	}
	return client.Do(req)
}

// FloatToString given a float value, it returns a full
// string representation of the value
func FloatToString(floatVal float64, precision int) string {
	var v big.Float
	v.SetFloat64(floatVal)
	return v.Text('f', precision)
}

// StringSliceMatchString takes a slice of regex pattern and a string to
// match and returns the pattern that matches the string to match. Returns
// empty string no match is found.
func StringSliceMatchString(strPatterns []string, strToMatch string) string {
	for _, pat := range strPatterns {
		if match, _ := regexp.MatchString(pat, strToMatch); match {
			return pat
		}
	}
	return ""
}

// JSONNumberToInt64 converts a json.Number object to Int64.
// Panics if object is not a json.Number type.
func JSONNumberToInt64(val interface{}) int64 {
	switch v := val.(type) {
	case json.Number:
		num, err := v.Int64()
		if err != nil {
			panic("JSONNumberToInt64: " + err.Error())
		}
		return num
		break
	default:
		panic("JSONNumberToInt64: unknown type. Expects json.Number")
	}
	return 0
}

//  ToJSON converts struct or map to json
func ToJSON(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

// FromJSON converts json to struct or map
func FromJSON(data []byte, container interface{}) error {
	return json.Unmarshal(data, container)
}

// GetRandFromSlice returns a random value from a slice of interface.
func GetRandFromSlice(slice []interface{}) interface{} {
	if len(slice) == 0 {
		return nil
	}
	return slice[RandNum(0, len(slice))]
}

// GetRandFromStringSlice returns a random value from a slice of string
func GetRandFromStringSlice(sliceOfString []string) string {
	if len(sliceOfString) == 0 {
		return ""
	}
	return sliceOfString[RandNum(0, len(sliceOfString))]
}

// DownloadURL downloads content from a url and returns buffer
func DownloadURL(url string) (*bytes.Buffer, int, error) {

	res, err := goreq.Request{
		Method:       "GET",
		Uri:          url,
		MaxRedirects: 3,
	}.Do()

	if err != nil {
		return nil, 0, err
	}

	defer res.Body.Close()
	if res.StatusCode < 200 && res.StatusCode > 299 {
		str, err := res.Body.ToString()
		if err != nil {
			return nil, res.StatusCode, pkgErrors.Wrap(err, "failed to read body")
		}
		return nil, res.StatusCode, fmt.Errorf(str)
	}

	var data bytes.Buffer
	buf := make([]byte, 0, 1*1024)

	for {

		n, err := res.Body.Read(buf[:cap(buf)])
		buf = buf[:n]
		if n == 0 {
			if err == nil {
				continue
			}
			if err == io.EOF {
				break
			}
			return nil, res.StatusCode, err
		}

		if err != nil && err != io.EOF {
			return nil, res.StatusCode, err
		}

		data.Write(buf)
	}

	return &data, res.StatusCode, nil
}

// DownloadURLToFunc fetches the content from a url and pass
// downloaded chunks to a callback function.
// If err is returned from the callback, the function returns
func DownloadURLToFunc(url string, f func([]byte, int) error) error {

	res, err := goreq.Request{
		Method:       "GET",
		Uri:          url,
		MaxRedirects: 3,
	}.Do()

	if err != nil {
		return err
	}

	defer res.Body.Close()
	if res.StatusCode < 200 && res.StatusCode > 299 {
		return f(nil, res.StatusCode)
	}

	buf := make([]byte, 0, 1*1024)

	for {
		n, err := res.Body.Read(buf[:cap(buf)])
		buf = buf[:n]
		if n == 0 {
			if err == nil {
				continue
			}
			if err == io.EOF {
				break
			}
			return err
		}

		if err != nil && err != io.EOF {
			return err
		}

		if err = f(buf, res.StatusCode); err != nil {
			return err
		}
	}

	return nil
}

// Printify prints pretty objects
func Printify(d interface{}) {
	bs, err := prettyjson.Marshal(d)
	if err != nil {
		panic(err)
	}
	Println(string(bs))
}

// RenderTemp takes a template string and parses it with the temp data.
func RenderTemp(data string, temp map[string]interface{}) string {
	data, err := mustache.Render(data, temp)
	if err != nil {
		panic(err)
	}
	return data
}

// FromBSONMap converts bson map to a different struct or map type
func FromBSONMap(d bson.M, container interface{}) {

	jsonBs, err := ToJSON(d)
	if err != nil {
		panic(err)
	}

	FromJSON(jsonBs, container)
}

// IfNilEmptyStringSlice returns an empty string slice.
func IfNilEmptyStringSlice() interface{} {
	return []string{}
}

// IfNil will run a callback function if val is nil. Otherwise, it
// return val
func IfNil(val interface{}, cb func() interface{}) interface{} {
	if val == nil {
		return cb()
	}
	return val
}

// GetDupItem takes a slice of maps or a slice of struct type and
// looks for items in the slice that have a specific field that
// already exists in the slice. It returns the duplicate item and its
// position in the slice.
func GetDupItem(obj interface{}, key string) (interface{}, int) {

	var sliceOfMap []map[string]interface{}
	var foundValues []string

	switch val := obj.(type) {
	case []map[string]interface{}:
		sliceOfMap = val
	default:
		bs, err := ToJSON(obj)
		if err != nil {
			panic(err)
		}
		if err = FromJSON(bs, &sliceOfMap); err != nil {
			panic(err)
		}
	}

	for i, m := range sliceOfMap {
		if !InStringSlice(foundValues, m[key].(string)) {
			foundValues = append(foundValues, m[key].(string))
		} else {
			return m, i
		}
	}

	return nil, 0
}

// RemoveEmptyInStringSlice takes a slice of strings and removes empty values
func RemoveEmptyInStringSlice(l []string) []string {
	var newSlice []string
	for _, s := range l {
		if len(strings.TrimSpace(s)) != 0 {
			newSlice = append(newSlice, s)
		}
	}
	return newSlice
}

// Spinner prints a loading spinner to the terminal.
// Returns a stop function to stop the spinner
func Spinner(txt string) func() {
	s := spin.New()
	stop := false
	go func() {
		for !stop {
			fmt.Printf("\r\033[36m\033[m %s %s", s.Next(), txt)
			time.Sleep(100 * time.Millisecond)
		}
		fmt.Printf("\r\033[2K")
	}()

	return func() {
		s.Done()
		stop = true
		time.Sleep(100 * time.Millisecond)
	}
}

// CryptoRandKey creates a random string of a specific length to be used
// as a cryptography key
func CryptoRandKey(length int) string {
	key := make([]byte, length)
	rand.Read(key)
	return hex.EncodeToString(key)
}

// UniqueStringSlice returns a new string slice with duplicates removed
func UniqueStringSlice(s []string) []string {
	var u = []string{}
	for _, _s := range s {
		if !InStringSlice(u, _s) {
			u = append(u, _s)
		}
	}
	return u
}
