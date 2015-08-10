package ak

import (
	"os"
	"path"
)



//var DefaultServer = NewServer()

func NewDefaultServer() *Server {
	wd, _ := os.Getwd()
	cfg := &serverConfig{}
	cfg.basePath = wd
	cfg.profiler = true
	cfg.leftDelim = "{{"
	cfg.rightDelim = "}}"
	cfg.defaultStaticDirs = append(cfg.defaultStaticDirs, path.Join(wd, "web"))
	cfg.tplPath = path.Join(wd, "web")
	return &Server{config: cfg, spool: newspool(), filterChain: make([]Filter, 0)}
}

////启动服务
//func Run(s *Server,addr string) {
//	wd, err := os.Getwd()
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Println(wd)
//	mux := http.NewServeMux()
//	mux.Handle("/", s)
//	err = http.ListenAndServe(addr, mux)
//	if err != nil {
//		log.Fatal(err)
//	}
//}
