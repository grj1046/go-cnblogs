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
	//ingID = "1128213" //是狼是狗
	ingID = "1127350"

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
	err = InsertIngToDB(*ingContent, *originContent)
	if err != nil {
		fmt.Println("Get IngInfo Error: ", err)
		os.Exit(1)
	}
}

//InsertIngToDB Insert or update Ing To sqlite3 db
func InsertIngToDB(ingContent ing.Content, originContent ing.OriginContent) error {
	sqlite, err := db.GetDB()
	if err != nil {
		return errors.New("open db error: " + err.Error())
	}
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
		trans.Prepare(sqlIngContent)
		_, err = trans.Exec(sqlIngContent, ingContent.IngID, ingContent.AuthorID, ingContent.AuthorUserName, ingContent.AuthorNickName,
			ingContent.Time, ingContent.Status, ingContent.Lucky, ingContent.IsPrivate, ingContent.AcquiredAt, ingContent.Body)
		if err != nil {
			trans.Rollback()
			return errors.New("insert ing table error: " + err.Error())
		}
	} else if err != nil {
		trans.Rollback()
		return errors.New("scan ingStatus error: " + err.Error())
	}
	if ingStatus == 404 {
		return nil
	}
	if ingContent.Status == 404 {
		//update status = 404 and return
		sqlIngUpdate := "update `Ing` set `Status` = 404 where `IngID` = ?"
		_, err := sqlite.Exec(sqlIngUpdate, 404)
		if err != nil {
			trans.Rollback()
			return errors.New("update ing Status error: " + err.Error())
		}
		return nil
	}

	//OriginContent
	stmt, err = sqlite.Prepare("select HTMLHash from OriginIng where IngID = ? and HTMLHash = ?")
	if err != nil {
		trans.Rollback()
		return errors.New("prepare select OriginIng hash error: " + err.Error())
	}
	defer stmt.Close()
	row = stmt.QueryRow(ingContent.IngID, md5String(originContent.HTML))
	var htmlHash string
	err = row.Scan(&htmlHash)
	if err == sql.ErrNoRows {
		sqlIngOriginContent := "insert into OriginIng (IngID, Status, AcquiredAt, Exception, HTMLHash, HTML) values (?, ?, ?, ?, ?, ?);"
		HTMLHash := md5String(originContent.HTML)
		trans.Prepare(sqlIngOriginContent)
		_, err = trans.Exec(sqlIngOriginContent, originContent.IngID, originContent.Status,
			originContent.AcquiredAt, originContent.Exception, HTMLHash, originContent.HTML)
		if err != nil {
			trans.Rollback()
			return errors.New("insert OriginContent error: " + err.Error())
		}
	} else if err != nil {
		trans.Rollback()
		return errors.New("scan htmlHash error: " + err.Error())
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
	trans.Prepare(sqlIngComment)
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
		_, err = trans.Exec(sqlIngComment, ingComment.IngID, ingComment.CommentID, ingComment.AuthorID, ingComment.AuthorUserName, ingComment.AuthorNickName,
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
		trans.Prepare(sqlStmt)
		_, err := trans.Exec(sqlStmt, ingContent.AcquiredAt, ingContent.IngID)
		if err != nil {
			trans.Rollback()
			return errors.New("update ing AcquiredAt error: " + err.Error())
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
