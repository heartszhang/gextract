package document

import (
	"code.google.com/p/go.net/html"
	"log"
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
	quality       float64

	body *html.Node
}

func NewBoilerpiper2(article *html.Node) *Boilerpiper {
  rtn := new_boilerpiper()
  flatten_paragraphs(article)
  return rtn
}

func (this *Boilerpiper) flatten_paragraphs(n *html.Node) {
  switch{
  case n.Data == "form" || n.Data == "input" || n.Data == "textarea" :
    this.content = append(this.content, make_readability_score_form(n))
  case hasInlineNodes(n) :
    this.content = append(this.content, make_readability_score(n))
  default:
    foreach_child(n, func(child *html.Node) {
      this.flatten_paragraphs(child)
    })
  }
}

func NewBoilerpiper(article *html.Node) *Boilerpiper {
	rtn := new_boilerpiper()

	foreach_child(article, func(child *html.Node) {
		rtn.content = append(rtn.content, make_paragraph(child))
	})
	return rtn
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
			log.Printf("remove %v, %v\n", p.link_density, p.text)
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
		if current.is_content && next.is_form && current.words < 16 {
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

func (this *boilerpipe_score) make_end() {
	this.lines = (this.words + max_words_per_line - 1) / max_words_per_line
	this.words_wrapped = int(this.words/max_words_per_line) * max_words_per_line
	if this.words > 0 {
		this.text_density = float64(this.words) / float64(this.lines)
		this.link_density = float64(this.anchor_words) / float64(this.words)
	}
}

func make_boilerpipe_score(p *html.Node) *boilerpipe_score {
	para := &boilerpipe_score{element: p}

	foreach_child(p, func(child *html.Node) {
		txt := get_inner_text(child)
		tks, wds := count_words(txt)
		_, w := count_words(anchor_text(child))
		para.chars += len(txt)
		para.words += wds
		para.tokens += tks
		para.text += txt
		para.anchor_words += w
		has_image := has_decendant_type(child, "img")
		para.has_object = has_decendant_object(child)
		if has_image {
			para.has_image = true
		}
		if child.Type == html.ElementNode && child.Data == "a" && has_image {
			para.words += 4
			para.anchor_words += 4
		}
	})
	if p.Type == html.ElementNode &&
    (p.Data == "form" || p.Data == "input" || p.Data == "textarea") {
		para.is_form = true
	}
	para.make_end()
	return para
}

func (this *Boilerpiper) classify(prev *boilerpipe_score, current *boilerpipe_score, next *boilerpipe_score) {
	if current.link_density > 0.333333 {
		current.is_content = false
	} else {
		c := (prev.link_density <= 0.5555556 &&
			(current.words > 20 || next.words > 15 || prev.words > 8)) ||
			(prev.link_density > 0.5555556 && (current.words > 40 || next.words > 17))
		current.is_content = current.is_content || c
	}
	if current.words < 8 && next.link_density > 0.555556 {
		current.is_content = false
	}
	if current.is_form {
		current.is_content = false
	}
	log.Printf("is_content: %v, words:%v,line:%v, isfomr:%v, lnk_dsnty: %v, anchor: %v, %v\n",
		current.is_content,
		current.words,
		current.lines, current.is_form, current.link_density, current.anchor_words, current.text)
}

const (
	max_words_per_line = 60 // 每行最大中文字符数，每个英文单词算等同一个中文字符
)
