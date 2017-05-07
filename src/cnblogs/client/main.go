package cnblogs

import (
	"cnblogs/ing"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

//"github.com/PuerkitoBio/goquery"

//Main main function
func Main() {
	conf := ReadConf()
	ingClient := &ing.Client{}
	ingClient.Init(conf.AuthCookie)
	//901567
	ingID := "1115171"
	ingID = "1125100"
	ingContent, err := ingClient.GetIngByID(ingID)
	if err != nil {
		fmt.Println("Get IngInfo Error: ", err)
		os.Exit(1)
	}

	strr, _ := json.Marshal(ingContent)
	log.Println("ccccccc==", string(strr))

	logFile, err := os.OpenFile(ingID+".html", os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		fmt.Println("Open log file", err)
		os.Exit(1)
	}
	defer logFile.Close()
	logger := log.New(logFile, "\r\n", log.Ldate|log.Ltime|log.Llongfile)
	logger.Println(string(strr))
}
