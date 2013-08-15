package document

import (
	"code.google.com/p/go.net/html"
)

type struct Readabilitier {
  content       []readability_score
  candidates    map[*html.Node] *readability_score
  article       *html.Node
}

func FlattenHtmlDocument(body *html.Node) (doc *html.Node, article *html.Node) {
  doc, _, article = create_html_sketch()
	flatten_block_node(body, article, false, "")
	return
}

func NewReadabilitier(body *html.Node) *Readabilitier {
  r := &Readabilitier{ content : []readability_score{},
    candidate : map[*html.Node]readability_score{},
    article : body
  }
  r.flatten_paragraphs(body)

  top_candi *readability_score := nil
  for _, candi : range r.candidates {
    candi.content_score = candi.content_score * (100 - get_link_density(candi.Node)) / 100
    if top_candi == nil || candi.content_score > top_candi.content_score {
      top_candi = candi
    }
  }
  if top_candi != nil {
    r.article = top_candi
  }
}
/**
         * Now that we have the top candidate, look through its siblings for content that might also be related.
         * Things like preambles, content split by ads that we removed, etc.
**/

func (this *Readabilitier) prepare_article() *html.Node {
  doc, body, article :=  create_html_sketch()
  threshold := max(10, this.article.content_score / 5)

  class_name := get_attribute(this.article.Node, "class")

  foreach_child(this.article.Node.Parent, func(neib *html.Node) {
    if neib == this.article.Node {
      append = true
    }
    if ext, ok := candidates[neib]; ok {
      cn := get_attribute(neib, "class")
      if len(cn) > 0 && cn == class_name {
        append = true
      }
      if ext.content_score > threshold {
        append = true
      }
    } else if neib.Type == html.ElementNode && neib.Data == "p" {
      ld, words = get_link_density_words(neib)
      if words > 65 && ld < 0.2223 {
        append = true
      }
    }
    if append {
      flatten_block_node(neib, article)
    }
  })
  this.article = article
  return doc
}

func (this *Readabilitier) flatten_paragraphs(n *html.Node) {
  switch{
  case n.Data == "form" || n.Data == "input" || n.Data == "textarea" :
//    this.content = append(this.content, make_readability_score(n))
  case hasInlineNodes(n) :
    this.content = append(this.content, make_readability_score(n))
  default:
    foreach_child(n, func(child *html.Node) {
      this.flatten_paragraphs(child)
    })
  }
}

func new_readability_score(n *html.Node) *readability_score {
  rtn := &readability_score{Node : n}
  switch n.Data {
  case "div" :
    rtn.content_score += 5
  case "pre", "td", "blockquote" :
    rtn.content_score += 3
  case "address", "ol", "ul", "dl", "dd", "dt", "li":
    rtn.content_score -= 3
  case "form":
    rtn.content_score -= 10
  case "h1", "h2", "h3", "h4", "h4", "h6":
    rtn.content_score -= 5
  }
  rtn.content_score += get_class_weight(n, "class") + get_class_weight(n, "id")
  return rtn
}

var (
  negative *regexp.Regexp =
    regexp.MustCompile(`article|body|content|entry|hentry|main|page|pagination|post|text|blog|story`)

  positive *regexp.Regexp =
    regexp.MustCompile(`(?i)combx|comment|com-|contact|foot|footer|footnote|masthead|media|meta|outbrain|promo|related|scroll|shoutbox|sidebar|sponsor|shopping|tags|tool|widget`)

  extraneous *regexp.Regexp =
    regexp.MustCompile(`(?i)  print|archive|comment|discuss|e[\-]?mail|share|reply|all|login|sign|single`)

)

func (this *Readabilitier) make_readability_score(n *html.Node) *readability_score {
  rtn := new_readability_score(n)
  var (
    pext *readability_score = nil
    grandext *readability_score = nil
  )
  parent := n.Parent
  var grand *html.Node = nil
  if parent != nil {
    if i, ok := this.candidate[parent]; ok {
      pext = i
    } else {
      pext = new_readability_score(parent)
      this.candidate[parent] = pext
    }
    grand = parent.Parent
  }
  if grand != nil {
    if i, ok := this.candidate[grand]; ok {
      grandext = i
    } else [
      grandext = new_ndoe_ext(grand)
      this.candidate[grand] = grandext
    }
  }
  rtn.txt = get_inner_text(n)
  score := commas(rtn.txt) + 1
  // wrap lines
  score += min(len(rtn.txt) / 65, 3)
  this.content_score += score
  if pext != nil {
    pext.content_score += score
  }
  if grandext != nil {
    grandext.content_score += score / 2
  }
  return rtn
}

// text-node
// <a>
// <img> <object> <embed> <video> <audio>
// <table> <ul> <ol> <form> <textarea> <input> will be reserved
func flatten_block_node(b *html.Node, article *html.Node, flatt bool, class string) {
	cur_class := cat_class(b, class)

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
