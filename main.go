package main

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const maxLength = 256

var outdir string

var queueSync sync.Mutex
var queueMap map[string]*sync.WaitGroup

var s3bucket = "spaceage-tts"
var s3ssession = session.Must(session.NewSessionWithOptions(session.Options{
	SharedConfigState: session.SharedConfigEnable,
}))
var s3client = s3.New(s3ssession)

func fileExists(fileName string) (bool, error) {
	_, err := s3client.HeadObject(&s3.HeadObjectInput{
		Key: aws.String(fileName),
		Bucket: aws.String(s3bucket),
	})
	if err == nil {
		return true, nil
	}

	awserr, ok := err.(awserr.Error)
	if ok && awserr.Code() == "NotFound" {
		return false, nil
	}

	return false, err
}

func mp3(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		return
	}

	text := req.Form.Get("q")

	if len(text) > maxLength {
		http.Error(w, "Message too long", 400)
		return
	}

	hash := sha512.New512_256()
	hash.Write([]byte(text))
	hashSum := hash.Sum([]byte{})

	filename := hex.EncodeToString(hashSum)

	filenameWAV := outdir + filename + ".wav"
	filenameMP3 := filename + ".mp3"
	localFilenameMP3 := outdir + filenameMP3

	exists, err := fileExists(filenameMP3)
	if !exists {
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
			exec.Command("lame", filenameWAV, localFilenameMP3).Run()
			os.Remove(filenameWAV)
			f, _ := os.Open(localFilenameMP3)
			_, err := s3client.PutObject(&s3.PutObjectInput{
				Key:    aws.String(filenameMP3),
				Bucket: aws.String(s3bucket),
				Body:   f,
				Bucket: ,
			})
			if err != nil {
				log.Printf("Issue uploading to S3: %v", err)
			}
			f.Close()
			os.Remove(localFilenameMP3)
			curQueue.Done()

			queueSync.Lock()
			delete(queueMap, filename)
			queueSync.Unlock()
		} else {
			curQueue.Wait()
		}
	}

	fmt.Fprintf(w, "https://d1x5a3iv2gxgba.cloudfront.net/%s", filenameMP3)
}

func main() {
	queueMap = make(map[string]*sync.WaitGroup)
	outdir = "out/"
	os.Mkdir(outdir, 0755)
	http.HandleFunc("/mp3", mp3)
	http.ListenAndServe("127.0.0.1:4001", nil)
}
