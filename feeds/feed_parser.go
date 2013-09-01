package feeds

import (
	"github.com/heartszhang/gextract/feeds/atom"
	. "github.com/heartszhang/gextract/feeds/meta"
	"github.com/heartszhang/gextract/feeds/rss"
)

func NewFeed(filepath string, uri string) (*Feed, []Entry, error) {
	f, entries, err := rss.NewRss2(filepath, uri)
	if err != nil {
		f, entries, err = atom.NewAtom(filepath, uri)
	}
	return f, entries, err
}
