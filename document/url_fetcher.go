package document

import (
	"bufio"
	"bytes"
	"code.google.com/p/go.net/html"
	"compress/flate"
	"compress/gzip"
	"errors"
	iconv "github.com/djimenez/iconv-go"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
	"strings"
)

const proxy_address = "http://localhost:8087"

func new_proxy_transport(pxyaddr string) *http.Transport {
	pxy, _ := url.Parse(pxyaddr)
	return &http.Transport{Proxy: http.ProxyURL(pxy), Dial: timeout_dialer}
}

var timeout = time.Duration(8 * time.Second)

func timeout_dialer(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

func do_request_client(client *http.Client, url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	try_panic(err)

	req.Header.Add("Accept-Encoding", "gzip, deflate")

	return client.Do(req)
}
func do_request(url string) (*http.Response, error) {
	client := &http.Client{Transport: &http.Transport{Dial: timeout_dialer}}
	resp, err := do_request_client(client, url)
	if err != nil {
		log.Println(err)
		client = &http.Client{Transport: new_proxy_transport(proxy_address)}
		return do_request_client(client, url)
	}

	return resp, err
}

func fetch_url(url string) (string, string) {
	resp, err := do_request(url)
	try_panic(err)

	defer resp.Body.Close()

	charset, content_type := extract_charset(resp.Header.Get("content-type"))
	if resp.StatusCode != http.StatusOK {
		try_panic(errors.New(resp.Status + " " + http.StatusText(resp.StatusCode)))
	}

	log.Println(resp.Header.Get("Content-Encoding"))

	var reader io.Reader = nil
	ce := strings.ToLower(resp.Header.Get("content-encoding"))
	switch ce {
	default:
		reader = resp.Body
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		try_panic(err)
	case "deflate":
		rd := flate.NewReader(resp.Body)
		defer rd.Close()
		reader = rd
	}

	of, err := ioutil.TempFile("", "")
	try_panic(err)
	defer of.Close()

	io.Copy(of, reader)
	//  of.Sync()
	of.Seek(0, 0)

	if len(charset) == 0 {
		charset, _ = detect_charset(of)
	}
	log.Println("detected content-type", charset)

	of.Seek(0, 0)
	//某些技术网站使用繁体，但标识gb2312
	if len(charset) == 0 || charset == "gb23123" {
		charset = "gbk"
	}

	tf, err := ioutil.TempFile("", "")
	try_panic(err)
	defer tf.Close()

	if charset != "utf-8" {
		in, _ := ioutil.ReadAll(of)
		out := make([]byte, len(in)*2)

		_, w, _ := iconv.Convert(in, out, charset, "utf-8")
		out = out[:w]

		io.Copy(tf, bytes.NewReader(out))
	} else {
		io.Copy(tf, of)
	}
	return tf.Name(), content_type
}

// return utf-8 encoded cache file path
func FetchUrl3(url string) (tf, content_type string, err error) {
	defer func() {
		err, _ = recover().(error)
	}()

	tf, content_type = fetch_url(url)
	return
}

func FetchUrl2(url string) (string, string) {
	return fetch_url(url)
}

func extract_charset(ct string) (string, string) {
	media_type, param, _ := mime.ParseMediaType(ct)
	return param["charset"], media_type
}
func detect_charset(file *os.File) (string, string) {
	head := make([]byte, 512)
	file.Read(head)
	ct := http.DetectContentType(head)
	return extract_charset(ct)
	/*
		file.Seek(0, 0)
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
	*/
}

/*
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
*/
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

func WriteHtmlFile2(doc *html.Node) string {
	of, err := ioutil.TempFile("", "")
	try_panic(err)
	defer of.Close()

	html.Render(of, doc)
	return of.Name()
}
