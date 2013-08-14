// gextract project main.go
package main

import (
	"bufio"
	"bytes"
	"code.google.com/p/go.net/html"
	"flag"
	"fmt"
	htmldoc "gextract/document"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	html_file = flag.String("f", "1.html", "the html file path")
	out       = flag.String("o", "o.html", "the cleaned-up filepath")
	url       = flag.String("u", "", "the online web page")
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Fatal(err)
		}
	}()
	flag.Parse()

	if len(*url) != 0 {
		fetch_url(*url, *html_file)
	}
	if len(*html_file) == 0 {
		fmt.Println("html_file is required")
		return
	}
	// data, err := ioutil.ReadFile(*html_file)
	f, err := os.Open(*html_file)
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(f)
	doc, err := html.Parse(reader)
	if err != nil {
		panic(err)
	}
	cleaner := htmldoc.NewHtmlCleaner()
	log.Println("prepare cleaning doc")
	cleaner.CleanHtml(doc)

	write_html(doc, "2.html")

	doc1, article := htmldoc.FlattenHtmlDocument(cleaner.Article)
	write_html(doc1, "3.html")

	boiler := htmldoc.NewBoilerpiper(article)
	boiler.NumberWordsRulesFilter()

	write_html(doc1, "4.html")

	boiler.FormPrefixFilter()
	write_html(doc1, "5.html")

}

func write_html(doc *html.Node, fp string) {
	data := new(bytes.Buffer)
	if err := html.Render(data, doc); err != nil {
		panic(err)
	}
	// fmt.Println(data.String())
	if err := ioutil.WriteFile(fp, data.Bytes(), 0644); err != nil {
		panic(err)
	}

}

func fetch_url(url string, ofile string) {
	log.Println("get from", url)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	of, err := os.Create(ofile)
	defer of.Close()

	io.Copy(of, resp.Body)
}
