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
func (s *Server) AddRoute(method, url string, f actionFunc) {
	s.router.AddRoute(method, url, f)
}

//添加过滤器
func (s *Server) AddFilter(filter Filter) {
	s.filterChain = append(s.filterChain, filter)
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
	ctx := &Context{
		Request:        req,
		ResponseWriter: w,
		Data:           make(map[string]interface{}),
		server:         s,
		closed:         false,
	}
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			ctx.Abort(500, fmt.Sprint(err))
		}
		ctx = nil
	}()
	rp := req.URL.Path
	log.Println(req.Method, rp)
	//	io.WriteString(w,"URL:" + rp)
	//静态文件请求处理
	if req.Method == "GET" || req.Method == "HEAD" {
		if s.tryServingFile(rp, req, w) {
			return
		}
	}

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
		return
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
