package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	FlagParse()
}

// 命令行参数结构体
type CommandLineArgument struct {
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

// 返回信息
type ResponseJson struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// 命令行参数
var Cla CommandLineArgument

// 解析命令行参数
func FlagParse() {
	flag.UintVar(&Cla.Port, "port", 9999, "端口")
	flag.StringVar(&Cla.Name, "name", "ups", "名称")
	flag.StringVar(&Cla.File, "file", "file", "单文件名称")
	flag.StringVar(&Cla.Files, "files", "files[]", "多文件名称")
	flag.StringVar(&Cla.Upload, "upload", "/var/www/", "保存目录")
	flag.StringVar(&Cla.Prefix, "prefix", "uploads", "保存目录前缀")
	flag.Int64Var(&Cla.Size, "size", 1024*1024*128, "单文件最大限制")
	flag.Int64Var(&Cla.Body, "body", 1024*1024*1280, "HTTP请求体最大限制")
	// daemon
	flag.BoolVar(&Cla.Daemon, "d", false, "以守护进程运行,使用 -d=true or -d")
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
		cmd.Start()
		fmt.Println("[PID]", cmd.Process.Pid)
		os.Exit(0)
	}
}

// JSON序列化
func Json(i interface{}) []byte {
	bytes, _ := json.Marshal(i)
	return bytes
}

// 返回成功信息
func OK(i interface{}) []byte {
	return Json(&ResponseJson{
		Msg:  "upload success",
		Data: i,
	})
}

// 返回错误信息
func Err(err error) []byte {
	return Json(&ResponseJson{
		Code: 2,
		Msg:  err.Error(),
	})
}

// 日期目录
func DateDir() string {
	return fmt.Sprintf("%s%s%s%s%s%s", time.Now().Format("2006"), Cla.Ds, time.Now().Format("01"), Cla.Ds, time.Now().Format("02"), Cla.Ds)
}

// 单文件上传
func Up(writer http.ResponseWriter, request *http.Request) {
	file, fileHeader, err := request.FormFile(Cla.File)
	if err != nil {
		writer.Write(Err(err))
		return
	}
	defer file.Close()
	if fileHeader.Size > Cla.Size {
		writer.Write(Err(errors.New(fmt.Sprintf("single file too large more than %d bytes", Cla.Size))))
		return
	}
	url := ""
	prefixDir := fmt.Sprintf("%s%s%s", Cla.Upload, Cla.Prefix, Cla.Ds)
	clientPrefixDir := request.Header.Get("prefix")
	if clientPrefixDir != "" {
		prefixDir = fmt.Sprintf("%s%s%s", prefixDir, clientPrefixDir, Cla.Ds)
		url = fmt.Sprintf("%s%s%s", url, clientPrefixDir, Cla.Ds)
	}
	dateDir := DateDir()
	dir := fmt.Sprintf("%s%s", prefixDir, dateDir)
	if _, err = os.Stat(dir); err != nil {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			writer.Write(Err(err))
			return
		}
	}
	suffix := path.Ext(fileHeader.Filename)
	saveName := string(Md5([]byte(fmt.Sprintf("%d%d", time.Now().UnixNano(), rand.Intn(10)))))
	saveFile := fmt.Sprintf("%s%s%s", dir, saveName, suffix)
	out, err := os.OpenFile(saveFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		writer.Write(Err(err))
		return
	}
	defer out.Close()
	if _, err = io.Copy(out, file); err != nil {
		writer.Write(Err(err))
		return
	}
	writer.Write(OK(fmt.Sprintf("%s%s%s%s", url, dateDir, saveName, suffix)))
	return
}

// <html>
// 	<head>
// 		<title>upload</title>
// 	</head>
// 	<body>
// 		<form enctype="multipart/form-data" action="http://127.0.0.1:8001/up" method="POST">
// 			<input type="file" name="file">
// 			<input type="hidden" name="token" value="{{.}}" />
// 			<input type="submit" value="upload" />
// 		</form>
// 	</body>
// </html>

// 多文件上传
func Ups(writer http.ResponseWriter, request *http.Request) {
	// 8bit(位)=1Byte(字节)
	// 1024Byte(字节)=1KB
	// 1024KB=1MB
	// 1024MB=1GB
	// 计算机存储单位一般用bit,B,KB,MB,GB,TB,PB,EB,ZB,YB,BB,NB,DB
	// 字节byte:8个二进制位为一个字节(B),最常用的单位
	err := request.ParseMultipartForm(Cla.Body)
	if err != nil {
		writer.Write(Err(err))
		return
	}
	form := request.MultipartForm
	files := form.File[Cla.Files]
	for _, file := range files {
		if file.Size > Cla.Size {
			writer.Write(Err(errors.New(fmt.Sprintf("single file too large more than %d bytes", Cla.Size))))
			return
		}
	}
	url := ""
	prefixDir := fmt.Sprintf("%s%s%s", Cla.Upload, Cla.Prefix, Cla.Ds)
	clientPrefixDir := request.Header.Get("prefix")
	if clientPrefixDir != "" {
		prefixDir = fmt.Sprintf("%s%s%s", prefixDir, clientPrefixDir, Cla.Ds)
		url = fmt.Sprintf("%s%s", clientPrefixDir, Cla.Ds)
	}
	dateDir := DateDir()
	dir := fmt.Sprintf("%s%s", prefixDir, dateDir)
	success := []string{}
	for _, file := range files {
		file.Filename = string(Md5([]byte(fmt.Sprintf("%d%d%d%d", time.Now().UnixNano(), rand.Intn(10), rand.Intn(10), rand.Intn(10))))) + path.Ext(file.Filename)
		if _, err := SaveUploadFile(file, dir); err == nil {
			success = append(success, fmt.Sprintf("%s%s%s", url, dateDir, file.Filename))
		}
	}
	writer.Write(OK(success))
	return
}

// <html>
// 	<head>
// 		<title>uploads</title>
// 	</head>
// 	<body>
// 		<form enctype="multipart/form-data" action="http://127.0.0.1:8001/ups" method="POST">
// 			<input type="file" name="files[]" multiple/>
// 			<input type="hidden" name="token" value="{{.}}" />
// 			<input type="submit" value="upload" />
// 		</form>
// 	</body>
// </html>

// 多文件上传
func SaveUploadFile(fh *multipart.FileHeader, destDirectory string) (int64, error) {
	src, err := fh.Open()
	if err != nil {
		return 0, err
	}
	defer src.Close()
	// dir is not found , create this dir
	if _, err = os.Stat(destDirectory); nil != err {
		err = os.MkdirAll(destDirectory, os.ModePerm)
		if err != nil {
			return 0, err
		}
	}
	out, err := os.OpenFile(filepath.Join(destDirectory, fh.Filename), os.O_WRONLY|os.O_CREATE, os.FileMode(0666))
	if err != nil {
		return 0, err
	}
	defer out.Close()
	return io.Copy(out, src)
}

// MD5加密
func Md5(plainText []byte) []byte {
	hash := md5.New()
	hash.Write(plainText)
	return []byte(hex.EncodeToString(hash.Sum(nil)))
}

func main() {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("file upload"))
	})
	http.HandleFunc("/up", Up)
	http.HandleFunc("/ups", Ups)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", Cla.Port), nil))
}
