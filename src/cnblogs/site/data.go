package site

import (
	"cnblogs/db"
	"cnblogs/ing"
	"database/sql"
	"log"
	"time"
)

func getIng(ingID int) (*ing.Content, error) {
	cnblogsdb, err := db.GetDB()
	if err != nil {
		return nil, err
	}
	defer cnblogsdb.Close()
	content := &ing.Content{}
	var acquireAt string
	err = cnblogsdb.QueryRow("select IngID, AuthorID, AuthorUserName, AuthorNickName, Time, Status, Lucky, AcquiredAt, Body from Ing where IngID = ?;",
		ingID).Scan(&content.IngID, &content.AuthorID, &content.AuthorUserName, &content.AuthorNickName, &content.Time,
		&content.Status, &content.Lucky, &acquireAt, &content.Body)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	content.AcquiredAt = convertStrTime(acquireAt)
	comments, err := getCommentByIngID(content.IngID)
	if err != nil {
		return nil, err
	}
	content.Comments = comments
	return content, nil
}

func getIngs(pageIndex, pageSize int) ([]ing.Content, error) {
	if pageIndex < 1 {
		pageIndex = 1
	}
	if pageSize < 1 {
		pageSize = 1
	}
	cnblogsdb, err := db.GetDB()
	if err != nil {
		return nil, err
	}
	defer cnblogsdb.Close()
	rows, err := cnblogsdb.Query("select IngID, AuthorID, AuthorUserName, AuthorNickName, Time, Status, Lucky, AcquiredAt, Body from Ing where Status = 200 order by IngID desc limit ?,?;",
		(pageIndex-1)*pageSize, pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []ing.Content{}
	for rows.Next() {
		content := &ing.Content{}
		var acquireAt string
		err = rows.Scan(&content.IngID, &content.AuthorID, &content.AuthorUserName, &content.AuthorNickName, &content.Time,
			&content.Status, &content.Lucky, &acquireAt, &content.Body)
		if err != nil {
			return nil, err
		}
		content.AcquiredAt = convertStrTime(acquireAt)
		comments, err := getCommentByIngID(content.IngID)
		if err != nil {
			return nil, err
		}
		content.Comments = comments
		list = append(list, *content)
	}
	return list, nil
}

func getCommentByIngID(ingID int) ([]ing.Comment, error) {
	var comments = []ing.Comment{}
	cnblogsdb, err := db.GetDB()
	if err != nil {
		return nil, err
	}
	defer cnblogsdb.Close()
	rows, err := cnblogsdb.Query("select IngID, CommentID, AuthorID, AuthorUserName, AuthorNickName, Body, Time, IsDelete from Comment where IngID = ?;",
		ingID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		comment := &ing.Comment{}
		err = rows.Scan(&comment.IngID, &comment.CommentID, &comment.AuthorID, &comment.AuthorUserName, &comment.AuthorNickName,
			&comment.Body, &comment.Time, &comment.IsDelete)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		//comment.AcquiredAt = convertStrTime(acquireAt)
		comments = append(comments, *comment)
	}
	return comments, nil
}

func convertStrTime(str string) time.Time {
	loc, _ := time.LoadLocation("Local")
	var t time.Time
	if len(str) == 33 {
		t, _ = time.ParseInLocation("2006-01-02 15:04:05.0000000-07:00", str, loc)
	}
	if len(str) == 32 {
		t, _ = time.ParseInLocation("2006-01-02 15:04:05.000000-07:00", str, loc)
	}
	return t
}
