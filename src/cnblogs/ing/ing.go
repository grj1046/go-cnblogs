package ing

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"errors"

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
	IngID          int
	AuthorID       string
	AuthorUserName string
	AuthorNickName string
	Time           string
	Status         int
	Lucky          bool
	IsPrivate      bool
	IsNewbie       bool
	AcquiredAt     time.Time
	Body           string
	Comments       []Comment
}

//Comment Ing.Content's Comment
type Comment struct {
	IngID          int
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
	IngID      int
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
func (client *Client) GetIngByID(ingID int) (*Content, *OriginContent, error) {
	//urlStr := "https://ing.cnblogs.com/u/grj1046/status/" + strconv.Itoa(ingID) + "/"
	//http://home.cnblogs.com/ing/1115171/
	urlStr := "https://ing.cnblogs.com/redirect/" + strconv.Itoa(ingID) + "/"
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

	nowTime := time.Now()

	originContent := &OriginContent{}
	originContent.IngID = ingID
	originContent.HTML = ""
	originContent.Status = 200
	originContent.AcquiredAt = nowTime

	content := &Content{}
	content.IngID = ingID
	content.AcquiredAt = nowTime
	content.Status = 200

	if resp.StatusCode != 200 {
		content.Status = resp.StatusCode
		originContent.Status = resp.StatusCode
		return content, originContent, nil
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, nil, err
	}
	feedBlock, err := doc.Find(".feed_block").Html()
	if err != nil {
		originContent.Exception += " Get feed_block error: " + err.Error()
	} else {
		originContent.HTML = feedBlock
	}
	errBody := doc.Find(".error_body")
	//if return 404
	if errBody.Text() != "" {
		content.Status = 404
		originContent.Status = 404
		return content, originContent, nil
	}

	if doc.Find("#Main form #Heading").Text() == "登录博客园 - 代码改变世界" {
		//need re acquired
		content.Status = 403
		originContent.Status = 403
		return content, originContent, nil
	}

	//AuthorID
	authorID, exists := doc.Find(".ing_item_face").Attr("src")
	if exists {
		//https://pic.cnblogs.com/face/sample_face.gif  test case: ingid=26
		//https://pic.cnblogs.com/face/289132/20130423092122.png
		if strings.Index(authorID, "sample_face.gif") != -1 {
			//replyToSpaceUserId=9931;isIngItem=true
			ret, err := doc.Html()
			if err != nil {
				originContent.Exception += " Get sample_face.gif error: " + err.Error()
			} else {
				start := strings.Index(ret, "replyToSpaceUserId=") + len("replyToSpaceUserId=")
				end := strings.Index(ret, ";isIngItem=")
				if start != -1 && end != -1 {
					authorID = ret[start:end]
				} else {
					originContent.Exception += " get AuthorID failed: sample_face.gif"
				}
			}
		} else if strings.Index(authorID, "https://pic.cnblogs.com/face/u") != -1 {
			tmplen := len("https://pic.cnblogs.com/face/u")
			if strings.Index(authorID, ".jpg") != -1 {
				authorID = authorID[tmplen:strings.Index(authorID, ".jpg")]
			} else if strings.Index(authorID, ".gif") != -1 {
				authorID = authorID[tmplen:strings.Index(authorID, ".gif")]
			} else if strings.Index(authorID, ".jpeg") != -1 {
				authorID = authorID[tmplen:strings.Index(authorID, ".jpeg")]
			} else if strings.Index(authorID, ".png") != -1 {
				authorID = authorID[tmplen:strings.Index(authorID, ".png")]
			} else if strings.Index(authorID, ".bmp") != -1 {
				authorID = authorID[tmplen:strings.Index(authorID, ".bmp")]
			} else {
				originContent.Exception += " get AuthorID failed: (face/u)" + authorID
			}
		} else {
			tmplen := len("https://pic.cnblogs.com/face/")
			if strings.LastIndex(authorID, "/") > tmplen {
				authorID = authorID[tmplen:strings.LastIndex(authorID, "/")]
			} else {
				originContent.Exception += " get AuthorID failed: (face)" + authorID
			}
		}
		content.AuthorID = authorID
	} else {
		originContent.Exception += " get ing_item_face failed"
	}

	//AuthorUserName
	authorUserName, exists := doc.Find(".ing_item_author").Attr("href")
	if exists {
		tmplen := len("//home.cnblogs.com/u/")
		if strings.Index(authorUserName, "//home.cnblogs.com/u/") != -1 {
			authorUserName = authorUserName[tmplen : len(authorUserName)-1]
			content.AuthorUserName = authorUserName
		} else {
			originContent.Exception += " get AuthorUserName failed"
		}
	} else {
		originContent.Exception += " Get ing_item_author error: " + err.Error()
	}

	//AuthorNickName
	authorNickName := doc.Find(".ing_item_author").Text()
	content.AuthorNickName = authorNickName

	publishTime := doc.Find(".ing_detail_title").Text()

	publishTime = publishTime[strings.Index(publishTime, "：")+3:]
	publishTime = strings.TrimSpace(publishTime)
	content.Time = publishTime
	//Lucky
	ingDetailBody := doc.Find("#ing_detail_body")
	luckyNode := ingDetailBody.Find("img[title='这是幸运闪']")
	_, exists = luckyNode.Attr("title")
	if exists {
		content.Lucky = true
		luckyNode.Remove()
	} else {
		content.Lucky = false
	}
	//Private
	privateNode := ingDetailBody.Find("img[title='私有闪存']")
	_, exists = privateNode.Attr("title")
	if exists {
		content.IsPrivate = true
		privateNode.Remove()
	} else {
		content.IsPrivate = false
	}
	//newbie
	newbieNode := ingDetailBody.Find("img[title='欢迎新人']")
	_, exists = newbieNode.Attr("title")
	if exists {
		content.IsNewbie = true
		newbieNode.Remove()
	} else {
		content.IsNewbie = false
	}

	//ingBody
	ingBody, err := ingDetailBody.Html()
	if err != nil {
		originContent.Exception += " Get ing_detail_body error: " + err.Error()
	} else {
		content.Body = ingBody
	}

	commentCount := doc.Find("#comment_block_" + strconv.Itoa(ingID) + " li").Length()
	content.Comments = make([]Comment, commentCount)

	doc.Find("#comment_block_" + strconv.Itoa(ingID) + " li").Each(func(index int, selection *goquery.Selection) {
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
		//commentReply(1129969,1415060,9487);return false
		authorID, exists := selection.Find(".gray3").Attr("onclick")
		//https://pic.cnblogs.com/face/sample_face.gif
		//https://pic.cnblogs.com/face/289132/20130423092122.png
		if !exists {
			//search with class="ing_comment_face".src
			selfAuthorID, selfExists := selection.Find(".ing_comment_face").Attr("src")
			if selfExists {
				if strings.Index(selfAuthorID, "https://pic.cnblogs.com/face/u") != -1 {
					tmplen := len("https://pic.cnblogs.com/face/u")
					if strings.Index(selfAuthorID, ".jpg") != -1 {
						selfAuthorID = selfAuthorID[tmplen:strings.Index(selfAuthorID, ".jpg")]
					} else if strings.Index(selfAuthorID, ".gif") != -1 {
						selfAuthorID = selfAuthorID[tmplen:strings.Index(selfAuthorID, ".gif")]
					} else if strings.Index(selfAuthorID, ".jpeg") != -1 {
						selfAuthorID = selfAuthorID[tmplen:strings.Index(selfAuthorID, ".jpeg")]
					} else if strings.Index(selfAuthorID, ".png") != -1 {
						selfAuthorID = selfAuthorID[tmplen:strings.Index(selfAuthorID, ".png")]
					} else if strings.Index(selfAuthorID, ".bmp") != -1 {
						selfAuthorID = selfAuthorID[tmplen:strings.Index(selfAuthorID, ".bmp")]
					} else {
						originContent.Exception += " get selfAuthorID failed: (face/u)" + selfAuthorID
					}
				} else {
					tmplen := len("https://pic.cnblogs.com/face/")
					if strings.LastIndex(selfAuthorID, "/") > tmplen {
						selfAuthorID = selfAuthorID[tmplen:strings.LastIndex(selfAuthorID, "/")]
					} else {
						originContent.Exception += " get selfAuthorID failed: (face)" + selfAuthorID
					}
				}
				if selfAuthorID != "" {
					comment.AuthorID = selfAuthorID
				}
			} else {
				originContent.Exception += " AuthorID not found by .gray3, ing_comment_face, index: " + string(index)
			}
		} else {
			start := strings.LastIndex(authorID, ",")
			end := strings.Index(authorID, ");")
			if start != -1 && end != -1 {
				authorID = authorID[start+1 : end]
				comment.AuthorID = authorID
			} else {
				originContent.Exception += "get comment AuthorID error"
			}
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

//GetLatestIngList get the latest ingList
//https://ing.cnblogs.com/ajax/ing/GetIngList?IngListType=all&PageIndex=1&PageSize=30&Tag=&_=1495616106104
func (client *Client) GetMaxIngID() (int, error) {
	urlStr := "https://ing.cnblogs.com/ajax/ing/GetIngList?IngListType=all&PageIndex=1&PageSize=1&Tag=&_=" + strconv.FormatInt(time.Now().Unix(), 10)
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return 0, err
	}
	//req.Header.Add("Accept", "text/plain, */*; q=0.01")
	req.Header.Add("Cookie", client.authCookie)
	//req.Header.Add("Referer", "https://ing.cnblogs.com/")
	//req.Header.Add("Host", "ing.cnblogs.com")
	//req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, errors.New("Response StatusCode" + strconv.Itoa(resp.StatusCode))
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return 0, err
	}

	maxIngID := doc.Find("#max_ing_id").Text()
	intMaxIngID, err := strconv.Atoi(maxIngID)
	if err != nil {
		return 0, err
	}
	return intMaxIngID, nil
}

//GetLatestIngFromComment get the latest ing's comment list ingID
//https://ing.cnblogs.com/ajax/ing/GetIngList?IngListType=recentcomment&PageIndex=1&PageSize=30&Tag=&_=1495616250086
func (client *Client) GetLatestIngFromComment(pageSize int) ([]int, error) {
	if pageSize <= 0 {
		pageSize = 30
	}
	urlStr := "https://ing.cnblogs.com/ajax/ing/GetIngList?IngListType=recentcomment&PageIndex=1&PageSize=" + strconv.Itoa(pageSize) +
		"&Tag=&_=" + strconv.FormatInt(time.Now().Unix(), 10)

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Cookie", client.authCookie)
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("Response StatusCode" + strconv.Itoa(resp.StatusCode))
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}

	ingList := make([]int, pageSize)
	//fmt.Println("length", doc.Find("#feed_list ul li .ing-item .ing_body").Length())

	doc.Find("#feed_list ul li .ing-item .ing_body").Each(func(index int, selection *goquery.Selection) {
		attrID, exists := selection.Attr("id")
		if exists {
			tmpLen := len("ing_body_")
			intVal, err := strconv.Atoi(attrID[tmpLen:])
			if err == nil {
				ingList[index] = intVal
			}
		}
	})
	return ingList, nil
}

func printToConsole(str string, v interface{}) {
	strr, _ := json.Marshal(v)
	log.Println(str, string(strr))
}
