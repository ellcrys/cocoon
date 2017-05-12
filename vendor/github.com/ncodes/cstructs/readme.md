### CStruct 

CStruct copies a source struct to another struct that share similar fields
and field types.  

#### Installation
```
go get github.com/ncodes/cstructs
```

#### Example

```go
package main

import "github.com/ncodes/cstructs"

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

func main(){ 
    
    ben := A{
		Name:  "benedith",
		Age:   20,
		About: []byte("Ben is a great sports person"),
	}

    var pete B
    cstructs.Copy(&ben, &pete)
}
```

#### CopySlice(src interface{}, dest interface{})

Copies a slice of struct to another slice of struct. Both slice must have the same length.If src is empty, `nil` is returned.

#### MakeSliceOf(of interface{}, size int)

A convenience method for creating a slice of an initialized type. Useful when there is need to create a slice of struct to use as destination when calling `CopySlice`.