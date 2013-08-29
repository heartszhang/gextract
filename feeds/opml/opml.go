package feeds

import (
	"encoding/xml"
	. "gextract/feeds/meta"
	//	"labix.org/v2/mgo"
	//	"labix.org/v2/mgo/bson"
	//	"log"
	"os"

	"time"
)

type opml struct {
	Body body `xml:"body,omitempty" json:"-" bson:"-"`
}

type body struct {
	Outline []outline `xml:"outline" bson:"outline,omitempty" json:"outline,omitempty"`
}

type Outline struct {
	Text     string   `xml:"text,attr" bson:"-" json:"-"`
	Title    string   `xml:"title,attr" bson:"title" json:"title"`
	Type     string   `xml:"type,attr" bson:"type" json:"type"`
	Link     string   `xml:"xmlUrl,attr" bson:"link" json:"link"`
	HtmlUrl  string   `xml:"htmlUrl,attr" bson:"htmlurl" json:"htmlurl"`
	Category []string `xml:"-" bson:"category,omitempty" json:"category,omitempty"`
}

type outline struct {
	Outline
	Children []outline `xml:"outline,omitempty" bson:"children,omitempty" json:"omitempty"`
}

func NewOpml(filepath string) ([]Feed, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var o opml
	d := xml.NewDecoder(f)
	err = d.Decode(&o)
	if err != nil {
		return nil, err
	}

	rtn := []Feed{}
	for _, ol := range o.Body.Outline {
		extract_feed(ol, []string{}, &rtn)
	}
	return rtn, nil
}

func extract_feed(o outline, cate []string, rtn *[]Feed) {
	cate = append(cate, o.Title)
	for _, ol := range o.Children {
		switch len(ol.Link) {
		default:
			*rtn = append(*rtn, ol.Outline.to_feed())
		case 0:
			extract_feed(ol, cate, rtn)
		}
	}
}

/*
func SaveOpml(outlines []Outline) error {
	err := do_in_session("channels", func(coll *mgo.Collection) {
		for _, ol := range outlines {
			ci, _ := coll.Upsert(bson.M{"link": ol.Link}, bson.M{"$set": bson.M{
				"title":    ol.Title,
				"enabled":  true,
				"ttl":      60,
				"refresh":  time.Now(),
				"category": ol.Category,
			}})
			// err would be ignored
		}
	})
	return err
}
*/

func (this *Outline) to_feed() Feed {
	return Feed{Title: this.Title,
		Link:     this.Link,
		TTL:      60,
		Refresh:  time.Now(),
		Category: this.Category,
	}
}
