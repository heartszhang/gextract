package document

import (
	"code.google.com/p/go.net/html"
	"regexp"
	"strings"
	"unicode"
)

func is_not_empty(n *html.Node) bool {
	switch n.Type {
	case html.TextNode:
		return len(n.Data) >= 0
	case html.ElementNode:
		switch n.Data {
		case "video", "audio", "object", "embed", "img", "a":
			return true
		case "form", "input", "textarea":
			return true
		default:
			rtn := false
			foreach_child(n, func(child *html.Node) {
				cne := is_not_empty(child)
				rtn = cne || rtn
			})
			return rtn
		}
	default:
		return false
	}
}

func hasBlockNodes(n *html.Node) bool {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		v := isBlockNode(child)
		if v {
			return true
		}
	}
	return false
}

func hasInlineNodes(n *html.Node) bool {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		v := isInlineNode(child)
		if v {
			return true
		}
	}
	return false
}

func isInlineNode(n *html.Node) bool {
	rtn := n.Type == html.TextNode ||
		is_inline_element(n) ||
		is_object(n)
	return rtn
}

func is_inline_element(n *html.Node) bool {
	switch n.Data {
	case "a", "font", "small", "span", "strong", "em", "dt", "dd", "br", "cite":
		return true
	case "li":
		return check_li_inline_mode(n)
	default:
		return false
	}
}
func is_object(n *html.Node) bool {
	switch n.Data {
	default:
		return false
	case "img", "embed", "object", "video", "audio":
		return true
	}
}

// ignorable: form
// block-level: div, p, h1-h6, body, html, object, embed, table, ol, ul, dl, video
// inline-level: a, span, strong, br, img, small, font, i
func isBlockNode(n *html.Node) bool {
	if n.Type != html.ElementNode && n.Type != html.DocumentNode {
		return false
	}

	switch n.Data {
	case "div", "p", "pre", "h1", "h2", "h3", "h4", "h5", "h6",
		"body", "html", "article", "section", "head", "ol", "ul", "dl",
		"tbody", "td", "tr", "table", "form", "textarea", "input":
		return true
	case "li":
		return !check_li_inline_mode(n)
	default:
		return false
	}

}

// number
// word
// zh-char
// punc
func tokenize(t string) []string {
	re := regexp.MustCompile(`\w+|\d+|[\W\D\S]`)
	rtn := re.FindAllString(t, -1)
	//	log.Println(t, rtn)
	return rtn
}

const (
	zh_stop_chars string = "。．？！，、；：“ ”﹃﹄‘ ’﹁﹂（）［］〔〕【】—…—-～·《》〈〉﹏＿."
)

func is_word(t string) bool {
	rs := []rune(t)
	if len(rs) == 0 {
		return false
	}
	return ((rs[0] > unicode.MaxLatin1 && strings.ContainsRune(zh_stop_chars, rs[0])) == false) || unicode.IsLetter(rs[0])
}

// n.Data has been lowered
/*
func is_anchor(n *html.Node) bool {
	return (n.Type == html.ElementNode && n.Data == "a")
}
*/

func has_children(this *html.Node) bool {
	return this.FirstChild != nil
}

func get_attribute(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if a.Key == name {
			return a.Val
		}
	}
	return ""
}

var (
	continue_spaces = regexp.MustCompile("[ \t]+$")
	lb_spaces       = regexp.MustCompile("[ \t]*[\r\n]+[ \t]*")
)

func merge_tail_spaces(txt string) string {
	txt = continue_spaces.ReplaceAllString(txt, "")
	txt = lb_spaces.ReplaceAllString(txt, "\n")
	return txt
}

func get_inner_text(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}

	// all comments has been removed
	rtn := ""
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		rtn += get_inner_text(child)
	}
	return rtn
}

func GetInnerText(n *html.Node) string {
	return get_inner_text(n)
}

func for_each_child(n *html.Node, f func(*html.Node)) {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		f(child)
	}
}

func count_words(txt string) (tokens int, words int, commas int) {

	for _, c := range txt {
		if unicode.IsPunct(c) {
			commas++
		}
	}

	tkns := tokenize(txt)
	tokens = len(tkns)
	for _, token := range tkns {
		if is_word(token) {
			words++
		}
	}
	return
}

func create_element(name string) (node *html.Node) {
	return &html.Node{Type: html.ElementNode, Data: name}
}

func create_text(txt string) (node *html.Node) {
	return &html.Node{Type: html.TextNode, Data: txt}
}

// 需要换行的li都认为不再是inline模式。这个函数主要使用来检查使用Li构造的menu
func check_li_inline_mode(li *html.Node) bool {
	if li.Parent == nil {
		return false
	}
	lis := 0
	txt := ""
	foreach_child(li.Parent, func(n *html.Node) {
		txt += get_inner_text(n)
		lis++
	})
	return len(txt) < 60
}

