package document

import (
	"code.google.com/p/go.net/html"
	"code.google.com/p/go.net/html/atom"
	"log"
	"strings"
	"unicode"
  "regexp"
)

type HtmlCleaner struct {
	may_be_html5 bool
	Article      *html.Node // body or article
	head         *html.Node
	header1s     []*html.Node
	header2s     []*html.Node
	header3s     []*html.Node
	header4s     []*html.Node
	uls          []*html.Node
	ols          []*html.Node
	forms        []*html.Node

	tables      []*html.Node
	pages       []string
	titles      []string
	keywords    []string
	author      []string
	description string
}

func NewHtmlCleaner() *HtmlCleaner {
	return &HtmlCleaner{}
}

func (cleaner *HtmlCleaner) grab_keywords(meta *html.Node) {
}

func (cleaner *HtmlCleaner) grab_description(meta *html.Node) {
}

func (cleaner *HtmlCleaner) grab_title(title *html.Node) {

}

var (
  unlikely *regexp.Regexp = regexp.MustCompile(`combx|comment|community|disqus|extra|foot|header|menu|remark|rss|shoutbox|sidebar|sponsor|ad-break|agegate|pagination|pager|popup|tweet|twitter`)
)
//CleanHtml 清洗掉所有的link/style/css
// 删除/html/head
// 转换所有的tag为小写字母
// 找到body/article节点
// 找到h1节点或者h2节点，根据数目设置body
func (cleaner *HtmlCleaner) CleanHtml(root *html.Node) {
	var (
		dropping []*html.Node = []*html.Node{}
	)
	cleaner.clean_unprintable_element(&dropping, root)

	for _, drop := range dropping {
		p := drop.Parent
		p.RemoveChild(drop)
	}

	if cleaner.head != nil {
		cleaner.head.Parent.RemoveChild(cleaner.head)
	}
	//文档中如果只有一个h1,通常这个h1所在的div就是文档内容
	if len(cleaner.header1s) == 1 { // only one h1
		cleaner.Article = find_article_via_header(cleaner.header1s[0], "h1")
		clean_element_before_header(cleaner.Article, "h1")
	}
	//如果文档中只有一个h2，这时又没有h1，h2就是其中的标题，所在的div就是文档内容
	if len(cleaner.header1s) == 0 && len(cleaner.header2s) == 1 {
		cleaner.Article = find_article_via_header(cleaner.header2s[0], "h2")
		clean_element_before_header(cleaner.Article, "h2")
	}

	if cleaner.Article == nil {
		cleaner.Article = &html.Node{Type: html.ElementNode,
			DataAtom: atom.Lookup([]byte("body")),
			Data:     "body"}
		root.AppendChild(cleaner.Article)
	}

	cleaner.clean_body()
	log.Println("begin cleanning empty nodes")
	cleaner.clean_empty_nodes(cleaner.Article)
	cleaner.clean_attributes(cleaner.Article)
}

func (cleaner *HtmlCleaner) clean_unprintable_element(dropping *[]*html.Node, n *html.Node) {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.CommentNode {
			*dropping = append(*dropping, child)
		} else if child.Type == html.ElementNode {
			drop := false
			child.Data = strings.ToLower(child.Data)
      idc := get_attribute(child, "class") + get_attribute(child, "id")
/*
      if regexp.MatchString(idc) {
        drop = true
        *dropping = append(*dropping, child)
      } else {
        switch child.Data {
        case "script", "link", "iframe", "nav", "aside", "noscript" :
          *dropping = append(*dropping, child)
          drop = true
        }
      }
      */
			if child.Data == "script" ||
				child.Data == "link" ||
				child.Data == "style" ||
				child.Data == "iframe" ||
				child.Data == "nav" ||
				child.Data == "aside" ||
				child.Data == "noscript" {
				*dropping = append(*dropping, child)
				drop = true
			} else if child.Data == "meta" {
				cleaner.grab_keywords(child)
				cleaner.grab_description(child)
			} else if child.Data == "title" {
				cleaner.grab_title(child)
			} else if child.Data == "head" {
				cleaner.head = child
			} else if child.Data == "body" {
				cleaner.Article = child
			} else if child.Data == "br" {
				child.Data = "p"
				child.DataAtom = atom.Lookup([]byte(child.Data))
			} else if child.Data == "article" {
				if cleaner.Article == nil || cleaner.Article.Data == "body" {
					cleaner.Article = child
				} else {
					pl := len(get_inner_text(cleaner.Article))
					cl := len(get_inner_text(child))
					if cl > pl {
						cleaner.Article = child
					}
				}

			} else if child.Data == "ul" {
				/*				if is_menu(child) {
									*dropping = append(*dropping, child)
									drop = true
								}
				*/
			} else if child.Data == "h1" {
				cleaner.header1s = append(cleaner.header1s, child)
			} else if child.Data == "h2" {
				cleaner.header2s = append(cleaner.header2s, child)
			} else if child.Data == "h3" {
				cleaner.header3s = append(cleaner.header3s, child)
			} else if child.Data == "h4" {
				cleaner.header4s = append(cleaner.header4s, child)
			} else if child.Data == "form" {
				cleaner.forms = append(cleaner.forms, child)
			} else if child.Data == "ul" {
				cleaner.uls = append(cleaner.uls, child)
			} else if child.Data == "ol" {
				cleaner.ols = append(cleaner.ols, child)
			} else if child.Data == "table" {
				cleaner.tables = append(cleaner.tables, child)
			} else if child.Data == "option" {
				child.Data = "a"
			} else {
				/* 有些菜单使用了这个属性，如果直接去除，菜单头会被保留下来*/
				st := get_attribute(child, "style")
				if strings.Contains(st, "display") && (strings.Contains(st, "none")) {
					//*dropping = append(*dropping, child)
					//drop = true
					log.Println(child)
					child.Data = "form"
				}
			}

			if !drop {
				cleaner.clean_unprintable_element(dropping, child)
			}
		} else if child.Type == html.TextNode {
			child.Data = merge_tail_spaces(child.Data)
		}
	}

	return
}

