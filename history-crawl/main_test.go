package main

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCrawlPage(t *testing.T) {
	t.Log("test")
}

func TestNgaPostRequest(t *testing.T) {
	client := http.Client{}
	req, err := ngaPostRequest(&client, "https://bbs.nga.cn/read.php?tid=27536822&page=57")
	if err != nil {
		t.Log(err)
	}
	guestJs, err := req.Cookie("guestJs")
	if err != nil {
		t.Error(err)
	}
	t.Log(guestJs)
}

func TestNgaExtractor(t *testing.T) {
	assert := assert.New(t)
	text := `
	<meta name='keywords' content=''>
	<title>[安科/安价] [原创] 我的女友是黄油女主这件事(心：女同竟在我身边) NGA玩家社区</title>
	<script type='text/javascript'>
		//loadscriptstart
var __CURRENT_UID = parseInt('',10),
class='pager_spacer'>下一页(23)</a><span id='pageBtnHere' class='x'></span>
<script>
	var __PAGE = {0:'/read.php?tid=29824736',1:82,2:22,3:20};commonui.pageBtn(document.getElementById('pageBtnHere').parentNode,__PAGE,true)
	
	
	class='pager_spacer'>下一页(23)</a><span id='pageBtnHere' class='x'></span>
<script>
	var __PAGE = {0:'/read.php?tid=29824736',1:82,2:22,3:20};commonui.pageBtn(document.getElementById('pageBtnHere').parentNode,__PAGE,true)
	
	sfas
`
	inf, err := ngaExtractor(text)
	assert.Equal(nil, err)
	assert.Equal("[安科/安价] [原创] 我的女友是黄油女主这件事(心：女同竟在我身边) NGA玩家社区", inf.Title)
	assert.Equal(82, inf.TotalPage)
}
