package main

type HisResp struct {
	Retcode string        `json:"retcode"`
	Errmsg  string        `json:"errmsg"`
	Card    []BalanceInfo `json:"card"`
}

type TodayResp struct {
	Retcode  string      `json:"retcode"`
	Errmsg   string      `json:"errmsg"`
	NextPage string      `json:"nextpage"`
	Total    []TodayInfo `json:"total"`
}

type BalanceInfo struct {
	DbBalance string `json:"db_balance"`
}

type TodayInfo struct {
	SignTranamt string `json:"sign_tranamt"`
	Tranamt     string `json:"tranamt"`
}
