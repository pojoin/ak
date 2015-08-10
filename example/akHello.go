package main

import (
	"github.com/hechuangqiang/ak"
)

func main() {
	s := ak.NewDefaultServer()
	s.AddRoute("/", func(c *ak.Context) {
		c.WriteTpl("text.html")
	})
	s.Run(":9000")
}
