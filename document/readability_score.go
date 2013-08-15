package document

import (
	"code.google.com/p/go.net/html"
)

type struct readability_score {
  *html.Node
  content_score int
  link_density  int     // percent
  words         int
  lines         int
  commas        int
  chars         int
  tokens        int

  images        int
  anchors       int
  anchor_images int
}
