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
	basePath   string
	leftDelim  string
	rightDelim string
	*template.Template
}

//创建模板
func newAktemplate(name, leftDelim, rightDelim string) *aktemplate {
	akt := &aktemplate{leftDelim: leftDelim, rightDelim: rightDelim}
	t := template.New(name).Funcs(template.FuncMap{"include": akt.includeTmplate})
	akt.Template = t.Delims(akt.leftDelim, akt.rightDelim)
	return akt
}

//解析主模板
func parseTpl(w io.Writer, tname string, data interface{}, leftDelim, rightDelim string) {
	s := parseTmplateToStr(tname)
	t := newAktemplate("main", leftDelim, rightDelim)
	t.basePath = path.Dir(tname)
	_, err := t.Parse(s)
	if err != nil {
		panic(err)
	}
	t.Execute(w, data)
}

//实现include方法
func (akt *aktemplate) includeTmplate(tname string, data interface{}) (template.HTML, error) {
	tname = path.Join(akt.basePath, tname)
	s := parseTmplateToStr(tname)
	var buf bytes.Buffer
	t := newAktemplate("include", akt.leftDelim, akt.rightDelim)
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
