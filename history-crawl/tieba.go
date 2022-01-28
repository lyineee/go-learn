package main

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strconv"
)

func tiebaPostRequest(client *http.Client, crawlUrl string) (*http.Request, error) {
	req, err := http.NewRequest("GET", crawlUrl, nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func tiebaExtractor(text string) (info postInformation, err error) {
	totalPageRegex := regexp.MustCompile(`"total_page":(\d+)`)
	titleRegex := regexp.MustCompile(`title:.?"(.+?)"`)

	totalPageRaw := totalPageRegex.FindStringSubmatch(text)
	if len(totalPageRaw) < 2 {
		return info, errors.New("can not find total page")
	}
	totalPage, err := strconv.Atoi(totalPageRaw[1])
	if err != nil {
		return info, errors.New("can not find total page")
	}
	totalPage = totalPage / 2 //登录后每页贴数翻倍

	titleRaw := titleRegex.FindStringSubmatch(text)
	if len(totalPageRaw) < 2 {
		return info, errors.New("can not find title")
	}
	title := titleRaw[1]
	if string([]rune(title)[:3]) == "回复：" {
		title = string([]rune(title)[3:])
	}

	info.TotalPage = totalPage
	info.Title = title
	return info, err
}

func tiebaProc(ctx context.Context, history *History) error {
	page, err := crawlPage(history.Url, tiebaPostRequest)
	if err != nil {
		return err
	}
	info, err := tiebaExtractor(page)
	if info.Title == "" && info.TotalPage == 0 {
		logger.Error("get tieba info fail", "crawl page", page, "history", history)
		return errors.New("get info fail")
	} else if info.Title == "" || info.TotalPage == 0 {
		logger.Warn("fail get all tieba data", "crawl page", page, "history", history)
	}
	if err != nil {
		return err
	}
	history.Title = info.Title
	history.TotalPage = info.TotalPage
	return nil
}
