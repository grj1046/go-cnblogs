package ing

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"strings"

	"encoding/json"

	"github.com/PuerkitoBio/goquery"
)

//Client ing.cnblogs.com
type Client struct {
	//IngID      int
	authCookie string
	urlStr     string
	httpClient *http.Client
}

//Content ing Content struct
type Content struct {
	IngID     string
	Time      string
	Body      string
	IsPrivate bool
	updateAt  time.Time
	Comments  []Comment
}

//Comment Ing.Content's Comment
type Comment struct {
	IngID          string
	CommentID      string
	AuthorID       string
	AuthorUserName string
	AuthorNickName string
	Body           string
	Time           string
	IsDelete       bool
	updateAt       time.Time
}

//OriginContent store the origin ing html
type OriginContent struct {
	IngID    string
	HTML     string
	updateAt time.Time
}

//Init Initialize httpClient with authCookie
func (client *Client) Init(authCookie string) {
	client.httpClient = &http.Client{}
	client.authCookie = authCookie
}

//GetIngByID Get Ing Html Document by ingID
func (client *Client) GetIngByID(ingID string) (*Content, error) {
	//urlStr := "https://ing.cnblogs.com/u/grj1046/status/" + strconv.Itoa(ingID) + "/"
	//http://home.cnblogs.com/ing/1115171/
	urlStr := "https://ing.cnblogs.com/redirect/" + ingID + "/"
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Cookie", client.authCookie)
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New("StatusCode: " + strconv.Itoa(resp.StatusCode))
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}
	publishTime := doc.Find(".ing_detail_title").Text()
	//log.Println("publishTime>>>" + publishTime)
	ingBody, err := doc.Find("#ing_detail_body").Html()

	nowTime := time.Now()

	content := &Content{}
	content.IngID = ingID
	content.updateAt = nowTime
	content.Time = publishTime
	content.Body = ingBody

	commentCount := doc.Find("#comment_block_" + ingID + " li").Length()
	content.Comments = make([]Comment, commentCount)

	doc.Find("#comment_block_" + ingID + " li").Each(func(index int, selection *goquery.Selection) {
		//IngID, CommentID, Comment, CommentTime, AuthorID AuthorUserName, AuthorNickName
		comment := &Comment{}
		//IngID
		comment.IngID = ingID
		//CommentID id="comment_1400623"
		commentID, _ := selection.Attr("id")
		tmplen := len("comment_")
		comment.CommentID = commentID[tmplen:]
		//CommentTime class="text_green"
		time, _ := selection.Find(".text_green").Attr("title")
		comment.Time = time
		//AuthorID
		authorID, _ := selection.Find(".ing_comment_face").Attr("src")
		//https://pic.cnblogs.com/face/289132/20130423092122.png
		tmplen = len("https://pic.cnblogs.com/face/")
		authorID = authorID[tmplen:strings.LastIndex(authorID, "/")]
		comment.AuthorID = authorID

		authorNode := selection.Find("#comment_author_" + comment.CommentID)
		//AuthorName //home.cnblogs.com/u/grj1046/
		authorUserName, _ := authorNode.Attr("href")
		tmplen = len("//home.cnblogs.com/u/")
		authorUserName = authorUserName[tmplen : len(authorUserName)-1]
		comment.AuthorUserName = authorUserName
		//AuthorNickName comment_author_1400623
		comment.AuthorNickName = authorNode.Text()

		//Comment
		tmpBody := selection.Find("div")
		/*<a target="_blank" href="//home.cnblogs.com/u/grj1046/">
		  	  <img src="https://pic.cnblogs.com/face/289132/20130423092122.png" class="ing_comment_face" alt="">
		  </a>
		*/
		//aa, _ := selection.Find("div a").Children().First().Html()
		//log.Println("tmpBody.First()>", aa)

		tmpBody.Find(".ing_comment_face").Parent().Remove()
		//<a target="_blank" id="comment_author_1400623" title="nil的闪存" href="//home.cnblogs.com/u/grj1046/">nil</a>
		tmpBody.Find("#comment_author_" + comment.CommentID).Remove()
		tmpBody = tmpBody.First().Remove()
		//if delete button exists, remove
		textGreenNode := tmpBody.Find(".text_green")
		textGreenNode.NextAll().Remove()
		textGreenNode.Remove()

		body, _ := tmpBody.Html()
		body = body[strings.Index(body, ": ")+1:]
		body = strings.TrimSpace(body)
		//log.Println("tmpBody======", body)
		comment.Body = body

		content.Comments[index] = *comment
		//str, _ := json.Marshal(comment)
		//log.Println("comment", string(str))

		//printToConsole("comment => ", comment)
	})
	//log.Println("asdf")

	return content, nil
}

func printToConsole(str string, v interface{}) {
	strr, _ := json.Marshal(v)
	log.Println(str, string(strr))
}
