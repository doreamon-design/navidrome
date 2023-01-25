package main

import (
	"runtime"

	"github.com/doreamon-design/navidrome/cmd"
)

func main() {
	runtime.MemProfileRate = 0
	cmd.Execute()
}
