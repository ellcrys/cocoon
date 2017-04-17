package util

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ellcrys/crypto"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2/bson"
)

// TestGenerateKeyPair tests that public and private key pairs are set
func TestGenerateKeyPair(t *testing.T) {
	keys, err := crypto.GenerateKeyPair()
	assert.Nil(t, err, "nil not expected")
	assert.NotNil(t, keys["private_key"], "private key should be set")
	assert.NotNil(t, keys["public_key"], "public key should be set")
}

// TestSha1 tests that sha1 hashing a string will return expected hash
func TestSha1(t *testing.T) {
	txt := "john"
	h := Sha1(txt)
	assert.Equal(t, h, "a51dda7c7ff50b61eaea0444371f4a6a9301e501")
}

// TestSha256 tests that sha1 hashing a string will return expected hash
func TestSha256(t *testing.T) {
	txt := "john"
	h := Sha256(txt)
	assert.Equal(t, h, "96d9632f363564cc3032521409cf22a852f2032eec099ed5967c0d000cec607a")
}

// TestDecodeJSONToMap tests that a json string is successfully decoded to a map
func TestDecodeJSONToMap(t *testing.T) {
	var s = `{"name":"john doe"}`
	m, err := DecodeJSONToMap(s)
	assert.Nil(t, err)
	assert.Equal(t, m["name"], "john doe")
}

// TestReadOneByteAtATime tests that the ReadReader can read one byte at a time
// and pass the read byte to the passed callback
func TestReadOneByteAtATime(t *testing.T) {
	var count = 0
	text := "hello"
	reader := strings.NewReader(text)
	ReadReader(reader, 1, func(err error, bs []byte, done bool) bool {
		assert.Nil(t, err)
		if !done {
			assert.Equal(t, string(bs), string(text[count]))
			count++
		}
		return true
	})
}

func TestJSONToSliceString(t *testing.T) {
	var d = `["kennedy", "odion"]`
	s, err := JSONToSliceString(d)
	assert.Nil(t, err)
	assert.Equal(t, len(s), 2)
	assert.Equal(t, s[0], "kennedy")
	assert.Equal(t, s[1], "odion")
}

// TestByteArrToString tests that a byte array is properly converted to string
func TestByteArrToString(t *testing.T) {
	var s = ByteArrToString([]byte("hello"))
	assert.Equal(t, s, "hello")
}

// TestNewID tests should create an id with 40 characters
func TestNewID(t *testing.T) {
	var id = NewID()
	assert.Equal(t, len(id), 40)
}

// TestValueNotInStringSlice tests that a string value is not contained in a string slice
func TestValueNotInStringSlice(t *testing.T) {
	var ss = []string{"john", "doe"}
	var r = InStringSlice(ss, "jane")
	assert.Equal(t, r, false)
}

// TestValueInStringSlice tests that a string value is contained in a string slice or not
func TestValueInStringSlice(t *testing.T) {
	var ss = []string{"john", "doe"}
	assert.Equal(t, InStringSlice(ss, "john"), true)
	assert.Equal(t, InStringSlice(ss, "jane"), false)
}

// TestGetMapKeys tests that GetMapKeys will return a list of all keys it contains
func TestGetMapKeys(t *testing.T) {
	var keys = GetMapKeys(map[string]interface{}{
		"key1": "0",
		"key2": "1",
	})
	r := InStringSlice(keys, "key1")
	r2 := InStringSlice(keys, "key2")
	assert.Equal(t, len(keys), 2)
	assert.Equal(t, r, true)
	assert.Equal(t, r2, true)
}

// TestHasKey tests that a key exist or doesn't exist in a map
func TestHasKey(t *testing.T) {
	var m = map[string]interface{}{
		"stuff_a": 2,
		"stuff_b": 3,
	}
	assert.Equal(t, HasKey(m, "stuff_b"), true)
	assert.Equal(t, HasKey(m, "stuff_a"), true)
	assert.Equal(t, HasKey(m, "stuff_c"), false)
}

// TestIsStringValue tests that a variable holds a string value or not
func TestIsStringValue(t *testing.T) {
	assert.Equal(t, IsStringValue("lorem"), true)
	assert.Equal(t, IsStringValue(20), false)
}

// TestIsMapOfAny tests that a variable's value type is a map of any type as value
func TestIsMapOfAny(t *testing.T) {
	var m = make(map[string]interface{})
	assert.Equal(t, IsMapOfAny(m), true)
	assert.Equal(t, IsMapOfAny(10), false)
}

