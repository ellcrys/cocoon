package main

import "github.com/ellcrys/util"
import "time"

func main() {
	s := util.Spinner("loading")
	time.Sleep(5 * time.Second)
	s()
	time.Sleep(5 * time.Second)
}
