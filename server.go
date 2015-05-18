package ak

import (
	"log"
	"net/http"
	"net"
	"reflect"
)

//定义简单路由
type route struct{
	r string
	handler reflect.Value //调用执行函数
}


//服务
type Server struct{
	routes []route
	l net.Listener
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request){
	io.WriteString(w,"URL:" + r.URL.Path)
}

//启动服务
func (s *Server) Run(addr string){
	mux := http.NewServeMux()
	mux.Handle("/", s)
	l,err := net.Listen("tcp",addr)
	if err != nil {
		log.Fatal(err)
	}
	s.l = l
	err = http.Serve(s.l,mux)
	s.l.Close()
}

