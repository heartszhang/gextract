package main

import (
	"flag"
	"fmt"
	d "gextract/document"
)

var uri = flag.String("uri", "", "http://xx.com.cn")

const (
	target_folder = "/home/hearts/Pictures/cache/"
	cocurrents    = 4
)

func main() {
	flag.Parse()
	if len(*uri) == 0 {
		fmt.Println(`image_crawler --uri http://www.baidu.com/`)
	}
	fmt.Println(*uri)
	_, sc, _ := d.ExtractHtml(*uri)
	fmt.Println(sc.WordCount, sc.LinkCount, len(sc.Images))

	//	curl := d.NewCurl(target_folder)

	capl := (len(sc.Images) + cocurrents - 1) / cocurrents
	if capl == 0 {
		return
	}
	cnt := (len(sc.Images) + capl - 1) / capl
	done := make(chan int, cnt)
	split_task(sc.Images, capl, done)
	for cnt > 0 {
		<-done
		cnt--
	}
	//	for _, img := range sc.Images {
	//		go download_image(img, done)
	//		imgf, _, _, err := curl.Download(img)
	//		fmt.Println(imgf, err)
	/*		if err == nil {
				fn := path.Base(imgf)
				ext := extension(mt)
				t := path.Join(target_folder, fn+ext)
				err = os.Rename(imgf, t)
				fmt.Println(t, err)
			}
	*/
	//	}
}

func download_images(uris []string, done chan<- int) {
	curl := d.NewCurl(target_folder)
	for _, uri := range uris {
		imgf, _, _, err := curl.Download(uri)
		fmt.Println(imgf, err)
	}
	done <- 0
}

func split_task(urls []string, cap int, done chan<- int) {
	if len(urls) < cap {
		go download_images(urls, done)
	} else {
		go download_images(urls[:cap], done)
		split_task(urls[cap:], cap, done)
	}
}
