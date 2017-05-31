package client

import (
	"cnblogs/conf"
	"cnblogs/db"
	"cnblogs/ing"
	"database/sql"
	"errors"
	"log"
	"strconv"
	"time"
)

//var ingClient *ing.Client

//Main main function
func Main(conf conf.Conf) {
	err := db.InitialDB()
	if err != nil {
		log.Println("Execute Sql Script Error: ", err)
		return
	}
	//http://home.cnblogs.com/ing/1115171/
	TaskDefault(conf)

	ingClient := &ing.Client{}
	ingClient.Init(conf.AuthCookie)

	//TaskSyncLatestIngToDB(ingClient)
	//TaskSyncSpecifyPageComment(ingClient, 1) //maxPageIndex
	//SyncSpecifyDateIng(ingClient, "2017-05-31")

	//SyncIngs(ingClient, 1131799, 1142048)
	//SyncIngs(ingClient, 1131799, 0)
}

//TaskDefault default task: sync ing & comment
func TaskDefault(conf conf.Conf) {
	go func() {
		ingClient := &ing.Client{}
		ingClient.Init(conf.AuthCookie)
		for {
			err := TaskSyncLatestIngToDB(ingClient)
			if err != nil {
				log.Println("TaskSyncLatestIngToDB", err)
			}
			time.Sleep(time.Second * 30)
		}
	}()

	go func() {
		ingClient := &ing.Client{}
		ingClient.Init(conf.AuthCookie)
		log.Println("run TaskSyncLatestCommentToDB")
		pageIndex := 1
		pageSize := 30
		for {
			err := TaskSyncLatestCommentToDB(ingClient, pageIndex, pageSize)
			if err != nil {
				log.Println("TaskSyncLatestCommentToDB", err)
			}
			time.Sleep(time.Second * 10)
		}
	}()

	if !conf.EnableSite {
		select {}
	}
}

//SyncSpecifyDateIng re-acquire the date specified, eg: 2017-05-25
func SyncSpecifyDateIng(ingClient *ing.Client, date string) error {
	sqlite, err := db.GetDB()
	if err != nil {
		return errors.New("open db error: " + err.Error())
	}
	defer sqlite.Close()
	statSQL := "select max(IngID), min(IngID) from Ing where Time between '" + date + " 00:00:00' and '" + date + " 23:59:59'"

	var maxIngCount int
	var minIngCount int
	err = sqlite.QueryRow(statSQL).Scan(&maxIngCount, &minIngCount)
	if err != nil {
		log.Println(err)
	}
	if maxIngCount > minIngCount && maxIngCount != 0 && minIngCount != 0 {
		//statSQL = "select count(1), max(IngID) - min(IngID) from Ing where IngID between " + strconv.Itoa(minIngCount) + " and " + strconv.Itoa(maxIngCount)
		for i := minIngCount; i <= maxIngCount; i++ {
			log.Println("sync ing:", i)
			err = GetIngAndSaveToDB(ingClient, i)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//SyncIngs Sync Ing Range from...to, eg 113248 to 1142048
//if to is zero, sync to maxIngID
func SyncIngs(ingClient *ing.Client, from int, to int) {
	if from < to || to != 0 {
		log.Println("from:", from, "to", to)
		return
	}

	maxIngID, err := ingClient.GetMaxIngID()
	if err != nil {
		log.Println("get maxIngID error", err)
		return
	}
	to = maxIngID
	for i := from; i <= to; i++ {
		log.Println("sync ing:", i)
		err = GetIngAndSaveToDB(ingClient, i)
		if err != nil {
			log.Println("GetIngAndSaveToDB err", err)
			return
		}
	}
}

//TaskSyncLatestIngToDB Sync latest Ing to Database
func TaskSyncLatestIngToDB(ingClient *ing.Client) error {
	maxIngID, err := ingClient.GetMaxIngID()
	if err != nil {
		return err
	}
	sqlite, err := db.GetDB()
	if err != nil {
		return errors.New("open db error: " + err.Error())
	}
	defer sqlite.Close()
	var currIngID int
	err = sqlite.QueryRow("select IngID from Ing order by IngID desc limit 1").Scan(&currIngID)
	if err != nil {
		if err == sql.ErrNoRows {
			currIngID = 1
		} else {
			return err
		}
	}
	log.Println("currIngID", currIngID, "maxIngID", maxIngID)
	if maxIngID == currIngID {
		log.Println("nothing to do")
		return nil
	}
	for i := currIngID + 1; i <= maxIngID; i++ {
		log.Println("Sync Ing", i)
		err = GetIngAndSaveToDB(ingClient, i)
		if err != nil {
			return errors.New(strconv.Itoa(i) + err.Error())
		}
	}
	return nil
}

//TaskSyncSpecifyPageComment sync specify page of comments
func TaskSyncSpecifyPageComment(ingClient *ing.Client, maxPageIndex int) {
	if maxPageIndex <= 0 {
		log.Println("maxPageIndex error", maxPageIndex)
		return
	}
	for i := 1; i <= maxPageIndex; i++ {
		pageIndex := i
		err := TaskSyncLatestCommentToDB(ingClient, pageIndex, 30)
		if err != nil {
			log.Println("TaskSyncSpecifyPageComment error", err)
			return
		}
	}
}

//TaskSyncLatestCommentToDB sync latest comment to database
func TaskSyncLatestCommentToDB(ingClient *ing.Client, pageIndex int, pageSize int) error {
	ingList, err := ingClient.GetLatestIngFromComment(pageIndex, pageSize)
	if err != nil {
		log.Println("GetLatestIngFromComment", err)
	}
	log.Println("ingList", len(ingList))
	for i := 0; i < pageSize; i++ {
		var currIngID = ingList[i]
		if currIngID > 0 {
			log.Println(i+1, " Sync Comment: IngID", currIngID)
			err = GetIngAndSaveToDB(ingClient, currIngID)
			if err != nil {
				return errors.New(strconv.Itoa(i) + err.Error())
			}
		}
	}
	return nil
}

//GetIngAndSaveToDB Get Ing Cotnent by IngID and save it to sqlite database
func GetIngAndSaveToDB(ingClient *ing.Client, ingID int) error {
	//search if current Ing in table && ingStatus is 404, do nothing.
	ingContent, originContent, err := ingClient.GetIngByID(ingID)
	if err != nil {
		return errors.New("Get IngInfo Error: " + err.Error())
	}

	if ingContent.Status == 403 {
		return errors.New("auth cookie invalid, please check")
	}
	//OriginContent
	//go call(*ingContent, *originContent)
	err = ingClient.InsertToOriginDB(ingContent.IngID, *originContent)
	if err != nil {
		return errors.New("InsertToOriginDB: " + err.Error())
	}
	err = ingClient.InsertIngToDB(*ingContent)
	if err != nil {
		return errors.New("InsertIngToDB: " + err.Error())
	}
	return nil
}
