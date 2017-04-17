package cstructs

import (
	"testing"

	"github.com/ellcrys/util"
	"github.com/stretchr/testify/assert"
)

// A
type A struct {
	Name  string
	Age   int
	About []byte
}

// B
type B struct {
	Name  string
	Age   int
	About []byte
}

// C
type C struct {
	Name   string
	Age    int
	About  []byte
	Gender string
}

// D
type D struct {
	NaMe   string
	AGe    int
	About  []byte
	Gender string
}

type E struct {
	Name       string
	Age        int
	About      []byte
	BestFriend *C
}

type F struct {
	Name       string
	Age        int
	About      []byte
	BestFriend *A
}

func TestCStruct1(t *testing.T) {
	ben := A{
		Name:  "benedith",
		Age:   20,
		About: []byte("Ben is a great sports person"),
	}

	var pete B
	Copy(&ben, &pete)

	assert.Equal(t, ben.Name, pete.Name)
	assert.Equal(t, ben.Age, pete.Age)
	assert.Equal(t, ben.About, pete.About)
}

func TestCStruct2(t *testing.T) {
	ben := A{
		Name:  "benedith",
		Age:   20,
		About: []byte("Ben is a great sports person"),
	}

	var pete C
	Copy(&ben, &pete)

	assert.Equal(t, ben.Name, pete.Name)
	assert.Equal(t, ben.Age, pete.Age)
	assert.Equal(t, ben.About, pete.About)
	assert.Empty(t, pete.Gender)
}

func TestCStruct3InnerStruct(t *testing.T) {
	ben := E{
		Name:  "benedith",
		Age:   20,
		About: []byte("Ben is a great sports person"),
		BestFriend: &C{
			Name:   "jon",
			Age:    19,
			About:  []byte("king of the north"),
			Gender: "male",
		},
	}
	var brad F
	err := Copy(&ben, &brad)
	assert.Nil(t, err)
	assert.Equal(t, brad.BestFriend.Name, "jon")
	assert.Equal(t, brad.BestFriend.Age, 19)
}

func BenchmarkCStruct3InnerStruct(b *testing.B) {
	ben := E{
		Name:  "benedith",
		Age:   20,
		About: []byte("Ben is a great sports person"),
		BestFriend: &C{
			Name:   "jon",
			Age:    19,
			About:  []byte("king of the north"),
			Gender: "male",
		},
	}
	var brad F
	err := Copy(&ben, &brad)
	assert.Nil(b, err)
}

func BenchmarkCStruct3InnerStructWithJSON(b *testing.B) {
	ben := E{
		Name:  "benedith",
		Age:   20,
		About: []byte("Ben is a great sports person"),
		BestFriend: &C{
			Name:   "jon",
			Age:    19,
			About:  []byte("king of the north"),
			Gender: "male",
		},
	}
	var brad F
	json, err := util.ToJSON(ben)
	assert.Nil(b, err)
	err = util.FromJSON(json, &brad)
	assert.Nil(b, err)
}

func TestIsSlice(t *testing.T) {
	assert.Equal(t, IsSlice([]string{}), true)
	assert.Equal(t, IsSlice("abc"), false)
}

func TestMakeSliceOf(t *testing.T) {
	assert.Equal(t, len(MakeSliceOf(A{}, 3)), 3)
	assert.Equal(t, len(MakeSliceOf(A{}, 4)), 4)
}

func TestCStructCopySliceSrcTypeErr(t *testing.T) {
	var sliceA = "aa"
	var sliceB = []*A{}
	expected := CopySlice(sliceA, sliceB)
	assert.Equal(t, expected.Error(), "src is not a slice")
}

func TestCStructCopySliceDestTypeErr(t *testing.T) {
	var sliceA = []*A{}
	var sliceB = "aa"
	expected := CopySlice(sliceA, sliceB)
	assert.Equal(t, expected.Error(), "dest is not a slice")
}

func TestCStructCopySliceUnequalLengthErr(t *testing.T) {
	var sliceA = []*A{&A{Name: "ben"}}
	var sliceB = []*A{}
	expected := CopySlice(sliceA, sliceB)
	assert.Equal(t, expected.Error(), "src and dest length are not equal")
}

func TestCStructCopySliceNilIfSrcIsEmpty(t *testing.T) {
	var sliceA = []*A{}
	var sliceB = []*A{&A{Name: "ben"}}
	expected := CopySlice(sliceA, sliceB)
	assert.Nil(t, expected)
}

func TestCStructCopySliceNonStructInSrc(t *testing.T) {
	var sliceA = []string{"abc"}
	var sliceB = []*A{&A{Name: "ben"}}
	expected := CopySlice(sliceA, sliceB)
	assert.Equal(t, expected.Error(), "found a non struct value in src. expects a slice of structs")
}

func TestCStructCopySliceNonStructInDest(t *testing.T) {
	var sliceA = []*A{&A{Name: "ben"}}
	var sliceB = []string{"abc"}
	expected := CopySlice(sliceA, sliceB)
	assert.Equal(t, expected.Error(), "found a non struct value in dest. expects a slice of structs")
}

func TestCStructCopySliceSuccess(t *testing.T) {
	var sliceA = []*A{&A{Name: "ben", Age: 12, About: []byte("cool person")}}
	var sliceB = []*A{&A{}}
	expected := CopySlice(sliceA, sliceB)
	assert.Nil(t, expected)
	assert.Equal(t, sliceA[0].Name, sliceB[0].Name)
	assert.Equal(t, sliceA[0].Age, sliceB[0].Age)
	assert.Equal(t, sliceA[0].About, sliceB[0].About)
}

func TestCStructCopySliceSuccess2(t *testing.T) {
	var sliceA = []*A{&A{Name: "ben", Age: 12, About: []byte("cool person")}}
	var sliceB = MakeSliceOf(A{}, 1)
	expected := CopySlice(sliceA, sliceB)
	assert.Nil(t, expected)
	assert.Equal(t, sliceA[0].Name, sliceB[0].(*A).Name)
	assert.Equal(t, sliceA[0].Age, sliceB[0].(*A).Age)
	assert.Equal(t, sliceA[0].About, sliceB[0].(*A).About)
}

func TestAAA(t *testing.T) {
	var a = F{
		Name:       "ken",
		Age:        15,
		About:      []byte("coder"),
		BestFriend: &A{Name: "pete"},
	}
	var b E
	Copy(a, &b)
}
