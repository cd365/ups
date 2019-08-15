package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

// init program initialization
func init() {
	rand.Seed(time.Now().UnixNano())
}

// Save
type Save struct {
	// html form file name 上传文件表单名
	HTMLFormFileName string
	// save upload directory 上传文件保存的磁盘目录
	SaveDir string
	// net directory 上传文件的网络访问地址[url.pathinfo] 如: 2019/08/15/405c3a237423c46a6fe47ec7ddc886c1.png
	NetDir string
	// verify 是否校验文件后后缀格式
	Verify bool
}

// Saves
type Saves struct {
	// html form file name 上传文件表单名
	HTMLFormFileName string
	// save upload directory 上传文件保存的磁盘目录
	SaveDir string
	// net directory 上传文件的网络访问地址[url.pathinfo] 如: ["2019/08/15/405c3a237423c46a6fe47ec7ddc886c1.png","2019/08/15/20a5a256c429aa8ae12ab7e4780f56d0.png","2019/08/15/ff001b1917c9457507b2ac070e0b26cc.png"]
	NetDir []string
	// verify 是否校验文件后后缀格式
	Verify bool
}

// Up processing request of single file upload
func Up(writer http.ResponseWriter, request *http.Request) {
	save := NewSave(&Save{Verify: false})
	file, fileHeader, err := request.FormFile(save.HTMLFormFileName)
	if err != nil {
		writer.Write(UploadError(err))
		return
	}
	defer file.Close()
	var maxSize int64 = 1024 * 1024 * 128
	if fileHeader.Size > maxSize {
		writer.Write(UploadError(errors.New(fmt.Sprintf("single file too large more than %d bytes", maxSize))))
		return
	}
	// 是否校验文件类型
	if save.Verify {
		fileType := request.Header.Get("type")
		if fileType == "" {
			writer.Write(UploadError(errors.New("missing parameters 'type' in the http request header")))
			return
		}
		err = UploadFileFormatVerify(fileType, path.Ext(fileHeader.Filename))
		if err != nil {
			writer.Write(UploadError(err))
			return
		}
	}
	dateDir := DateDir()
	dir := fmt.Sprintf("%s%s", save.SaveDir, dateDir)
	_, err = os.Stat(dir)
	if err != nil {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			writer.Write(UploadError(err))
			return
		}
	}
	suffix := path.Ext(fileHeader.Filename)
	saveName := string(Md5([]byte(fmt.Sprintf("%d%d", time.Now().UnixNano(), rand.Intn(10)))))
	saveFile := fmt.Sprintf("%s%s%s", dir, saveName, suffix)
	out, err := os.OpenFile(saveFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		writer.Write(UploadError(err))
		return
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		writer.Write(UploadError(err))
		return
	}
	save.NetDir = fmt.Sprintf("%s%s%s", dateDir, saveName, suffix)
	writer.Write([]byte(fmt.Sprintf(`{"code":0,"msg":"success","data":%s}`, save.NetDir)))
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

