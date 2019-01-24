package main

import (
	"encoding/json"
	"flag"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/disintegration/imaging"
)

type RespData struct {
	URL       string `json:"url,omitempty"`
	Mime      string `json:"mime,omitempty"`
	Original  File   `json:"original,omitempty"`
	Thumbnail File   `json:"thumbnail,omitempty"`
	Status    string `json:"status"`
}

type File struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
	Size int64  `json:"size,omitempty"`
}

var (
	maxUploadSize = flag.Int64("maxsize", 9437184, "max allowed size uploaded files in Byte")
	uploadPath    = flag.String("path", "/files/", "root upload path")
	host          = flag.String("host", "127.0.0.1:8081", "server address")
	domain        = flag.String("domain", "", "domain address (with http://)")
	thumwidth     = flag.Int64("thumwidth", 300, "thumbnail width")
	thumhigher    = flag.Int64("thumhigher", 200, "thumbnail higher")
)

func main() {
	flag.Parse()

	http.Handle("/files/", http.StripPrefix("/files", http.FileServer(http.Dir("."+*uploadPath))))
	http.HandleFunc("/upload", uploadFileHandler())
	log.Print("Server started on " + *host + " ...")
	log.Fatal(http.ListenAndServe(*host, nil))
}

func uploadFileHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rdata RespData

		// validate max Reguest size
		r.Body = http.MaxBytesReader(w, r.Body, *maxUploadSize)
		cSize, err := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64)
		if err != nil {
			response(w, http.StatusCreated, "Invalid request", rdata)
			return
		}
		if cSize > *maxUploadSize {
			response(w, http.StatusCreated, "File too big", rdata)
		}
		//rdata.Original.Size = cSize
		err = r.ParseMultipartForm(*maxUploadSize)
		if err != nil {
			return
		}

		if r.Method != http.MethodPost || len(r.FormValue("pathFile")) == 0 {
			response(w, http.StatusCreated, "Error post data", rdata)
			return
		}
		rdata.Original.Path = *uploadPath + r.FormValue("pathFile") + "/"

		file, handle, err := r.FormFile("uploadFile")
		if err != nil {
			response(w, http.StatusCreated, "Invalid file", rdata)
			return
		}
		defer file.Close()
		rdata.Original.Name = handle.Filename
		rdata.Mime = handle.Header.Get("Content-Type")
		switch rdata.Mime {
		case "image/jpeg", "image/jpg", "image/gif", "image/png", "image/bmp", "image/tif":
			saveFile(file, rdata.Original.Path, rdata.Original.Name)
			rdata.Thumbnail.Name = handle.Filename
			rdata.Thumbnail.Path = rdata.Original.Path + "thumbnail/"
			saveThumbnail(rdata)
		case "application/pdf", "application/msword", "application/octet-stream":
			saveFile(file, rdata.Original.Path, rdata.Original.Name)
			rdata.Thumbnail.Name = "application.svg"
			rdata.Thumbnail.Path = *uploadPath
		default:
			log.Println("Err! Invalid file type:", rdata.Mime)
			response(w, http.StatusCreated, "Invalid file type", rdata)
			return
		}

		rdata.Original.Size = fileSize(rdata.Original.Path + rdata.Original.Name)
		rdata.Thumbnail.Size = fileSize(rdata.Thumbnail.Path + rdata.Thumbnail.Name)

		if len(*domain) > 0 {
			rdata.URL = *domain
		} else {
			rdata.URL = "http://" + *host
		}
		response(w, http.StatusCreated, "Success", rdata)
	})
}

//func saveFile(file multipart.File, rdata respData) {
func saveFile(file multipart.File, pathFile string, fname string) {
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println("Err!", err)
		return
	}

	// create directory if not exist
	_, err = ioutil.ReadDir("." + pathFile)
	if err != nil {
		pathErr := os.MkdirAll("."+pathFile, 0777)
		if pathErr != nil {
			log.Println("Err!", err)
			return
		}
	}

	// write file
	err = ioutil.WriteFile("."+pathFile+fname, data, 0666)
	if err != nil {
		log.Println("Err!", err)
		return
	}
}

func saveThumbnail(rdata RespData) {
	_, err := ioutil.ReadDir("." + rdata.Thumbnail.Path)
	if err != nil {
		pathErr := os.MkdirAll("."+rdata.Thumbnail.Path, 0777)
		if pathErr != nil {
			log.Println("Err!", err)
			return
		}
	}
	src, err := imaging.Open("." + rdata.Original.Path + rdata.Original.Name)
	if err != nil {
		log.Println("Err!", err)
	}
	src = imaging.Fill(src, int(*thumwidth), int(*thumhigher), imaging.Center, imaging.Lanczos)
	dst := imaging.New(int(*thumwidth), int(*thumhigher), color.NRGBA{0, 0, 0, 0})
	dst = imaging.Paste(dst, src, image.Pt(0, 0))
	err = imaging.Save(dst, "."+rdata.Thumbnail.Path+rdata.Thumbnail.Name)
	if err != nil {
		log.Println("Err!", err)
	}
}

func response(w http.ResponseWriter, code int, message string, rd RespData) {
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

func fileSize(path string) int64 {
	fi, err := os.Stat("." + path)
	if err != nil {
		log.Println("Err!", err)
	}
	return fi.Size()
}
