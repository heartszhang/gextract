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
	quality                  float64

	//经过初次处理后得到的html/body节点，或者更准确的article节点
	body *html.Node
}

func NewBoilerpiper(article *html.Node) *Boilerpiper {
	rtn := new_boilerpiper()
	rtn.evaluate_score(article)
	return rtn
}

func (this *Boilerpiper) evaluate_score(n *html.Node) {
	switch {
	case n.Data == "form" || n.Data == "input" || n.Data == "textarea":
		//form中的内容，仍然需要进行统计
		bs := new_boilerpipe_score(n)
		this.content = append(this.content, &bs)
		//经过前面的整理，如果节点包含inline-nodes，则所有子节点必然都是inline-nodes
	case hasInlineNodes(n):
		bs := new_boilerpipe_score(n)
		this.content = append(this.content, &bs)
	default:
		foreach_child(n, func(child *html.Node) {
			this.evaluate_score(child)
		})
	}
}

//http://www.l3s.de/~kohlschuetter/boilerplate/
//implement
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
	//表单前面的一段内容如果很短，基本上可以认定是form或者menu的标题，可以进行清除
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

const (
	ld_link_para_t = 13
	ld_link_group_t       = 33
	ld_link_group_title_t = 55
	w_current_line_l      = 20
	w_next_line_l         = 15
	w_prev_line_l         = 8
)

//链接密度高于0.33的段落，直接认为不是正文
//当前段不多于一定字符，后续段落的链接密度很高，认为这一段落是后续段落的标题，可以进行清除
//form组成的段落，直接抛弃
// many magic numbers in this function
func (this *Boilerpiper) classify(prev *boilerpipe_score,
	current *boilerpipe_score,
	next *boilerpipe_score) {
	if current.link_density() > ld_link_group_t {
		current.is_content = false
	} else {
		c := (prev.link_density() <= ld_link_group_title_t &&
			(current.words > w_current_line_l || next.words > w_next_line_l || prev.words > w_prev_line_l)) ||
			(prev.link_density() > ld_link_group_title_t && (current.words > 40 || next.words > 17))
		current.is_content = current.is_content || c
		//images between content paragraphs
		if prev.link_density() <= ld_link_group_t && next.link_density() <= ld_link_group_t &&
		current.words == 0 && current.imgs > 0 && current.anchor_imgs == 0 {
			current.is_content = true
		}
		// short paragraphs
		if prev.link_density() < ld_link_para_t && next.link_density() < ld_link_para_t && 
		current.link_density() < ld_link_para_t && current.words < 40 {
			current.is_content = true
		}
	}
	if current.words < w_prev_line_l && next.link_density() > ld_link_group_title_t {
		current.is_content = false
	}
	if current.forms > 0 && current.words == 0 {
		current.is_content = false
	}
	
	log.Println("is-content:", current.is_content,
		"words:", current.words,
		"imgs:", current.imgs,
		"lines:", current.lines(),
		"density:", current.link_density(),
		"links:", current.anchors,
		"aimgs:", current.anchor_imgs)
}
