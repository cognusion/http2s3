package main

import (
	"fmt"
	"github.com/spf13/cast"
	"github.com/zenazn/goji/web"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// This is the "catchall" used to redirect everything else the URLbase
func defaultHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	newUrl := GlobalConfig.Get("urlBase")

	sep := ""
	if !strings.HasPrefix(r.URL.Path, "/") {
		sep = "/"
	}

	debugOut.Printf("Default redirect to %s\n", fmt.Sprintf("%s%s%s", newUrl, sep, r.URL.Path))
	http.Redirect(w, r, fmt.Sprintf("%s%s%s", newUrl, sep, r.URL.Path), 301)
}

// The load-balancer needs a zippy, page to land on. This is that page.
// TODO: Make this better
func healthHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Healthy as a gopher")
}

// Grabs the values from the form, downloads the file from the user,
// uploads it to S3, and communicates the results via HipChat
func uploadHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	defer Track("Upload-Total", Now(), debugOut)
	// After we're done, redirect them elsewhere
	defer http.Redirect(w, r, GlobalConfig.Get("thanksURL"), 301)

	// Roughly 256M max memory consumption
	var mem int64 = 256
	if GlobalConfig.IsNotNull("maxFormMemMB") {
		mem = cast.ToInt64(GlobalConfig.Get("maxFormMemMB"))
	}

	err := r.ParseMultipartForm(mem * 1000000) // Yes, that's < 256M. Intentional.
	if err != nil {
		// Form parse failed
		// go errorMessage(fmt.Sprintf("Error parsing form: '%v'", err)
		return
	}

	// Grab the name bits
	name := r.FormValue(GlobalConfig.Get("formNameField"))
	email := r.FormValue(GlobalConfig.Get("formEmailField"))
	to := r.FormValue(GlobalConfig.Get("formToField"))
	from := fmt.Sprintf("%s <%s>", name, email)

	// Grab the file from the browser
	fileHandle, headers, err := r.FormFile(GlobalConfig.Get("formFileField"))
	if err != nil {
		// Form file was empty. Spam or bad submit.
		//go errorMessage(fmt.Sprintf("Error acquiring file from %s: '%v'", from, err))
		return
	}
	defer fileHandle.Close()
	defer Track("Upload-Downloaded", Now(), debugOut)

	// See if it's worth grabbing
	if isBadFileMaybe(headers.Filename) {
		go errorMessage(fmt.Sprintf("Rejecting file '%s' from %s", headers.Filename, from))
		return
	}

	// Create temp folder
	newFolder := GlobalConfig.Get("tempFolderRoot") + "/" + randString(10)
	newFile := newFolder + "/" + headers.Filename

	derr := os.Mkdir(newFolder, 0777)
	if derr != nil {
		go errorMessage(fmt.Sprintf("Error creating folder %s for file from %s: '%v'", newFolder, from, derr))
		return
	}

	// Write the file to disk
	f, err := os.OpenFile(newFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		go errorMessage(fmt.Sprintf("Error writing file %s from %s: '%v'", newFile, from, err))
		return
	}
	defer f.Close()

	// Copy the file, possibly from memory) to the temp file
	io.Copy(f, fileHandle)
	defer Track("Upload-Copied", Now(), debugOut)

	// NOTE: At this point, we let the browser go. The upload to S3 will happen
	// asynchronously, and then whomever needs to be notified, will be. No need
	// to keep them on the line

	// Copy the file off to S3, let the support room know, and clean up
	go func(file, folder, sender, to string) {
		defer Track("Upload-Deleted", Now(), debugOut)
		defer deleteFiles([]string{file, folder})

		size, err := fileToBucket(file, GlobalConfig.Get("awsS3Bucket"))
		if err != nil {
			go errorMessage(fmt.Sprintf("Error copying file %s to S3 for %s: '%v'", file, sender, err))
			return
		}

		fsize := byteFormat(size)
		baseFilename := filepath.Base(file)
		url := "s3://" + GlobalConfig.Get("awsS3Bucket") + "/" + baseFilename

		// Hit the message callback
		// TODO: We have chans for a reason
		go message_callback(sender, to, url, fsize)

	}(newFile, newFolder, from, to)
}
