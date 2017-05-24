package client

import (
	"cnblogs/conf"
	"cnblogs/db"
	"cnblogs/ing"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"log"
	"strconv"

	"time"
)

var ingClient *ing.Client

//Main main function
func Main(conf conf.Conf) {
	ingClient = &ing.Client{}
	ingClient.Init(conf.AuthCookie)
	err := db.InitialDB()
	if err != nil {
		log.Println("Execute Sql Script Error: ", err)
		return
	}
	//http://home.cnblogs.com/ing/1115171/
	go func() {
		for {
			err = TaskSyncLatestIngToDB()
			if err != nil {
				log.Println("TaskSyncLatestIngToDB", err)
			}
			time.Sleep(time.Second * 30)
		}
	}()

	go func() {
		log.Println("run TaskSyncLatestCommentToDB")
		pageSize := 30
		err = TaskSyncLatestCommentToDB(pageSize)
		if err != nil {
			log.Println("TaskSyncLatestCommentToDB", err)
		}
		time.Sleep(time.Second * 10)
	}()

	if !conf.EnableSite {
		select {}
	}
}

//LeakFinding check if ingID not in database, update it.
func LeakFinding(ingID int) error {
	sqlite, err := db.GetDB()
	if err != nil {
		return errors.New("open db error: " + err.Error())
	}
	defer sqlite.Close()

	row := sqlite.QueryRow("select `Status` from `Ing` where IngID = ?", ingID)
	var ingStatus int
	err = row.Scan(&ingStatus)

	if ingStatus == 0 || err == sql.ErrNoRows {
		log.Println("Current IngID Not Exist In db, Update it. ", ingID)
		err = GetIngAndSaveToDB(ingID)
		if err != nil {
			return err
		}
	}
	return nil
}

//TaskSyncLatestIngToDB Sync latest Ing to Database
func TaskSyncLatestIngToDB() error {
	if ingClient == nil {
		return errors.New("ingClient is not initial")
	}
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
		return err
	}
	log.Println("maxIngID", maxIngID, "currIngID", currIngID)
	for i := currIngID; i <= maxIngID; i++ {
		log.Println("Sync Ing", i)
		err = GetIngAndSaveToDB(i)
		if err != nil {
			return errors.New(strconv.Itoa(i) + err.Error())
		}
	}
	return nil
}

//TaskSyncLatestCommentToDB sync latest comment to database
func TaskSyncLatestCommentToDB(pageSize int) error {
	if ingClient == nil {
		return errors.New("ingClient is not initial")
	}

	ingList, err := ingClient.GetLatestIngFromComment(pageSize)
	if err != nil {
		log.Println("GetLatestIngFromComment", err)
	}
	for i := 0; i < pageSize; i++ {
		var currIngID = ingList[i]
		if currIngID > 0 {
			log.Println("Sync Ing Comment", i)
			err = GetIngAndSaveToDB(currIngID)
			if err != nil {
				return errors.New(strconv.Itoa(i) + err.Error())
			}
		}
	}
	return nil
}

//GetIngAndSaveToDB Get Ing Cotnent by IngID and save it to sqlite database
func GetIngAndSaveToDB(ingID int) error {
	if ingClient == nil {
		return errors.New("ingClient is not initial")
	}
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
	err = InsertToOriginDB(ingContent.IngID, *originContent)
	if err != nil {
		return errors.New("InsertToOriginDB: " + err.Error())
	}
	err = InsertIngToDB(*ingContent)
	if err != nil {
		return errors.New("InsertIngToDB: " + err.Error())
	}
	return nil
}

