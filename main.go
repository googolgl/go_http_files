package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
)

type File struct {
	Mime     string `json:"mime,omitempty"`
	URL      string `json:"url,omitempty"`
	PathFile string `json:"pathfile,omitempty"`
	NameFile string `json:"namefile,omitempty"`
	Size     int64  `json:"size,omitempty"`
	Status   string `json:"status"`
}

const maxRequestSize = 60 * 1024 * 1024 // 60 mb
const maxUploadSize = 1 * 1024 * 1024   // 9 mb
const uploadPath = "/files/"
const host = "localhost:8081"

func main() {
	http.Handle(uploadPath, http.StripPrefix(uploadPath, http.FileServer(http.Dir("."+uploadPath))))

	http.HandleFunc("/upload", uploadFileHandler())
	log.Print("Server started on " + host + "...")
	log.Fatal(http.ListenAndServe(host, nil))
}

func uploadFileHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var origFile File
		if r.Method != http.MethodPost || len(r.FormValue("pathFile")) == 0 {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		origFile.PathFile = uploadPath + r.FormValue("pathFile") + "/"

		// validate max Reguest size
		r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)

		fileSize, err := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64)
		origFile.Size = fileSize
		if err != nil {
			Response(w, http.StatusCreated, "Strange client", origFile)
			return
		}

		if fileSize > maxUploadSize {
			Response(w, http.StatusCreated, "File too big", origFile)
			return
		}

		// parse and validate file and post parameters
		file, handle, err := r.FormFile("uploadFile")
		if err != nil {
			Response(w, http.StatusCreated, "Invalid file", origFile)
			return
		}
		defer file.Close()

		origFile.NameFile = handle.Filename
		origFile.Mime = handle.Header.Get("Content-Type")
		log.Println("Type file:", origFile.Mime)
		switch origFile.Mime {
		case "image/jpeg", "image/jpg", "image/gif", "image/png":
			saveFile(w, file, origFile)
		case "application/pdf", "application/msword":
			saveFile(w, file, origFile)
		default:
			Response(w, http.StatusCreated, "Invalid file type", origFile)
			return
		}
	})
}

func saveFile(w http.ResponseWriter, file multipart.File, ofile File) {
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println("Err!", err)
		return
	}

	// create directory if not exist
	_, err = ioutil.ReadDir("." + ofile.PathFile)
	if err != nil {
		pathErr := os.MkdirAll("."+ofile.PathFile, 0777)
		if pathErr != nil {
			log.Println("Err!", err)
			return
		}
	}

	err = ioutil.WriteFile("."+ofile.PathFile+ofile.NameFile, data, 0666)
	if err != nil {
		log.Println("Err!", err)
		return
	}
	ofile.URL = "http://" + host + "/" + ofile.PathFile + "/" + ofile.NameFile
	Response(w, http.StatusCreated, "Success", ofile)
}

func Response(w http.ResponseWriter, code int, message string, of File) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	of.Status = message
	jsonData, err := json.Marshal(of)
	if err != nil {
		log.Println("Err!", err)
		return
	}
	w.WriteHeader(code)
	log.Println(code)
	w.Write([]byte(jsonData))
}
