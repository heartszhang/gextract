package rss

import (
	"encoding/xml"
	"errors"
	iconv "github.com/djimenez/iconv-go"
	"github.com/heartszhang/gextract/document"
	. "github.com/heartszhang/gextract/feeds/meta"
	"io"
	"log"
	"os"
	"time"
)

//filepath : utf-8 encoded xml file, or gbk encoded xml file
func NewRss2(filepath string, uri string) (*Feed, []Entry, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	var v rss
	decoder := xml.NewDecoder(f)
	decoder.CharsetReader = charset_reader

	err = decoder.Decode(&v)
	if err != nil {
		return nil, nil, err
	}
	if len(v.Channel.Link) == 0 {
		return nil, nil, errors.New("invalid rss file")
	}
	//v.Channel.preprocess(url)
	feed := v.Channel.to_feed(uri)

	entries := make([]Entry, len(v.Channel.Items))
	for i, item := range v.Channel.Items {
		entries[i] = *item.to_entry(feed.Link, feed.Category)
	}
	return feed, entries, nil
}

func (this *item) to_entry(cid string, category []string) *Entry {
	entry := &Entry{Title: this.Title,
		Updated: parse_time1(this.PubDate),
		Author:  this.Author,
		Guid:    this.Guid,
		Link:    this.Link,
		Created: time.Now(),
		Readed:  false,
		Feed:    cid,
	}
	if len(entry.Guid) == 0 {
		entry.Guid = entry.Link
	}

	entry.Category = make([]string, len(this.Category)+len(category))
	copy(entry.Category, this.Category)
	copy(entry.Category[len(this.Category):], category)

	var sc *document.SummaryScore
	entry.Summary, entry.Status, sc = clean_summary(select_text(this.Summary, this.Content), this.Link)
	entry.Statis.Words = sc.WordCount
	entry.Statis.Imgs = len(sc.Images)
	entry.Image = sc.Images

	return entry
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

const (
	minutes_hour  = 60
	minutes_day   = minutes_hour * 24
	minutes_week  = minutes_day * 7
	minutes_month = minutes_day * 30
	minutes_year  = minutes_day * 365
)

func (this *channel) to_feed(uri string) *Feed {
	feed := &Feed{Title: this.Title,
		Link:     this.link_one(uri),
		TTL:      this.TTL,
		Category: this.Category,
		Refresh:  time.Now(),
	}

	switch this.UpdatePeriod {
	case "hourly":
		feed.TTL = minutes_hour
	case "daily":
		feed.TTL = minutes_day
	case "weekly":
		feed.TTL = minutes_week
	case "monthly":
		feed.TTL = minutes_month
	case "yearly":
		feed.TTL = minutes_year
	}

	if feed.TTL == 0 {
		feed.TTL = minutes_hour // minutes
	}
	//	feed.Refresh = time.Now()
	return feed
}

// private wrapper around the RssFeed which gives us the <rss>..</rss> xml
type rss struct {
	XMLName xml.Name `xml:"rss"`
	Channel channel  `xml:"channel" json:"channel,omitempty"`
}

type channel struct {
	Title         string `xml:"title"`             // required  unique?
	Link          []link `xml:"link"`              // required
	PubDate       string `xml:"pubDate,omitempty"` // created or updated
	LastBuildDate string `xml:"lastBuildDate"`

	TTL          int    `xml:"ttl"` // minitues
	UpdatePeriod string `xml:"http://www.w3.org/2005/Atom updatePeriod"`

	Category []string `xml:"category,omitempty"`
	Items    []item   `xml:"item"`
}

type item struct {
	Title    string   `xml:"title"`             // required
	Link     string   `xml:"link"`              // required, unique , used as primary key
	PubDate  string   `xml:"pubDate,omitempty"` // created or updated
	Category []string `xml:"category"`

	Author  string `xml:"creator,omitempty"`
	Summary string `xml:"description"` // required
	Content string `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`

	Guid string `xml:"guid,omitempty"`
}

const time_format_1 = "Mon, 2 Jan 2006 15:04:05 -0700"

func parse_time1(times ...string) time.Time {
	for _, t := range times {
		x, err := time.Parse(time_format_1, t)
		if err == nil {
			return x
		}
	}
	return time.Now()
}

type link struct {
	Rel   string `xml:"rel,attr,omitempty" json:"rel"`
	Href  string `xml:"href,attr" json:"href"`
	Href2 string `xml:",chardata" json:"-"` // just for rss xmltextnode
}

func (this *channel) link_one(uri string) string {
	for _, l := range this.Link {
		if l.Rel == "self" {
			return l.Href
		}
	}
	return uri
}

func charset_reader(charset string, input io.Reader) (io.Reader, error) {
	log.Println("charset-reader", charset)
	switch charset {
	default: // any other encoding should be ignored
		rdr, err := iconv.NewReader(input, charset, "UTF-8")
		return rdr, err
	case "gbk", "gb2312":
		rdr, err := iconv.NewReader(input, "gbk", "UTF-8")
		return rdr, err
	case "utf-8":
		return input, nil
	}
}

/*
func try_panic(err error) {
	if err != nil {
		panic(err)
	}
}
*/
