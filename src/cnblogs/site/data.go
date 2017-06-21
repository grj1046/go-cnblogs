package site

import (
	"cnblogs/db"
	"cnblogs/ing"
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
	list := make([]ing.Content, count)
	var i = 0
	loc, _ := time.LoadLocation("Local")
	for rows.Next() {
		content := &ing.Content{}
		var acquireAt string
		err = rows.Scan(&content.IngID, &content.AuthorID, &content.AuthorUserName, &content.AuthorNickName, &content.Time,
			&content.Lucky, &acquireAt, &content.Body)
		if err != nil {
			return nil, err
		}
		if len(acquireAt) == 33 {
			t, _ := time.ParseInLocation("2006-01-02 15:04:05.0000000-07:00", acquireAt, loc)
			content.AcquiredAt = t
		}
		if len(acquireAt) == 32 {
			t, _ := time.ParseInLocation("2006-01-02 15:04:05.000000-07:00", acquireAt, loc)
			content.AcquiredAt = t
		}
		list[i] = *content
		i++
	}
	return list, nil
}