//InsertIngToDB Insert or update Ing To sqlite3 db
func InsertIngToDB(ingContent ing.Content) error {
	sqlite, err := db.GetDB()
	if err != nil {
		return errors.New("open db error: " + err.Error())
	}
	defer sqlite.Close()

	//if error is database is locked repeat 10 times
	for i := 1; i <= 10; i++ {
		err = sqlite.Ping()
		if err != nil {
			if err.Error() == "database is locked" {
				log.Println("Ping occured database is locked, try times:" + strconv.Itoa(i) +
					" IngID: " + strconv.Itoa(ingContent.IngID))
				time.Sleep(time.Millisecond * 500)
				continue
			}
			return errors.New("Ping error: " + err.Error())
		}
		break
	}
	trans, err := sqlite.Begin()
	if err != nil {
		return errors.New("begin trans error: " + err.Error())
	}
	//http://go-database-sql.org/prepared.html
	defer trans.Rollback()
	//Content
	stmt, err := trans.Prepare("select `Status` from `Ing` where IngID = ?")
	if err != nil {
		return errors.New("prepare select IngStatus error: " + err.Error())
	}
	defer stmt.Close()
	row := stmt.QueryRow(ingContent.IngID)
	var ingStatus int
	err = row.Scan(&ingStatus)

	if ingStatus == 0 || err == sql.ErrNoRows {
		sqlIngContent := "insert into `Ing` (`IngID`, `AuthorID`, `AuthorUserName`, `AuthorNickName`, `Time`, `Status`, `Lucky`, `IsPrivate`, `IsNewbie`, `AcquiredAt`, `Body`) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);"
		stmt, err = trans.Prepare(sqlIngContent)
		if err != nil {
			return errors.New("prepare ing sql error: " + err.Error())
		}
		defer stmt.Close()
		//if error is database is locked repeat 10 times
		for i := 1; i <= 10; i++ {
			_, err = stmt.Exec(ingContent.IngID, ingContent.AuthorID, ingContent.AuthorUserName, ingContent.AuthorNickName,
				ingContent.Time, ingContent.Status, ingContent.Lucky, ingContent.IsPrivate, ingContent.IsNewbie,
				ingContent.AcquiredAt, ingContent.Body)
			if err != nil {
				if err.Error() == "database is locked" {
					log.Println("insert ing table occured database is locked, try times:" + strconv.Itoa(i) +
						" IngID: " + strconv.Itoa(ingContent.IngID))
					time.Sleep(time.Millisecond * 500)
					continue
				}
				return errors.New("insert ing table error: " + err.Error())
			}
			break
		}
	} else if err != nil {
		return errors.New("scan ingStatus error: " + err.Error())
	}
	if ingStatus == 404 {
		trans.Commit()
		return nil
	}
	if ingContent.Status == 404 {
		//update status = 404 and return
		sqlIngUpdate := "update `Ing` set `Status` = 404 where `IngID` = ?"
		stmt, err = trans.Prepare(sqlIngUpdate)
		if err != nil {
			return errors.New("prepare update status sql error: " + err.Error())
		}
		defer stmt.Close()
		_, err := stmt.Exec(ingContent.IngID)
		if err != nil {
			return errors.New("update ing Status error: " + err.Error())
		}
		trans.Commit()
		return nil
	}
	//Comments
	stmt, err = trans.Prepare("select ID, CommentID from Comment where IngID = ? and IsDelete = 0")
	if err != nil {
		return errors.New("prepare select CommentID error: " + err.Error())
	}
	defer stmt.Close()
	rows, err := stmt.Query(ingContent.IngID)
	if err != nil {
		return errors.New("get CommentID error: " + err.Error())
	}
	defer rows.Close()
	// update IsDelete = 1, insert
	unDeletedCommentIDs := make([]string, 0)
	for rows.Next() {
		var ID int
		var commentID string
		err = rows.Scan(&ID, &commentID)
		if err != nil {
			return errors.New("get commentID error: " + err.Error())
		}
		unDeletedCommentIDs = append(unDeletedCommentIDs, commentID)
	}
	commentUpdated := false
	sqlIngComment := "insert into `Comment` (`IngID`, `CommentID`, `AuthorID`, `AuthorUserName`, `AuthorNickName`, `Body`, `Time`, `IsDelete`) values (?, ?, ?, ?, ?, ?, ?, ?);"
	stmt, err = trans.Prepare(sqlIngComment)
	if err != nil {
		return errors.New("prepare insert ingComment sql error: " + err.Error())
	}
	defer stmt.Close()
	for _, ingComment := range ingContent.Comments {
		//if CommentID in savedCommentIDs, remove it.
		currIndex := -1
		for i := 0; i < len(unDeletedCommentIDs); i++ {
			if unDeletedCommentIDs[i] == ingComment.CommentID {
				currIndex = i
				break
			}
		}
		if currIndex != -1 {
			unDeletedCommentIDs[currIndex] = ""
			continue
		}
		_, err = stmt.Exec(ingComment.IngID, ingComment.CommentID, ingComment.AuthorID, ingComment.AuthorUserName, ingComment.AuthorNickName,
			ingComment.Body, ingComment.Time, ingComment.IsDelete)
		if err != nil {
			return errors.New("insert comment error: " + err.Error())
		}
		if !commentUpdated {
			commentUpdated = true
		}
	}

	// set to Deleted
	sqlIngCommentUpdate := "update `Comment` set IsDelete = 1 where IngID = ? and CommentID = ?"
	stmt, err = trans.Prepare(sqlIngCommentUpdate)
	if err != nil {
		return errors.New("prepare update set IsDelete sql error: " + err.Error())
	}
	defer stmt.Close()
	if err != nil {
		return errors.New("prepare delete sql error: " + err.Error())
	}
	for _, willDeletedCommentID := range unDeletedCommentIDs {
		if willDeletedCommentID == "" {
			continue
		}
		if !commentUpdated {
			commentUpdated = true
		}
		_, err = stmt.Exec(ingContent.IngID, willDeletedCommentID)
		if err != nil {
			return errors.New("update Comment IsDelete flag error: " + err.Error())
		}
	}
	if commentUpdated && ingStatus == 200 {
		sqlStmt := "update `Ing` set `AcquiredAt` = ? where `IngID` = ?"
		stmt, err = trans.Prepare(sqlStmt)
		if err != nil {
			return errors.New("prepare ing AcquiredAt error: " + err.Error())
		}
		defer stmt.Close()
		_, err := stmt.Exec(ingContent.AcquiredAt, ingContent.IngID)
		if err != nil {
			return errors.New("update ing AcquiredAt error: " + err.Error())
		}
	}
	trans.Commit()
	return nil
}

