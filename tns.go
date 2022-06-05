package main

import (
	"bytes"
	"crypto/rand"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	//"path"
	"strings"
	"time"

	"flag"
	"fmt"
	"math/big"
	"path"
)

const (
	download_path string = "./download/"
)

var Port = flag.String("p", "9999", "服务端口号")
var del = flag.Int("d", 1, "是否删除上传的文件,默认 1 表示删除,非 1 表示删除")

func RandomString(len int) string {
	var container string
	var str = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	b := bytes.NewBufferString(str)
	length := b.Len()
	bigInt := big.NewInt(int64(length))
	for i := 0; i < len; i++ {
		randomInt, _ := rand.Int(rand.Reader, bigInt)
		container += string(str[randomInt.Int64()])
	}
	return container
}

func main() {

	flag.Parse()

	if !checkFileIsExist("download") {
		os.Mkdir("download", 0755)
	}

	if !checkFileIsExist("download/1M") {
		createFileBySize("download/1M", 1024*1024)
	}
	if !checkFileIsExist("download/10M") {
		createFileBySize("download/10M", 1024*1024*10)
	}
	if !checkFileIsExist("download/100M") {
		createFileBySize("download/100M", 1024*1024*100)
	}
	if !checkFileIsExist("download/1G") {
		createFileBySize("download/1G", 1024*1024*1024)
	}
	if !checkFileIsExist("download/10G") {
		createFileBySize("download/10G", 1024*1024*1024*10)
	}

	http.HandleFunc("/", indexHandle)
	http.HandleFunc("/uploads", uploadHandle)
	http.HandleFunc("/download/", downloadHandle)
	//http.HandleFunc("/target/", downloadHandle)

	fmt.Println("服务器服务于端口:", *Port)
	err := http.ListenAndServe(":"+*Port, nil)
	if err != nil {
		fmt.Println("服务器启动失败")
		return
	}
}
func indexHandle(w http.ResponseWriter, r *http.Request) {
	files, _ := Listdir("./download")

	htmlString := `<html>


  <head>
    <title>http网速测试</title></head>

  <body>
  	<br/>
  	<br/>
	<br/>
	<div class="file-box">
		<form id="uploadForm" action='/uploads' method="post" enctype="multipart/form-data">

			<input type="file" name="file" class="file" id="fileField" >
			<br/>
			<br/>
			<input type="submit" class="btn" value="上传文件" />
		</form>
	</div>
	<br/>
	<br/> 
	<br/>
	点击文件名下载文件 </br></br>`

	for i := 0; i < len(files); i++ {

		htmlString = htmlString + `<a href="./download/` + files[i] + `" download>` + files[i] + `</a> </br> `
	}

	htmlString = htmlString + `
  </body>


</html>

`
	io.WriteString(w, htmlString)

	fmt.Println("访问时间:", time.Now().Format("2006-01-02 15:04:05"), "访问IP:", r.RemoteAddr)
}
func uploadHandle(w http.ResponseWriter, r *http.Request) {

	htmlTemp := `<html> <body>`
	htmlTemp = htmlTemp + "开始上传文件时间：" + time.Now().Format("2006-01-02 15:04:05") + "</br>"
	fmt.Println("***************************************************************************************")
	fmt.Println("开始上传文件时间：" + time.Now().Format("2006-01-02 15:04:05"))

	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println(err)
			break
		}

		fmt.Printf("FileName=[%s], FormName=[%s]\n", part.FileName(), part.FormName())
		if part.FileName() == "" { // this is FormData
			data, _ := ioutil.ReadAll(part)
			fmt.Printf("FormData=[%s]\n", string(data)) //其中一次接收上传文件执行的输出：FileName=[notepad.exe], FormName=[file]

		} else { // This is FileData
			dst, _ := os.Create("download/" + part.FileName() + RandomString(12) + path.Ext(part.FileName()))
			defer dst.Close()

			//实际的文件传送过程消耗在这里
			fmt.Println("开始io.Copy时间：" + time.Now().Format("2006-01-02 15:04:05"))
			io.Copy(dst, part)
			fmt.Println("io.Copy完成时间：" + time.Now().Format("2006-01-02 15:04:05"))

			//根据用户输入的参数决定是否删除上传的文件
			if *del == 1 {
				dst.Close()
				err := os.Remove("download/" + part.FileName())
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println("完成删除文件时间：" + time.Now().Format("2006-01-02 15:04:05"))
			}

		}
	}
	fmt.Println("完成接收文件时间：" + time.Now().Format("2006-01-02 15:04:05"))
	htmlTemp = htmlTemp + "完成接收文件时间" + time.Now().Format("2006-01-02 15:04:05") + "</br>"
	htmlTemp = htmlTemp + `<a href="./">` + "返回首页" + `</a> </br> </body> </html>`

	fmt.Println("--------------------------------------------------------------------------\n")

	w.Write([]byte(htmlTemp))

}

func showErrorToClient(w http.ResponseWriter, strError string) {
	httpStr := `
<html>
  
  <head>
    <title>下载后文件</title></head>

  <body>

<h3>` + strError + `</h3>
<br/>
<a href="./" >返回重新上传数据文件</a>

  </body>

</html>	
	`
	w.Write([]byte(httpStr))
}

func downloadHandle(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println("==========================================")

	urlSourceUrl, _ := url.QueryUnescape(r.RequestURI)

	baseFileName := urlSourceUrl[strings.LastIndex(urlSourceUrl, "/")+1:]

	downloadFileName := download_path + baseFileName

	fmt.Println("开始下载文件时间", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("文件名", downloadFileName, "远端IP:", r.RemoteAddr)

	f, _ := os.Open(downloadFileName)
	io.Copy(w, f)
	fmt.Println("完成下载文件时间", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("文件名", downloadFileName, "远端IP:", r.RemoteAddr)
	//下载完成后，删除后文件。
	time.Sleep(time.Duration(1) * time.Second)

	fmt.Println("==========================================")
}

func Listdir(dirPth string) (files []string, err error) {
	files = make([]string, 0, 10)

	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}

	files = make([]string, 0)
	for _, fi := range dir {
		if !fi.IsDir() { // 忽略目录
			files = append(files, fi.Name())
		}
	}
	return files, nil
}

func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

//创建指定大小文件
func createFileBySize(fileName string, size int64) bool {
	f, err := os.Create(fileName)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer f.Close()

	if err := f.Truncate(size); err != nil {
		fmt.Println(err)
		return false
	}
	return true
}
