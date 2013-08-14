package document

import (
	"code.google.com/p/go.net/html"
	"log"
)

const (
	max_words_per_line = 60 // 每行最大中文字符数，每个英文单词算等同一个中文字符
)

type block struct {
	element       *html.Node
	chars         int // alpha, digits, punc, zh-char
	words         int // word or zh-char
	tokens        int // number, punc, zh-char or word
	lines         int
	anchor_words  int
	words_wrapped int
	text_density  float64
	link_density  float64

	text string

	is_content bool
	has_image  bool
	has_object bool // object, embed, video, audio
	is_form    bool
}

type Boilerpiper struct {
	titles      []string
	description string
	authors     []string
	keywords    []string

	content []*block

	words       int
	lines       int
	chars       int
	inner_chars int
	outer_chars int
	parags      int

	words_p_block float64
	quality       float64

	body *html.Node
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
			prev *block = &block{}
			next *block = &block{}
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
		var next = &block{}
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
		content:  []*block{}}
	return &rtn
}

func (this *block) make_end() {
	this.lines = (this.words + max_words_per_line - 1) / max_words_per_line
	this.words_wrapped = int(this.words/max_words_per_line) * max_words_per_line
	if this.words > 0 {
		this.text_density = float64(this.words) / float64(this.lines)
		this.link_density = float64(this.anchor_words) / float64(this.words)
	}
}

func make_paragraph(p *html.Node) *block {
	para := &block{element: p}
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
	if p.Type == html.ElementNode && (p.Data == "form" || p.Data == "input" || p.Data == "textarea") {
		para.is_form = true
	}
	para.make_end()
	return para
}

func (this *Boilerpiper) classify(prev *block, current *block, next *block) {
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
