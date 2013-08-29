package main

import (
	"fmt"
	"gextract/feeds"
)

func main() {
	ols, _ := feeds.NewOpml("opml.xml")
	feeds.NewFeedOperator().Save(ols)
	fmt.Println(ols)
}