// reserve id, class, href, src
func (this *HtmlCleaner) clean_attributes(n *html.Node) {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		this.clean_attributes(child)
	}
	attrs := []html.Attribute{}
	for _, attr := range n.Attr {
		if attr.Key == "id" || attr.Key == "class" || attr.Key == "href" || attr.Key == "src" {
			attrs = append(attrs, attr)
		}
	}
	if len(attrs) != len(n.Attr) {
		n.Attr = attrs
	}
}

// clean-body wraps text-node with p
func (this *HtmlCleaner) clean_body() {
	log.Println("beginning cleanning body", this.Article.Data)
	this.clean_block_node(this.Article)
}

//整理html文档，将block-level/inline-level混合的节点改成只有block-level的节点
//对已只有inline-level的节点，删除行前后的空白符
//将包含inline-level的节点展开成更为简单的形式，去掉想<font><span><strong>等等格式节点
func (this *HtmlCleaner) clean_block_node(n *html.Node) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()
	blks := hasBlockNodes(n)
	inlines := hasInlineNodes(n)

	// has bocks and inlines
	if blks && inlines {
		child := n.FirstChild
		for child != nil {
			if isInlineNode(child) {
				p := child.PrevSibling
				if p == nil || p.Data != "p" {
					p = NewHtmlElement("p")
					n.InsertBefore(p, child)
				}
				n.RemoveChild(child)
				p.AppendChild(child)
				child = p.NextSibling
			} else {
				child = child.NextSibling
			}
		}
		inlines = false
	}

	// only inlines
	if blks == false && inlines {
		this.clean_inline_node(n)
		this.trim_empty_spaces(n)
	}

	// only blocks
	if blks && !inlines {
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			this.clean_block_node(child)
		}
	}
}

// flatten inlines text image a object video audio seq
// n is element-node
// inline node may have div element
func (this *HtmlCleaner) clean_inline_node(n *html.Node) {
	inlines := this.flatten_inline_node(n)

	for child := n.FirstChild; child != nil; child = n.FirstChild {
		n.RemoveChild(child)
	}
	for _, inline := range inlines {
		p := inline.Parent
		if p != nil {
			p.RemoveChild(inline) //			this.article.RemoveChild(child)

		}
		n.AppendChild(inline)
	}
}

//img video audio object embed保留原内容
//text-node保持原内容
//如果inline-level节点包含table/div/ul/ol等等block-level的节点，将这些节点保留
//其他inline-level的节点都直接使用text-node代替
func (this *HtmlCleaner) flatten_inline_node(n *html.Node) []*html.Node {
	inlines := []*html.Node{}
	for i := n.FirstChild; i != nil; i = i.NextSibling {
		if i.Type == html.TextNode {
			inlines = append(inlines, i)
		} else if i.Data == "img" || i.Data == "object" || i.Data == "video" || i.Data == "audio" {
			inlines = append(inlines, i)

			// may be div
		} else if isBlockNode(i) {
			this.clean_inline_node(i)
			inlines = append(inlines, i)
		} else if i.Type == html.ElementNode && is_a_has_valid_href(i) {
			this.clean_inline_node(i)
			inlines = append(inlines, i)
		} else if i.Type == html.ElementNode {
			x := this.flatten_inline_node(i)
			t := make([]*html.Node, len(inlines)+len(x))
			copy(t, inlines)
			copy(t[len(inlines):], x)
			inlines = t
		}
	}
	return inlines
}

//节点中没有可显示内容，也没有form等等后续需要处理的节点就是空节点
func (this *HtmlCleaner) clean_empty_nodes(n *html.Node) {
	child := n.FirstChild
	for child != nil {
		next := child.NextSibling
		this.clean_empty_nodes(child)
		child = next
	}

	if isEmptyNode(n) {
		parent := n.Parent
		parent.RemoveChild(n)
	}
}

//删除行前后空白
func (this *HtmlCleaner) trim_empty_spaces_func(n *html.Node, trim func(string) string) {
	child := n.FirstChild
	for child != nil {
		if child.Type == html.TextNode {
			child.Data = trim(child.Data)
		} else {
			this.trim_empty_spaces_func(child, trim)
		}
		if !isEmptyNode(child) {
			break
		}
		next := child.NextSibling
		n.RemoveChild(child)
		child = next
	}
}

func (this *HtmlCleaner) trim_empty_spaces(n *html.Node) {
	this.trim_empty_spaces_func(n, func(o string) string {
		return strings.TrimLeftFunc(o, unicode.IsSpace)
	})

	this.trim_empty_spaces_func(n, func(o string) string {
		return strings.TrimRightFunc(o, unicode.IsSpace)
	})

}
