package main

import (
	"encoding/json"
	"fmt"
	htmldoc "gextract/document"
	"gextract/feeds"
	"io"
	"log"
	"net/http"
	//	"net/url"
	"sync"
	"time"
)

/*	at_url_template = `https://api.weibo.com/oauth2/access_token?client_id=%v&client_secret=%v&grant_type=%v&redirect_uri=%v&code=`
 */

const (
	client_id     = `4277294736`
	client_secret = `c840165b9fd5d08a8353393d94b8d808`
	grant_type    = `authorization_code`
	redirect_uri  = `http://gohearts.duapp.com/sina/callback`
	oauth2_url    = `https://api.weibo.com/oauth2/access_token`
)

type conf struct {
	Last   int64                  `json:"last"`
	Token  map[int64]access_token `json:"token,omitempty"`
	Status status                 `json:"status, omitempty"`
	Code   string                 `json:"code,omitempty"`
}

type access_token struct {
	Token   string `json:"token,omitempty"`
	Secret  string `json:"secret,omitempty"`
	Expires int64  `json:"expires,omitempty"`
	Refresh int64  `json:"refresh"`
	UserId  int64  `json:"userid"`
}

type status struct {
	code int64
	last int64
	text string `json:"omitempty"`
}

type global_status struct {
	conf
	locker sync.Mutex
}

func (this *global_status) config() conf {
	this.locker.Lock()
	defer this.locker.Unlock()
	return this.conf
}

func (this *global_status) update(updater func(c *conf)) {
	this.locker.Lock()
	defer this.locker.Unlock()
	updater(&this.conf)
}

func get_at_url(code string) string {
	return oauth2_url + "?client_id=" + client_id +
		"&client_secret=" + client_secret +
		"&grant_type=" + grant_type +
		"&redirect_uri=" + redirect_uri +
		"&code=" + code
}

func oauth2_callback2(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	_status.update(func(c *conf) {
		c.Code = r.FormValue("code")
		c.Status.text = r.FormValue("error")
	})
	write_json(w, r.Form)
}
func oauth2_callback(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ocode := r.FormValue("code")
	if len(ocode) == 0 {
		http.Error(w, r.Form.Encode(), http.StatusBadGateway)
		return
	}
	ac_url := get_at_url(ocode)
	c := &http.Client{}
	resp, err := c.Post(ac_url, "", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	w.WriteHeader(resp.StatusCode)
	for k, v := range resp.Header {
		w.Header().Set(k, v[0])
		log.Println("remote", k, v)
	}
	io.Copy(w, resp.Body)
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "works")
}

func tick_tack(w http.ResponseWriter, r *http.Request) {
	write_json(w, _status.config())
}

func extract_html(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	url := r.FormValue("url")
	if len(url) == 0 {
		write_json(w, _status.config())
	}
	tf := htmldoc.ExtractHtml(url)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, tf)
}
func extract_json(w http.ResponseWriter, r *http.Request) {
	write_json(w, _status.config())
}
func extract_simple_json(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	uri := r.FormValue("url")
	if len(uri) == 0 {
		write_json(w, _status.config())
	}
	tf := htmldoc.ExtractHtml(uri)
	val := struct {
		Url    string `json:"url,omitempty"`
		Target string `json:"target,omitempty"`
	}{Url: uri, Target: tf}

	write_json(w, val)
}

func fetch_channels(w http.ResponseWriter, r *http.Request) {
	write_json(w, _status.config())
}

func fetch_catetorys(w http.ResponseWriter, r *http.Request) {
	write_json(w, _status.config())
}

func fetch_feeds_channel_refresh(w http.ResponseWriter, r *http.Request) {
	write_json(w, _status.config())
}

func fetch_feeds_catetory_refresh(w http.ResponseWriter, r *http.Request) {
	write_json(w, _status.config())
}

func fetch_feeds_all_refresh(w http.ResponseWriter, r *http.Request) {
	write_json(w, _status.config())
}
func fetch_feeds_channel(w http.ResponseWriter, r *http.Request) {
	write_json(w, _status.config())
}
func fetch_feeds_category(w http.ResponseWriter, r *http.Request) {
	write_json(w, _status.config())
}
func fetch_feeds_all(w http.ResponseWriter, r *http.Request) {
	write_json(w, _status.config())
}
func unsubscribe_rss2(w http.ResponseWriter, r *http.Request) {
	write_json(w, _status.config())
}
func subscribe_rss2(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			http.Error(w, err.(error).Error(), http.StatusBadGateway)
			fmt.Println(err)
		}
	}()

	r.ParseForm()
	url := r.FormValue("url")
	fmt.Println("fetching ", url)
	rss_file := htmldoc.FetchUrl2(url)
	fmt.Println("rss-file", rss_file)
	channel, _ := feeds.NewRss2(rss_file, url)
	feeds.InsertChannel(channel)

	write_json(w, channel)
}
func write_json(w http.ResponseWriter, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(body)
}

func init() {
	http.HandleFunc("/", index)
	http.HandleFunc("/tick.json/", tick_tack)
	http.HandleFunc("/extract.json", extract_json)
	http.HandleFunc("/extract_simp.json", extract_simple_json)
	http.HandleFunc("/extract.html", extract_html)

	http.HandleFunc("/subscribe", subscribe_rss2)       // ?url=
	http.HandleFunc("/unsubscribe", unsubscribe_rss2)   // ?url= || ?id=
	http.HandleFunc("/feeds.json/all", fetch_feeds_all) //skip=0&limit=10
	http.HandleFunc("/feeds.json/catetory", fetch_feeds_category)
	http.HandleFunc("/feeds.json/channel", fetch_feeds_channel)
	http.HandleFunc("/feeds.json/all/refresh", fetch_feeds_channel) //pubdate=
	http.HandleFunc("/feeds.json/catetory/refresh", fetch_feeds_channel)
	http.HandleFunc("/feeds.json/channel/refresh", fetch_feeds_channel)

	http.HandleFunc("/catetory.json", fetch_catetorys)
	http.HandleFunc("/channel.json", fetch_channels)

	http.HandleFunc("/api/callback.json/sina/callback", oauth2_callback2)
}

func main() {
	log.Fatal(http.ListenAndServe("localhost:1212", nil))
}

var _status = global_status{conf: conf{Last: time.Now().Unix()}}
