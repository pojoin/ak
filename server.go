package ak

import (
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
	spool *spool
}

//添加路由
func (s *Server) addRoute(url string, f actionFunc){
	for _, route := range s.routes {
		if route.r == url {
			log.Fatal(url, "is exists")
			return
		}
	}
	s.routes = append(mainServer.routes, route{r: url, handler: f})
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
	ctx := Context{Request:req,
			ResponseWriter: w, 
			Params:params,
			Data:make(map[string]interface{}),
			server:mainServer}
	cookie,err := req.Cookie(sessionIdKey)
	if err == nil {
		session,ok := s.spool.getSession(cookie.Value)
		if ok {
			ctx.Session = session
		}else{
			ctx.Session = newSession()
			s.spool.addSession(ctx.Session)
			cookie = &http.Cookie{Name:sessionIdKey,Value:ctx.Session.sessionId}
			http.SetCookie(w,cookie)
		}
	}else{
		ctx.Session = newSession()
		s.spool.addSession(ctx.Session)
		cookie = &http.Cookie{Name:sessionIdKey,Value:ctx.Session.sessionId}
		http.SetCookie(w,cookie)
	}
	
	
	//路由配置查询
	for _, route := range s.routes {
		if rp == route.r {
			invoke(route.handler, &ctx)
			return
		}
	}
	//请求不存在，404错误
	ctx.Abort(404,"page not fond")
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
func invoke(function actionFunc, ctx *Context) {
	defer func(){
		if err := recover(); err != nil {
			ctx.Abort(500,"server error")
		}
	}()
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
	s.l = l
	err = http.Serve(s.l, mux)
	s.l.Close()
}
