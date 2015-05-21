package main

import (
	"io/ioutil"
	"os"
)

import(
	"github.com/hechuangqiang/ak"
)

type loginFilter struct{}

func (l *loginFilter) Execute(ctx *ak.Context) (ok bool) {
	ok = true
	return 
}

func main(){
	
	ak.AddFilter(&loginFilter{})
	
	ak.AddRoute("/hello",func(ctx *ak.Context) {
		ctx.WriteJson("哈哈")
	})
	
	ak.AddRoute("/test",func(ctx *ak.Context){
		ctx.Data["name"] = "张三"
		ctx.WriteTpl("text.html")
	})
	
	ak.AddRoute("/user",func(ctx *ak.Context) {
		ctx.Data["users"] = []string{"张三","李四","王五"}
		ctx.WriteTpl("user/user.htm")
	})
	
	ak.AddRoute("/redirct",func(ctx *ak.Context){
		ctx.Redirect("/user")
	})
	
	ak.AddRoute("/panic",func(ctx *ak.Context) {
		panic("ok")
	})
	
	ak.AddRoute("/download",func(ctx *ak.Context){
		f,err := os.Open("hello.go")
		if err != nil {
			ctx.Abort(500,"open file fail")
			return
		}
		defer f.Close()
		buf,_ := ioutil.ReadAll(f)
		ctx.WriteStream("1t.html","application/octet-stream",buf)
	})
	
	ak.Run(":9000")
}