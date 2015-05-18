package ak

import (
	"os"
	"log"
	"net/http"
)

//启动服务
func Run(){
	wd,err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(wd)
	mux := http.NewServeMux()
	mux.Handle("/",&actionHandler{})
	err = http.ListenAndServe(":9000",mux)
	if err != nil {
		log.Fatal(err)
	}
}