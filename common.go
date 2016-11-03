package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Determine how to best output error information
func errorMessage(message string) {
	debugOut.Println(message)

	if GlobalConfig.IsNotNull("hipChatRoom") {
		// We're using hipchat
		hipchatErrorMessage(message)
	}
}

// Return a human-readable string representation of a byte count
func byteFormat(num_in int64) string {
	suffix := "B"
	num := float64(num_in)
	units := []string{"", "K", "M", "G", "T", "P", "E", "Z"} // "Y" caught  below
	for _, unit := range units {
		if num < 1024.0 {
			return fmt.Sprintf("%3.1f%s%s", num, unit, suffix)
		}
		num = (num / 1024)
	}
	return fmt.Sprintf("%.1f%s%s", num, "Y", suffix)
}

// Return a randomish string of the specified size
func randString(size int) string {
	chars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	bytes := make([]byte, size)

	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = chars[v%byte(len(chars))]
	}
	return string(bytes)
}

// "Hail Mary" deleting the list of files, in order
func deleteFiles(files []string) {
	for _, f := range files {
		os.Remove(f)
	}
}

// Guesses if a file is "bad" based on its name
func isBadFileMaybe(filename string) (isBad bool) {
	isBad = false // Default good

	if GlobalConfig.Exists("badFileExts") {
		badExts := GlobalConfig.GetArray("badFileExts")
		ext := strings.ToLower(filepath.Ext(filename))

		debugOut.Printf("IBFM for %s (%s): ", filename, ext)
		for _, e := range badExts {
			if strings.ToLower(strings.TrimSpace(e)) == ext {
				debugOut.Printf(" Triggered on %s", e)
				isBad = true
				break
			}
		}
		debugOut.Printf("\n")
	}
	return isBad
}
