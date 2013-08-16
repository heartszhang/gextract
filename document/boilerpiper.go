package document

import (
	"code.google.com/p/go.net/html"
	"log"
	"strings"
)

type Boilerpiper struct {
	titles      []string
	description string
	authors     []string
	keywords    []string

	content []*boilerpipe_score

	words       int
	lines       int
	chars       int
	inner_chars int
	outer_chars int
	parags      int

	words_p_boilerpipe_score float64
	quality                  float64

	body *html.Node
}

func NewBoilerpiper(article *html.Node) *Boilerpiper {
	rtn := new_boilerpiper()
	rtn.flatten_paragraphs(article)
	return rtn
}

func (this *Boilerpiper) flatten_paragraphs(n *html.Node) {
	switch {
	case n.Data == "form" || n.Data == "input" || n.Data == "textarea":
		bs := new_boilerpipe_score(n)
		this.content = append(this.content, &bs)
	case hasInlineNodes(n):
		bs := new_boilerpipe_score(n)
		this.content = append(this.content, &bs)
	default:
		foreach_child(n, func(child *html.Node) {
			this.flatten_paragraphs(child)
		})
	}
}

func (this *Boilerpiper) NumberWordsRulesFilter() {
	for idx, current := range this.content {
		var (
			prev *boilerpipe_score = &boilerpipe_score{}
			next *boilerpipe_score = &boilerpipe_score{}
		)
		if idx != 0 {
			prev = this.content[idx-1]
		}
		if idx < len(this.content)-1 {
			next = this.content[idx+1]
		}
		this.classify(prev, current, next)
	}
	for _, p := range this.content {
		if !p.is_content {
			p.element.Parent.RemoveChild(p.element)
		}
	}
}

// 清除表单前的提示行
func (this *Boilerpiper) FormPrefixFilter() {
	for idx, current := range this.content {
		var next = &boilerpipe_score{}
		if idx < len(this.content)-1 {
			next = this.content[idx+1]
		}
		if current.is_content && next.forms > 0 && current.words < 16 {
			current.is_content = false
			current.element.Parent.RemoveChild(current.element)
		}
	}

}

func new_boilerpiper() *Boilerpiper {
	rtn := Boilerpiper{titles: []string{},
		authors:  []string{},
		keywords: []string{},
		content:  []*boilerpipe_score{}}
	return &rtn
}

func (this *Boilerpiper) classify(prev *boilerpipe_score,
	current *boilerpipe_score,
	next *boilerpipe_score) {
	if current.link_density() > 33 {
		current.is_content = false
	} else {
		c := (prev.link_density() <= 55 &&
			(current.words > 20 || next.words > 15 || prev.words > 8)) ||
			(prev.link_density() > 55 && (current.words > 40 || next.words > 17))
		current.is_content = current.is_content || c
	}
	if current.words < 8 && next.link_density() > 55 {
		current.is_content = false
	}
	if current.forms > 0 && current.words == 0 {
		current.is_content = false
	}
	if current.link_density() > 25 {
		log.Printf("is_content: %v, words:%v,line:%v, isfomr:%v, lnk_dsnty: %v, anchor: %v, %v\n",
			current.is_content,
			current.words,
			current.lines(), current.forms, current.link_density(), current.anchors, strings.TrimSpace(current.inner_text))
	}
}
