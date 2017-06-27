package stub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/hoisie/mustache"
)

// TODO: Write test

var (
	// ErrViewNotFound indicates an unknown view file
	ErrViewNotFound = fmt.Errorf("view file not found")
)

// View represents a structure of a view
type View struct {
	Markup string `json:"markup"`
}

// ToJSON encodes the view to json
func (v *View) ToJSON() []byte {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.Encode(v)
	return buffer.Bytes()
}

// Render creates a View from am eml view file and
// a data source. Data source can be a map or a struct type.
// The created View is json encoded.
func Render(name string, dataSource interface{}) ([]byte, error) {
	if !HasView(name) {
		return nil, ErrViewNotFound
	}

	content, err := ioutil.ReadFile(path.Join(ViewDir, name+".html"))
	if err != nil {
		return nil, err
	}

	return (&View{
		Markup: mustache.Render(string(content), dataSource),
	}).ToJSON(), nil
}

// RenderString creates a view from an eml component and a data source.
// Data source can be a map or a struct type.The created View is json encoded.
func RenderString(eml string, dataSource interface{}) (content []byte, err error) {
	return (&View{
		Markup: mustache.Render(string(eml), dataSource),
	}).ToJSON(), nil
}

// HasView checks whether a view exists
func HasView(name string) bool {
	_, err := os.Stat(path.Join(ViewDir, name+".html"))
	return err == nil || os.IsExist(err)
}
