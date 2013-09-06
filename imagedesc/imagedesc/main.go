package main

import (
	"flag"
	"fmt"
	"github.com/heartszhang/gextract/imagedesc"
)

var uri = flag.String("uri", "", "image-url")

func main() {
	flag.Parse()
	if *uri == "" {
		fmt.Println("imagedesc --uri http://www.google.com/logo.png")
		return
	}
	mt, w, h, cl, err := imagedesc.DetectImageType(*uri)
	fmt.Println(mt, w, h, cl, err)
	fp, w, h, err := imagedesc.NewJpegThumbnail(*uri, 120, 0)
	fmt.Println(fp, mt, w, h, err)
	/*
		const header_size = 4
		header := make([]byte, header_size)
		n, err := io.ReadAtLeast(resp.Body, header, header_size)
		if n < header_size {
			fmt.Println("insufficient body size")
			return
		} else {
			fmt.Println(header)
			for n, v := range image_hdr {
				if byte_prefix_equal(header, v) {
					fmt.Println(n)
					return
				}
			}
		}
	*/
}

func byte_prefix_equal(lhs, rhs []byte) bool {
	for i, v := range rhs {
		if lhs[i] != v {
			return false
		}
	}
	return true
}

var (
	image_hdr map[string][]byte = map[string][]byte{
		"bmp":   []byte{'B', 'M'},
		"gif":   []byte{'G', 'I', 'F'},
		"png":   []byte{0x89, 0x50, 0x4e, 0x47},
		"tiff":  []byte{73, 73, 42},
		"tiff2": []byte{77, 77, 42},
		"jpeg":  []byte{255, 216, 255, 224},
		"jpeg2": []byte{255, 216, 255, 225},
	}
)
