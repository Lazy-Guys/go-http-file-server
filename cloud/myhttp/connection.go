package myhttp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Lazy-Guys/controller"
)

type Worker struct {
	Ct        *controller.Controller
	ipPort    string
	rootdir   string
	tmppath   string
	nameSpace string
	AlgPath   map[string]map[string]string
}

var instance *Worker
var once sync.Once

func GetWorkerInstance(ipPort string, nameSpace string, rootpath string, tmppath string) *Worker {
	once.Do(func() {
		ct, err := controller.NewController(false, nameSpace)
		if err != nil {
			panic(err)
			return
		}
		myWker := &Worker{
			Ct:        ct,
			ipPort:    ipPort,
			nameSpace: nameSpace,
			rootdir:   rootpath,
			tmppath:   tmppath,
			AlgPath:   make(map[string]map[string]string),
		}
		instance = myWker
		http.HandleFunc("/pod", podhandler)
		http.HandleFunc("/upload_file", upload)
		http.HandleFunc("/docker_auth", dockerauth)
	})
	return instance
}

func http_resp(code int, msg string, w http.ResponseWriter) {
	var Result map[string]interface{}
	Result = make(map[string]interface{})

	Result["code"] = code
	Result["msg"] = msg

	data, err := json.Marshal(Result)
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	w.Write([]byte(string(data)))
}

func podhandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err.Error())
	}
	operator := r.PostForm.Get("operator")
	if string(operator) == "openfaas" {
		Name := r.PostForm.Get("name")
		version := r.PostForm.Get("version")
		cmd := r.PostForm.Get("cmd")
		isCreateFromFile := r.PostForm.Get("fromfile")
		instance.Ct.Ofc.FuncOpt = string(cmd)
		instance.Ct.Ofc.Name = string(Name)
		instance.Ct.Ofc.IsCreateFromFile = true
		instance.Ct.Ofc.YmlFileAddress = instance.AlgPath[Name][version]
		fmt.Println(cmd, Name, version, isCreateFromFile)
		if Name == "" || cmd == "" || isCreateFromFile == "" {
			w.Write([]byte("Param Error!"))
			fmt.Println(Name, cmd)
			return
		}
		if string(isCreateFromFile) == "y" {
			err = instance.Ct.ExecOpenFaasCmd()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println("Done!")
			w.Write([]byte("Done!"))
		} else {
			fmt.Println("Warning: wait for development in future.")
			w.Write([]byte("Wait for development in future..."))
		}

	} else if string(operator) == "pod" {
		Name := r.PostForm.Get("name")
		cmd := r.PostForm.Get("cmd")
		fmt.Println(cmd, Name)
		if Name == "" || cmd == "" {
			w.Write([]byte("Param Error!"))
			fmt.Println(Name, cmd)
			return
		}
		if string(cmd) == "apply" {
			yamlName := instance.Ct.GenerateYamlAddress(Name)
			err = instance.Ct.GenerateYamlFile(instance.nameSpace, Name, false)
			if err != nil {
				fmt.Println(err.Error())
				w.Write([]byte("Get yaml file address error!"))
				return
			}
			err = instance.Ct.ApplyPod(instance.nameSpace, yamlName)
			if err != nil {
				fmt.Println(err.Error())
				w.Write([]byte("Apply pods error!" + err.Error()))
				return
			}
		} else if string(cmd) == "delete" {
			err = instance.Ct.DeletePod(instance.nameSpace, Name)
			if err != nil {
				w.Write([]byte("Delete pods error!" + err.Error()))
				fmt.Println(err.Error())
				return
			}
		}
		fmt.Println("Done!")
		w.Write([]byte("Done!"))
	} else {
		fmt.Println("Error: cannot parse the form!")
		w.Write([]byte("Param Error!"))
	}

}

func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Write([]byte("Error: cannot access by other method."))
		return
	}

	contentType := r.Header.Get("content-type")
	contentLen := r.ContentLength

	if !strings.Contains(contentType, "multipart/form-data") {
		http_resp(-1001, "The content-type must be multipart/form-data", w)
		return
	}
	// Max file size 50MB
	if contentLen >= 50*1024*1024 {
		http_resp(-1002, "Failed to large,limit 50MB", w)
		return
	}

	err := r.ParseMultipartForm(50 * 1024 * 1024)
	if err != nil {
		http_resp(-1003, "Failed to ParseMultipartForm", w)
		return
	}

	if len(r.MultipartForm.File) == 0 {
		http_resp(-1004, "File is NULL", w)
		return
	}

	var Result map[string]interface{}
	Result = make(map[string]interface{})

	Result["code"] = 0

	algname := r.PostForm.Get("name")
	version := r.PostForm.Get("version")
	algpath := instance.rootdir + "/" + algname + "/" + version + "/"
	err = instance.Ct.CreatePath(algpath)
	if err != nil {
		fmt.Println("Error: cannot create ALG directory!")
		return
	}
	FileCount := 0
	for _, headers := range r.MultipartForm.File {
		for _, header := range headers {
			fmt.Printf("recv file: %s\n", header.Filename)
			filePath := filepath.Join(algpath, header.Filename)
			dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0777)
			if err != nil {
				fmt.Printf("Create file %s error: %s\n", filePath, err)
				return
			}
			srcFile, err := header.Open()
			if err != nil {
				fmt.Printf("Open header failed: %s\n", err)
			}
			_, err = io.Copy(dstFile, srcFile)
			if err != nil {
				fmt.Printf("Write file %s error: %s\n", filePath, err)
			}
			_, _ = srcFile.Close(), dstFile.Close()
			FileCount++
			name := fmt.Sprintf("file%d_url", FileCount)
			Result[name] = (header.Filename)
		}
	}
	err = instance.Ct.CopyTmpFile(instance.tmppath, algpath)
	if err != nil {
		fmt.Println("Error: cannot copy template file: ", err.Error())
		return
	}
	instance.AlgPath[algname] = map[string]string{
		version: algpath,
	}
	fmt.Println(instance.AlgPath[algname][version])
	err = instance.Ct.BuildImage(instance.AlgPath[algname][version], algname+":"+version)
	if err != nil {
		fmt.Printf("Error: cannot build image %s: ", algname+":"+version)
		fmt.Println(err.Error())
		return
	}
	err = instance.Ct.PushImage(algname, version)
	if err != nil {
		w.Write([]byte("Error: push images failed:" + err.Error()))
		fmt.Println("Error: push images failed: ", err.Error())
		return
	}
	w.Write([]byte("Done!"))
}

func dockerauth(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err.Error())
	}
	username := r.PostForm.Get("username")
	passwd := r.PostForm.Get("passwd")
	if username == "" || passwd == "" {
		fmt.Println("Error: username or passwd is empty!")
		w.Write([]byte("Error: username or passwd is empty!"))
		return
	}
	instance.Ct.Docker.Auth.Username = username
	instance.Ct.Docker.Auth.Password = passwd
	fmt.Println(username)
	fmt.Println(passwd)
	w.Write([]byte("Done!"))
}

func (wk *Worker) RunServer() {
	http.ListenAndServe(wk.ipPort, nil)
}
