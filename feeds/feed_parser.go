package feeds

import (
  . "gextract/feeds/meta"
	"gextract/feeds/atom"
	"gextract/feeds/rss"
)


func NewFeed(filepath string, uri string) (*Feed, []Entry, error) {
	f, entries, err := rss.NewRss2(filepath, uri)
	if err != nil {
		f, entries, err = atom.NewAtom(filepath, uri)
	}
	return f, entries, err
}
