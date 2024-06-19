package main

import (
	"log"
	"os"
	"path"
	"time"
)

func cleanupTask() {
	for {
		log.Printf("Running cleanup")
		runCleanup()
		log.Printf("Cleanup done")
		time.Sleep(time.Hour)
	}
}

func runCleanup() {
	files, err := os.ReadDir(OUT_DIR)
	if err != nil {
		log.Fatalf("Error reading OUTDIR: %v", err)
	}

	cutoffTime := time.Now().Add(-MAX_CACHE_AGE)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		info, err := file.Info()
		if err != nil {
			log.Printf("Error reading file info (%s): %v", file.Name(), err)
			continue
		}

		if info.ModTime().Before(cutoffTime) {
			fileLockSync.Lock()
			if info.ModTime().Before(cutoffTime) {
				log.Printf("Removing file: %s", file.Name())
				err = os.Remove(path.Join(OUT_DIR, file.Name()))
				if err != nil {
					log.Printf("Error removing file (%s): %v", file.Name(), err)
				}
			}
			fileLockSync.Unlock()
		}
	}
}