// TestContainsOnlyMapType tests that a variable's value type is a slice containing only map type
func TestContainsOnlyMapType(t *testing.T) {
	s := []interface{}{
		map[string]interface{}{"a": "b"},
	}
	assert.Equal(t, true, ContainsOnlyMapType(s))
}

func TestUnixToTime(t *testing.T) {
	expected := time.Now().Unix()
	tm := UnixToTime(expected)
	assert.Equal(t, expected, tm.Unix())
}

func TestIsNumberValue(t *testing.T) {
	assert.Equal(t, IsNumberValue(1), true)
	assert.Equal(t, IsNumberValue(1.5), true)
	assert.Equal(t, IsNumberValue("abc"), false)
	assert.Equal(t, IsNumberValue([]string{}), false)
}

func TestIsInt(t *testing.T) {
	assert.Equal(t, IsInt(1), true)
	assert.Equal(t, IsInt(int64(1)), true)
	assert.Equal(t, IsInt(1.5), false)
	assert.Equal(t, IsInt("1"), false)
	assert.Equal(t, IsInt([]string{}), false)
}

func TestIsJSONNumber(t *testing.T) {
	var jn json.Number
	jn = "1"
	assert.Equal(t, IsJSONNumber(jn), true)
	assert.Equal(t, IsJSONNumber(1), false)
}

func TestIntToFloat64(t *testing.T) {
	expected := interface{}(IntToFloat64(1))
	switch expected.(type) {
	case float64:
		assert.Equal(t, expected, 1.0)
		break
	default:
		assert.Fail(t, "expected value type must be float64")
	}
}

func TestIsMapEmpty(t *testing.T) {
	expected := IsMapEmpty(map[string]interface{}{})
	expected2 := IsMapEmpty(map[string]interface{}{"a": "b"})
	assert.Equal(t, expected, true)
	assert.Equal(t, expected2, false)
}

func TestIntToString(t *testing.T) {
	assert.Equal(t, IntToString(1), "1")
}

func TestMapToJSON(t *testing.T) {
	m := map[string]interface{}{"a": "b"}
	v, err := MapToJSON(m)
	assert.Nil(t, err)
	assert.Equal(t, v, `{"a":"b"}`)
}

func TestJSONToMap(t *testing.T) {
	s := `{"a":"b"}`
	expected := map[string]interface{}{"a": "b"}
	m, err := JSONToMap(s)
	assert.Nil(t, err)
	assert.Exactly(t, m, expected)
}

func TestGetJWSPayload(t *testing.T) {
	var validJWS = `eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZGRyZXNzX2lkIjoiNTZmZDBiZTNmNjhkYWE0MGEzMDAwMDAyIiwiYXNzZXRfaWRzIjpbIjRiYTM0NDNiNjE1Nzc1M2ZiNjc4YTU5M2U1ZDc2ODRjODRkMWIyMDciXSwiZW5kcG9pbnQiOiIvdjEvYWRkcmVzc2VzLzU2ZmQ4YTE4ZjY4ZGFhNWMzZDAwMDAwMS9kZXBvc2l0IiwiaWF0IjoxNDU5Njc4NjkyfQ.hKWb7KtMCTf4VI5CCkJMS1r33DDDmyKiwuKU_DVPl8HRg152mUtqzRrdCUC2VXlryBJWtrXwZWu0NcAnJcraTSJ8IW60Lgi6Bd2pGTp-eDSkeR8kDYG4sMBMHKjm-fJaO0sYlEdgeZSp9bpV24417na_e6Ct7mRkNverQo0bL9M`
	var expected = strings.Split(validJWS, ".")
	p, err := GetJWSPayload(validJWS)
	assert.Nil(t, err)
	assert.Equal(t, p, expected[1])
}

func TestGetJWSPayloadWithInvalidJWSToken(t *testing.T) {
	var invalidJWS = `aa.aa`
	_, err := GetJWSPayload(invalidJWS)
	assert.NotNil(t, err)
}

func TestIsSliceOfStrings(t *testing.T) {
	var strSlice = []interface{}{"a", "b"}
	var nonStrSlice = []interface{}{"a", 2, "b"}
	assert.Equal(t, IsSliceOfStrings(strSlice), true)
	assert.Equal(t, IsSliceOfStrings(nonStrSlice), false)
}

func TestFloatToString(t *testing.T) {
	assert.Equal(t, FloatToString(0.0001, 3), "0.000")
	assert.Equal(t, FloatToString(0.0001, 4), "0.0001")
	assert.Equal(t, FloatToString(433, 3), "433.000")
}

