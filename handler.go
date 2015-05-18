package ak

import(
	"io"
	"net/http"
)

type actionHandler struct{
	*Action
}

func (h *actionHandler) ServeHTTP(w http.ResponseWriter,r *http.Request){
	io.WriteString(w,"URL:" + r.URL.Path)
}