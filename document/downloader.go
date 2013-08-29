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
	"strings"
	"time"
)

const (
	proxy_address = "http://localhost:8087"
	conn_timeout  = 8  // seconds
	resp_timeout  = 15 // seconds
	tmp_dir       = ""
	prefix        = ""
)

type Curl interface {
	Download(uri string) (filepath, media_type, charset string, err error)
	DownloadHtml(uri string) (filepath, media_type string, err error)
}

type Curler struct {
	TempDir string
	Prefix  string
	UseExt  bool
}

func DefaultCurl() Curl {
	return &Curler{TempDir: tmp_dir, Prefix: prefix, UseExt: false}
}

func NewCurl(tmpdir string) Curl {
	return &Curler{TempDir: tmpdir, Prefix: prefix, UseExt: true}
}

func (this *Curler) Download(uri string) (filepath, media_type, charset string, err error) {
	var cd string
	filepath, media_type, charset, cd, err = this.download_imp(uri)
	if this.UseExt {
		filepath = try_rename_file(filepath, media_type, cd)
	}
	return
}

func try_rename_file(filepath, media_type, content_disposition string) string {
	if ext := filename_suffix(media_type, content_disposition); len(ext) > 0 {
		nf := filepath + ext
		if err := os.Rename(filepath, nf); err == nil {
			filepath = nf
		}
	}
	return filepath
}

func filename_suffix(media_type, content_disposition string) string {
	_, params, _ := mime.ParseMediaType(content_disposition)
	fn, _ := url.QueryUnescape(strings.Trim(params["filename"], ` \t"'`))
	if len(fn) > 0 {
		return fn
	}

	fields := strings.Split(media_type, "/")
	switch len(fields) {
	case 2:
		fn = "." + fields[1]
	}
	return fn
}

func (this *Curler) DownloadHtml(uri string) (filepath, media_type string, err error) {
	filepath, media_type, cd, err := this.download_html(uri)
	if this.UseExt {
		filepath = try_rename_file(filepath, media_type, cd)
	}
	return
}

/*
func ext_from_media_type(media_type string) (ext string, ok bool) {
	ok = false
	fields := strings.Split(media_type, "/")
	switch len(fields) {
	case 2:
		ext = "." + fields[1]
		ok = true
	}
	return
}
*/
func (this *Curler) download_imp(uri string) (filepath, media_type, charset, content_dispos string, err error) {
	resp, err := do_request(uri)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	media_type, charset = extract_charset(resp.Header.Get("content-type"))
	if resp.StatusCode != http.StatusOK {
		err = errors.New(resp.Status + " " + http.StatusText(resp.StatusCode))
		return
	}

	content_dispos = resp.Header.Get("content-disposition")
	var reader io.Reader = nil
	ce := strings.ToLower(resp.Header.Get("content-encoding"))
	switch ce {
	default:
		reader = resp.Body
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return
		}
		defer reader.(*gzip.Reader).Close()
	case "deflate":
		rd := flate.NewReader(resp.Body)
		defer rd.Close()
		reader = rd
	}

	of, err := ioutil.TempFile(this.TempDir, this.Prefix)
	if err != nil {
		return
	}
	defer of.Close()

	io.Copy(of, reader)

	return of.Name(), media_type, charset, content_dispos, nil
}

//return filepath, media_type, content_disposition, error
func (this *Curler) download_html(uri string) (filepath, media_type, content_dispos string, err error) {
	//	var charset string
	ofn, media_type, charset, content_dispos, err := this.download_imp(uri)
	if err != nil {
		return
	}
	of, err := os.Open(ofn)
	if err != nil {
		return
	}
	defer of.Close()

	if len(charset) == 0 {
		_, charset = extract_charset(detect_content_type(of))
		of.Seek(0, 0)
	}
	//	log.Println("detected content-type", charset)

	//某些技术网站使用繁体，但标识gb2312
	if len(charset) == 0 || charset == "gb2312" {
		charset = "gbk"
	}

	tf, err := ioutil.TempFile(this.TempDir, this.Prefix)
	if err != nil {
		return
	}
	defer tf.Close()

	if charset != "utf-8" {
		in, _ := ioutil.ReadAll(of)
		out := make([]byte, len(in)*2)

		var w int
		_, w, err = iconv.Convert(in, out, charset, "utf-8")
		if err != nil {
			return
		}
		out = out[:w]

		io.Copy(tf, bytes.NewReader(out))
	} else {
		io.Copy(tf, of)
	}
	return tf.Name(), media_type, content_dispos, nil
}

func new_proxy_transport(pxyaddr string) *http.Transport {
	pxy, _ := url.Parse(pxyaddr)
	return &http.Transport{
		Proxy: http.ProxyURL(pxy),
		Dial:  timeout_dialer,
		ResponseHeaderTimeout: resp_timeout * time.Second}
}

var timeout = time.Duration(conn_timeout * time.Second)

