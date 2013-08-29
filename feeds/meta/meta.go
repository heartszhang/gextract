package meta

import (
	"time"
)

const (
	ContentStatusUnresolved = iota
	ContentStatusReady      // content has been stored in a file
	ContentStatusFailed     // fetch content failed
	ContentStatusUnavail    // fetch content ok, buy extracted-body is low quality
	ContentStatusSummary    // content stored in Entry.Summary
)

type Category struct {
	Title   string    `title,omitempty"`
	Refresh time.Time `refresh`
}

type Feed struct { // generic struct
	Title    string    `title,omitempty`
	Link     string    `link,omitempty`     // rss url, not the web site, as primary key
	Refresh  time.Time `refresh`            // next refresh time
	TTL      int       `ttl`                // minutes      refresh period
	Category []string  `catetory,omitempty` // Category
}

type Entry struct {
	Title   string    `title`
	Updated time.Time `updated`
	Author  string    `author,omitempty`
	Summary string    `summary,omitempty`

	Guid     string    `guid,omitempty` //guid or link
	Link     string    `link`
	Created  time.Time `created` // fetch date, not the publish date
	Readed   bool      `readed`
	Feed     string    `feed,omitempty` // feed's link
	Category []string  `catetory`       // feed's category + items category

	//	ContentPath   string `content_path,omitempty` // local html path, relative or abslute?
	//	ContentStatus int    `content_status`         // content_status_...

	Status  int           `status`
	Content ContentStatis `content`
	Statis  EntryStatis   `statis`
	Image   []string      `image`
}

type EntryStatis struct {
	Words int `words`
	Imgs  int `imgs`
	//	Image []string `image,omitempty`
}
type ContentStatis struct {
	Local string `local,omitempty`
	Words int    `words`
	Imgs  int    `imgs`
}
type Image struct {
	Link    string    `link,omitempty`
	Local   string    `local,omitempty`
	Ready   int       `ready`
	Created time.Time `created`
}
