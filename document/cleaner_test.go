package document

import (
	"code.google.com/p/go.net/html"
	"log"
	"testing"
)

func TestCleaner(t *testing.T) {
	doc := NewHtmlDocument("../2.html")
	cleaner := NewHtmlCleaner()

	cleaner.CleanHtml(doc)

	WriteHtmlFile(doc, "../3.html.txt")

}

func print_html_doc(node *html.Node) {
	foreach_child(node, func(child *html.Node) {
		print_html_doc(child)
	})
	log.Println(trim(node.Data))
}

func trim(d string) string {
	if len(d) < 10 {
		return d
	}
	return d[:10]
}
