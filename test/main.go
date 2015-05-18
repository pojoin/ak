package main

import (
	"log"
	"os"
	"io"
	"net/http"
)

type myHandle struct{}
func (*myHandle) ServeHTTP(w http.ResponseWriter,r *http.Request){
	io.WriteString(w,"URL:" + r.URL.Path)
}

func sayHello(w http.ResponseWriter,r *http.Request){
	io.WriteString(w,"hello world,this is version 1")
}

func main(){
	mux := http.NewServeMux()
	mux.Handle("/",&myHandle{})
	mux.HandleFunc("/hello",sayHello)
	
	wd,err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	
	mux.Handle("/static/",
			http.StripPrefix("/s",
					http.FileServer(http.Dir(wd))))
	err = http.ListenAndServe(":9000",mux)
	if err != nil {
		log.Fatal(err)
	}
}