/*
func is_menu(ul *html.Node) bool {
	lis := get_element_by_tag_name(ul, "li")
	as := get_element_by_tag_name(ul, "a")

	// like 频道: <a href=xxxx>商业</a>
	rtn := len(lis) > 0 && len(as)*10/len(lis) >= 5
	return rtn
}
*/

func get_element_by_tag_name2(n *html.Node, tag string, set []*html.Node) []*html.Node {
	foreach_child(n, func(child *html.Node) {
		if child.Type == html.ElementNode && child.Data == tag {
			set = append(set, child)
		} else if child.Type == html.ElementNode {
			set = get_element_by_tag_name2(child, tag, set)
		}
	})
	return set
}

func get_element_by_tag_name(n *html.Node, tag string) []*html.Node {
	return get_element_by_tag_name2(n, tag, []*html.Node{})
}

func clean_element_before_header(body *html.Node, name string) {
	child := body.FirstChild
	for child != nil {
		if child.Type == html.ElementNode && child.Data != name {
			next := child.NextSibling
			body.RemoveChild(child)
			child = next
		} else {
			break
		}
	}
}

func find_article_via_header_i(h *html.Node) *html.Node {
	parent := h.Parent
	pcl := 0
	if parent != nil {
		pcl = len(get_inner_text(parent))
	} else {
		return nil
	}
	// 内容超过3行才行，每行大概又65个字符
	if pcl > 195 {
		return parent
	}
	return find_article_via_header_i(parent)
}

func is_unflatten_node(b *html.Node) bool {
	return b.Data == "form" || b.Data == "textarea" || b.Data == "input"
}

func clone_element_deep(n *html.Node) (inline *html.Node) {
	inline = clone_element(n)
	foreach_child(n, func(child *html.Node) {
		i := clone_element_deep(child)
		inline.AppendChild(i)
	})
	return
}

func clone_element(n *html.Node) (inline *html.Node) {
	inline = &html.Node{Type: n.Type, Data: n.Data}
	inline.Attr = make([]html.Attribute, len(n.Attr))
	copy(inline.Attr, n.Attr)
	return
}

func foreach_child(n *html.Node, dof func(*html.Node)) {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		dof(child)
	}
}

func append_children(src *html.Node, target *html.Node) {
	foreach_child(src, func(child *html.Node) {
		switch {
		case child.Type == html.TextNode:
			target.AppendChild(create_text(child.Data))
		case child.Data == "a" || child.Data == "img" || is_object(child):
			// ommit all children elements
			a := clone_element(child)
			append_children(child, a)
			target.AppendChild(a)
		default:
			append_children(child, target)
		}
	})
}

func create_p(n *html.Node) (p *html.Node) {
	p = create_element("p")
	append_children(n, p)
	return
}

func try_update_class_attr(b *html.Node, class string) {
	if len(class) > 0 {
		ca := make([]html.Attribute, len(b.Attr)+1)
		copy(ca, b.Attr)
		ca[len(b.Attr)] = html.Attribute{Key: "class", Val: class}
		b.Attr = ca
	}
}

func cat_class(b *html.Node, class string) (rtn string) {
	c := get_attribute(b, "class")
	id := get_attribute(b, "id")
	rtn = class
	if len(c) > 0 {
		rtn = class + "/" + c
	}
	if len(id) > 0 {
		rtn = class + "#" + id
	}
	return
}
func create_html_sketch() (doc *html.Node, body *html.Node, article *html.Node) {
	doc = &html.Node{Type: html.DocumentNode}
	dt := &html.Node{Type: html.DoctypeNode, Data: "html"}
	root := create_element("html")
	body = create_element("body")
	article = create_element("article")
	doc.AppendChild(dt)
	doc.AppendChild(root)
	root.AppendChild(body)
	body.AppendChild(article)
	return
}

func remove_decentant(n *html.Node, tag string) {
	child := n.FirstChild
	for child != nil {
		if child.Type == html.ElementNode && child.Data == tag {
			next := child.NextSibling
			n.RemoveChild(child)
			child = next
		} else {
			remove_decentant(child, tag)
			child = child.NextSibling
		}
	}
}

func is_ownered_by_a(a *html.Node) bool {
	for p := a.Parent; p != nil; p = p.Parent {
		if p.Type == html.ElementNode && p.Data == "a" {
			return true
		}
	}
	return false
}

func max(l int, r int) int {
	if l > r {
		return l
	}
	return r
}
func min(l int, r int) int {
	if l < r {
		return l
	}
	return r
}

func get_classid(n *html.Node) string {
	return get_attribute(n, "class") + ":" + get_attribute(n, "id")
}

func update_attribute(n *html.Node, key string, val string) {
	for idx, attr := range n.Attr {
		if attr.Key == key {
			n.Attr[idx].Val = val
		}
	}
}
