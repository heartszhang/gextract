package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	input = "/home/hearts/Downloads/gfwlistd.txt"
)

var uri = flag.String("uri", "", "url with scheme")

func main() {
	flag.Parse()
	if len(*uri) == 0 {
		fmt.Println("gfwlist --uri http://www.google.com")
	}
	/*
		//	fmt.Println(domain_patterns, "\n", prefix_patterns, "\n", exceptional_patterns, "\n", wildcard_patterns)
		for k, v := range domain_patterns.pattern_group_m {
			if len(v) > 1 {
				fmt.Println(k, len(v))
			}
		}
	*/
	f, _ := NewGfwChecker(input, 0)
	fmt.Println(f.IsBlocked(*uri), "match-count:", match_count)
}

type GfwChecker interface {
	IsBlocked(uri string) bool
}

func NewGfwChecker(gfwlist_file string, min_pattern_length int) (GfwChecker, error) {
	gc := &gfw_checker{
		exceptional_patterns: &wildcard_group{pattern_group_m: make(map[uint32][]pattern)},
		wildcard_patterns:    &wildcard_group{pattern_group_m: make(map[uint32][]pattern)},
		prefix_patterns:      &prefix_group{pattern_group_m: make(map[uint32][]pattern)},
		domain_patterns:      &wildcard_group{pattern_group_m: make(map[uint32][]pattern)},
		regexp_patterns:      &regexp_group{exps: make([]*regexp.Regexp, 0)},
		min_pattern_length:   min_pattern_length,
	}
	if gc.min_pattern_length == 0 {
		gc.min_pattern_length = min_pattern_len
	}
	if len(gfwlist_file) == 0 {
		gfwlist_file = input
	}
	f, err := os.Open(gfwlist_file)
	if err != nil {
		fmt.Println(err)
		return gc, err
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		sline := string(bytes.TrimSpace(line))
		switch {
		case strings.HasPrefix(sline, `||`):
			if key, p, ok := new_noscheme_pattern(sline[2:]); ok {
				append_p(gc.domain_patterns.pattern_group_m, key, p)
			} else {
				gc.domain_patterns.shorts = append(gc.domain_patterns.shorts, *p)
			}
		case strings.HasPrefix(sline, `!`):
		case strings.HasPrefix(sline, `|`):
			if key, p, ok := new_scheme_pattern(sline[1:]); ok {
				append_p(gc.prefix_patterns.pattern_group_m, key, p)
			} else {
				gc.prefix_patterns.shorts = append(gc.prefix_patterns.shorts, *p)
			}
		case strings.HasPrefix(sline, `@@||`):
			if key, r, ok := new_noscheme_pattern(sline[4:]); ok {
				append_p(gc.exceptional_patterns.pattern_group_m, key, r)
			} else {
				gc.exceptional_patterns.shorts = append(gc.exceptional_patterns.shorts, *r)
			}
		case strings.HasPrefix(sline, `[`):
		case strings.HasPrefix(sline, `/`) && strings.HasSuffix(sline, `/`):
			re, err := regexp.Compile(strings.Trim(sline, "/"))
			if err == nil {
				gc.regexp_patterns.exps = append(gc.regexp_patterns.exps, re)
			}
		case len(sline) == 0:
		default:
			if key, p, ok := new_scheme_pattern(sline); ok {
				append_p(gc.wildcard_patterns.pattern_group_m, key, p)
			}
		}
	}
	return gc, nil
}

type gfw_checker struct {
	exceptional_patterns *wildcard_group //= &wildcard_group{pattern_group_m: make(map[uint32][]pattern)}
	wildcard_patterns    *wildcard_group //= &wildcard_group{pattern_group_m: make(map[uint32][]pattern)}
	prefix_patterns      *prefix_group   //= &prefix_group{pattern_group_m: make(map[uint32][]pattern)}
	domain_patterns      *wildcard_group //= &wildcard_group{pattern_group_m: make(map[uint32][]pattern)}
	regexp_patterns      *regexp_group   //= &regexp_group{exps: make([]*regexp.Regexp, 0)}
	min_pattern_length   int
}

var match_count int = 0

const prime_rk = 16777619
const min_pattern_len = 5 // 前两个字符中存在通配符的时候，结果就会出现错误

func (this *gfw_checker) IsBlocked(uri string) bool {
	scheme, hostpath, host := url_parse(uri)

	if r := this.exceptional_patterns.match(scheme, host); r {
		return false
	}
	if r := this.prefix_patterns.match(scheme, hostpath); r {
		return true
	}
	if r := this.domain_patterns.match(scheme, host); r {
		return true
	}
	if r := this.wildcard_patterns.match(scheme, hostpath); r {
		return true
	}
	/*  //效率降低太多
	if r := regexp_patterns.match(scheme, uri); r {
		return true
	}
	*/
	return false
}

