package main

import "fmt"

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func toolVersion() string {
	return fmt.Sprintf("%s (%s, %s)", version, commit, date)
}
