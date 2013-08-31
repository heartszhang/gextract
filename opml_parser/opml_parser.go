package main

import (
	"fmt"
	feeds "github.com/heartszhang/gextract/feeds"
	opml "github.com/heartszhang/gextract/feeds/opml"
)

func main() {
	ols, _ := opml.NewOpml("opml.xml")
	feeds.NewFeedOperator().Save(ols)
	fmt.Println(ols)
}
