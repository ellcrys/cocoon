// Code generated by go-bindata.
// sources:
// bindata.go
// cocoon.job.json
// DO NOT EDIT!

package data

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _bindataGo = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x01\x00\x00\xff\xff\x00\x00\x00\x00\x00\x00\x00\x00")

func bindataGoBytes() ([]byte, error) {
	return bindataRead(
		_bindataGo,
		"bindata.go",
	)
}

func bindataGo() (*asset, error) {
	bytes, err := bindataGoBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "bindata.go", size: 0, mode: os.FileMode(420), modTime: time.Unix(1487179671, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cocoonJobJson = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x8c\x55\x5d\x8f\xea\x36\x10\x7d\xdf\x5f\x61\x59\xf7\xb1\x0b\x6c\xab\xab\x4a\x48\x7d\x48\x09\x5d\x51\x01\x89\xf8\x78\xa8\xaa\x15\x9a\x35\xb3\x59\x0b\xc7\x8e\xec\x81\xbb\x34\xca\x7f\xaf\x9c\x84\x6e\x48\x02\xdb\x7d\x59\x33\xe7\x8c\xe7\x9c\x89\x3d\xce\x1f\x18\xe3\x7f\x9a\x57\x3e\x66\x7e\xc9\x18\x5f\x61\x22\x8d\xe6\x63\xc6\x13\x65\x5e\x41\xf1\x9f\xaa\xf8\x2c\xf4\x31\x61\x84\x31\xfa\x31\xcf\x67\x61\x51\x5c\xa0\x25\xa4\x78\x13\xdc\x9c\xb3\x12\x74\x68\x4f\x52\xe0\x25\x1c\x5b\x69\xac\xa4\x33\x1f\xb3\xef\xa3\x3a\x16\x28\x15\x50\xa4\x85\xe7\xbf\x81\x72\x58\xc7\x43\x20\x10\xa8\x09\xad\xe3\x63\xf6\x77\x19\x64\x8c\xef\xc5\x13\x2f\xd7\x2f\x35\x6f\x62\xb4\x23\x0b\x52\x53\x93\x97\xd7\xff\x19\xe3\xf3\x0d\xd8\x04\xc9\xcb\xf9\x96\x03\x91\x1d\x1c\xd0\x6a\x54\x03\x0d\x29\x5e\x14\x57\x5d\xf8\x64\x2a\xa9\x8f\x1f\x4d\x2c\xca\xd0\x82\xde\x7b\xec\x37\x5e\x87\x8b\x2b\x21\x1b\x70\x87\x67\x6b\x8e\xd9\x0d\x1d\xf7\x1a\x56\x3b\x39\x6a\x5f\x3c\xcf\xcb\x55\x51\x5c\x61\x4d\x97\xfa\xa8\x54\x03\xf4\x95\x9b\x45\xaf\x0b\xf7\x15\x27\x70\x87\xae\x82\xaa\xef\x56\x9e\xd0\x7a\xee\xde\x88\x03\xda\x36\x3e\x31\xfa\x4d\x26\xff\x1d\x9d\x06\x22\x53\x48\xca\x22\x79\x3e\xf3\xcb\xce\xde\xcc\xd7\x4f\xd3\xba\x8d\xaf\xe0\xde\xbb\x04\xb0\x49\x69\x85\x7f\xcb\x97\xd1\x22\x08\x77\x8b\xe9\x26\xd8\xad\x27\xab\x59\xbc\x59\xef\xc2\xd9\xaa\x18\x5e\x21\xe1\x34\x9e\x47\x7f\xd5\x84\xdd\x32\x58\x4c\x0b\xfe\x72\xb5\x69\xd1\x32\x30\xd5\xa7\x3e\xf5\x93\x68\x12\x45\xcb\xdd\xed\x13\xdf\xe5\x4e\xa2\x70\xba\xdb\xae\xe6\x95\xe7\x49\x99\x33\x31\x7b\xdc\xae\xe6\x5f\x65\x6d\x82\xe7\x76\xd6\x06\x92\xaf\xb2\xe6\xc1\xb2\x93\x36\x07\xed\xf3\xee\x5a\x5e\x57\xd7\xb0\x6c\xec\x4b\x0b\x5b\x20\x41\x5f\x3f\xba\x8d\x2d\xcf\x04\x9e\x1e\xf7\x98\x29\x73\x7e\x14\x46\x6b\x14\x64\xec\xe3\x68\x30\x1a\xfc\x3a\xe8\xfb\x9a\x8d\xef\xe6\xb3\x87\xca\x08\x50\x43\x27\xac\xcc\xc8\xdd\xd7\x3c\x37\xc9\xed\xa3\xb6\x80\x8f\x3f\xa4\x2a\x0d\x3d\x8d\x3a\x55\x6b\x74\x2d\xff\xc1\xc5\xef\x25\xe5\x6e\xa5\x0d\xa6\x99\x02\xea\x6f\x4f\x60\x49\xbe\x81\xa8\x86\x4b\x5b\x07\x63\xfc\x19\x89\xd0\xae\xcd\xd1\x96\x13\x8c\xbf\x13\x65\x6e\x3c\x1c\x5a\xf8\x31\x48\x24\xbd\x1f\x5f\x8f\x0e\xad\x30\x9a\x50\xd3\x40\x98\x74\xa8\x85\xd9\xa3\x1b\x56\x47\x6c\x98\x82\x23\xb4\x97\x96\x7c\x7d\xb6\xdb\x66\xcb\xe1\xad\x80\xe4\x09\x43\x74\xf4\x65\x97\xdb\xfe\x56\xe8\x4a\xed\xae\xf7\x4e\xc4\xdb\x6a\x1e\xc5\xdb\xa2\xe8\xb6\x19\x53\x63\xcf\x65\x87\xf3\xfc\xf2\xa3\x87\x37\x8b\xe2\x35\x1f\xb3\xee\x77\x5a\x22\xfd\x30\xb6\x1a\x5d\xf7\xef\x6c\x28\x5d\x06\x24\xde\x63\x38\x2b\x03\x7e\x7c\xe4\x45\x83\xf1\xb9\x7e\x69\x8e\xf3\x1b\xd6\x6e\xdb\xfa\x3f\x96\xbc\x96\x43\xcd\xa8\x96\x2d\xbc\xcf\xee\x0d\xab\xc5\xb5\x5a\x02\x4b\xb1\x51\x52\x9c\xdb\x8a\x67\xfe\x1d\x3c\x81\xe2\x63\xf6\xcb\xe8\xf3\xef\xaa\x44\x40\x84\x69\x46\xdd\x3b\xc1\x43\x54\xe0\xb7\xfc\xf9\x7b\x7f\xe6\xc2\xec\xb1\xba\xdc\x9e\xd7\xab\xee\x32\x25\x8a\xde\xa7\x6f\x9b\xed\x81\xb0\x21\x9a\xaf\x09\x92\xa4\x7c\x46\x9e\xba\x72\xfd\x05\x8d\xc1\x82\x52\xe8\x1d\x3d\x3d\x5c\x76\x2c\x1e\x8a\x7f\x03\x00\x00\xff\xff\x25\xaa\x28\x6e\x9c\x08\x00\x00")

func cocoonJobJsonBytes() ([]byte, error) {
	return bindataRead(
		_cocoonJobJson,
		"cocoon.job.json",
	)
}

func cocoonJobJson() (*asset, error) {
	bytes, err := cocoonJobJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cocoon.job.json", size: 2204, mode: os.FileMode(420), modTime: time.Unix(1487179658, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"bindata.go": bindataGo,
	"cocoon.job.json": cocoonJobJson,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}
var _bintree = &bintree{nil, map[string]*bintree{
	"bindata.go": &bintree{bindataGo, map[string]*bintree{}},
	"cocoon.job.json": &bintree{cocoonJobJson, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

