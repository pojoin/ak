package main

import (
	"io/ioutil"
	"log"
	"os"
)

import (
	"github.com/hechuangqiang/ak"
)

type loginFilter struct{}

func (l *loginFilter) Execute(ctx *ak.Context) (ok bool) {
	ok = false
	log.Println("loginFilter")
	ctx.WriteJson("没有权限")
	return
}

func main() {

	ak.AddFilter(&loginFilter{})

	ak.AddRoute("GET", "/hello/:name/ok/", func(ctx *ak.Context) {
		ctx.WriteJson("hello , " + ctx.Params["name"])
	})

	ak.AddRoute("GET", "/test/", func(ctx *ak.Context) {
		ctx.Data["name"] = "张三"
		ctx.WriteTpl("text.html")
	})

	ak.AddRoute("GET", "/user/", func(ctx *ak.Context) {
		ctx.Data["users"] = []string{"张三", "李四", "王五"}
		ctx.WriteTpl("user/user.htm")
	})

	ak.AddRoute("GET", "/redirct/", func(ctx *ak.Context) {
		ctx.Redirect("/user")
	})

	ak.AddRoute("GET", "/panic/", func(ctx *ak.Context) {
		panic("ok")
	})

	ak.AddRoute("GET", "/download/", func(ctx *ak.Context) {
		f, err := os.Open("hello.go")
		if err != nil {
			ctx.Abort(500, "open file fail")
			return
		}
		defer f.Close()
		buf, _ := ioutil.ReadAll(f)
		ctx.WriteStream("1t.html", "application/octet-stream", buf)
	})

	ak.RunSimpleServer(":9000")
}
