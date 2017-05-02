package cnblogs

import (
	"cnblogs/ing"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

//GetIng Get Ing Html Document and save to file
func GetIng(ing *ing.Ing, ingID int) (string, error) {
	bodyString, err := ing.GetIngByID(ingID)
	if err != nil {
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(bodyString))
	return bodyString, nil
}

//Main main function
func Main() {
	conf := ReadConf()
	ing := &ing.Ing{}
	ing.Init(conf.AuthCookie)
	//901567
	bodyString, err := GetIng(ing, 901567)
	if err != nil {
		fmt.Println("Get IngInfo Error: ", err)
		os.Exit(1)
	}

	logFile, err := os.OpenFile(strconv.Itoa(901567)+".html", os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		fmt.Println("Open log file", err)
		os.Exit(1)
	}
	defer logFile.Close()
	logger := log.New(logFile, "\r\n", log.Ldate|log.Ltime|log.Llongfile)
	logger.Println(bodyString)
	//1117187
	bodyString, err = GetIng(ing, 1117187)
	logFile, err = os.OpenFile(strconv.Itoa(1117187)+".html", os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		fmt.Println("Open log file", err)
		os.Exit(1)
	}
	defer logFile.Close()
	logger = log.New(logFile, "\r\n", log.Ldate|log.Ltime|log.Llongfile)
	logger.Println(bodyString)
}
