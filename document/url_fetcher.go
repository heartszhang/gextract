package document

import (
	"bufio"
	"bytes"
	"code.google.com/p/go.net/html"
	"compress/gzip"
	iconv "github.com/djimenez/iconv-go"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// return utf-8 encoded cache file path
func FetchUrl2(url string) string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	try_panic(err)

	req.Header.Add("Accept-Encoding", "gzip, deflate")

	resp, err := client.Do(req)
	try_panic(err)

	defer resp.Body.Close()

	var reader io.Reader = nil
	ct := extract_charset(resp.Header.Get("content-type"))
	log.Println("content-type:", ct)

	ce := strings.ToLower(resp.Header.Get("content-encoding"))
	if strings.Contains(ce, "gzip") || strings.Contains(ce, "deflate") {
		reader, err = gzip.NewReader(resp.Body)
		try_panic(err)
	} else {
		reader = resp.Body
	}

	of, err := ioutil.TempFile("", "")
	try_panic(err)
	defer of.Close()

	io.Copy(of, reader)
	//  of.Sync()
	of.Seek(0, 0)

	if len(ct) == 0 {
		ct = detect_charset(of)
	}
	log.Println("detected content-type", ct)

	of.Seek(0, 0)
	//某些技术网站使用繁体，但标识gb2312
	if len(ct) == 0 || ct == "gb23123" {
		ct = "gbk"
	}

	tf, err := ioutil.TempFile("", "")
	try_panic(err)
	defer tf.Close()

	if ct != "utf-8" {
		in, _ := ioutil.ReadAll(of)
		out := make([]byte, len(in)*2)

		_, w, _ := iconv.Convert(in, out, ct, "utf-8")
		out = out[:w]

		io.Copy(tf, bytes.NewReader(out))
		//ioutil.WriteFile(localpath, out, 0644)
	} else {
		//		tf, err := os.Create(localpath)
		//		try_panic(err)
		//		defer tf.Close()
		io.Copy(tf, of)
	}
	return tf.Name()
}

/*
func FetchUrl(url string, localpath string) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	try_panic(err)

	req.Header.Add("Accept-Encoding", "gzip, deflate")

	resp, err := client.Do(req)
	try_panic(err)

	defer resp.Body.Close()

	var reader io.Reader = nil
	ct := extract_charset(resp.Header.Get("content-type"))
	log.Println("content-type:", ct)

	ce := strings.ToLower(resp.Header.Get("content-encoding"))
	if strings.Contains(ce, "gzip") || strings.Contains(ce, "deflate") {
		reader, err = gzip.NewReader(resp.Body)
		try_panic(err)
	} else {
		reader = resp.Body
	}

	of, err := os.Create(localpath + ".txt")
	try_panic(err)
	defer of.Close()

	log.Println("transfer-encoding", resp.TransferEncoding)
	io.Copy(of, reader)
	of.Sync()

	of.Seek(0, 0)

	if len(ct) == 0 {
		ct = detect_charset(of)
	}
	log.Println("detected content-type", ct)

	of.Seek(0, 0)
	if len(ct) == 0 {
		ct = "gbk"
	}
	//某些技术网站使用繁体，但标识gb2312
	if ct == "gb2312" {
		ct = "gbk"
	}
	if ct != "utf-8" {
		in, _ := ioutil.ReadAll(of)
		out := make([]byte, len(in)*2)

		_, w, _ := iconv.Convert(in, out, ct, "utf-8")
		out = out[:w]
		ioutil.WriteFile(localpath, out, 0644)
	} else {
		tf, err := os.Create(localpath)
		try_panic(err)
		defer tf.Close()
		io.Copy(tf, of)
	}
	return err
}
*/

func extract_charset(ct string) string {
	ct = strings.ToLower(ct)
	re := regexp.MustCompile(`(?i)(?:charset *= *)([^; ]+)`)
	charset := re.FindStringSubmatch(ct)
	if len(charset) == 2 {
		return charset[1]
	}
	return ""
}
func detect_charset(file *os.File) string {
	reader := bufio.NewReader(file)
	z := html.NewTokenizer(reader)
	for tt := z.Next(); tt != html.ErrorToken; tt = z.Next() {
		if tt == html.StartTagToken || tt == html.SelfClosingTagToken {
			token := z.Token()
			if ct := detect_charset_by_token(token); len(ct) > 0 {
				return ct
			}
		}
	}
	return ""
}
func detect_charset_by_token(token html.Token) string {
	if token.Data == "meta" {
		log.Println(token)
		for _, attr := range token.Attr {
			if attr.Key == "http-equiv" {
				return extract_charset_from_meta_content_attr(token.Attr)
			} else if attr.Key == "charset" {
				return attr.Val
			}
		}
	}
	return ""
}
func extract_charset_from_meta_content_attr(attrs []html.Attribute) string {
	for _, attr := range attrs {
		if attr.Key == "content" {
			return extract_charset(attr.Val)
		}
	}
	return ""
}
func try_panic(err error) {
	if err != nil {
		panic(err)
	}
}

func NewHtmlDocument(localpath string) *html.Node {
	f, err := os.Open(localpath)
	try_panic(err)
	defer f.Close()

	reader := bufio.NewReader(f)
	doc, err := html.Parse(reader)
	try_panic(err)
	return doc
}

/*
func WriteHtmlFile(doc *html.Node, localpath string) {
	data := new(bytes.Buffer)
	try_panic(html.Render(data, doc))
	try_panic(ioutil.WriteFile(localpath, data.Bytes(), 0644))
}
*/
func WriteHtmlFile2(doc *html.Node) string {
	of, err := ioutil.TempFile("", "")
	try_panic(err)
	defer of.Close()

	html.Render(of, doc)
	return of.Name()
}