func TestJSONToSliceOfMap(t *testing.T) {

	var tests = [][]interface{}{
		[]interface{}{`[ { "name": "john"  }, { "name": "john" } ]`, nil},
		[]interface{}{`[ { "name": "john"  }, 3 ]`, errors.New("unable to parse json string")},
	}

	for _, test := range tests {
		_, err := JSONToSliceOfMap(test[0].(string))
		assert.Equal(t, test[1], err)
	}
}

func TestInStringSliceRx(t *testing.T) {
	strs := []string{"container", "content", "COOl"}
	assert.Equal(t, InStringSliceRx(strs, ".*tent$"), true)
	assert.Equal(t, InStringSliceRx(strs, "cool"), false)
	assert.Equal(t, InStringSliceRx(strs, "(?i)cool"), true)
}

func TestStringSliceMatchString(t *testing.T) {
	patterns := []string{"(?i)central", "man"}
	assert.Equal(t, StringSliceMatchString(patterns, "CENTRAL"), "(?i)central")
	assert.Equal(t, StringSliceMatchString(patterns, "MAN"), "")
}

func TestJSONNumberToInt64(t *testing.T) {
	assert.Equal(t, int64(10), JSONNumberToInt64(json.Number("10")))
}

func TestDownloadURL(t *testing.T) {
	buf, status, err := DownloadURL("https://google.com.ng")
	assert.Nil(t, err)
	assert.Equal(t, status, 200)
	assert.NotNil(t, buf)
}

func TestDownloadURLToFunc(t *testing.T) {
	err := DownloadURLToFunc("https://google.com.ng", func(d []byte, status int) error {
		assert.Equal(t, 200, status)
		return nil
	})
	assert.Nil(t, err)
}

func TestRenderTemp(t *testing.T) {
	str := "My name is {{name}}"
	data := map[string]interface{}{
		"name": "Jeff",
	}
	assert.Equal(t, "My name is Jeff", RenderTemp(str, data))
}

func TestFromBSONMap(t *testing.T) {
	bsonM := bson.M{"_id": bson.NewObjectId(), "name": "John"}
	var container = struct {
		ID   string `json:"_id"`
		Name string `json:"name"`
	}{}
	FromBSONMap(bsonM, &container)
	assert.Equal(t, bsonM["_id"].(bson.ObjectId).Hex(), container.ID)
	assert.Equal(t, bsonM["name"].(string), container.Name)
}

func TestIfNil(t *testing.T) {
	val := IfNil(2, func() interface{} {
		return "a string"
	})
	assert.IsType(t, int(1), val)

	val = IfNil(nil, func() interface{} {
		return "a string"
	})
	assert.IsType(t, "string", val)

	val = IfNil(nil, IfNilEmptyStringSlice)
	assert.IsType(t, []string{}, val)
}

func TestGetDupItem(t *testing.T) {

	mapSlice := []map[string]interface{}{
		{"name": "john", "age": 20},
		{"name": "jane", "age": 20},
	}

	obj, pos := GetDupItem(&mapSlice, "name")
	assert.Equal(t, obj, nil)
	assert.Equal(t, pos, 0)

	mapSlice = []map[string]interface{}{
		{"name": "john", "age": 20},
		{"name": "john", "age": 50},
	}

	obj, pos = GetDupItem(&mapSlice, "name")
	assert.Equal(t, obj.(map[string]interface{})["age"].(float64), float64(50))
	assert.Equal(t, pos, 1)
}

func TestGetDupItemUsingStructType(t *testing.T) {

	var data = []struct {
		Name string
		Age  int
	}{
		{Name: "john", Age: 20},
		{Name: "jane", Age: 20},
	}

	obj, pos := GetDupItem(&data, "Name")
	assert.Equal(t, obj, nil)
	assert.Equal(t, pos, 0)

	data = []struct {
		Name string
		Age  int
	}{
		{Name: "john", Age: 20},
		{Name: "john", Age: 50},
	}

	obj, pos = GetDupItem(&data, "Name")
	assert.Equal(t, obj.(map[string]interface{})["Age"].(float64), float64(50))
	assert.Equal(t, pos, 1)
}

func TestRemoveEmptyInStringSlice(t *testing.T) {
	assert.Equal(t, RemoveEmptyInStringSlice([]string{"a", "", "c"}), []string{"a", "c"})
}