//InsertToOriginDB store Origin Ing Info to seperator database
func InsertToOriginDB(ingID int, originContent ing.OriginContent) error {
	originDB, err := db.GetDBOrigin()
	if err != nil {
		return errors.New("open origin db error:" + err.Error())
	}
	defer originDB.Close()
	/*
		err = originDB.Ping()
		if err != nil {
			// do something here
		}
	*/
	originDB.SetMaxOpenConns(1)
	md5Hash := md5String(originContent.HTML)
	var htmlHash string
	//if error is database is locked repeat 10 times
	for i := 1; i <= 10; i++ {
		err = originDB.QueryRow("select `HTMLHash` from `OriginIng` where `IngID` = ? and `HTMLHash` = ?",
			ingID, md5Hash).Scan(&htmlHash)
		if err != nil {
			if err == sql.ErrNoRows {
				//sql: no rows in result set
				break
			}
			if err.Error() == "database is locked" {
				log.Println("scan htmlHash occured database is locked, try times:" + strconv.Itoa(i) + " IngID: " + strconv.Itoa(originContent.IngID))
				time.Sleep(time.Millisecond * 500)
				continue
			}
			return errors.New("scan htmlHash error: " + err.Error())
		}
		break
	}

	if htmlHash == "" || err == sql.ErrNoRows {
		sqlIngOriginContent := "insert into OriginIng (IngID, Status, AcquiredAt, Exception, HTMLHash, HTML) values (?, ?, ?, ?, ?, ?);"
		//if error is database is locked repeat 10 times
		for i := 1; i <= 10; i++ {
			_, err := originDB.Exec(sqlIngOriginContent, originContent.IngID, originContent.Status, originContent.AcquiredAt,
				originContent.Exception, md5Hash, originContent.HTML)
			if err != nil {
				if err.Error() == "database is locked" {
					log.Println("scan htmlHash occured database is locked, try times:" + strconv.Itoa(i) + " IngID: " + strconv.Itoa(originContent.IngID))
					time.Sleep(time.Millisecond * 500)
					continue
				}
				return errors.New("insert OriginContent error: " + err.Error())
			}
			break
		}
		/*
			id, err := result.LastInsertId()
			if err != nil {
				return errors.New("get LastInsertId error: " + err.Error())
			}
			log.Println("id", id)
		*/
	}
	return nil
}

func md5String(originString string) string {
	md5 := md5.New()
	md5.Write([]byte(originString))
	hashString := hex.EncodeToString(md5.Sum(nil))
	return hashString
}
