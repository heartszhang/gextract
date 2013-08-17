package document

import (
	"code.google.com/p/go.net/html"
	"fmt"
	"regexp"
)

type readability_score struct {
	boilerpipe_score
	content_score int
}

func new_readability_score(n *html.Node) *readability_score {
	rtn := &readability_score{boilerpipe_score: new_boilerpipe_score(n)}
	switch n.Data {
	case "div", "p":
		rtn.content_score += 5
	case "pre", "td", "blockquote":
		rtn.content_score += 3
	case "address", "ol", "ul", "dl", "dd", "dt", "li":
		rtn.content_score -= 3
	case "form":
		rtn.content_score -= 10
	case "h1", "h2", "h3", "h4", "h5", "h6":
		rtn.content_score -= 5
	}
	rtn.content_score += get_class_weight(n, "class") + get_class_weight(n, "id")
	return rtn
}

var (
	negative *regexp.Regexp = regexp.MustCompile(`article|body|content|entry|hentry|main|page|pagination|post|text|blog|story`)

	positive *regexp.Regexp = regexp.MustCompile(`(?i)combx|comment|com-|contact|foot|footer|footnote|masthead|media|meta|outbrain|promo|related|scroll|shoutbox|sidebar|sponsor|shopping|tags|tool|widget`)

	extraneous *regexp.Regexp = regexp.MustCompile(`(?i)  print|archive|comment|discuss|e[\-]?mail|share|reply|all|login|sign|single`)
)

func (score *readability_score) String() string {
	return fmt.Sprintf("%v, score: %v, linkd: %v, words: %v, lines: %v, commas: %v, imgs: %v, a: %v, aimg: %v",
		score.element.Data, score.content_score,
		score.link_density(),
		score.words, score.lines(), score.commas,
		score.imgs, score.anchors, score.anchor_imgs)
}
