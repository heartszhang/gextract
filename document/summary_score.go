package document

import (
	"code.google.com/p/go.net/html"

//	"log"

//	"code.google.com/p/go.net/html/atom"
)

type SummaryScore struct {
	WordCount int `json:"word_count" bson:"word_count"`
	//	ImageCount int      `json:"image_count" bson:"image_count"`
	LinkCount int      `json:"link_count" bson:"link_count"`
	Images    []string `json:"image,omitempty" bson:"image,omitempty"`
}

func NewSummaryScore(n *html.Node) *SummaryScore {
	rtn := &SummaryScore{Images: []string{}}
	if n == nil {
		return rtn
	}
	foreach_child(n, func(child *html.Node) {
		switch {
		case child.Type == html.CommentNode:
		case child.Type == html.DoctypeNode:
		case child.Type == html.TextNode:
			_, c, _ := count_words(child.Data)
			rtn.WordCount +=c
		case child.Data == "img":
			rtn.Images = append(rtn.Images, get_attribute(child, "src"))
		case child.Data == "a":
			rtn.LinkCount++
//			rtn.add(NewSummaryScore(child))
		default:
			sc := NewSummaryScore(child)
			rtn.add(sc)
		}
	})
	return rtn
}
func (this *SummaryScore) add(l *SummaryScore) {
	if l == nil {
		return
	}
	this.WordCount += l.WordCount
	//	this.ImageCount += l.ImageCount
	this.LinkCount += l.LinkCount
	imgs := make([]string, len(this.Images)+len(l.Images))
	copy(imgs, this.Images)
	copy(imgs[len(this.Images):], l.Images)
	this.Images = imgs
	//	this.Images = append(this.Images, l.Images)
}
