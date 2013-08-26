package document

import (
	"code.google.com/p/go.net/html"
	"fmt"
	"log"
)

type Readabilitier struct {
	content    []readability_score
	candidates map[*html.Node]*readability_score
	article    *readability_score
	body       *html.Node
}

func FlattenHtmlDocument(body *html.Node) (doc *html.Node, article *html.Node) {
	doc, _, article = create_html_sketch()
	flatten_block_node(body, article, false, "")
	return
}

func NewReadabilitier(body *html.Node) *Readabilitier {
	r := &Readabilitier{
		content:    []readability_score{},
		candidates: make(map[*html.Node]*readability_score),
		body:       body}

	r.extract_paragraphs(body)

	var top_candi *readability_score = nil
	for _, candi := range r.candidates {
		candi.content_score = candi.content_score * (100 - candi.link_density()) / 100
		if top_candi == nil || candi.content_score > top_candi.content_score {
			top_candi = candi
		}
	}
	if top_candi != nil {
		r.article = top_candi
	}

	return r
}

/**
         * Now that we have the top candidate, look through its siblings for content that might also be related.
         * Things like preambles, content split by ads that we removed, etc.
**/

func (this *Readabilitier) CreateArticle() (*html.Node, *html.Node) {
	doc, _, article := create_html_sketch()
	if this.article == nil {
		return doc, article
	}
	threshold := max(10, this.article.content_score/5)

	class_name := get_attribute(this.article.element, "class")

	foreach_child(this.article.element.Parent, func(neib *html.Node) {
		append := false
		if neib == this.article.element {
			append = true
		} else if ext, ok := this.candidates[neib]; ok {
			cn := get_attribute(neib, "class")
			if len(cn) > 0 && cn == class_name {
				append = true
				log.Println("append same class", ext)
			}
			if ext.content_score > threshold {
				append = true
				log.Println("append high score neib", ext)
			}
		} else if neib.Type == html.ElementNode && neib.Data == "p" {
			sc := new_boilerpipe_score(neib)
			if sc.words > 65 && sc.link_density() < 22 {
				append = true
				log.Println("append high p", neib)
			}
		}
		if append {
			flatten_block_node(neib, article, false, "")
		}
	})
	return doc, article
}
func (this *Readabilitier) make_readability_score(n *html.Node) *readability_score {
	rtn := new_readability_score(n)
	var (
		pext     *readability_score = nil
		grandext *readability_score = nil
	)
	parent := n.Parent
	var grand *html.Node = nil

	//parent isnt nil
	if i, ok := this.candidates[parent]; ok {
		pext = i
	} else {
		pext = new_readability_score(parent)
		this.candidates[parent] = pext
	}
	if parent != this.body {
		grand = parent.Parent
	}
	if grand != nil {
		if i, ok := this.candidates[grand]; ok {
			grandext = i
		} else {
			grandext = new_readability_score(grand)
			this.candidates[grand] = grandext
		}
	}
	bc := new_boilerpipe_score(n)
	score := bc.commas + 1
	// wrap lines
	score += min(bc.lines(), 3)
	rtn.content_score += score
	if pext != nil {
		pext.content_score += score
	}
	if grandext != nil {
		grandext.content_score += score / 2
	}
	return rtn
}

func (this *Readabilitier) extract_paragraphs(n *html.Node) {
	switch {
	case n.Data == "form" || n.Data == "input" || n.Data == "textarea":
		//    this.content = append(this.content, make_readability_score(n))

		// has only inlines here
	case hasInlineNodes(n):
		this.content = append(this.content, *this.make_readability_score(n))
	default:
		foreach_child(n, func(child *html.Node) {
			this.extract_paragraphs(child)
		})
	}
}

// text-node
// <a>
// <img> <object> <embed> <video> <audio>
// <ul> <ol> <form> <textarea> <input> will be reserved
func flatten_block_node(b *html.Node, article *html.Node, flatt bool, class string) {
	cur_class := cat_class(b, class)
	switch {
	case b.Data == "form" || b.Data == "inputbox" || b.Data == "textarea":
	case flatt && is_unflatten_node(b):
		nb := create_element(b.Data)
		try_update_class_attr(nb, cur_class)
		flatten_block_node(b, nb, false, class)
		article.AppendChild(nb)
	case hasInlineNodes(b):
		p := create_p(b)
		try_update_class_attr(p, cur_class)
		article.AppendChild(p)
	default:
		foreach_child(b, func(child *html.Node) {
			flatten_block_node(child, article, true, cur_class)
		})
	}
}

func get_class_weight(n *html.Node, attname string) int {
	c := get_attribute(n, attname)

	weight := 0
	if negative.MatchString(c) {
		weight -= 25
	}
	if positive.MatchString(c) {
		weight += 25
	}
	return weight
}

func (this *Readabilitier) String() string {
	return fmt.Sprint("readerabilitier content:", len(this.content), ", candidates:", len(this.candidates))
}
