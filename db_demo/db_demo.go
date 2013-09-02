package main

import (
	"fmt"
	"github.com/heartszhang/gextract/feeds"
)

func main() {
	/*
		fo := feeds.NewFeedOperator()
		fds, err := fo.TimeoutFeeds()
		fmt.Println(len(fds), err)

			for _, feed := range fds {
				err := fo.Touch(feed.Link, feed.TTL)
				if err != nil {
					fmt.Println(err)
				}
			}
	*/
	eo := feeds.NewEntryOperator()
	etris, err := eo.TopN(0, 10)
	fmt.Println(len(etris), err)
	for _, entry := range etris {
		err = eo.SetContent(entry.Link, "", 1, []string{"a.dumy", "b.dmy"})
	}
}
