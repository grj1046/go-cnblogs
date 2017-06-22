package site

import (
	"cnblogs/conf"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
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
	r := mux.NewRouter()
	r.HandleFunc("/api/manage", manageHandler)
	r.HandleFunc("/api/ings/{p}", ingsHandler)
	r.HandleFunc("/api/ing/{ingID}", ingHandler)
	r.HandleFunc("/api/latest", latestHandler)
	http.Handle("/api/", r)

	err := http.ListenAndServe(":"+strconv.Itoa(HTTPPort), nil)
	if err != nil {
		log.Fatal("ListenAndServe", err)
	}
}

func manageHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello World")
}

func latestHandler(w http.ResponseWriter, r *http.Request) {
	list, err := getIngs(1, 30)
	if err != nil {
		log.Println(err)
		return
	}
	io.WriteString(w, objectToJSONString(list))
	//http.Redirect(w, r, "/ing", http.StatusFound)
}

func ingsHandler(w http.ResponseWriter, r *http.Request) {
	pageIndex := 1
	pageSize := 30

	vars := mux.Vars(r)
	pageIndex, err := strconv.Atoi(vars["p"])
	if err != nil || pageIndex < 1 {
		pageIndex = 1
	}

	list, err := getIngs(pageIndex, pageSize)
	if err != nil {
		log.Println(err)
		return
	}
	io.WriteString(w, objectToJSONString(list))
}

func ingHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ingID, err := strconv.Atoi(vars["ingID"])
	if err != nil || ingID < 1 {
		ingID = 1
	}
	content, err := getIng(ingID)
	if err != nil {
		log.Println(err)
		return
	}
	if content != nil {
		io.WriteString(w, objectToJSONString(content))
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 Not Found"))
	}
}

func objectToJSONString(obj interface{}) string {
	strBytes, _ := json.Marshal(obj)
	return string(strBytes)
}
