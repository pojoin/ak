package ak

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"reflect"
	"strconv"
	"io/ioutil"
)

//定义方法类型
type actionFunc func(*Context)

type Context struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Params         map[string]string
	Data           map[string]interface{} //返回参数定义
	server         *Server
	Session        *Session
}

//返回json
func (ctx *Context) WriteJson(content interface{}) {
	ctx.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	cv := reflect.ValueOf(content)
	if cv.Type().Kind() == reflect.String {
		ctx.ResponseWriter.Write([]byte(cv.String()))
	}
}

//字符流
func (ctx *Context) WriteStream(filename string, contentType string, data []byte) {
	ctx.ResponseWriter.Header().Add("Content-Disposition", "attachment;filename="+filename)
	ctx.ResponseWriter.Header().Set("Content-Type", contentType)
	ctx.ResponseWriter.Header().Set("Content-Length", strconv.Itoa(len(data)))
	ctx.ResponseWriter.Write(data)
}

//返回html
func(ctx *Context) WriteHtml(html string){
	ctx.ResponseWriter.Write([]byte(html))
}

//返回模板
func (ctx *Context) WriteTpl(tplName string) {
	tplPath := path.Join(ctx.server.config.tplPath, tplName)
	if !fileExists(tplPath) {
		ctx.Abort(404, tplName+" not fond")
	}
	s := parseTmplateToStr(tplPath)
	t, err := template.New("index").Funcs(template.FuncMap{"include": includeTmplate }).Parse(s)
//	tpl, err := template.ParseFiles(tplPath)
	if err != nil {
		panic(err.Error())
	}
	t.Execute(ctx.ResponseWriter, ctx.Data)
}

func includeTmplate(tname string) template.HTML{
	tname = path.Join(mainServer.config.tplPath,tname)
	return template.HTML(parseTmplateToStr(tname))
}

func parseTmplateToStr(tname string) string {
	log.Println("tname = ",tname)
	b, err := ioutil.ReadFile(tname) 
    if err != nil {
            log.Println(err)
    }
    s := string(b)  
    return s
}

//重定向 3xx
func (ctx *Context) Redirect(url_ string) {
	ctx.ResponseWriter.Header().Set("Location", url_)
	ctx.ResponseWriter.WriteHeader(301)
	ctx.ResponseWriter.Write([]byte("Redirecting to: " + url_))
}

//错误信息
func (ctx *Context) Abort(status int, body string) {
	ctx.ResponseWriter.WriteHeader(status)
	ctx.ResponseWriter.Write([]byte(body))
}

//定义过滤器
type Filter interface{
	Execute(ctx *Context) bool
}

var mainServer = NewServer()

func NewServer() *Server {
	wd, _ := os.Getwd()
	cfg := &serverConfig{}
	cfg.profiler = true
	cfg.defaultStaticDirs = append(cfg.defaultStaticDirs, path.Join(wd, "static"))
	cfg.tplPath = path.Join(wd, "views")
	return &Server{config: cfg, spool: newspool(),filterChain: make([]Filter,0)}
}

//添加路由
func AddRoute(url string, f actionFunc) {
	mainServer.addRoute(url, f)
}

//添加过滤器
func AddFilter(filter Filter){
	mainServer.filterChain = append(mainServer.filterChain,filter)
}

//启动服务
func Run(addr string) {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(wd)
	mux := http.NewServeMux()
	mux.Handle("/", mainServer)
	err = http.ListenAndServe(addr, mux)
	if err != nil {
		log.Fatal(err)
	}
}
