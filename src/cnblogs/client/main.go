package cnblogs

import (
	"cnblogs/db"
	"cnblogs/ing"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/robfig/cron"
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
	//http://home.cnblogs.com/ing/1115171/
	//901567
	ingID := "1115171"
	ingID = "1125100"
	ingID = "1127096"
	ingID = "1127498"
	//ingID = "901567" //private ing
	//ingID = "1128213" //是狼是狗
	ingID = "1127350"
	ingID = "1129270"
	ingID = "26"
	ingID = "0"
	c := cron.New()
	spec := "*/1 * * * * *"
	c.AddFunc(spec, func() {
		//ingID++
		i, err := strconv.Atoi(ingID)
		if err != nil {
			fmt.Println("convert to int error", err.Error())
			os.Exit(1)
		}
		i++
		ingID = strconv.Itoa(i)
		fmt.Println("start", ingID)

		//search if current Ing in table && ingStatus is 404, do nothing.
		ingContent, originContent, err := ingClient.GetIngByID(ingID)
		if err != nil {
			fmt.Println("Get IngInfo Error: ", err)
			os.Exit(1)
		}

		if ingContent.Status == 403 {
			fmt.Println("auth cookie invalid, please check.")
			os.Exit(1)
		}
		//OriginContent
		//go call(*ingContent, *originContent)
		err = InsertToOriginDB(ingContent.IngID, *originContent)
		if err != nil {
			fmt.Println("Insert OriginIngInfo Error: ", err)
			os.Exit(1)
		}
		//err = InsertIngToDB(*ingContent, *originContent)
		err = InsertIngToDB(*ingContent)
		if err != nil {
			fmt.Println("Get IngInfo Error: ", err)
			os.Exit(1)
		}
	})
	c.Start()
	select {} //阻塞主线程不退出
}

func call(ingContent ing.Content, originContent ing.OriginContent) {
	err := InsertToOriginDB(ingContent.IngID, originContent)
	if err != nil {
		fmt.Println("Insert OriginIngInfo Error: ", err)
		os.Exit(1)
	}
}