func timeout_dialer(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

func do_request_client(client *http.Client, url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept-Encoding", "gzip,deflate")
	return client.Do(req)
}

func do_request(url string) (*http.Response, error) {
	client := &http.Client{Transport: &http.Transport{
		Dial: timeout_dialer,
		ResponseHeaderTimeout: resp_timeout * time.Second},
	}
	resp, err := do_request_client(client, url)
	if err != nil {
		log.Println(err)
		client = &http.Client{Transport: new_proxy_transport(proxy_address)}
		resp, err = do_request_client(client, url)
	}
	return resp, err
}

// html has been converted to utf-8
// return local_filepath, media_type, error
func DownloadHtml(uri string) (string, string, error) {
	return DefaultCurl().DownloadHtml(uri)
	/*
		ofn, media_type, charset, err := DownloadFile(url)
		if err != nil {
			return "", "", err
		}
		of, err := os.Open(ofn)
		if err != nil {
			return "", "", err
		}
		defer of.Close()

		if len(charset) == 0 {
			_, charset = extract_charset(detect_content_type(of))
			of.Seek(0, 0)
		}
		//	log.Println("detected content-type", charset)

		//某些技术网站使用繁体，但标识gb2312
		if len(charset) == 0 || charset == "gb2312" {
			charset = "gbk"
		}

		tf, err := ioutil.TempFile(tmp_dir, "")
		if err != nil {
			return "", media_type, err
		}
		defer tf.Close()

		if charset != "utf-8" {
			in, _ := ioutil.ReadAll(of)
			out := make([]byte, len(in)*2)

			_, w, err := iconv.Convert(in, out, charset, "utf-8")
			if err != nil {
				return "", media_type, err
			}
			out = out[:w]

			io.Copy(tf, bytes.NewReader(out))
		} else {
			io.Copy(tf, of)
		}
		return tf.Name(), media_type, nil
	*/
}

// download to tmp path, ungzipped already
// filepath, mediate-type, charset, error
func DownloadFile(uri string) (string, string, string, error) {
	return DefaultCurl().Download(uri)
	/*	resp, err := do_request(uri)
		if err != nil {
			return "", "", "", err
		}
		defer resp.Body.Close()

		media_type, charset := extract_charset(resp.Header.Get("content-type"))
		if resp.StatusCode != http.StatusOK {
			return "", media_type, charset, errors.New(resp.Status + " " + http.StatusText(resp.StatusCode))
		}

		var reader io.Reader = nil
		ce := strings.ToLower(resp.Header.Get("content-encoding"))
		switch ce {
		default:
			reader = resp.Body
		case "gzip":
			reader, err = gzip.NewReader(resp.Body)
			if err != nil {
				return "", media_type, charset, err
			}
			defer reader.(*gzip.Reader).Close()
		case "deflate":
			rd := flate.NewReader(resp.Body)
			defer rd.Close()
			reader = rd
		}

		of, err := ioutil.TempFile(tmp_dir, "")
		if err != nil {
			return "", media_type, charset, err
		}
		defer of.Close()

		io.Copy(of, reader)
		//	log.Println(of.Name(), media_type, charset)
		return of.Name(), media_type, charset, nil
	*/
}

// media-type, charset
func extract_charset(ct string) (string, string) {
	media_type, param, _ := mime.ParseMediaType(ct)
	return media_type, param["charset"]
}

//<meta charset='gb2312'/>
//<meta http-equiv='Content-type' content='text/html;charset=utf-8'/>
// unsupport xml encoded with gbk or non-utf8 endecoders
// media-type, charset
func detect_content_type(file *os.File) string {
	reader := bufio.NewReader(file)
	z := html.NewTokenizer(reader)
	expect_html_root := true
	for tt := z.Next(); tt != html.ErrorToken; tt = z.Next() {
		t := z.Token()
		switch {
		case t.Data == "meta" && (tt == html.StartTagToken || tt == html.SelfClosingTagToken):
			if ct, ok := detect_charset_by_token(t.Attr); ok == true {
				return ct
			}
		case t.Data == "head" && tt == html.EndTagToken:
			break
			// un-html file
		case expect_html_root && (tt == html.StartTagToken || tt == html.SelfClosingTagToken):
			if t.Data == "html" {
				expect_html_root = false
			} else {
				break
			}
		}
	}
	return ""
}

// <meta http-equiv="" content=xxx/>...
// <meta charset=''/>
// return content-type
func detect_charset_by_token(attrs []html.Attribute) (string, bool) {
	var http_equiv, content, charset string
	for _, attr := range attrs {
		switch attr.Key {
		case "http-equiv":
			http_equiv = attr.Val
		case "content":
			content = attr.Val
		case "charset":
			charset = attr.Val
		}
	}
	switch {
	default:
		return "", false
	case strings.ToLower(http_equiv) == "content-type":
		return content, true
	case len(charset) > 0:
		return "text/html; charset=" + charset, true
	}
}

// read utf-8 html file
func NewHtmlDocument(localpath string) (*html.Node, error) {
	f, err := os.Open(localpath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	return html.Parse(reader)
}

// write html.Node to tmp file
// return tmp_filename, utf-8 encoded
func WriteHtmlFile2(doc *html.Node) (string, error) {
	of, err := ioutil.TempFile(tmp_dir, prefix)
	if err != nil {
		return "", err
	}
	defer of.Close()

	html.Render(of, doc)
	return of.Name(), nil
}
