package ak

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"path"
	"time"
)

type Method int

const (
	GET     = 0x00002000
	POST    = 0x00004000
	PUT     = 0x00008000
	PATCH   = 0x00020000
	DELETE  = 0x00040000
	HEAD    = 0x00080000
	OPTIONS = 0x00200000
	CONNECT = 0x00400000
	TRACE   = 0x00800000
)

func (m Method) String() string {
	var r string
	switch int(m) {
	case GET:
		r = http.MethodGet
	case POST:
		r = http.MethodPost
	case PUT:
		r = http.MethodPut
	case PATCH:
		r = http.MethodPatch
	case DELETE:
		r = http.MethodDelete
	case HEAD:
		r = http.MethodHead
	case OPTIONS:
		r = http.MethodOptions
	case CONNECT:
		r = http.MethodConnect
	case TRACE:
		r = http.MethodTrace
	}
	return r
}

//服务配置
type serverConfig struct {
	basePath          string
	tplPath           string
	leftDelim         string
	rightDelim        string
	defaultStaticDirs []string
	sessionProc       bool //是否开始session处理
	profiler          bool
}

//服务
type Server struct {
	router      *router
	l           net.Listener
	config      *serverConfig
	spool       *spool
	filterChain []Filter
}

//添加路由
func (s *Server) AddRoute(methods int, url string, f actionFunc) {
	if GET == GET&methods {
		s.router.AddRoute(http.MethodGet, url, f)
	}
	if POST == POST&methods {
		s.router.AddRoute(http.MethodPost, url, f)
	}
	if PUT == PUT&methods {
		s.router.AddRoute(http.MethodPut, url, f)
	}

	if PATCH == PATCH&methods {
		s.router.AddRoute(http.MethodPatch, url, f)
	}
	if DELETE == DELETE&methods {
		s.router.AddRoute(http.MethodDelete, url, f)
	}
	if HEAD == HEAD&methods {
		s.router.AddRoute(http.MethodHead, url, f)
	}
	if OPTIONS == OPTIONS&methods {
		s.router.AddRoute(http.MethodOptions, url, f)
	}
	if CONNECT == CONNECT&methods {
		s.router.AddRoute(http.MethodConnect, url, f)
	}
	if TRACE == TRACE&methods {
		s.router.AddRoute(http.MethodTrace, url, f)
	}
}

//添加过滤器
func (s *Server) AddFilter(filter Filter) {
	s.filterChain = append(s.filterChain, filter)
}

//设置模板路径
func (s *Server) SetTplPath(tplPath string) {
	if s.config == nil {
		s.config = &serverConfig{}
		wd, _ := os.Getwd()
		s.config.basePath = wd
	}
	s.config.tplPath = tplPath
}

//添加静态资源文件夹
func (s *Server) AddStaticDir(staticDir string) {
	if s.config == nil {
		s.config = &serverConfig{}
		wd, _ := os.Getwd()
		s.config.basePath = wd
	}
	s.config.defaultStaticDirs = append(s.config.defaultStaticDirs, path.Join(s.config.basePath, staticDir))
}

//设置模板标签边界
func (s *Server) SetTplDelim(leftDelim, rightDelim string) {
	if s.config == nil {
		s.config = &serverConfig{}
	}
	s.config.leftDelim = leftDelim
	s.config.rightDelim = rightDelim
}

//是否开始session处理
func (s *Server) StartSession(state bool) {
	s.config.sessionProc = state
	if s.config.sessionProc {
		if s.spool == nil {
			s.spool = newspool()
		}
	}
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
	st := time.Now()
	isStatic := true
	ctx := &Context{
		Request:        req,
		ResponseWriter: w,
		Data:           make(map[string]interface{}),
		server:         s,
		status:         http.StatusOK,
		closed:         false,
	}
	rp := req.URL.Path
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			ctx.status = http.StatusInternalServerError
			ctx.Abort(ctx.status, fmt.Sprint(err))
		}
		if isStatic {
		} else {
			dis := time.Now().Sub(st).Seconds() * 1000
			log.Printf("%s [%d] [%.3f ms] %s\n", req.Method, ctx.status, dis, rp)
		}
		ctx = nil
	}()
	//	io.WriteString(w,"URL:" + rp)
	//静态文件请求处理
	if req.Method == http.MethodGet || req.Method == http.MethodHead {
		if s.tryServingFile(rp, req, w) {
			return
		}
	}
	isStatic = false
	//log.Println(req.Method, rp)
	ctx.Params = parseParam(req)
	//session处理
	//获取cookie
	if s.config.sessionProc {
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
	}

	//路由配置查询
	params, fn := s.router.find(rp, req.Method)
	if params != nil && fn != nil {
		for k, v := range params {
			ctx.Params[k] = v
		}
		s.invoke(fn, ctx)
	} else {
		//请求不存在，404错误
		ctx.Abort(404, "["+rp+"] page not fond")
	}

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
	log.Println("server is start runing : ", addr)
	s.l = l
	err = http.Serve(s.l, mux)
	s.l.Close()
	log.Println("server is stoped")
}
