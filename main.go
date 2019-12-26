package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/xooooooox/sea"
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
	Port       uint
	Name       string
	File       string
	Files      string
	Upload     string
	Prefix     string
	Size       int64
	Body       int64
	Ds         string
	Daemon     bool
	DataSource string
}

// Cla command line arguments
var Cla CommandLineArguments

func init() {
	// rand seed
	rand.Seed(time.Now().UnixNano())
	// flag parse
	flag.UintVar(&Cla.Port, "port", 8001, "端口")
	flag.StringVar(&Cla.Name, "name", "ups", "名称")
	flag.StringVar(&Cla.File, "file", "file", "单文件名称")
	flag.StringVar(&Cla.Files, "files", "files[]", "多文件名称")
	flag.StringVar(&Cla.Upload, "upload", "/var/www/", "保存目录")
	flag.StringVar(&Cla.Prefix, "prefix", "uploads", "保存目录前缀")
	flag.Int64Var(&Cla.Size, "size", 1024*1024*128, "单文件最大限制")
	flag.Int64Var(&Cla.Body, "body", 1024*1024*1280, "HTTP请求体最大限制")
	// daemon run
	flag.BoolVar(&Cla.Daemon, "d", false, "以守护进程运行,使用 -d=true or -d")
	// database data source
	flag.StringVar(&Cla.DataSource, "db", "root:root@tcp(127.0.0.1:3306)/xooooooox?charset=utf8mb4", "数据库连接源")
	flag.Parse()
	db, err := sql.Open("mysql", Cla.DataSource)
	if err != nil {
		log.Fatalln(err)
	}
	db.SetMaxIdleConns(3)
	db.SetMaxOpenConns(5)
	err = sea.Instance(db)
	if err != nil {
		log.Fatalln(err)
	}
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
