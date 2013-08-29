package main

import (
	"flag"
	d "gextract/document"
	"gextract/feeds"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	max_cocurrents = 2
	chan_buf_size  = 100
	target_folder  = "/home/hearts/Pictures/cache/"
)

func main() {
	defer recover_panic()
	flag.Parse()
	images := feeds.ImagesUnready()
	log.Println(len(images))

	cocurrents := max_cocurrents
	echans := make([]chan *feeds.Image, cocurrents)
	done := make(chan int)

	for i := 0; i < cocurrents; i++ {
		echan := make(chan *feeds.Image, chan_buf_size)
		echans[i] = echan
		go touch_entry(echan, done)
	}

	for i, e := range images {
		x := e
		echans[i%cocurrents] <- &x
	}

	for i := 0; i < cocurrents; i++ {
		echans[i] <- nil
	}

	for i := 0; i < cocurrents; i++ {
		<-done
	}
}

func init() {
	log.SetOutput(os.Stderr)
}

func recover_panic() {
	if err := recover(); err != nil {
		log.Fatalln(err)
	}
}

func touch_entry(echan <-chan *feeds.Image, done chan<- int) {
	for e := <-echan; e != nil; e = <-echan {
		tf, mt, _, _ := d.DownloadFile2(e.Link)
		targetf, ok := copy_to_target(target_folder, tf, mt)
		log.Println(targetf, ok, e.Link)
		if ok == false {
			targetf = ""
		}
		feeds.ImageUpdateState(e.Link, targetf)
	}
	done <- 0
}

func copy_to_target(folder, src, media_type string) (string, bool) {
	fields := strings.Split(media_type, "/")
	if len(src) == 0 || len(fields) != 2 || fields[0] != "image" {
		return "", false
	}
	ext := fields[1]
	_, fn := filepath.Split(src)
	t := filepath.Join(folder, fn+"."+ext)
	ok := copy_file(t, src)
	return t, ok
}

func copy_file(target, src string) bool {
	sf, err := os.Open(src)
	if err != nil {
		return false
	}
	defer sf.Close()

	df, err := os.Create(target)
	if err != nil {
		log.Println(err)
		return false
	}
	defer df.Close()
	_, err = io.Copy(df, sf)
	if err != nil {
		return false
	}
	return true
}
