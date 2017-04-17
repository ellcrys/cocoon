package cstructs

import (
	"fmt"
	"reflect"

	"github.com/fatih/structs"
	"github.com/ncodes/mapstructure"
)

// Copy copies the value of fields from src to similar
// fields of dest
func Copy(src interface{}, dest interface{}) error {

	if !structs.IsStruct(src) {
		return fmt.Errorf("src is not a struct")
	} else if !structs.IsStruct(dest) {
		return fmt.Errorf("dest is not a struct")
	}

	_src := structs.New(src)
	_srcAsMap := _src.Map()
	return mapstructure.Decode(_srcAsMap, dest)
}

// IsSlice checks if an interface holds a slice.
func IsSlice(s interface{}) bool {
	return reflect.TypeOf(s).Kind() == reflect.Slice
}

// MakeSliceOf creates a slice of values. It returns a slice of interfaces
// of length `size`` with the same underlying type as the value `of``
func MakeSliceOf(of interface{}, size int) []interface{} {
	var s = make([]interface{}, size)
	for i := 0; i < size; i++ {
		s[i] = reflect.New(reflect.TypeOf(of)).Interface()
	}
	return s
}

// CopySlice copies the field values of every struct in src to
// a corresponding struct in dest with matching id, field name and type.
func CopySlice(src interface{}, dest interface{}) error {

	if !IsSlice(src) {
		return fmt.Errorf("src is not a slice")
	}

	if !IsSlice(dest) {
		return fmt.Errorf("dest is not a slice")
	}

	_src := reflect.ValueOf(src)
	_dest := reflect.ValueOf(dest)

	if _src.Len() == 0 {
		return nil
	}

	if _src.Len() != _dest.Len() {
		return fmt.Errorf("src and dest length are not equal")
	}

	for i := 0; i < _src.Len(); i++ {
		srcSrt := _src.Index(i).Interface()

		if !structs.IsStruct(srcSrt) {
			return fmt.Errorf("found a non struct value in src. expects a slice of structs")
		}

		destSrt := _dest.Index(i).Interface()
		if !structs.IsStruct(destSrt) {
			return fmt.Errorf("found a non struct value in dest. expects a slice of structs")
		}

		if err := Copy(srcSrt, destSrt); err != nil {
			return err
		}
	}
	return nil
}
