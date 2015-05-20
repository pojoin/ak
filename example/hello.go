package main

import(
	"github.com/hechuangqiang/ak"
)

func main(){
	ak.Get("/hello",func(ctx *ak.Context) ak.Render{
		return ak.JsonRender("哈哈")
	})
	ak.Get("/test",func(ctx *ak.Context) ak.Render{
		data := ak.RenderData{}
		data["name"] = "张三"
		return ak.TplRender("text.html",data)
	})
	
	ak.Get("/user",func(ctx *ak.Context) ak.Render{
		data := ak.RenderData{}
		data["users"] = []string{"张三","李四","王五"}
		return ak.TplRender("user/user.htm",data)
	})
	ak.Get("/redirct",func(ctx *ak.Context) ak.Render{
		return ak.RedirectRender("/user")
	})
	ak.Run(":9000")
}