package ak

import (
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"path"
	"time"
)

//定义简单路由
type route struct {
	r       string
	handler actionFunc
}

//服务配置
type serverConfig struct {
	basePath		string
	tplPath           string
	leftDelim	string
	rightDelim	string
	defaultStaticDirs []string
	profiler          bool
}

//服务
type Server struct {
	routes      []route
	l           net.Listener
	config      *serverConfig
	spool       *spool
	filterChain []Filter
}

//添加路由
func (s *Server) AddRoute(url string, f actionFunc) {
	for _, route := range s.routes {
		if route.r == url {
			log.Fatal(url, "is exists")
			return
		}
	}
	s.routes = append(s.routes, route{r: url, handler: f})
}

//批量添加路由
func (s *Server) AddRoutes(routeMap map[string]actionFunc){
	for k,v := range routeMap{
		s.AddRoute(k,v)
	}
}

//添加过滤器
func (s *Server) AddFilter(filter Filter){
	s.filterChain = append(s.filterChain, filter)
}

//添加静态资源文件夹
func (s *Server) AddStaticDir(staticDir string){
	if s.config == nil {
		s.config = &serverConfig{}
		wd, _ := os.Getwd()
		s.config.basePath = wd
	}
	s.config.defaultStaticDirs = append(s.config.defaultStaticDirs, path.Join(s.config.basePath, staticDir))
}

//设置模板标签边界
func(s *Server) SetTplDelim(leftDelim,rightDelim string){
	if s.config == nil{
		s.config = &serverConfig{}
	}
	s.config.leftDelim = leftDelim
	s.config.rightDelim = rightDelim
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
	params := parseParam(req)
	//session处理
	ctx := Context{Request: req,
		ResponseWriter: w,
		Params:         params,
		Data:           make(map[string]interface{}),
		server:         s}
	//获取cookie
	cookie, err := req.Cookie(sessionIdKey)
	if err == nil { //如果cookie 存在则将cookie转成session
		session, ok := s.spool.getSession(cookie.Value)
		if ok {
			ctx.Session = session
			ctx.Session.t = time.Now()
		} else {
			ctx.Session = newSession()
			s.spool.addSession(ctx.Session)
			cookie = &http.Cookie{Name: sessionIdKey, Value: ctx.Session.sessionId}
			http.SetCookie(w, cookie)
		}
	} else { //如果cookie不存在这个创建cookie
		ctx.Session = newSession()
		s.spool.addSession(ctx.Session)
		cookie = &http.Cookie{Name: sessionIdKey, Value: ctx.Session.sessionId}
		http.SetCookie(w, cookie)
	}

	//路由配置查询
	for _, route := range s.routes {
		log.Println("rp = ", rp, ",route.r = ", route.r)
		if rp == route.r {
			log.Println("路由解析成功 开始执行路由...")
			s.invoke(route.handler, &ctx)
			return
		}
	}
	//请求不存在，404错误
	ctx.Abort(404, "page not fond")
}

//提取参数 将请求参数转换成map类型
func parseParam(req *http.Request) map[string]string {
	params := make(map[string]string)
	req.ParseForm()
	for k, v := range req.Form {
		params[k] = v[0]
	}
	return params
}

//调用自定义方法
func (s *Server) invoke(function actionFunc, ctx *Context) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			ctx.Abort(500, "server error")
		}
	}()

	//执行过滤器
	for _, filter := range s.filterChain {
		if !filter.Execute(ctx) {
			return
		}
	}
	function(ctx)
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
	log.Println("server is start runing : ",addr)
	s.l = l
	err = http.Serve(s.l, mux)
	s.l.Close()
	log.Println("server is stoped")
}
