package document

import (
	"code.google.com/p/go.net/html"
)

func FlattenHtmlDocument(body *html.Node) (doc *html.Node, article *html.Node) {
	doc = &html.Node{Type: html.DocumentNode}
	dt := &html.Node{Type: html.DoctypeNode, Data: "html"}
	root := create_element("html")
	bd := create_element("body")
	article = create_element("article")
	doc.AppendChild(dt)
	doc.AppendChild(root)
	root.AppendChild(bd)
	bd.AppendChild(article)

	flatten_block_node(body, article, false, "")
	return
}

// text-node
// <a>
// <img> <object> <embed> <video> <audio>
// <table> <ul> <ol> <form> <textarea> <input> will be reserved
func flatten_block_node(b *html.Node, article *html.Node, flatt bool, class string) {
	cur_class := append_class(b, class)

	if flatt && is_unflatten_node(b) {
		nb := create_element(b.Data)
		try_update_class_attr(nb, cur_class)

		flatten_block_node(b, nb, false, class)
		article.AppendChild(nb)
	} else if hasInlineNodes(b) {
		p := create_p(b)
		try_update_class_attr(p, cur_class)
		article.AppendChild(p)
	} else {
		foreach_child(b, func(child *html.Node) {
			flatten_block_node(child, article, true, cur_class)
		})
	}
	/*
		if flatt && is_unflatten_node(b) {
			nb := create_element(b.Data)
			try_update_class_attr(nb, cur_class)

			flatten_block_node(b, nb, false, cur_class)
			article.AppendChild(nb)
		} else if hasInlineNodes(b) {
			create_inlines(b, cur_class, article)
		} else {
			foreach_child(b, func(child *html.Node) {
				flatten_block_node(child, article, true, cur_class)
			})
		}
	*/
}

func append_class(b *html.Node, class string) (rtn string) {
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

/*
func create_inlines(n *html.Node, class string, article *html.Node) {
	p := create_element("p")

	foreach_child(n, func(child *html.Node) {
		if child.Type == html.TextNode {
			p.AppendChild(create_text(child.Data))
		} else if child.Type == html.ElementNode && child.Data == "img" {
			if p.FirstChild != nil {
				article.AppendChild(p)
				p = create_element("p")
			}
			p.AppendChild(clone_inline(child))
			article.AppendChild(p)
			log.Println("creating img", p.FirstChild.Attr)
			p = create_element("p")

		} else if child.Type == html.ElementNode {
			p.AppendChild(clone_inline(child))
		}
	})
	if p.FirstChild != nil {
		article.AppendChild(p)
	}
}
*/
