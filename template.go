package ak

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"path"
)

type aktemplate struct {
	basePath string
	*template.Template
}

//创建模板
func newAktemplate(name string) *aktemplate {
	t := template.New(name).Funcs(template.FuncMap{"include": includeTmplate})
	akt := &aktemplate{}
	akt.Template = t
	return akt
}

//解析主模板
func parseTpl(w io.Writer, tname string, data interface{}) {
	s := parseTmplateToStr(tname)
	t := newAktemplate("main")
	t.basePath = path.Dir(tname)
	_, err := t.Parse(s)
	if err != nil {
		panic(err)
	}
	t.Execute(w, data)
}

//实现include方法
func includeTmplate(tname string, data interface{}) (template.HTML, error) {
	tname = path.Join(mainServer.config.tplPath, tname)
	s := parseTmplateToStr(tname)
	var buf bytes.Buffer
	t := newAktemplate("include")
	_, err := t.Parse(s)
	if err != nil {
		return "", err
	}
	t.Execute(&buf, data)
	return template.HTML(buf.String()), nil
}

//解析文件
func parseTmplateToStr(tname string) string {
	b, err := ioutil.ReadFile(tname)
	if err != nil {
		log.Println(err)
	}
	s := string(b)
	return s
}
