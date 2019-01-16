package main

import (
	"os"

	"github.com/txgruppi/run/cli"
)

func main() {
	args := os.Args
	if len(args) == 1 {
		args = append(args, "--help")
	}
	cli.NewApp().Run(args)
}
