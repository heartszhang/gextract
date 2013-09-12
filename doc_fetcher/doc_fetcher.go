package main

import (
	"flag"
	d "github.com/heartszhang/gextract/document"
	"github.com/heartszhang/gextract/feeds"
	"github.com/heartszhang/gextract/feeds/meta"
	"log"
	"os"
)

const (
	max_cocurrents = 8
	chan_buf_size  = 100
)

func main() {
	defer recover_panic()
	flag.Parse()
	/*
			entries := feeds.EntriesContentUnready()
			log.Println(len(entries))

		cocurrents := max_cocurrents
		echans := make([]chan *feeds.Entry, cocurrents)
		done := make(chan int)

		for i := 0; i < cocurrents; i++ {
			echan := make(chan *feeds.Entry, chan_buf_size)
			echans[i] = echan
			go touch_entry(echan, done)
		}

		for i, e := range entries {
			x := e
			echans[i%cocurrents] <- &x
		}

		for i := 0; i < cocurrents; i++ {
			echans[i] <- nil
		}

		for i := 0; i < cocurrents; i++ {
			<-done
		}
	*/
}

func init() {
	log.SetOutput(os.Stderr)
}

func recover_panic() {
	if err := recover(); err != nil {
		log.Fatalln(err)
	}
}

func touch_entry(echan <-chan *meta.Entry, done chan<- int) {
	for e := <-echan; e != nil; e = <-echan {
		tf, sc, _ := d.ExtractHtml(e.Link)
		//		cf := feeds.NewContentFile(tf, sc.Words, sc.Imgs)
		feeds.NewEntryOperator().SetContent(e.Link, tf, sc.WordCount, sc.Images)
		//		feeds.EntryUpdateContent(tf, e.Link)
		//		feeds.FetchEntryImagesExternal(e)
	}
	done <- 0
}
