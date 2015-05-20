package ak

import (
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"net/http/pprof"
)

//定义简单路由
type route struct {
	r       string
	handler actionFunc
}

//服务配置
type serverConfig struct {
	tplPath string
	defaultStaticDirs []string
	profiler bool
}

//服务
type Server struct {
	routes []route
	l      net.Listener
	config *serverConfig
}

//查看是否是静态请求
func (s *Server) tryServingFile(name string, req *http.Request, w http.ResponseWriter) bool {
	for _, staticDir := range s.config.defaultStaticDirs {
		staticFile := path.Join(staticDir, name)
		if fileExists(staticFile) {
			http.ServeFile(w, req, staticFile)
			return true
		}
	}
	return false
}

//判断文件是否存在
func fileExists(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

//继承http.Handler实现ServeHTTP方法
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.process(w, req)
}

//请求处理
func (s *Server) process(w http.ResponseWriter, req *http.Request) {
	rp := req.URL.Path
	log.Println(rp)
	//	io.WriteString(w,"URL:" + rp)
	//静态文件请求处理
	if req.Method == "GET" || req.Method == "HEAD" {
		if s.tryServingFile(rp, req, w) {
			return
		}
	}
	//路由配置查询
	params := parseParam(req)
	ctx := Context{req, params}
	for _, route := range s.routes {
		if rp == route.r {
			invoke(route.handler, &ctx,w)
			return
		}
	}
	//请求不存在，404错误
	io.WriteString(w, "404")
}

//渲染结果
func render(r Render, w http.ResponseWriter,ctx *Context) {
	switch r.renderType {
	case Tpl: //模板渲染
		renderTpl(r,w)
	case Json: //json渲染
		io.WriteString(w, r.jsonStr)
	case Redirect: // 重定向到
		w.Header().Set("Location", r.rUrl)
	    w.WriteHeader(301)
	    w.Write([]byte("Redirecting to: " + r.rUrl))
//		renderRedirect(r.rUrl,ctx,w)
	}
}

//渲染模板
func renderTpl(r Render,w http.ResponseWriter){
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tplPath := path.Join(mainServer.config.tplPath,r.rUrl)
	if !fileExists(tplPath) {
		io.WriteString(w, "404")
	}
	tpl,err := template.ParseFiles(tplPath)
	if err != nil {
		panic(err.Error())
	}
	tpl.Execute(w,r.data)
}

//渲染重定向
func renderRedirect(url string,ctx *Context,w http.ResponseWriter){
	for _, route := range mainServer.routes {
		if url == route.r {
			invoke(route.handler, ctx,w)
			return
		}
	}
	//请求不存在，404错误
	io.WriteString(w, "404")
}

//提取参数
func parseParam(req *http.Request) map[string]string {
	params := make(map[string]string)
	req.ParseForm()
	for k, v := range req.Form {
		params[k] = v[0]
	}
	return params
}

//调用自定义方法
func invoke(function actionFunc, ctx *Context,w http.ResponseWriter) {
	defer func(){
		if err := recover(); err != nil {
			io.WriteString(w,"500")
			log.Println(err)
		}
	}()
	r := function(ctx)
	render(r, w,ctx)
}

func Get(url string, f actionFunc) {
	for _, route := range mainServer.routes {
		if route.r == url {
			log.Fatal(url, "is exists")
			return
		}
	}
	mainServer.routes = append(mainServer.routes, route{r: url, handler: f})
}

//启动服务
func (s *Server) Run(addr string) {
	mux := http.NewServeMux()
	if s.config.profiler {
        mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
        mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
        mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
        mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
    }
	mux.Handle("/", s)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	s.l = l
	err = http.Serve(s.l, mux)
	s.l.Close()
}
