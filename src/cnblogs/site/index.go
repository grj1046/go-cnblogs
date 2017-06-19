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
	fmt.Println("site started on port: " + strconv.Itoa(HTTPPort))
	//Static Html
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("www"))))
	//API
	http.HandleFunc("/api/manage", manageHandler)
	http.HandleFunc("/api/ing", ingHandler)

	err := http.ListenAndServe(":"+strconv.Itoa(HTTPPort), nil)
	if err != nil {
		log.Fatal("ListenAndServe", err)
	}
}

func manageHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello World")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/ing", http.StatusFound)
}

func ingHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "hello this is ing page")
}
