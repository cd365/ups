package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CommandLineArguments command line arguments
type CommandLineArguments struct {
	Port   uint
	Name   string
	File   string
	Files  string
	Upload string
	Prefix string
	Size   int64
	Body   int64
	Ds     string
	Daemon bool
}

// Cla command line arguments
var Cla CommandLineArguments

func init() {
	// rand seed
	rand.Seed(time.Now().UnixNano())
	// flag parse
	flag.UintVar(&Cla.Port, "port", 8001, "service port")
	flag.StringVar(&Cla.Name, "name", "ups", "name")
	flag.StringVar(&Cla.File, "file", "file", "single file name in html form")
	flag.StringVar(&Cla.Files, "files", "files[]", "more files name in html form")
	flag.StringVar(&Cla.Upload, "upload", "/var/www/", "save directory")
	flag.StringVar(&Cla.Prefix, "prefix", "uploads", "prefix save directory")
	flag.Int64Var(&Cla.Size, "size", 1024*1024*128, "max size of single file")
	flag.Int64Var(&Cla.Body, "body", 1024*1024*1280, "max size of http request body")
	// daemon run
	flag.BoolVar(&Cla.Daemon, "d", false, "run as a daemon,use -d=true or -d")
	flag.Parse()
	Cla.Ds = string(filepath.Separator)
	if !strings.HasSuffix(Cla.Upload, Cla.Ds) {
		Cla.Upload = Cla.Upload + Cla.Ds
	}
	// daemon
	if Cla.Daemon {
		args := os.Args[1:]
		for i := 0; i < len(args); i++ {
			if args[i] == "-d=true" || args[i] == "-d" {
				args[i] = "-d=false"
				break
			}
		}
		cmd := exec.Command(os.Args[0], args...)
		err := cmd.Start()
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println("[PID]", cmd.Process.Pid)
		os.Exit(0)
	}
}

func main() {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte("file upload"))
	})
	http.HandleFunc("/up", Up)
	http.HandleFunc("/ups", Ups)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", Cla.Port), nil))
}
