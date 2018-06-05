package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/femaref/log"
)

var traceFile io.Writer

func init() {
	var err error
	traceFile, err = os.Create("trace")
	if err != nil {
		panic(err)
	}
}

func tln(args ...interface{}) (int, error) {
	return fmt.Fprintln(traceFile, args)
}

func tf(f string, args ...interface{}) (int, error) {
	return fmt.Fprintf(traceFile, f, args)
}

func main() {
	log.Setup(log.Config{
		Base: log.BaseConfig{
			Name:   "foo",
			File:   true,
			Stdout: true,
		},
	})
	log.RedirectStderrToFile("stderr", true)

	go func() {

		for {
			log.Logger.Error("foo")
		}
	}()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	<-c
}
