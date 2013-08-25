// gextract project main.go
package main

import (
	"bufio"
	"code.google.com/p/go.net/html"
	"flag"
	htmldoc "gextract/document"
	"log"
	"os"
)

var (
	//	html_file = flag.String("f", "1.html", "the html file path")
	//	out = flag.String("o", "5.html", "the cleaned-up filepath")
	url = flag.String("u", "", "the online web page")
)

func main() {
	flag.Parse()
	htmlfile := ""
	if len(*url) != 0 {
		htmlfile, _ = htmldoc.FetchUrl2(*url)
		log.Println("writing step 1", htmlfile)

	}

	f, err := os.Open(htmlfile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	doc, err := html.Parse(reader)
	if err != nil {
		log.Println(err)
		return
	}

	cleaner := htmldoc.NewHtmlCleaner(*url)
	log.Println("prepare cleaning doc")
	cleaner.CleanHtml(doc)
	log.Println(cleaner)

	//	h2ml := write_html(doc)
	log.Println("writing step 2", htmldoc.WriteHtmlFile2(doc))

	//	doc1, _ := htmldoc.FlattenHtmlDocument(cleaner.Article)
	//	write_html(doc1, "3.html")

	rdr := htmldoc.NewReadabilitier(cleaner.Article)

	doc1, article := rdr.CreateArticle()

	//	write_html(doc1, "3.html")
	log.Println("writing step 3", htmldoc.WriteHtmlFile2(doc1))

	boiler := htmldoc.NewBoilerpiper(article)
	boiler.NumberWordsRulesFilter()

	//	write_html(doc1, "4.html")
	h4ml := htmldoc.WriteHtmlFile2(doc1)
	log.Println("writing step 4", h4ml)

	boiler.FormPrefixFilter()
	//	write_html(doc1, *out)
	h5ml := htmldoc.WriteHtmlFile2(doc1)
	log.Println("writing step 5", h5ml)
}
