package main

import (
	"fmt"
	d "github.com/heartszhang/gextract/document"
	"github.com/heartszhang/gextract/feeds"
	"github.com/heartszhang/gextract/feeds/meta"
)

const cocurrents = 8

func main() {
	fds, _ := feeds.NewFeedOperator().AllFeeds()
	capl := (len(fds) + cocurrents - 1) / cocurrents
	if capl == 0 {
		return
	}
	cnt := (len(fds) + capl - 1) / capl
	done := make(chan int, cnt)
	split_task(fds, capl, done)
	for cnt > 0 {
		<-done
		cnt--
	}
}

func fetch_entries(feds []meta.Feed, done chan<- int) {
	fmt.Println("fetch entries", len(feds))
	curl := d.DefaultCurl()
	for _, feed := range feds {
		rsfile, _, _, err := curl.Download(feed.Link)
		fmt.Println(rsfile, err)
		if err != nil {
			fmt.Println(err)
			continue
		}
		f, entries, err := feeds.NewFeed(rsfile, feed.Link)
		if err != nil {
			fmt.Println(feed.Link, len(entries), err)
			feeds.NewFeedOperator().Remove(feed.Link)
		} else {
			fmt.Println(f.Title, len(entries))
		}

		if err == nil {
			feeds.NewEntryOperator().Save(entries)
		}

	}
	done <- 0
}

func split_task(fds []meta.Feed, cap int, done chan<- int) {
	if len(fds) < cap {
		go fetch_entries(fds, done)
	} else {
		go fetch_entries(fds[:cap], done)
		split_task(fds[cap:], cap, done)
	}
}
