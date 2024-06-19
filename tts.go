package main

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"sync"
)

var fileLockSync sync.Mutex
var queueSync sync.Mutex
var queueMap map[string]*sync.WaitGroup

func initTTS() {
	queueMap = make(map[string]*sync.WaitGroup)
}

func fileExists(localFilenameMP3 string) (exists bool) {
	exists = false
	fileLockSync.Lock()
	stat, err := os.Stat(localFilenameMP3)
	if err == nil && stat != nil {
		exists = true
		fh, err := os.OpenFile(localFilenameMP3, os.O_APPEND, 0666)
		if err != nil {
			log.Printf("Error refreshing file: %v", err)
			exists = false
		}
		_ = fh.Close()
	}
	fileLockSync.Unlock()
	return
}

func mp3(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		return
	}

	text := req.Form.Get("q")

	if len(text) > MAX_LENGTH {
		http.Error(w, "Message too long", 400)
		return
	}

	hash := sha512.New512_256()
	hash.Write([]byte(text))
	hashSum := hash.Sum([]byte{})

	filename := hex.EncodeToString(hashSum)

	filenameWAV := path.Join(OUT_DIR, filename+".wav")
	filenameMP3 := filename + ".mp3"
	localFilenameMP3 := path.Join(OUT_DIR, filenameMP3)
	localFilenameMP3Temp := localFilenameMP3 + ".tmp.mp3"

	if !fileExists(localFilenameMP3) {
		queueSync.Lock()
		curQueue, hadQueue := queueMap[filename]
		if !hadQueue {
			curQueue = &sync.WaitGroup{}
			curQueue.Add(1)
			queueMap[filename] = curQueue
		}
		queueSync.Unlock()

		if !hadQueue {
			defer func() {
				_ = os.Remove(filenameWAV)
				_ = os.Remove(localFilenameMP3Temp)
			}()

			defer func() {
				curQueue.Done()

				queueSync.Lock()
				delete(queueMap, filename)
				queueSync.Unlock()
			}()

			err = exec.Command("espeak", "-v", "en", "-w", filenameWAV, text).Run()
			if err != nil {
				log.Printf("Error converting text to speech: %v", err)
				http.Error(w, "Error converting text to speech", 500)
				return
			}
			err = exec.Command("lame", filenameWAV, localFilenameMP3Temp).Run()
			if err != nil {
				log.Printf("Error converting WAV to MP3: %v", err)
				http.Error(w, "Error converting WAV to MP3", 500)
				return
			}
			err = os.Rename(localFilenameMP3Temp, localFilenameMP3)
			if err != nil {
				log.Printf("Error moving MP3 file: %v", err)
				http.Error(w, "Error moving MP3 file", 500)
				return
			}
		} else {
			curQueue.Wait()
			if !fileExists(localFilenameMP3) {
				log.Printf("Error generating MP3")
				http.Error(w, "Error generating MP3", 500)
				return
			}
		}
	}

	fmt.Fprintf(w, "https://tts.spaceage.mp/files/%s", filenameMP3)
}
