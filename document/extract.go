package document

import (
	"bufio"
	"code.google.com/p/go.net/html"
	"log"
	"os"
)

func ExtractHtml(url string) string {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	htmlfile := FetchUrl2(url)
	log.Println("writing step 1", htmlfile)

	f, err := os.Open(htmlfile)
	try_panic(err)
	defer f.Close()

	reader := bufio.NewReader(f)
	doc, err := html.Parse(reader)
	try_panic(err)

	cleaner := NewHtmlCleaner(url)
	cleaner.CleanHtml(doc)
	log.Println(cleaner)

	log.Println("writing step 2", WriteHtmlFile2(doc))

	rdr := NewReadabilitier(cleaner.Article)
	doc1, article := rdr.CreateArticle()

	//	write_html(doc1, "3.html")
	log.Println("writing step 3", WriteHtmlFile2(doc1))

	boiler := NewBoilerpiper(article)
	boiler.NumberWordsRulesFilter()

	//	write_html(doc1, "4.html")
	h4ml := WriteHtmlFile2(doc1)
	log.Println("writing step 4", h4ml)

	boiler.FormPrefixFilter()
	h5ml := WriteHtmlFile2(doc1)
	log.Println("writing step 5", h5ml)
	return h5ml
}
