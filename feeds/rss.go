package feeds

import (
	"encoding/xml"
	"gextract/document"
	"log"
	"os"
	"time"
)

const (
	content_status_unresolved = iota
	content_status_ready      // content has been stored in a file
	content_status_failed     // fetch content failed
	content_status_unavail    // fetch content ok, buy extracted-body is low quality
	content_status_summary    // content stored in Entry.Summary
	content_status_content    // content stored in Entry.Content
)

type Category struct {
	//	Id      bson.ObjectId `xml:"-" bson:"_id,omitempty" json:"-"`
	Title   string    `bson:"title,omitempty" json:"title,omitempty"`
	Cid     int64     `bson:"cid,minsize" json:"cid,minsize"`
	Refresh time.Time `bson:"refresh" json:"refresh"`
}

type Channel struct {
	//	Id    bson.ObjectId `xml:"-" bson:"_id,omitempty" json:"_id,omitempty"`
	Title string `xml:"title" json:""`                                 // required  unique?
	Link  []link `xml:"link" json:"-" bson:"-"`                        // required
	Href  string `xml:"-" json:"link,omitempty" bson:"link,omitempty"` // unique, used as primary key

	PubDate       string `xml:"pubDate,omitempty" json:"-" bson:"-"` // created or updated
	LastBuildDate string `xml:"lastBuildDate" json:"-" bson:"-"`

	Date         time.Time `xml:"-" json:"pubdate" bson:"pubdate"`
	Ttl          int32     `xml:"ttl" json:"ttl" bson:"ttl"` // minitues
	UpdatePeriod string    `xml:"http://www.w3.org/2005/Atom updatePeriod" json:"-" bson:"-"`

	Enabled  bool      `xml:"-" json:"enabled" bson:"enabled"`
	Refresh  time.Time `xml:"-" bson:"refresh" json:"refresh"`
	Category []string  `xml:"category,omitempty" json:"catetory,omitempty" bson:"catetory,omitempty"`
}

type Entry struct {
	//	Id       bson.ObjectId `xml:"-" bson:"_id,omitempty" json:"-"`
	Title   string    `xml:"title" json:"title"`                  // required
	Link    string    `xml:"link" json:"link" bson:"link"`        // required, unique , used as primary key
	PubDate string    `xml:"pubDate,omitempty" json:"-" bson:"-"` // created or updated
	Date    time.Time `xml:"-" json:"pubdate" bson:"pubdate"`
	Created time.Time `xml:"-" json:"created" bson:"created"`

	//	Digest int64  `xml:"-" json:"digest,minsize" bson:"digest,minsize"` // fnv1.sum64(summary+content)
	Read bool   `xml:"-" json:"read" bson:"read"`
	Cid  string `xml:"-" bson:"cid" json:"cid"`

	Category []string `xml:"category" bson:"category,omitempty" json:"category,omitempty"`
	//http://purl.org/dc/elements/1.1/
	Author        string `xml:"//http://purl.org/dc/elements/1.1/ creator,omitempty" bson:"author,omitempty" json:"author,omitempty"`
	Summary       string `xml:"description" json:"summary,omitempty" bson:"summary"` // required
	Content       string `xml:"http://purl.org/rss/1.0/modules/content/ encoded" json:"content,omitempty" bson:"content,omitempty"`
	Guid          string `xml:"guid,omitempty" json:"guid,omitempty" bson:"guid,omitempty"`
	ContentStatus int    `xml:"-" json:"content_status" bson:"content_status"` // full content is ready
	ContentPath   string `xml:"" bson:"content_path" json:"content_path"`
}

func (this *Entry) PreProcess(cid string, category []string) {
	//	this.Id = bson.NewObjectId()
	this.Date = parse_time1(this.PubDate)
	this.Created = time.Now()
	this.Cid = cid
	if len(this.Guid) == 0 {
		this.Guid = this.Link
	}
	/*	hash := fnv.New64a()
		hash.Write([]byte(this.Summary))
		hash.Write([]byte(this.Content))
		this.Digest = int64(hash.Sum64())
	*/
	log.Println(this.Link)
	c := make([]string, len(this.Category)+len(category))
	copy(c, this.Category)
	copy(c[len(this.Category):], category)
	this.Category = c

	this.clean_summary()

	switch {
	default:
		this.ContentStatus = content_status_unresolved
	case len(this.Summary) > 512:
		this.ContentStatus = content_status_summary
	case len(this.Content) > 512:
		this.ContentStatus = content_status_content
	}
}

func (this *Entry) clean_summary() {
	this.Summary, _ = document.CleanFragment(this.Summary, this.Link)
	this.Content, _ = document.CleanFragment(this.Content, this.Link)
}

func (this *Channel) PreProcess(url string) {
	//	this.Id = bson.NewObjectId()
	this.Date = parse_time1(this.PubDate, this.LastBuildDate)
	this.Href = this.LinkOne()
	if len(this.Href) == 0 {
		this.Href = url
	}
	this.Enabled = true
	//  hourly, daily, weekly, monthly, yearly
	switch this.UpdatePeriod {
	case "hourly":
		this.Ttl = 60
	case "daily":
		this.Ttl = 24 * 60
	case "weekly":
		this.Ttl = 7 * 24 * 60
	case "monthly":
		this.Ttl = 30 * 24 * 60
	case "yearly":
		this.Ttl = 365 * 24 * 60
	}

	if this.Ttl == 0 {
		this.Ttl = 24 * 60 // minutes
	}
	this.Refresh = this.Date.Add(time.Duration(this.Ttl) * time.Minute)
}

// private wrapper around the RssFeed which gives us the <rss>..</rss> xml
type rss struct {
	Channel channel `xml:"channel" json:"channel,omitempty"`
}

type channel struct {
	Channel
	Items []Entry `xml:"item" json:"item"`
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

func (this *Channel) LinkOne() string {
	for _, l := range this.Link {
		if len(l.Href) > 0 {
			return l.Href
		}
	}
	return ""
}

func NewRss2(filepath string, url string) (Channel, []Entry) {
	f, err := os.Open(filepath)
	try_panic(err)
	defer f.Close()

	var v rss
	err = xml.NewDecoder(f).Decode(&v)
	try_panic(err)
	v.Channel.PreProcess(url)
	for i, _ := range v.Channel.Items {
		v.Channel.Items[i].PreProcess(v.Channel.Href, v.Channel.Category)
	}
	return v.Channel.Channel, v.Channel.Items
}

func try_panic(err error) {
	if err != nil {
		panic(err)
	}
}
