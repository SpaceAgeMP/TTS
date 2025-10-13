package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

var OUT_DIR = os.Getenv("OUT_DIR")
var LISTEN_ADDR = os.Getenv("LISTEN_ADDR")

const MAX_LENGTH = 256
const MAX_CACHE_AGE = time.Hour * 24 * 14

func health(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("OK"))
}

func main() {
	os.Mkdir(OUT_DIR, 0755)
	initTTS()

	http.Handle("/files/", http.StripPrefix("/files", http.FileServer(http.Dir(OUT_DIR))))
	http.HandleFunc("/mp3", mp3)
	http.HandleFunc("/health", health)

	go cleanupTask()
	log.Printf("Listening on %s", LISTEN_ADDR)
	http.ListenAndServe(LISTEN_ADDR, nil)
}
