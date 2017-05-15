package cnblogs

import (
	"encoding/json"
	"fmt"
	"os"
)

//Conf config.json
type Conf struct {
	StartIngID int
	EndIngID   int
	AuthCookie string
}

//ReadConf convert config.json file to conf struct
func ReadConf() Conf {
	confFile := "conf.json"
	_, err := os.Stat(confFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("please rename conf.default.json to conf.json")
			os.Exit(1)
		}
	}
	conf := &Conf{}
	file, err := os.Open(confFile)
	if err != nil {
		fmt.Println("get conf.json file error", err)
		os.Exit(1)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)

	err = decoder.Decode(&conf)
	if err != nil {
		fmt.Println("decode conf file error", err)
		os.Exit(1)
	}
	return *conf
}
