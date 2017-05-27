package ak

import (
	"encoding/json"
	"net/http"
	"path"
	"reflect"
	"strconv"
)

//定义方法类型
type actionFunc func(*Context)

//请求上下文
type Context struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Params         map[string]string
	Data           map[string]interface{} //返回参数定义
	server         *Server
	Session        *Session
	status         int //http状态
	closed         bool
}

func (ctx *Context) GetArgStr(param, def string) string {
	if v, ok := ctx.Params[param]; ok {
		return v
	} else {
		return def
	}
}

func (ctx *Context) GetArgInt(param string, def int) int {
	if v, ok := ctx.Params[param]; ok {
		i, err := strconv.Atoi(v)
		if err != nil {
			return def
		}
		return i
	} else {
		return def
	}
}

func (ctx *Context) IsArgExists(param string) bool {
	_, ok := ctx.Params[param]
	return ok
}

func (ctx *Context) SetStatus(status int) {
	ctx.status = status
}

//获取参数int
func (ctx *Context) ArgumentInt(name string) (int, error) {
	return strconv.Atoi(ctx.Params[name])
}

//获取参数string
func (ctx *Context) ArgumentString(name string) string {
	return ctx.Params[name]
}

//返回json
func (ctx *Context) WriteJson(content interface{}) {
	if ctx.closed {
		return
	}
	ctx.ResponseWriter.WriteHeader(ctx.status)
	ctx.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	cv := reflect.ValueOf(content)
	vkind := cv.Type().Kind()
	if vkind == reflect.String {
		ctx.ResponseWriter.Write([]byte(cv.String()))
	} else if vkind == reflect.Struct || vkind == reflect.Slice || vkind == reflect.Ptr || vkind == reflect.Map {
		jsonData, err := json.Marshal(content)
		if err != nil {
			panic(err)
		}
		ctx.ResponseWriter.Write(jsonData)
	}
	ctx.closed = true
}

//字符流
func (ctx *Context) WriteStream(filename string, contentType string, data []byte) {
	if ctx.closed {
		return
	}
	//ctx.ResponseWriter.WriteHeader(ctx.status)
	ctx.ResponseWriter.Header().Add("Content-Disposition", "attachment;filename="+filename)
	ctx.ResponseWriter.Header().Set("Content-Type", contentType)
	ctx.ResponseWriter.Header().Set("Content-Length", strconv.Itoa(len(data)))
	ctx.ResponseWriter.Write(data)
	ctx.closed = true
}

//返回html
func (ctx *Context) WriteHtml(htmlStr string) {
	if ctx.closed {
		return
	}
	ctx.ResponseWriter.WriteHeader(ctx.status)
	ctx.ResponseWriter.Write([]byte(htmlStr))
	ctx.closed = true
}

//返回模板
func (ctx *Context) WriteTpl(tplName string) {
	if ctx.closed {
		return
	}
	ctx.ResponseWriter.WriteHeader(ctx.status)
	tplPath := path.Join(ctx.server.config.tplPath, tplName)
	if !fileExists(tplPath) {
		ctx.Abort(http.StatusNotFound, tplName+" not fond")
	}
	parseTpl(ctx.ResponseWriter, tplPath, ctx.Data, ctx.server.config.leftDelim, ctx.server.config.rightDelim)
	//	s := parseTmplateToStr(tplPath)
	//	t, err := template.New("index").Funcs(template.FuncMap{"include": includeTmplate}).Parse(s)
	//	tpl, err := template.ParseFiles(tplPath)
	//	if err != nil {
	//		panic(err.Error())
	//	}
	//	t.Execute(ctx.ResponseWriter, ctx.Data)
	ctx.closed = true
}

//重定向 3xx
func (ctx *Context) Redirect(url_ string) {
	if ctx.closed {
		return
	}
	ctx.status = http.StatusMovedPermanently
	ctx.ResponseWriter.Header().Set("Location", url_)
	ctx.ResponseWriter.WriteHeader(http.StatusMovedPermanently)
	ctx.ResponseWriter.Write([]byte("Redirecting to: " + url_))
	ctx.closed = true
}

//错误信息
func (ctx *Context) Abort(status int, body string) {
	if ctx.closed {
		return
	}
	ctx.status = status
	ctx.ResponseWriter.WriteHeader(status)
	ctx.ResponseWriter.Write([]byte(body))
	ctx.closed = true
}

//将请求数据的json格式信息转换成结构
func (ctx *Context) ParseReqBodyJson(obj interface{}) error {
	decoder := json.NewDecoder(ctx.Request.Body)
	k := reflect.ValueOf(obj).Type().Kind()
	var err error
	switch k {
	case reflect.Struct:
		err = decoder.Decode(&obj)
	case reflect.Ptr:
		err = decoder.Decode(obj)
	case reflect.Array:
		err = decoder.Decode(obj)
	case reflect.String:
		err = decoder.Decode(&obj)
	}
	return err
}
