# go-cnblogs
cnblogs ing services write by golang https://ing.cnblogs.com/

http://home.cnblogs.com/ing/1/

PRAGMA busy_timeout=4000;

```bash
go get github.com/PuerkitoBio/goquery
go get github.com/mattn/go-sqlite3
go get github.com/robfig/cron
go get -u github.com/gorilla/mux
```


```sql
CREATE INDEX IF NOT EXISTS index_IngID ON OriginIng(IngID);
CREATE INDEX IF NOT EXISTS index_IngID ON Comment(IngID);
CREATE INDEX IF NOT EXISTS index_IngID_HTMLHash ON OriginIng(IngID, HTMLHash);
```


https://github.com/revel/revel