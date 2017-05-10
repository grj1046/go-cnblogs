package ing

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	IngID          string
	AuthorID       string
	AuthorUserName string
	AuthorNickName string
	Time           string
	Status         int
	Lucky          bool
	IsPrivate      bool
	AcquiredAt     time.Time
	Body           string
	Comments       []Comment
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
}

//OriginContent store the origin ing html
type OriginContent struct {
	IngID      string
	URL        string
	Status     int //200 404
	AcquiredAt time.Time
	Exception  string
	HTML       string
}

//Init Initialize httpClient with authCookie
func (client *Client) Init(authCookie string) {
	client.httpClient = &http.Client{}
	client.authCookie = authCookie
}

//GetIngByID Get Ing Html Document by ingID
func (client *Client) GetIngByID(ingID string) (*Content, *OriginContent, error) {
	//urlStr := "https://ing.cnblogs.com/u/grj1046/status/" + strconv.Itoa(ingID) + "/"
	//http://home.cnblogs.com/ing/1115171/
	urlStr := "https://ing.cnblogs.com/redirect/" + ingID + "/"
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Add("Cookie", client.authCookie)
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, nil, errors.New("StatusCode: " + strconv.Itoa(resp.StatusCode))
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, nil, err
	}
	nowTime := time.Now()

	originContent := &OriginContent{}
	originContent.IngID = ingID
	originContent.HTML = ""
	originContent.URL = req.URL.String()
	originContent.Status = 200
	originContent.AcquiredAt = nowTime

	feedBlock, err := doc.Find(".feed_block").Html()
	if err != nil {
		originContent.Exception += " Get feed_block error: " + err.Error()
	} else {
		originContent.HTML = feedBlock
	}

	content := &Content{}
	content.IngID = ingID
	content.AcquiredAt = nowTime
	content.Status = 200

	errBody := doc.Find(".error_body")
	//if return 404
	if errBody.Text() != "" {
		content.Status = 404
		originContent.Status = 404
		return content, originContent, nil
	}
	//AuthorID
	authorID, exists := doc.Find(".ing_item_face").Attr("src")
	if exists {
		//https://pic.cnblogs.com/face/289132/20130423092122.png
		if strings.Index(authorID, "https://pic.cnblogs.com/face/u") != -1 {
			tmplen := len("https://pic.cnblogs.com/face/u")
			authorID = authorID[tmplen:strings.Index(authorID, ".jpg")]
		} else {
			tmplen := len("https://pic.cnblogs.com/face/")
			authorID = authorID[tmplen:strings.LastIndex(authorID, "/")]
		}
		content.AuthorID = authorID
	} else {
		originContent.Exception += " get ing_item_face failed"
	}

	//AuthorUserName
	authorUserName, exists := doc.Find(".ing_item_author").Attr("href")
	if exists {
		tmplen := len("//home.cnblogs.com/u/")
		authorUserName = authorUserName[tmplen : len(authorUserName)-1]
		content.AuthorUserName = authorUserName
	} else {
		originContent.Exception += " Get ing_item_author error: " + err.Error()
	}

	//AuthorNickName
	authorNickName := doc.Find(".ing_item_author").Text()
	content.AuthorNickName = authorNickName

	publishTime := doc.Find(".ing_detail_title").Text()
	//log.Println("publishTime>>>", publishTime)
	publishTime = publishTime[strings.Index(publishTime, "：")+3:]
	publishTime = strings.TrimSpace(publishTime)
	content.Time = publishTime
	//Lucky
	ingDetailBody := doc.Find("#ing_detail_body")
	_, exists = ingDetailBody.Find(".ing-icon").Attr("title")
	if exists {
		content.Lucky = true
		ingDetailBody.Find(".ing-icon").Remove()
	} else {
		content.Lucky = false
	}
	//Private
	privateNode := ingDetailBody.Find("img[title='私有闪存']")
	_, exists = privateNode.Attr("title")
	if exists {
		content.IsPrivate = true
	} else {
		content.IsPrivate = false
	}
	privateNode.Remove()

	//ingBody
	ingBody, err := ingDetailBody.Html()
	if err != nil {
		originContent.Exception += " Get ing_detail_body error: " + err.Error()
	} else {
		content.Body = ingBody
	}

	commentCount := doc.Find("#comment_block_" + ingID + " li").Length()
	content.Comments = make([]Comment, commentCount)

	doc.Find("#comment_block_" + ingID + " li").Each(func(index int, selection *goquery.Selection) {
		//IngID, CommentID, Comment, CommentTime, AuthorID AuthorUserName, AuthorNickName
		comment := &Comment{}
		//IngID
		comment.IngID = ingID
		//CommentID id="comment_1400623"
		commentID, exists := selection.Attr("id")
		if !exists {
			originContent.Exception += " commentID not found by id='comment_1400623', index: " + string(index)
		} else {
			tmplen := len("comment_")
			comment.CommentID = commentID[tmplen:]
		}

		//CommentTime class="text_green"
		time, exists := selection.Find(".text_green").Attr("title")
		if !exists {
			originContent.Exception += " comment time not found by .text_green, index: " + string(index)
		} else {
			comment.Time = time
		}

		//AuthorID
		authorID, exists := selection.Find(".ing_comment_face").Attr("src")
		if !exists {
			originContent.Exception += " AuthorID not found by .ing_comment_face, index: " + string(index)
		} else {
			//https://pic.cnblogs.com/face/289132/20130423092122.png
			if strings.Index(authorID, "https://pic.cnblogs.com/face/u") != -1 {
				tmplen := len("https://pic.cnblogs.com/face/u")
				authorID = authorID[tmplen:strings.Index(authorID, ".jpg")]
			} else {
				tmplen := len("https://pic.cnblogs.com/face/")
				authorID = authorID[tmplen:strings.LastIndex(authorID, "/")]
			}
			comment.AuthorID = authorID
		}

		authorNode := selection.Find("#comment_author_" + comment.CommentID)
		//AuthorName //home.cnblogs.com/u/grj1046/
		authorUserName, exists := authorNode.Attr("href")
		if !exists {
			originContent.Exception += " AuthorName not found by #comment_author_.href, index: " + string(index)
		} else {
			tmplen := len("//home.cnblogs.com/u/")
			authorUserName = authorUserName[tmplen : len(authorUserName)-1]
			comment.AuthorUserName = authorUserName
		}

		//AuthorNickName comment_author_1400623
		comment.AuthorNickName = authorNode.Text()

		//Comment
		tmpBody := selection.Find("div")
		/*<a target="_blank" href="//home.cnblogs.com/u/grj1046/">
		  	  <img src="https://pic.cnblogs.com/face/289132/20130423092122.png" class="ing_comment_face" alt="">
		  </a>
		*/
		tmpBody.Find(".ing_comment_face").Parent().Remove()
		//<a target="_blank" id="comment_author_1400623" title="nil的闪存" href="//home.cnblogs.com/u/grj1046/">nil</a>
		tmpBody.Find("#comment_author_" + comment.CommentID).Remove()
		tmpBody = tmpBody.First().Remove()
		//if delete button exists, remove
		textGreenNode := tmpBody.Find(".text_green")
		textGreenNode.NextAll().Remove()
		textGreenNode.Remove()

		body, err := tmpBody.Html()
		if err != nil {
			originContent.Exception += " Get comment detail exception, index: " + err.Error()
		} else {
			body = body[strings.Index(body, ": ")+1:]
			body = strings.TrimSpace(body)
			comment.Body = body
		}
		content.Comments[index] = *comment
		//printToConsole("comment => ", comment)
	})
	return content, originContent, nil
}

func printToConsole(str string, v interface{}) {
	strr, _ := json.Marshal(v)
	log.Println(str, string(strr))
}
