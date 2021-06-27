package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/lyineee/go-learn/utils"
)

func main() {
	// date := time.Now().Format("2006-01-02")
	date := "2021-06-27"
	log.Printf("----- start generate daily feed (%s) -----", date)
	env := utils.GetEnv()
	if env["COLLECTION_ID"] == "" {
		log.Fatal("Dont provide collection id")
	}
	if env["HYPOTHESIS_TOKEN"] == "" {
		log.Fatal("Dont provide hypothesis token")
	}
	if env["OUTLINE_TOKEN"] == "" {
		log.Fatal("Dont provide outline token")
	}
	result, err := getNotations(date, env["HYPOTHESIS_TOKEN"])
	if err != nil {
		log.Fatal(err)
	}
	if len(result) == 0 {
		log.Printf("No notations today on %s", date)
		return
	}
	markdown := genMarkdown(result)
	err = postArticle(markdown, date, env["COLLECTION_ID"], env["OUTLINE_TOKEN"])
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("----- end generate daily feed (%s) -----", date)

}

func ReadBody(body io.ReadCloser) {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))
}

func getNotations(dateStr string, token string) (hypoData []HypothesisNotation, err error) {
	//
	timeLayout := "2006-01-02"
	timer, err := time.Parse(timeLayout, dateStr)
	if err != nil {
		return
	}
	dua, err := time.ParseDuration("24h")
	if err != nil {
		return
	}
	timer = timer.Add(dua)
	endTime := timer.Format(timeLayout)
	//
	query := map[string]string{
		"user":         "acct:liuzhengyuan@hypothes.is",
		"search_after": dateStr,
	}
	reqUrl := fmt.Sprintf("https://hypothes.is/api/search?order=asc&user=%s&sort=updated&search_after=%s", query["user"], query["search_after"])
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	respJson := HypothesisResponse{}
	json.NewDecoder(resp.Body).Decode(&respJson)
	for _, notation := range respJson.Rows {
		if notation.Updated > endTime {
			continue
		}
		var dataPt *HypothesisNotation
		for dataIndex, data := range hypoData {
			if notation.Document.Title[0] == data.Title {
				dataPt = &hypoData[dataIndex]
			}
		}
		if dataPt == nil {
			hypoData = append(hypoData, HypothesisNotation{})
			dataPt = &hypoData[len(hypoData)-1]
			dataPt.Title = notation.Document.Title[0]
			dataPt.Url = notation.Uri
		}
		dataPt.Tags = append(dataPt.Tags, notation.Tags...)
		var note, quote string
		note = notation.Text
		for _, target := range notation.Target {
			for _, selector := range target.Selector {
				if text := selector.Exact; text != "" {
					quote = text
				}
			}
		}
		dataPt.Cite = append(dataPt.Cite, struct {
			Note  string
			Quote string
		}{note, quote})
	}
	return
}

func genMarkdown(notations []HypothesisNotation) (text string) {
	for _, notation := range notations {
		text += fmt.Sprintf("# %s\n", notation.Title)
		text += fmt.Sprintf("链接: %s\n", notation.Url)
		for _, tag := range notation.Tags {
			text += fmt.Sprintf("#%s ", tag)
		}
		text += "\n"
		for _, cite := range notation.Cite {
			if cite.Quote != "" {
				text += fmt.Sprintf("> %s \n\n", cite.Quote)
			}
			if cite.Note != "" {
				if cite.Note[0] == '#' {
					text += fmt.Sprintf("#%s \n", cite.Note)
				} else {
					text += fmt.Sprintf("%s \n", cite.Note)
				}
			}
		}
		text += "\n\n---\n"
	}
	return
}

func postArticle(text, title, collectionId, token string) (err error) {
	payload := OutlineArticle{
		Title:        title,
		Text:         text,
		CollectionId: collectionId,
		Publish:      "true",
	}
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", "https://wiki.lyine.pw:444/api/documents.create", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	respJson := OutlineResponse{}
	json.NewDecoder(resp.Body).Decode(&respJson)
	if respJson.Ok != "true" {
		err = errors.New(respJson.Error)
		return
	}
	return
}
