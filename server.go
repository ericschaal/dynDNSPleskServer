package main

import (
	"log"
	"net/http"
	"os"
)


var stdlog, errlog *log.Logger
var config Config

func init() {
	stdlog = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "", log.Ldate|log.Ltime)
	config = readConfig()
}

func main() {

	router := NewRouter()

	log.Fatal(http.ListenAndServe(":" + config.ListenPort, router))
}

