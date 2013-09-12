package imagedesc

import (
	"github.com/heartszhang/gextract/document"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"io/ioutil"
	"os"
)

func NewThumbnail(uri string, width, height uint) (filepath, mediatype string, w, h int, err error) {
	fp, mediatype, _, err := document.DefaultCurl().Download(uri)
	if err != nil {
		return
	}
	f, err := os.Open(fp)
	if err != nil {
		return
	}
	defer f.Close()
	img, mediatype, err := image.Decode(f)
	if err != nil {
		return
	}
	imgnew := resize.Resize(width, height, img, resize.MitchellNetravali)
	w = imgnew.Bounds().Max.X
	h = imgnew.Bounds().Max.Y

	of, err := ioutil.TempFile("", "")
	if err != nil {
		return
	}
	defer of.Close()
	err = jpeg.Encode(of, imgnew, &jpeg.Options{90})
	if err != nil {
		return
	}
	filepath = of.Name()
	return
}

func NewJpegThumbnail(uri string, width, height uint) (filepath string, w, h int, err error) {
	fp, mt, w, h, err := NewThumbnail(uri, width, height)
	if err != nil {
		return
	}
	filepath = fp + "." + mt
	os.Rename(fp, filepath)
	return
}
