# MoDo - Simple Docker Command Executor
[![GoDoc](https://godoc.org/github.com/ncodes/modo?status.svg)](https://godoc.org/github.com/ncodes/modo)

MoDo provides the ability to run a series of commands in a docker container while also allowing flexible behaviours like attaching a simple function to collect logs, stopping the series of commands if one fails or continuing regardless of a command failing and the ability to enable privileged mode, collect full output and exit code per command.

### Installation
```go
go get github.com/ncodes/modo
```

### Example

```go

// Create a MoDo instance and set container id
m := modo.NewMoDo("container_id", true, false, func(d []byte, stdout bool){
    fmt.Printf("%s", d)
})

// Add a command
m.Add(&Do{
    Cmd: []string{"echo", "hello world"}, 
    AbortSeriesOnFail: false,
    Privileged: false,
    OutputCB: nil,
    KeepOutput: false,
})

// Run the commands
// 'errs' hold any error resulting from running each command
// 'err' will hold any general error
errs, err := m.Do()  
```

### Full Documentation
See [Documentation](https://godoc.org/github.com/ncodes/modo)

### ToDo
- Support parallel execution
- More tests