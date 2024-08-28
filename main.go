package main

import (
	"github.com/nv-gha-runners/gh-nv-gha-aws/cmd"
)

var (
	version = "undefined"
)

func main() {
	cmd.Execute(version)
}
