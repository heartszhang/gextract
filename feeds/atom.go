package feeds

import (
	"time"
)

type atom_feed struct {
	//	XMLName Name      `xml:"http://www.w3.org/2005/Atom feed" json:""`
	Title   string       `xml:"title" json:"title"`
	Id      string       `xml:"id" json:"id"`
	Link    []link       `xml:"link" json:"link"`
	PubDate time.Time    `xml:"pubDate,attr" json:"pubdata"`
	Author  atom_person  `xml:"author" json:"author"`
	Entry   []atom_entry `xml:"entry" json:"entry"`
}

type atom_entry struct {
	Title   string      `xml:"title" json:"title"`
	Id      string      `xml:"id" json:"id"`
	Link    []link      `xml:"link" json:"link"`
	Updated time.Time   `xml:"updated" json:"updated"`
	Author  atom_person `xml:"author" json:"author"`
	Summary atom_text   `xml:"summary" json:"summary"`
}

type atom_person struct {
	Name  string `xml:"name" json:"name"`
	URI   string `xml:"uri" json:"uri"`
	Email string `xml:"email" json:"email"`
	//	InnerXML string `xml:",innerxml" json:""`
}

type atom_text struct {
	Type string `xml:"type,attr,omitempty" json:"type"`
	Body string `xml:",chardata" json:"body"`
}
