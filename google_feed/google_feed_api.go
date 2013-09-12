package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/heartszhang/gextract/document"
	"net/url"
	"os"
)

type feed_find_response struct {
	Data    find_data `json:"responseData"`
	Details string    `json:"responseDetails"`
	Status  int       `json:"responseStatus"`
}

type feed_error struct {
	Code    int    `code`    // http-style error code
	Message string `message` // human-readable description
}

type find_data struct {
	Query   string `query,omitempty`
	Entries []feed `entries`
}

type feed struct {
	Title          string `title`
	Link           string `link` // html version of feed
	ContentSnippet string `contentSnippet`
	Url            string `url` // feed url
}

type feed_load_response struct {
	Data    load_data `json:"responseData"`
	Details string    `json:"responseDetails"`
	Status  int       `json:"responseStatus"`
}

type load_data struct {
	Feed feed_entries `feed`
}

type feed_entries struct {
	FeedUrl     string       `feedUrl,omitempty` // rss version
	Title       string       `title`
	Link        string       `link` // html version
	Description string       `description,omitempty`
	Author      string       `author,omitempty`
	Type        string       `type,omitempty` // rss20
	Entries     []feed_entry `entries`
}

type feed_entry struct {
	//	MediaGroup     media_group `mediaGroup, omitempty`
	Title          string `title,omityempty`
	Link           string `link`
	Author         string `author,omitempty`
	Content        string `content`
	ContentSnippet string `contentSnippet`
	PublishedDate  string `publishedDate`
	Categories     []string
}

// https://ajax.googleapis.com/ajax/services/feed/find?v=1.0&num=100&q=

// https://ajax.googleapis.com/ajax/services/feed/load?v=1.0&num=100&q=
var q = flag.String("q", "", "query word")
var load = flag.Bool("load", false, "load entries")

func main() {
	flag.Parse()
	if *q == "" {
		fmt.Println("google_feed -q caoliu")
		return
	}
	u, e := url.Parse(*q)
	switch {
	case e != nil:
		*load = false
	case u.IsAbs():
		*load = true
	default:
		*load = false
	}
	if *load {
		load_entries(*q)
	} else {
		find_feeds(*q)
	}
}

func find_feeds(q string) (v feed_find_response, e error) {
	uri := "http://ajax.googleapis.com/ajax/services/feed/find?v=1.0&num=20&q=" + q
	f, _, _, e := document.DownloadFile(uri)
	if e != nil {
		fmt.Println(e)
		return
	}
	fmt.Println(f)
	lf, e := os.Open(f)
	if e != nil {
		fmt.Println(e)
		return
	}
	defer lf.Close()

	dec := json.NewDecoder(lf)
	e = dec.Decode(&v)
	fmt.Println(len(v.Data.Entries), e)
	return
}

func load_entries(q string) (v feed_load_response, e error) {
	uri := "http://ajax.googleapis.com/ajax/services/feed/load?v=1.0&num=20&q=" + q // "http%3A%2F%2Fwww.digg.com%2Frss%2Findex.xml"
	f, _, _, e := document.DownloadFile(uri)
	if e != nil {
		fmt.Println(e)
		return
	}
	fmt.Println(f)
	lf, e := os.Open(f)
	if e != nil {
		fmt.Println(e)
		return
	}
	defer lf.Close()

	dec := json.NewDecoder(lf)
	e = dec.Decode(&v)
	fmt.Println(v.Data.Feed.Title, e, v.Details)
	return
}
