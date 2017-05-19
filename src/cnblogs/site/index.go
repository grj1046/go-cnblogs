package site

import (
	"cnblogs/conf"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
)

//Main cnblogs Site main function
func Main(conf conf.Conf) {
	HTTPPort := conf.HTTPPort
	if HTTPPort <= 0 {
		HTTPPort = 8080
	}
	http.HandleFunc("/manage", manageHandler)
	err := http.ListenAndServe(":"+strconv.Itoa(HTTPPort), nil)
	if err != nil {
		log.Fatal("ListenAndServe", err)
	}
	fmt.Println("site started on port: " + strconv.Itoa(HTTPPort))
}

func manageHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello World")
}
