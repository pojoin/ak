package ak

import (
	"path"
	"os"
	"log"
	"net/http"
)



//定义方法类型
type actionFunc func(*Context)Render

type Context struct {
    Request *http.Request
    Params  map[string]string
}

var mainServer = NewServer()

func NewServer() *Server{
	wd, _ := os.Getwd()
	cfg := &serverConfig{}
	cfg.defaultStaticDirs = append(cfg.defaultStaticDirs,path.Join(wd,"static"))
	cfg.tplPath = path.Join(wd,"views")
	return &Server{config:cfg}
}

//启动服务
func Run(addr string){
	wd,err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(wd)
	mux := http.NewServeMux()
	mux.Handle("/",mainServer)
	err = http.ListenAndServe(addr,mux)
	if err != nil {
		log.Fatal(err)
	}
}