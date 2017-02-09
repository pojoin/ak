package ak

import (
	"fmt"
	"log"
	"net/url"
	"path"
	"regexp"
	"strings"
)

type route struct {
	regex  *regexp.Regexp
	method string
	params []string
	fn     actionFunc
}

type router struct {
	routes []*route
}

func newRouter() *router {
	rt := new(router)
	rt.routes = make([]*route, 0)
	return rt
}

func (rt *router) AddRoute(method, pattern string, fn actionFunc) {
	if rt.isRouteExists(pattern, method) {
		log.Fatalf(fmt.Sprintf("error:addRoute url[%s %s] is exist", method, pattern))
	}
	log.Printf("addRoute [%s]\t[%s]\t [%v]\n", method, pattern, fn)
	r := &route{fn: fn, method: method}
	r.regex, r.params = rt.parsePattern(pattern)
	rt.routes = append(rt.routes, r)
}

func (rt *router) parsePattern(pattern string) (regex *regexp.Regexp, params []string) {
	params = make([]string, 0)
	segments := strings.Split(url.QueryEscape(pattern), "%2F")
	for i, v := range segments {
		if strings.HasPrefix(v, "%3A") {
			segments[i] = `([\w-%]+)`
			params = append(params, strings.TrimPrefix(v, "%3A"))
		}
	}
	regex, _ = regexp.Compile("^" + strings.Join(segments, "/") + "$")
	return
}

func (rt *router) isRouteExists(pattern, method string) bool {
	sfx := path.Ext(pattern)
	pattern = strings.Replace(pattern, sfx, "", -1)
	pattern = url.QueryEscape(pattern)

	if !strings.HasSuffix(pattern, "%2F") && sfx == "" {
		pattern += "%2F"
	}
	pattern = strings.Replace(pattern, "%2F", "/", -1)
	for _, r := range rt.routes {
		if r.regex.MatchString(pattern) && r.method == method {
			return true
		}
	}
	return false
}

func (rt *router) find(pattern string, method string) (params map[string]string, fn actionFunc) {
	sfx := path.Ext(pattern)
	pattern = strings.Replace(pattern, sfx, "", -1)
	pattern = url.QueryEscape(pattern)

	if !strings.HasSuffix(pattern, "%2F") && sfx == "" {
		pattern += "%2F"
	}
	pattern = strings.Replace(pattern, "%2F", "/", -1)
	for _, r := range rt.routes {
		if r.regex.MatchString(pattern) && r.method == method {
			p := r.regex.FindStringSubmatch(pattern)
			if len(p) != len(r.params)+1 {
				continue
			}
			params = make(map[string]string)
			for i, v := range r.params {
				params[v] = p[i+1]
			}
			fn = r.fn
			return
		}
	}
	return nil, nil
}
