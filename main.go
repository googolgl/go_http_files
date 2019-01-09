package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
)

type respData struct {
	Mime     string `json:"mime,omitempty"`
	URL      string `json:"url,omitempty"`
	PathFile string `json:"pathfile,omitempty"`
	NameFile string `json:"namefile,omitempty"`
	Size     int64  `json:"size,omitempty"`
	Status   string `json:"status"`
}

var (
	maxUploadSize = flag.Int64("maxsize", 9437184, "max allowed size uploaded files")
	uploadPath    = flag.String("path", "/files/", "root upload path")
	host          = flag.String("host", "127.0.0.1:8081", "server address")
	domain        = flag.String("domain", "", "domain address (with http://)")
)

func main() {
	flag.Parse()

	http.Handle("/files", http.StripPrefix("/files", http.FileServer(http.Dir("."+*uploadPath))))
	http.HandleFunc("/upload", uploadFileHandler())
	log.Print("Server started on " + *host + " ...")
	log.Fatal(http.ListenAndServe(*host, nil))
}

func uploadFileHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rdata respData

		// validate max Reguest size
		r.Body = http.MaxBytesReader(w, r.Body, *maxUploadSize)
		err := r.ParseMultipartForm(*maxUploadSize)
		if err != nil {
			response(w, 413, "File too big", rdata)
			return
		}

		if r.Method != http.MethodPost || len(r.FormValue("pathFile")) == 0 {
			response(w, http.StatusCreated, "Error post data", rdata)
			return
		}
		rdata.PathFile = *uploadPath + r.FormValue("pathFile") + "/"

		rdata.Size, err = strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64)
		if err != nil {
			response(w, http.StatusCreated, "Strange client", rdata)
			return
		}

		file, handle, err := r.FormFile("uploadFile")
		if err != nil {
			response(w, http.StatusCreated, "Invalid file", rdata)
			return
		}
		defer file.Close()
		rdata.NameFile = handle.Filename

		rdata.Mime = handle.Header.Get("Content-Type")
		switch rdata.Mime {
		case "image/jpeg", "image/jpg", "image/gif", "image/png":
			saveFile(w, file, rdata)
		case "application/pdf", "application/msword":
			saveFile(w, file, rdata)
		default:
			log.Println("Err! Invalid file type:", rdata.Mime)
			response(w, http.StatusCreated, "Invalid file type", rdata)
			return
		}
	})
}

func saveFile(w http.ResponseWriter, file multipart.File, rdata respData) {
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println("Err!", err)
		return
	}

	// create directory if not exist
	_, err = ioutil.ReadDir("." + rdata.PathFile)
	if err != nil {
		pathErr := os.MkdirAll("."+rdata.PathFile, 0777)
		if pathErr != nil {
			log.Println("Err!", err)
			return
		}
	}

	// write file
	err = ioutil.WriteFile("."+rdata.PathFile+rdata.NameFile, data, 0666)
	if err != nil {
		log.Println("Err!", err)
		return
	}

	if len(*domain) > 0 {
		rdata.URL = *domain + rdata.PathFile + rdata.NameFile
	} else {
		rdata.URL = "http://" + *host + rdata.PathFile + rdata.NameFile
	}
	response(w, http.StatusCreated, "Success", rdata)
}

func response(w http.ResponseWriter, code int, message string, rd respData) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	rd.Status = message
	jsonData, err := json.Marshal(rd)
	if err != nil {
		log.Println("Err!", err)
		return
	}
	w.WriteHeader(code)
	w.Write([]byte(jsonData))
}
