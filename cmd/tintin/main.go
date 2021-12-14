package main

import (
	"log"
	"os"

	"github.com/datatok/tintin/pkg/utils/cli"
)

var (
	settings = cli.New()
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	cmd := newRootCmd(os.Stdout, os.Args[1:])

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
