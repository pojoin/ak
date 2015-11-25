package ak

//定义过滤器
type Filter interface {
	Execute(ctx *Context) bool
}
