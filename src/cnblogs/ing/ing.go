package ing

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
)

//Ing ing.cnblogs.com
type Ing struct {
	//IngID      int
	authCookie string
	urlStr     string
	client     *http.Client
}

//Init Initialize httpClient with authCookie
func (ing *Ing) Init(authCookie string) {
	ing.client = &http.Client{}
	ing.authCookie = authCookie
}

//GetIngByID Get Ing Html Document by ingID
func (ing *Ing) GetIngByID(ingID int) (string, error) {
	//urlStr := "https://ing.cnblogs.com/u/grj1046/status/" + strconv.Itoa(ingID) + "/"
	//http://home.cnblogs.com/ing/645536/
	urlStr := "https://ing.cnblogs.com/redirect/" + strconv.Itoa(ingID) + "/"
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Cookie", ing.authCookie)
	resp, err := ing.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", errors.New("StatusCode: " + strconv.Itoa(resp.StatusCode))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	bodystring := string(body)
	return bodystring, nil
}
