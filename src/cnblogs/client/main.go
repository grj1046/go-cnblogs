package cnblogs

import (
	"cnblogs/db"
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
	err := db.InitialDB()
	if err != nil {
		fmt.Println("Execute Sql Script Error: ", err)
		os.Exit(1)
	}
	//901567
	ingID := "1115171"
	ingID = "1125100"
	ingID = "1127096"
	ingID = "1127498"
	//ingID = "901567" //private ing
	ingContent, originContent, err := ingClient.GetIngByID(ingID)
	if err != nil {
		fmt.Println("Get IngInfo Error: ", err)
		os.Exit(1)
	}

	strr, _ := json.Marshal(ingContent)
	log.Println("ingContent==", string(strr))

	strrr, _ := json.Marshal(originContent)
	log.Println("originContent==", string(strrr))

	/*logFile, err := os.OpenFile(ingID+".html", os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		fmt.Println("Open log file", err)
		os.Exit(1)
	}
	defer logFile.Close()
	logger := log.New(logFile, "\r\n", log.Ldate|log.Ltime|log.Llongfile)
	logger.Println(string(strr))
	*/
}
