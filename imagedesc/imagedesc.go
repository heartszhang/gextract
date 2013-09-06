package imagedesc

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"mime"
	"net"
	"net/http"
	"time"
)

const (
	dial_timeo = time.Duration(3) * time.Second
	resp_timeo = time.Duration(5) * time.Second
	rw_timeo   = time.Duration(5) * time.Second
)

func dial_timeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, dial_timeo)
}

func new_timeout_httpclient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{Dial: dial_timeout, ResponseHeaderTimeout: resp_timeo},
	}
}

func NewTimeoutHttpClient(dial_timeo, resp_timeo time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: resp_timeo,
			Dial: func(network, addr string) (net.Conn, error) {
				return net.DialTimeout(network, addr, dial_timeo)
			},
		},
	}
}

func DetectImageType(uri string) (mediatype string, width, height int, filelength int64, err error) {
	resp, err := new_timeout_httpclient().Get(uri)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("%v: %v", resp.StatusCode, http.StatusText(resp.StatusCode))
		return
	}
	filelength = resp.ContentLength
	if resp.ContentLength < 16 && resp.ContentLength >= 0 {
		fmt.Errorf("%v: %v", resp.ContentLength, "content-length insufficient info")
		return
	}
	ct := resp.Header.Get("Content-Type")
	mediatype, _, _ = mime.ParseMediaType(ct)

	ic, mediatype, err := image.DecodeConfig(resp.Body)
	width = ic.Width
	height = ic.Height
	return
}
