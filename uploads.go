package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/xooooooox/arc"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

// Success
func Success(writer http.ResponseWriter,msg string, data ...interface{}){
	bytes, _ := json.Marshal(arc.Success(msg, data...))
	_, _ = writer.Write([]byte(bytes))
}

// Failure
func Failure(writer http.ResponseWriter,msg string, data ...interface{}){
	bytes, _ := json.Marshal(arc.Failure(msg, data...))
	_, _ = writer.Write([]byte(bytes))
}

// Unusual
func Unusual(writer http.ResponseWriter,msg string, data ...interface{}){
	bytes, _ := json.Marshal(arc.Unusual(msg, data...))
	_, _ = writer.Write([]byte(bytes))
}

// Up single file upload
func Up(writer http.ResponseWriter, request *http.Request) {
	file, fileHeader, err := request.FormFile(Cla.File)
	if err != nil {
		Unusual(writer,"file not found")
		return
	}
	defer func(file multipart.File) {
		_ = file.Close()
	}(file)
	if fileHeader.Size > Cla.Size {
		Failure(writer,fmt.Sprintf("single file too large more than %d bytes", Cla.Size))
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
			Failure(writer,err.Error())
			return
		}
	}
	suffix := path.Ext(fileHeader.Filename)
	saveName := string(Md5([]byte(fmt.Sprintf("%d%d", time.Now().UnixNano(), rand.Intn(10)))))
	saveFile := fmt.Sprintf("%s%s%s", dir, saveName, suffix)
	out, err := os.OpenFile(saveFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		Failure(writer,err.Error())
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(out)
	if _, err = io.Copy(out, file); err != nil {
		Failure(writer,err.Error())
		return
	}
	Success(writer,"",fmt.Sprintf("%s%s%s%s", url, dateDir, saveName, suffix))
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

// Ups more files uploads
func Ups(writer http.ResponseWriter, request *http.Request) {
	// 8bit(位)=1Byte(字节)
	// 1024Byte(字节)=1KB
	// 1024KB=1MB
	// 1024MB=1GB
	// 计算机存储单位一般用bit,B,KB,MB,GB,TB,PB,EB,ZB,YB,BB,NB,DB
	// 字节byte:8个二进制位为一个字节(B),最常用的单位
	err := request.ParseMultipartForm(Cla.Body)
	if err != nil {
		Failure(writer,err.Error())
		return
	}
	form := request.MultipartForm
	files := form.File[Cla.Files]
	for _, file := range files {
		if file.Size > Cla.Size {
			Failure(writer,fmt.Sprintf("single file too large more than %d bytes", Cla.Size))
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
	ok := []string{}
	for _, file := range files {
		file.Filename = string(Md5([]byte(fmt.Sprintf("%d%d%d%d", time.Now().UnixNano(), rand.Intn(10), rand.Intn(10), rand.Intn(10))))) + path.Ext(file.Filename)
		if _, err := MoreFileUploads(file, dir); err == nil {
			ok = append(ok, fmt.Sprintf("%s%s%s", url, dateDir, file.Filename))
		}
	}
	Success(writer,"",ok)
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

// MoreFileUploads more files uploads
func MoreFileUploads(fh *multipart.FileHeader, destDirectory string) (int64, error) {
	src, err := fh.Open()
	if err != nil {
		return 0, err
	}
	defer func(file multipart.File) {
		_ = file.Close()
	}(src)
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
	defer func(file *os.File) {
		_ = file.Close()
	}(out)
	return io.Copy(out, src)
}

// DateDir date directory
func DateDir() string {
	return fmt.Sprintf("%s%s%s%s%s%s", time.Now().Format("2006"), Cla.Ds, time.Now().Format("01"), Cla.Ds, time.Now().Format("02"), Cla.Ds)
}

// Md5 encrypt
func Md5(bytes []byte) []byte {
	hash := md5.New()
	hash.Write(bytes)
	return []byte(hex.EncodeToString(hash.Sum(nil)))
}
