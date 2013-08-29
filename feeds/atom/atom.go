package atom

import (
	"encoding/xml"
	"errors"
	"gextract/document"
	. "gextract/feeds/meta"
	"net/http"
	"os"
	"time"
)

func NewAtom(filepath string, uri string) (*Feed, []Entry, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	var v feed
	decoder := xml.NewDecoder(f)
	//	decoder.CharsetReader = charset_reader

	err = decoder.Decode(&v)
	if err != nil {
		return nil, nil, err
	}
	if len(v.Link) == 0 {
		return nil, nil, errors.New("invalid atom file")
	}
	//v.Channel.preprocess(url)
	fd := v.to_feed(uri)

	entries := make([]Entry, len(v.Entry))
	for i, item := range v.Entry {
		entries[i] = *item.to_entry(fd.Link, fd.Category)
	}
	return fd, entries, nil
}

type feed struct {
	XMLName  xml.Name   `xml:"http://www.w3.org/2005/Atom feed"`
	Title    string     `xml:"title"`
	Id       string     `xml:"id"`
	Link     []link     `xml:"link"`
	Updated  time.Time  `xml:"updated,attr"`
	Author   person     `xml:"author"`
	Entry    []entry    `xml:"entry"`
	Category []category `xml:"category"`
}

func (this *feed) to_feed(uri string) *Feed {
	f := &Feed{Title: this.Title,
		Link:    this.link_one(uri),
		Refresh: time.Now(),
		TTL:     60, // minutes
	}
	f.Category = make([]string, len(this.Category))
	for i, c := range this.Category {
		f.Category[i] = c.Term
	}
	return f
}

type entry struct {
	Title   string `xml:"title"`
	Id      string `xml:"id"`
	Link    []link `xml:"link"`
	Updated string `xml:"updated"`
	Author  person `xml:"author,omitempty"`
	Summary text   `xml:"summary,omitempty"`
	Content text   `xml:"content,omitempty"`
}

func (this *entry) to_entry(fid string, category []string) *Entry {
	e := &Entry{
		Title:   this.Title,
		Updated: parse_time2(this.Updated),
		Author:  this.Author.Name,
		Guid:    this.Id,
		Link:    this.link_one(),
		Created: time.Now(),
		Readed:  false,
		Feed:    fid,
	}
	if len(e.Guid) == 0 {
		e.Guid = e.Link
	}
	e.Category = category

	var sc *document.SummaryScore
	e.Summary, e.Status, sc = clean_summary(select_text(this.Summary.Body, this.Content.Body), e.Link)
	e.Statis.Words = sc.WordCount
	e.Statis.Imgs = len(sc.Images)
	e.Image = sc.Images

	return e
}

func (this *feed) link_one(uri string) string {
	for _, l := range this.Link {
		if l.Rel == "self" {
			return l.Href
		}
	}
	return uri
}

func (this *entry) link_one() string {
	var self, alt string
	for _, l := range this.Link {
		switch l.Rel {
		case "self":
			self = l.Href
		case "alternate":
			alt = l.Href
		}
	}
	switch len(self) > 0 {
	case true:
		return self
	default:
		return alt
	}
}
func parse_time2(txt string) time.Time {
	t, err := http.ParseTime(txt)
	if err == nil {
		t = time.Now()
	}
	return t
}

type link struct {
	Rel  string `xml:"rel,attr,omitempty"`
	Href string `xml:"href,attr,omitempty"`
}

type person struct {
	Name  string `xml:"name,omitempty"`
	URI   string `xml:"uri,omitempty"`
	Email string `xml:"email,omitempty"`
	//	InnerXML string `xml:",innerxml"`
}

type text struct {
	Type string `xml:"type,attr,omitempty"`
	Body string `xml:",chardata"` // omitempty cannot be used
}

type category struct {
	Term string `xml:"term,attr,omitempty"`
}

func select_text(a, b string) string {
	switch len(a) > len(b) {
	case true:
		return a
	default:
		return b
	}
}

func clean_summary(orig, baseuri string) (summary string, status int, score *document.SummaryScore) {
	summary, score = document.CleanFragment(orig, baseuri)

	status = ContentStatusUnresolved
	if score.WordCount+len(score.Images)*128 > 128 {
		status = ContentStatusSummary
	}
	return
}
