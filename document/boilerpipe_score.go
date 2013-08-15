package document
import (
	"code.google.com/p/go.net/html"
)

type boilerpipe_score struct {
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
