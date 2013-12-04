package flog 

import (
	"os"
)

const (
	Dev  string = "development"
	Prod string = "production"
	Test string = "test"
)

// Env is the environment that Martini is executing in. The FLOG_ENV is read on initialization to set this variable.
var Env string = Dev

func init() {
	e := os.Getenv("FLOG_ENV")
	if len(e) > 0 {
		Env = e
	}
}
