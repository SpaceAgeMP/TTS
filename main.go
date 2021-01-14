package main

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"
)

var outdir string

var queueSync sync.Mutex
var queueMap map[string]*sync.WaitGroup

func mp3(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		return
	}

	text := req.Form.Get("q")

	hash := sha512.New512_256()
	hash.Write([]byte(text))
	hashSum := hash.Sum([]byte{})

	filename := hex.EncodeToString(hashSum)

	filenameWAV := outdir + filename + ".wav"
	filenameMP3 := outdir + filename + ".mp3"

	_, err = os.Stat(filenameMP3)
	if os.IsNotExist(err) {
		queueSync.Lock()
		curQueue, hadQueue := queueMap[filename]
		if !hadQueue {
			curQueue = &sync.WaitGroup{}
			curQueue.Add(1)
			queueMap[filename] = curQueue
		}
		queueSync.Unlock()

		if !hadQueue {
			exec.Command("espeak", "-v", "en", "-w", filenameWAV, text).Run()
			exec.Command("lame", filenameWAV, filenameMP3).Run()
			os.Remove(filenameWAV)
			curQueue.Done()

			queueSync.Lock()
			delete(queueMap, filename)
			queueSync.Unlock()
		} else {
			curQueue.Wait()
		}
	} else {
		now := time.Now().Local()
		os.Chtimes(filenameMP3, now, now)
	}

	fmt.Fprintf(w, "https://api.spaceage.mp/out/%s.mp3", filename)
}

func main() {
	queueMap = make(map[string]*sync.WaitGroup)
	outdir = "./out/"
	http.HandleFunc("/tts/mp3", mp3)
	http.ListenAndServe(":4001", nil)
}
