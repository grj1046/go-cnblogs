# go-cnblogs
cnblogs ing services write by golang https://ing.cnblogs.com/

http://home.cnblogs.com/ing/1/

PRAGMA busy_timeout=4000;

```bash
go get github.com/PuerkitoBio/goquery
go get github.com/mattn/go-sqlite3
go get github.com/robfig/cron
```


```sql
CREATE INDEX index_IngID ON OriginIng(IngID);
CREATE INDEX index_IngID_HTMLHash ON OriginIng(IngID, HTMLHash);
```