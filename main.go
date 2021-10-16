package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const MAX_UPLOAD_SIZE int64 = 1024 * 1024 * 10 //maximum of 10mb per file
type Progress struct {
	TotalBytesToRead int64
	TotalBytesRead   int64
}

// create the write method so satisfy the i0.Writer interface
func (p *Progress) Write(pb []byte) (int, error) {
	n := len(pb) //length of the file read so far
	// set error to nil since no error will be handled

	p.TotalBytesRead = int64(n)

	if p.TotalBytesRead == p.TotalBytesToRead {
		fmt.Println("Done")
		return n, nil
	}
	fmt.Printf("File upload still in progress -- %d\n", p.TotalBytesRead)

	return n, nil
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	// this neccessary else you will not be able to access other properties needed to get the file
	// this method returns an err
	err := r.ParseForm()
	// always handle error
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusInternalServerError)
		return
	}

	file, fileHeader, err := r.FormFile("image")
	// get the file by entering its form name.
	// handle errors as neccessary
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}
	//postpone closing the file until the function is complete
	defer file.Close()

	if fileHeader.Size > MAX_UPLOAD_SIZE {
		http.Error(w, "file sizes cannot be bigger than 10mb", http.StatusBadRequest)
		return
	}
	// create file on the local server, filepath.Ext() will get the extension out of the filename
	localfile, err := os.Create(time.Now().UTC().String() + filepath.Ext(fileHeader.Filename))
	//handle errors as neccessary
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}

	// Create a new instance of progress
	progress := &Progress{
		TotalBytesToRead: fileHeader.Size,
	}

	// use the io.Copy() method to copy bytes to the localbyte from the uploaded file,
	// this method returns the number of bytes copied and an error, we would be ignoring the bytes
	_, err = io.Copy(localfile, io.TeeReader(file, progress))
	//handle errors as neccessary
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}

	//after which we can simply respond back to the client
	fmt.Fprint(w, "File upload successfull")
}
func uploadMultipleFiles(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(1024 * 1024 * 10) //maxMemory if 10mb

	//we get the slice of files
	files := r.MultipartForm.File["images"]
	// looping through the files to upload each file individually
	for _, file := range files {

		uploadFile, err := file.Open()
		// handle error as neccessary
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// postpone file closure until the function is complete
		defer uploadFile.Close()

		if file.Size > MAX_UPLOAD_SIZE {
			http.Error(w, "file sizes cannot be bigger than 10mb", http.StatusBadRequest)
			return
		}
		//create a new localfile
		localfile, err := os.Create(time.Now().UTC().String() + filepath.Ext(file.Filename))
		//handle errors as neccessary
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = io.Copy(localfile, uploadFile)
		//handle errors as neccessary
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	//after which we can simply respond back to the client
	fmt.Fprint(w, "File upload successfull")

}
func indexHandler(res http.ResponseWriter, r *http.Request) {
	res.Header().Add("Content-Type", "text/html")
	http.ServeFile(res, r, "index.html")
}

func main() {

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/uploadmultiple", uploadMultipleFiles)

	http.ListenAndServe(":9000", nil)
}
