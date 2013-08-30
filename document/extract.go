package document

import (
	"bufio"
	"code.google.com/p/go.net/html"
	"log"
	"os"
)

// cleaned html doc by utf-8 encoded
// return filepath, *SummaryScore, error
func ExtractHtml(url string) (string, *SummaryScore, error) {
	htmlfile, _, err := DownloadHtml(url)
	if err != nil {
		return "", &SummaryScore{}, err
	}
	//	log.Println("writing step 1", htmlfile)

	f, err := os.Open(htmlfile)
	if err != nil {
		return "", &SummaryScore{}, err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	doc, err := html.Parse(reader)
	if err != nil {
		return "", &SummaryScore{}, err
	}

	cleaner := NewHtmlCleaner(url)
	cleaner.CleanHtml(doc)
	//	log.Println(cleaner)

	s2, _ := WriteHtmlFile2(doc)
	log.Println("writing step 2", s2)

	rdr := NewReadabilitier(cleaner.Article)
	doc1, article := rdr.CreateArticle()

	s2, _ = WriteHtmlFile2(doc1)
	log.Println("writing step 3", s2)

	boiler := NewBoilerpiper(article)
	boiler.NumberWordsRulesFilter()

	h4ml, _ := WriteHtmlFile2(doc1)
	log.Println("writing step 4", h4ml)

	boiler.FormPrefixFilter()
	h5ml, err := WriteHtmlFile2(doc1)
	log.Println("writing step 5", h5ml)
	return h5ml, NewSummaryScore(doc1), err
}