// Ups processing request of more files upload
func Ups(writer http.ResponseWriter, request *http.Request) {
	saves := NewSaves(&Saves{NetDir: []string{}, Verify: false})
	// 8bit(位)=1Byte(字节)
	// 1024Byte(字节)=1KB
	// 1024KB=1MB
	// 1024MB=1GB
	// 计算机存储单位一般用bit,B,KB,MB,GB,TB,PB,EB,ZB,YB,BB,NB,DB
	// 字节byte:8个二进制位为一个字节(B),最常用的单位
	err := request.ParseMultipartForm(1024 * 1024 * 256)
	if err != nil {
		// HTTP请求大小超过256M
		writer.Write(UploadError(err))
		return
	}
	form := request.MultipartForm
	files := form.File[saves.HTMLFormFileName]
	var singleFileMaxSize int64 = 1024 * 1024 * 128
	for _, file := range files {
		if file.Size > singleFileMaxSize {
			writer.Write(UploadError(errors.New(fmt.Sprintf("single file too large more than %d bytes", singleFileMaxSize))))
			return
		}
	}
	// 是否校验文件类型
	if saves.Verify {
		fileType := request.Header.Get("type")
		if fileType == "" {
			writer.Write(UploadError(errors.New("missing parameters 'type' in the http request header")))
			return
		}
		for _, file := range files {
			err = UploadFileFormatVerify(fileType, path.Ext(file.Filename))
			if err != nil {
				writer.Write(UploadError(err))
				return
			} else {
				continue
			}
		}
	}
	dateDir := DateDir()
	for _, file := range files {
		file.Filename = string(Md5([]byte(fmt.Sprintf("%d%d", time.Now().UnixNano(), rand.Intn(10))))) + path.Ext(file.Filename)
		_, err := SaveUploadFile(file, fmt.Sprintf("%s%s", saves.SaveDir, dateDir))
		if err == nil {
			saves.NetDir = append(saves.NetDir, fmt.Sprintf("%s%s", dateDir, file.Filename))
		}
	}
	bytes, _ := json.Marshal(saves.NetDir)
	writer.Write([]byte(fmt.Sprintf(`{"code":0,"msg":"success","data":%s}`, string(bytes))))
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

// SaveUploadFile save upload file
func SaveUploadFile(fh *multipart.FileHeader, destDirectory string) (int64, error) {
	src, err := fh.Open()
	if err != nil {
		return 0, err
	}
	defer src.Close()
	// dir is not found , create this dir
	_, err = os.Stat(destDirectory)
	if nil != err {
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

// NewSave new save
func NewSave(s *Save) *Save {
	save := &Save{}
	// html form file name
	if s.HTMLFormFileName == "" {
		save.HTMLFormFileName = "file"
	} else {
		save.HTMLFormFileName = s.HTMLFormFileName
	}
	// upload file save directory
	if s.SaveDir == "" {
		save.SaveDir = DefaultSaveDir()
	} else {
		save.SaveDir = s.SaveDir
	}
	save.Verify = s.Verify
	return save
}

// NewSaves new saves
func NewSaves(s *Saves) *Saves {
	saves := &Saves{NetDir: []string{}}
	// html form file name
	if s.HTMLFormFileName == "" {
		saves.HTMLFormFileName = "files[]"
	} else {
		saves.HTMLFormFileName = s.HTMLFormFileName
	}
	// upload file save directory
	if s.SaveDir == "" {
		saves.SaveDir = DefaultSaveDir()
	} else {
		saves.SaveDir = s.SaveDir
	}
	saves.Verify = s.Verify
	return saves
}

// DefaultSaveDir default save dir
func DefaultSaveDir() string {
	return "/tmp/upload/"
}

// Date Dir date directory
func DateDir() string {
	return fmt.Sprintf("%s/%s/%s/", time.Now().Format("2006"), time.Now().Format("01"), time.Now().Format("02"))
}

// Md5 md5
func Md5(plainText []byte) []byte {
	hash := md5.New()
	hash.Write(plainText)
	return []byte(hex.EncodeToString(hash.Sum(nil)))
}

// UploadFileFormatVerify upload file format verify
func UploadFileFormatVerify(uploadFileType string, uploadFileSuffix string) error {
	unknownFileFormat := errors.New(`unknown file format`)
	if uploadFileSuffix == "" {
		return unknownFileFormat
	}
	switch uploadFileType {
	case "image":
		return UploadFileFormatVerifyJudge([]string{".png", ".jpg", ".jpeg", ".gif", ".psd", ".swf", ".bmp", ".emf"}, uploadFileSuffix)
	case "icon":
		return UploadFileFormatVerifyJudge([]string{".icon"}, uploadFileSuffix)
	case "audio":
		return UploadFileFormatVerifyJudge([]string{".mp3", ".midi", ".wma", ".wave", ".emf"}, uploadFileSuffix)
	case "video":
		return UploadFileFormatVerifyJudge([]string{".mp4", ".avi", ".wmv", ".mpeg", ".dv", ".rm", ".rmvb", ".mov"}, uploadFileSuffix)
	case "office":
		return UploadFileFormatVerifyJudge([]string{".doc", ".docx", ".xls", ".ppt", ".pdf"}, uploadFileSuffix)
	case "text":
		return UploadFileFormatVerifyJudge([]string{".txt"}, uploadFileSuffix)
	case "binary":
		return UploadFileFormatVerifyJudge([]string{".msi", ".exe", ".out", ".so"}, uploadFileSuffix)
	case "archive":
		return UploadFileFormatVerifyJudge([]string{".zip", ".rar"}, uploadFileSuffix)
	case "other":
		return UploadFileFormatVerifyJudge([]string{".html", ".chm"}, uploadFileSuffix)
	}
	return errors.New(`unknown file format`)
}

// UploadFileFormatVerifyJudge upload file format verify judge
func UploadFileFormatVerifyJudge(lists []string, uploadFileSuffix string) error {
	unknownFileFormat := errors.New(`unknown file format`)
	length := len(lists)
	for i := 0; i < length; i++ {
		if i == length-1 {
			if uploadFileSuffix != lists[i] {
				return unknownFileFormat
			}
		}
		if uploadFileSuffix != lists[i] {
			continue
		} else {
			break
		}
	}
	return nil
}

// UploadError upload error
func UploadError(err error) []byte {
	return []byte(fmt.Sprintf(`{"code":-1,"msg":"%s"}`, err.Error()))
}

// main program entry
func main() {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("ups"))
	})
	http.HandleFunc("/up", Up)
	http.HandleFunc("/ups", Ups)
	log.Fatal(http.ListenAndServe(":8001", nil))
}
