package site

import (
	"cnblogs/db"
	"cnblogs/ing"
	"log"
	"time"
)

func getLatest(count int) ([]ing.Content, error) {
	cnblogsdb, err := db.GetDB()
	if err != nil {
		return nil, err
	}
	defer cnblogsdb.Close()
	rows, err := cnblogsdb.Query("select IngID, AuthorID, AuthorUserName, AuthorNickName, Time, Lucky, AcquiredAt, Body from Ing where Status = 200 order by IngID desc limit 30;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []ing.Content{}
	for rows.Next() {
		content := &ing.Content{}
		var acquireAt string
		err = rows.Scan(&content.IngID, &content.AuthorID, &content.AuthorUserName, &content.AuthorNickName, &content.Time,
			&content.Lucky, &acquireAt, &content.Body)
		if err != nil {
			return nil, err
		}
		content.AcquiredAt = convertStrTime(acquireAt)
		comments, _ := getCommentByIngID(content.IngID)
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