//InsertIngToDB Insert or update Ing To sqlite3 db
func InsertIngToDB(ingContent ing.Content) error {
	/*
		//OriginContent , originContent ing.OriginContent
		err := InsertToOriginDB(ingContent.IngID, originContent)
		if err != nil {
			return err
		}
	*/
	sqlite, err := db.GetDB()
	if err != nil {
		return errors.New("open db error: " + err.Error())
	}
	defer sqlite.Close()

	trans, err := sqlite.Begin()
	if err != nil {
		trans.Rollback()
		return errors.New("begin trans error: " + err.Error())
	}
	//Content
	stmt, err := sqlite.Prepare("select `Status` from `Ing` where IngID = ?")
	if err != nil {
		trans.Rollback()
		return errors.New("prepare select IngStatus error: " + err.Error())
	}
	defer stmt.Close()
	row := stmt.QueryRow(ingContent.IngID)
	var ingStatus int
	err = row.Scan(&ingStatus)

	if err == sql.ErrNoRows {
		sqlIngContent := "insert into `Ing` (`IngID`, `AuthorID`, `AuthorUserName`, `AuthorNickName`, `Time`, `Status`, `Lucky`, `IsPrivate`, `AcquiredAt`, `Body`) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);"
		stmt, err = trans.Prepare(sqlIngContent)
		if err != nil {
			trans.Rollback()
			return errors.New("prepare ing sql error: " + err.Error())
		}
		defer stmt.Close()
		_, err = stmt.Exec(ingContent.IngID, ingContent.AuthorID, ingContent.AuthorUserName, ingContent.AuthorNickName,
			ingContent.Time, ingContent.Status, ingContent.Lucky, ingContent.IsPrivate, ingContent.AcquiredAt, ingContent.Body)
		if err != nil {
			fmt.Println("err", err)
			trans.Rollback()
			return errors.New("insert ing table error: " + err.Error())
		}
	} else if err != nil {
		trans.Rollback()
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
			trans.Rollback()
			return errors.New("prepare update status sql error: " + err.Error())
		}
		defer stmt.Close()
		_, err := stmt.Exec(ingContent.IngID)
		if err != nil {
			trans.Rollback()
			return errors.New("update ing Status error: " + err.Error())
		}
		trans.Commit()
		return nil
	}
	//Comments
	stmt, err = sqlite.Prepare("select ID, CommentID from Comment where IngID = ? and IsDelete = 0")
	if err != nil {
		trans.Rollback()
		return errors.New("prepare select CommentID error: " + err.Error())
	}
	defer stmt.Close()
	rows, err := stmt.Query(ingContent.IngID)
	if err != nil {
		trans.Rollback()
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
			trans.Rollback()
			return errors.New("get commentID error: " + err.Error())
		}
		unDeletedCommentIDs = append(unDeletedCommentIDs, commentID)
	}
	commentUpdated := false
	sqlIngComment := "insert into `Comment` (`IngID`, `CommentID`, `AuthorID`, `AuthorUserName`, `AuthorNickName`, `Body`, `Time`, `IsDelete`) values (?, ?, ?, ?, ?, ?, ?, ?);"
	stmt, err = trans.Prepare(sqlIngComment)
	if err != nil {
		trans.Rollback()
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
			trans.Rollback()
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
		trans.Rollback()
		return errors.New("prepare update set IsDelete sql error: " + err.Error())
	}
	defer stmt.Close()
	if err != nil {
		trans.Rollback()
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
			trans.Rollback()
			return errors.New("update Comment IsDelete flag error: " + err.Error())
		}
	}
	if commentUpdated && ingStatus == 200 {
		sqlStmt := "update `Ing` set `AcquiredAt` = ? where `IngID` = ?"
		stmt, err = trans.Prepare(sqlStmt)
		if err != nil {
			trans.Rollback()
			return errors.New("prepare ing AcquiredAt error: " + err.Error())
		}
		defer stmt.Close()
		_, err := stmt.Exec(ingContent.AcquiredAt, ingContent.IngID)
		if err != nil {
			trans.Rollback()
			return errors.New("update ing AcquiredAt error: " + err.Error())
		}
	}
	trans.Commit()
	return nil
}

//InsertToOriginDB store Origin Ing Info to seperator database
func InsertToOriginDB(ingID string, originContent ing.OriginContent) error {
	originDB, err := db.GetDBOrigin()
	if err != nil {
		return errors.New("open origin db error:" + err.Error())
	}
	defer originDB.Close()

	trans, err := originDB.Begin()
	stmt, err := originDB.Prepare("select `HTMLHash` from `OriginIng` where `IngID` = ? and `HTMLHash` = ?")
	if err != nil {
		trans.Rollback()
		return errors.New("prepare select OriginIng hash error: " + err.Error())
	}
	defer stmt.Close()
	md5Hash := md5String(originContent.HTML)
	var htmlHash string
	err = stmt.QueryRow(ingID, md5Hash).Scan(&htmlHash)
	if err != nil && err != sql.ErrNoRows {
		trans.Rollback()
		return errors.New("scan htmlHash error: " + err.Error())
	}
	if htmlHash == "" {
		sqlIngOriginContent := "insert into OriginIng (IngID, Status, AcquiredAt, Exception, HTMLHash, HTML) values (?, ?, ?, ?, ?, ?);"
		stmt, err = trans.Prepare(sqlIngOriginContent)
		if err != nil {
			trans.Rollback()
			return errors.New("prepare OriginContent error: " + err.Error())
		}
		defer stmt.Close()
		_, err := stmt.Exec(originContent.IngID, originContent.Status, originContent.AcquiredAt,
			originContent.Exception, md5Hash, originContent.HTML)
		if err != nil {
			trans.Rollback()
			return errors.New("insert OriginContent error: " + err.Error())
		}
	}
	trans.Commit()
	return nil
}

func md5String(originString string) string {
	md5 := md5.New()
	md5.Write([]byte(originString))
	hashString := hex.EncodeToString(md5.Sum(nil))
	return hashString
}
