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
func is_anchor(n *html.Node) bool {
	return (n.Type == html.ElementNode && n.Data == "a")
}

func has_children(this *html.Node) bool {
	return this.FirstChild != nil
}

func print_line(n *html.Node) string {
	line := ""
	if n.Type == html.TextNode {
		return n.Data
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		line += print_line(c)
	}
	return line
}

func is_a_has_valid_href(a *html.Node) bool {
	isa := a.Type == html.ElementNode && a.Data == "a"
	//	href := get_attribute(a, "href")
	//	return isa && len(href) > 0 && !strings.Contains(href, "javascript:")
	return isa
}

func get_attribute(n *html.Node, name string) (rtn string) {
	for _, a := range n.Attr {
		if a.Key == name {
			rtn = a.Val
			return
		}
	}
	return
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
	rtn := ""
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		rtn += get_inner_text(child)
	}
	return rtn
}

func anchor_text(n *html.Node) (rtn string) {
	if n.Type == html.ElementNode && n.Data == "a" {
		rtn = get_inner_text(n)
		//		log.Println("anchor: ", rtn)
		return
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		rtn += " " + anchor_text(child)
	}
	return
}

func has_decendant_type(n *html.Node, name string) bool {
	if n.Type == html.ElementNode && n.Data == name {
		return true
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if has_decendant_type(child, name) {
			return true
		}
	}
	return false
}

func for_each_child(n *html.Node, f func(*html.Node)) {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		f(child)
	}
}

func count_words(txt string) (tokens int, words int, commas int) {
	tkns := tokenize(txt)
	tokens = len(tkns)
	for _, token := range tkns {
		if is_word(token) {
			words++
		}
	}
	return
}

func has_decendant_object(n *html.Node) bool {
	if n.Type == html.ElementNode && is_object(n) {
		return true
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if has_decendant_object(child) {
			return true
		}
	}
	return false
}

func create_element(name string) (node *html.Node) {
	return &html.Node{Type: html.ElementNode, Data: name}
}

func create_text(txt string) (node *html.Node) {
	return &html.Node{Type: html.TextNode, Data: txt}
}

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

func is_menu(ul *html.Node) bool {
	lis := get_element_by_tag_name(ul, "li")
	as := get_element_by_tag_name(ul, "a")

	// like 频道: <a href=xxxx>商业</a>
	rtn := len(lis) > 0 && len(as)*10/len(lis) >= 5
	return rtn
}

func get_element_by_tag_name(n *html.Node, tag string) []*html.Node {
	rtn := []*html.Node{}
	foreach_child(n, func(child *html.Node) {
		if child.Type == html.ElementNode && child.Data == tag {
			rtn = append(rtn, child)
		} else if child.Type == html.ElementNode {
			x := get_element_by_tag_name(child, tag)
			if len(x) > 0 {
				tmp := make([]*html.Node, len(x)+len(rtn))
				copy(tmp, rtn)
				copy(tmp[len(rtn):], x)
				rtn = tmp
			}
		}
	})
	return rtn
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

func find_article_via_header(h *html.Node, name string) *html.Node {
	return find_article_via_header_i(h, name, 0)
}

func find_article_via_header_i(h *html.Node, name string, cl int) *html.Node {
	parent := h.Parent
	if cl == 0 {
		cl = len(get_inner_text(h))
	}
	if cl == 0 {
		return nil
	}
	pcl := 0
	if parent != nil {
		pcl = len(get_inner_text(parent))
	}
	if pcl*10/cl > 55 {
		return parent
	}
	return find_article_via_header_i(parent, name, cl)
}
func is_unflatten_node(b *html.Node) bool {
	return b.Data == "form" || b.Data == "textarea" || b.Data == "input"
}

func clone_inline(n *html.Node) (inline *html.Node) {
	if n.Type == html.TextNode {
		inline = create_text(n.Data)
	} else {
		inline = create_element(n.Data)
	}
	inline.Attr = []html.Attribute{}
	for _, attr := range n.Attr {
		if attr.Key == "src" || attr.Key == "href" {
			inline.Attr = append(inline.Attr, attr)
		}
	}
	foreach_child(n, func(child *html.Node) {
		i := clone_inline(child)
		inline.AppendChild(i)
	})
	return
}

func foreach_child(n *html.Node, dof func(*html.Node)) {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		dof(child)
	}
}

func create_p(n *html.Node) (p *html.Node) {
	p = create_element("p")
	foreach_child(n, func(child *html.Node) {
		if child.Type == html.TextNode {
			p.AppendChild(create_text(child.Data))
		} else if child.Type == html.ElementNode {
			p.AppendChild(clone_inline(child))
		}
	})

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

/*
func get_link_density(n *html.Node) int{
  ld,_ := get_link_density_words(n)
  return ld
}

func get_link_density_words(n *html.Node) int,int {
  ll := 0
  wl := 0

  for _, t := range tokenize(get_inner_text(n)) {
    if is_word(t) {
      wl++
    }
  }
  if wl == 0 {
    return 0,0
  }
  for _, t : range tokenize(anchor_text(n)) {
    if is_word(t) {
      ll++
    }
  }
  return ll * 100 / wl,0
}
*/
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
