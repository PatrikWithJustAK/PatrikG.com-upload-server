package main

//handlers.go contains all the http handlers for our server
//these take in an HTTP request *http.Request, and return HTML templates that have context data applied
import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

// PageData is a generic struct containing a string for development purposes
type PageData struct {
	Link    string
	Imglist []string
}

// This handles files uploaded to the uploadform.html
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		uploadPageHandler(w, r)
		return
	}

	// Parse the form with a max upload size of 100 MB
	err := r.ParseMultipartForm(100 << 20) // 100 MB
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// My S3 bucket name is "patrikguploads"
	bucketName := "patrikguploads"

	// Upload the file to S3
	fileURL, err := uploadFileToS3(file, fileHeader, bucketName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//prepare the template with the navbar, icon, uploadform, and main container componenets
	tmpl, err := template.ParseFiles("assets/index.html", "assets/nav.html", "assets/icon.html", "assets/uploadform.html", "assets/imagelist.html")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}
	imgs := listS3Objects(bucketName)
	data := PageData{
		Link:    fileURL,
		Imglist: imgs}
	err = tmpl.Execute(w, data)
	if err != nil {
		fmt.Printf("error: %s", err)
	}

}

func uploadPageHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the template file

	tmpl, err := template.ParseFiles("assets/index.html", "assets/nav.html", "assets/icon.html", "assets/uploadform.html", "assets/imagelist.html")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	// Create a new PageData to pass to the template

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
	}

}
func extensionChecker(p string) PageData {
	ext := strings.ToUpper(p)
	if strings.HasSuffix(ext, ".PNG") || strings.HasSuffix(ext, ".SVG") {
		return PageData{Link: p}
	}
	return PageData{Link: p}
}
