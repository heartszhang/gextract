package main

import (
	"flag"
	d "gextract/document"
	"log"
	"os"
)

var uri = flag.String("uri", "", "rss 2.0 url")

func init() {
	log.SetOutput(os.Stderr)
}

func main() {
	flag.Parse()
	if len(*uri) == 0 {
		log.Println("html_extractor --uri http://xxxx.com")
		return
	}
	tf := d.ExtractHtml(*uri)
	log.Println(tf)
}