func make_noscheme_pattern(line string) *noscheme_pattern {
	p := &noscheme_pattern{sep: strings.Trim(line, "*"), fields: []string{}}
	//根据通配符，将模式串转换为数个字符段，消除其中的连续空串
	for _, fd := range strings.Split(p.sep, "*") {
		if len(fd) > 0 {
			p.fields = append(p.fields, fd)
		}
	}
	if len(p.fields) > 0 {
		p.has_misterik = true
	}
	return p
}

func new_noscheme_pattern(line string) (key uint32, p *noscheme_pattern, ok bool) {
	p = &noscheme_pattern{sep: strings.Trim(line, "*")}
	if len(line) < min_pattern_len {
		fmt.Println(line)
		return
	}
	return make_key(p.sep), p, true
}

func new_scheme_pattern(line string) (key uint32, p *scheme_pattern, ok bool) {
	scheme, hostpath, _ := url_parse(line)
	if len(scheme) == 0 {
		scheme = "http"
	}
	p = &scheme_pattern{noscheme_pattern: *make_noscheme_pattern(hostpath), scheme: scheme}

	if len(p.sep) < min_pattern_len {
		fmt.Println("ignore-scheme:", line)
		return
	}

	return make_key(p.sep), p, true
}

type pattern interface {
	match(scheme, hostpath string) bool
}

type scheme_pattern struct {
	noscheme_pattern
	scheme string
}

func (this *scheme_pattern) match(scheme, hostpath string) bool {
	if scheme != this.scheme {
		//		fmt.Println("scheme unmatch", hostpath, this.sep)
		return false
	}
	ok := this.noscheme_pattern.match(scheme, hostpath)
	//	fmt.Println("sch", this.sep, "vs", hostpath, "=", ok)
	return ok
}

type noscheme_pattern struct {
	sep          string
	has_misterik bool
	fields       []string
}

func (this *noscheme_pattern) match(scheme, hostpath string) bool {
	match_count++
	ok := false
	switch {
	case !this.has_misterik:
		ok = strings.HasPrefix(hostpath, this.sep)
	case strings.HasPrefix(hostpath, this.fields[0]):
		hostpath = hostpath[len(this.fields[0]):]
		ok = match_fields(hostpath, this.fields[1:])
	default:
		ok = false
	}
	//	fmt.Println("noscheme", this.sep, "vs", hostpath, "=", ok)
	return ok
}

func match_fields(s string, fields []string) bool {
	if len(fields) == 0 {
		return true
	}
	idx := strings.Index(s, fields[0])
	if idx < 0 {
		return false
	}
	s = s[idx+len(fields[0]):]
	return match_fields(s, fields[1:])
}

type pattern_group_m map[uint32][]pattern

func append_p(group pattern_group_m, key uint32, p pattern) {
	group[key] = append(group[key], p)
}

/*
type exceptional_group struct {
  pattern_group
}
*/

type wildcard_group struct {
	pattern_group_m
	shorts []noscheme_pattern
}

type prefix_group struct {
	pattern_group_m
	shorts []scheme_pattern
}
type regexp_group struct {
	exps []*regexp.Regexp
}

func (this *regexp_group) match(scheme, hostpath string) bool {
	for _, re := range this.exps {
		match_count++
		if re.MatchString(hostpath) {
			return true
		}
	}
	return false
}

func (this *prefix_group) match(scheme, hostpath string) bool {
	// ignore shorts
	if len(hostpath) < min_pattern_len {
		for _, s := range this.shorts {
			if s.match(scheme, hostpath) {
				return true
			}
		}
		return false
	}
	k := make_key(hostpath)
	for _, p := range this.pattern_group_m[k] {
		if p.match(scheme, hostpath) {
			return true
		}
	}
	return false
}

func (this *wildcard_group) match(scheme, hostpath string) bool {
	// ignore shorts
	if len(hostpath) < min_pattern_len {
		for _, s := range this.shorts {
			if s.match(scheme, hostpath) {
				return true
			}
		}
		return false
	}
	for i := 0; i <= len(hostpath)-min_pattern_len; i++ {
		s := hostpath[i:]
		k := make_key(s)
		for _, p := range this.pattern_group_m[k] {
			if p.match(scheme, s) {
				return true
			}
		}
	}
	return false
}

func url_parse(uri string) (scheme, hostpath, host string) {
	idx := strings.Index(uri, "://")
	if idx < 0 {
		hostpath = uri
	} else {
		scheme = uri[:idx]
		hostpath = uri[idx+3:]
	}
	idx = strings.Index(hostpath, "/")
	if idx < 0 {
		host = hostpath
	} else {
		host = hostpath[:idx]
	}
	return
}

//rabin-k hash algorithm
func make_key(sep string) uint32 {
	h := uint32(0)
	for i := 0; i < min_pattern_len; i++ {
		h = h*prime_rk + uint32(sep[i])
	}
	return h
}
