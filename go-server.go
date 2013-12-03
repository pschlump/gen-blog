
// Simple HTTP server

package main

import (
	"log"
	"net/http"
)

func main() {
	port := "8765"
	dir := "/home/pschlump/www/sketchground/pjs/99-static/_site"
	log.Fatal(http.ListenAndServe(":"+port, http.FileServer(http.Dir(dir))))
}

