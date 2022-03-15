package main

import (
	"context"
	"encoding/json"
	"testing"
)

func TestGetRoomStatus(t *testing.T) {
	status, err := getRoomStatus(context.Background(), 92613)
	if err != nil {
		t.Log(err)
	}
	t.Log(status)
}

func TestAny(t *testing.T) {
	data := `
	{
		"code": 0,
		"message": "0",
		"ttl": 1,
		"data": {
		  "room_info": {
			"uid": 13046,
			"room_id": 92613,
			"short_id": 0,
			"title": "老头环，爆金币",
			"cover": "http://i0.hdslb.com/bfs/live/new_room_cover/d5ea683cce62b0753d7b90a4e2d3d227bfc87fcb.jpg",
			"tags": "Pi",
			"background": "http://i0.hdslb.com/bfs/live/room_bg/93f5baa781d70f43dbc84d83b520f90d65a5f614.jpg",
			"description": "<p>直播通知群号：24994464(新群) 432907038 561053359 183261781（排名不分先后）</p>\n<p>禁言准则</p>\n<p><span style=\"font-size: 14px; color: #ff0000;\"><strong><span style=\"font-family: 微软雅黑, Tahoma, 宋体; font-styl line !important; float: none; background-color: #ffffff;\">传话筒：1h</span></strong></span><br style=\"color: #808080; font-family: 微软雅黑, Tahoma, 宋体; font-siz one; white-space: normal; widows: 1; word-spacing: 0px; -webkit-text-stroke-width: 0px; background-color: #ffffff;\"><span style=\"font-size: 14px; color: #ff0000;\"><strong><span style=\"font-family: 微软雅黑, Tahoma, 宋体; font-styl line !important; float: none; background-color: #ffffff;\">点播，催更：1h</span></strong></span><br style=\"color: #808080; font-family: 微软雅黑, Tahoma, 宋体; font-siz one; white-space: normal; widows: 1; word-spacing: 0px; -webkit-text-stroke-width: 0px; background-color: #ffffff;\"><span style=\"font-size: 14px; color: #ff0000;\"><strong><span style=\"font-family: 微软雅黑, Tahoma, 宋体; font-styl line !important; float: none; background-color: #ffffff;\">带节奏：1h</span></strong></span><br style=\"color: #808080; font-family: 微软雅黑, Tahoma, 宋体; font-siz one; white-space: normal; widows: 1; word-spacing: 0px; -webkit-text-stroke-width: 0px; background-color: #ffffff;\"><span style=\"font-size: 14px; color: #ff0000;\"><strong><span style=\"font-family: 微软雅黑, Tahoma, 宋体; font-styl line !important; float: none; background-color: #ffffff;\">跟节奏警告后：1h</span></strong></span><br style=\"color: #808080; font-family: 微软雅黑, Tahoma, 宋体; font-siz one; white-space: normal; widows: 1; word-spacing: 0px; -webkit-text-stroke-width: 0px; background-color: #ffffff;\"><span style=\"font-size: 14px; color: #ff0000;\"><strong><span style=\"font-family: 微软雅黑, Tahoma, 宋体; font-styl line !important; float: none; background-color: #ffffff;\">恶意刷屏：24h</span></strong></span><br style=\"color: #808080; font-family: 微软雅黑, Tahoma, 宋体; font-siz one; white-space: normal; widows: 1; word-spacing: 0px; -webkit-text-stroke-width: 0px; background-color: #ffffff;\"><span style=\"font-size: 14px; color: #ff0000;\"><strong><span style=\"font-family: 微软雅黑, Tahoma, 宋体; font-styl line !important; float: none; background-color: #ffffff;\">恶意带节奏,玩梗，ky：72h</span></strong></span><br style=\"color: #808080; font-family: 微软雅黑, Tahoma, 宋体; font-siz one; white-space: normal; widows: 1; word-spacing: 0px; -webkit-text-stroke-width: 0px; background-color: #ffffff;\"><span style=\"font-size: 14px; color: #ff0000;\"><strong><span style=\"font-family: 微软雅黑, Tahoma, 宋体; font-styl line !important; float: none; background-color: #ffffff;\">广告,骂人：720h</span></strong></span></p>\n<p><span style=\"font-size: 14px; color: #ff0000;\"><strong><span style=\"font-family: 微软雅黑, Tahoma, 宋体; font-styl line !important; float: none; background-color: #ffffff;\"><span style=\"color: #ff9900;\">不过希望观众们对弹幕宽容一些，不守规矩的弹幕可能是并不知道这么回事，房管来处理就好了，没必要跟风骂的</span><br></span></strong></span></p>",
			"live_status": 1,
			"live_start_time": 1647259166,
			"live_screen_type": 0,
			"lock_status": 0,
			"lock_time": 0,
			"hidden_status": 0,
			"hidden_time": 0,
			"area_id": 555,
			"area_name": "艾尔登法环",
			"parent_area_id": 6,
			"parent_area_name": "单机游戏",
			"keyframe": "http://i0.hdslb.com/bfs/live-key-frame/keyframe03142211000000092613ys1hmm.jpg",
			"special_type": 0,
			"up_session": "218681231993301445",
			"pk_status": 0,
			"is_studio": false,
			"pendants": {
			  "frame": {
				"name": "春风拂动",
				"value": "https://i0.hdslb.com/bfs/live/9c8e45ec3673296952f41bc7a1ab7b8ad93ea6e4.png",
				"desc": ""
			  }
			},
			"on_voice_join": 0,
			"online": 678412,
			"room_type": {
			  "3-21": 0
			}
		  },
		  "anchor_info": {
			"base_info": {
			  "uname": "少年Pi",
			  "face": "http://i2.hdslb.com/bfs/face/6e3b84c1fe71caf523ed87d264f9026013af1c2c.jpg",
			  "gender": "男",
			  "official_info": {
				"role": -1,
				"title": "",
				"desc": "",
				"is_nft": 0,
				"nft_dmark": "https://i0.hdslb.com/bfs/live/9f176ff49d28c50e9c53ec1c3297bd1ee539b3d6.gif"
			  }
			},
			"live_info": {
			  "level": 40,
			  "level_color": 16746162,
			  "score": 425024676,
			  "upgrade_score": 0,
			  "current": [
				25000000,
				147013810
			  ],
			  "next": [],
			  "rank": "31"
			},
			"relation_info": {
			  "attention": 589907
			},
			"medal_info": {
			  "medal_name": "帅Pi",
			  "medal_id": 1786,
			  "fansclub": 16045
			}
		  },
		  "news_info": {
			"uid": 13046,
			"ctime": "2020-11-20 12:36:52",
			"content": "建了舰长群：639259920    发送加群信息以后请在B站私信也发个1，方便对照"
		  },
		  "rankdb_info": {
			"roomid": 92613,
			"rank_desc": "小时总榜",
			"color": "#FB7299",
			"h5_url": "https://live.bilibili.com/p/html/live-app-rankcurrent/index.html?is_live_half_webview=1&hybrid_half_ui=1,5,85p,70p,FFE293,0,30,100,10;2,2,320,100p,FFE293,0,30,100,0;4,2,320,100p,FFE293,0,30,100,0;6,5,65p,60p,FFE293,0,30,100,10;5,5,55p,60p,FFE293,0,30,100,10;3,5,85p,70p,FFE293,0,30,100,10;7,5,65p,60p,FFE293,0,30,100,10;&anchor_uid=13046&rank_type=master_realtime_hour_room&area_hour=1&area_v2_id=555&area_v2_parent_id=6",
			"web_url": "https://live.bilibili.com/blackboard/room-current-rank.html?rank_type=master_realtime_hour_room&area_hour=1&area_v2_id=555&area_v2_parent_id=6",
			"timestamp": 1647267406
		  },
		  "area_rank_info": {
			"areaRank": {
			  "index": 1,
			  "rank": "30"
			},
			"liveRank": {
			  "rank": "31"
			}
		  },
		  "battle_rank_entry_info": {
			"first_rank_img_url": "",
			"rank_name": "尚无段位",
			"show_status": 1
		  },
		  "tab_info": {
			"list": [
			  {
				"type": "seven-rank",
				"desc": "高能榜",
				"isFirst": 1,
				"isEvent": 0,
				"eventType": "",
				"listType": "",
				"apiPrefix": "",
				"rank_name": "room_7day"
			  },
			  {
				"type": "guard",
				"desc": "大航海",
				"isFirst": 0,
				"isEvent": 0,
				"eventType": "",
				"listType": "top-list",
				"apiPrefix": "",
				"rank_name": ""
			  }
			]
		  },
		  "activity_init_info": {
			"eventList": [],
			"weekInfo": {
			  "bannerInfo": null,
			  "giftName": null
			},
			"giftName": null,
			"lego": {
			  "timestamp": 1647267407,
			  "config": "[{\"name\":\"frame-mng\",\"url\":\"https:\\/\\/live.bilibili.com\\/p\\/html\\/live-web-mng\\/index.html?roomid=#roomid#&arae_id=#area_id#&parent_area_id=#parent_area_id#&ruid=#ruid#\",\"startTime\":1559544736,\"endTime\":1877167950,\"type\":\"frame-mng\"},{\"name\":\"s10-fun\",\"target\":\"sidebar\",\"icon\":\"https:\\/\\/i0.hdslb.com\\/bfs\\/activity-plat\\/static\\/20200908\\/3435f7521efc759ae1f90eae5629a8f0\\/HpxrZ7SOT.png\",\"text\":\"\\u7545\\u73a9s10\",\"url\":\"https:\\/\\/live.bilibili.com\\/s10\\/fun\\/index.html?room_id=#roomid#&width=376&height=600&source=sidebar\",\"color\":\"#2e6fc0\",\"startTime\":1600920000,\"endTime\":1604721600,\"parentAreaId\":2,\"areaId\":86},{\"name\":\"lottery-gift\",\"target\":\"sidebar\",\"icon\":\"https:\\/\\/i0.hdslb.com\\/bfs\\/activity-plat\\/static\\/20220127\\/3435f7521efc759ae1f90eae5629a8f0\\/HHm0Sw5ZTk.png\",\"text\":\"\\u661f\\u8fd0\\u793e\",\"url\":\"https:\\/\\/live.bilibili.com\\/activity\\/live-activity-grand\\/shopping-store\\/mobile.html?room_id=#roomid#&width=376&height=480&source=sidebar#\\/shopping\",\"color\":\"#2e6fc0\",\"startTime\":1643450400,\"endTime\":1644940799},{\"name\":\"genshin-avatar\",\"target\":\"sidebar\",\"icon\":\"https:\\/\\/i0.hdslb.com\\/bfs\\/activity-plat\\/static\\/20210721\\/fa538c98e9e32dc98919db4f2527ad02\\/qWxN1d0ACu.jpg\",\"text\":\"\\u539f\\u77f3\\u798f\\u5229\",\"url\":\"https:\\/\\/live.bilibili.com\\/activity\\/live-activity-full\\/genshin_avatar\\/mobile.html?no-jump=1&room_id=#roomid#&width=376&height=550#\\/\",\"color\":\"#2e6fc0\",\"frameAllowNoBg\":\"1\",\"frameAllowDrag\":\"1\",\"startTime\":1627012800,\"endTime\":1630425540,\"parentAreaId\":3,\"areaId\":321}]"
			}
		  },
		  "voice_join_info": {
			"status": {
			  "open": 0,
			  "anchor_open": 0,
			  "status": 0,
			  "uid": 0,
			  "user_name": "",
			  "head_pic": "",
			  "guard": 0,
			  "start_at": 0,
			  "current_time": 1647267407
			},
			"icons": {
			  "icon_close": "https://i0.hdslb.com/bfs/live/a176d879dffe8de1586a5eb54c2a08a0c7d31392.png",
			  "icon_open": "https://i0.hdslb.com/bfs/live/70f0844c9a12d29db1e586485954290144534be9.png",
			  "icon_wait": "https://i0.hdslb.com/bfs/live/1049bb88f1e7afd839cc1de80e13228ccd5807e8.png",
			  "icon_starting": "https://i0.hdslb.com/bfs/live/948433d1647a0704f8216f017c406224f9fff518.gif"
			},
			"web_share_link": "https://live.bilibili.com/h5/92613"
		  },
		  "ad_banner_info": {
			"data": [
			  {
				"id": 138779,
				"title": "永劫无间激励计划（第八期）",
				"location": "room_advertisement",
				"position": 3,
				"pic": "https://i0.hdslb.com/bfs/live/5131ae5e49be829e5145c4e48d00fadd46c96092.jpg",
				"link": "https://www.bilibili.com/blackboard/activity-GQXe26f6fM.html",
				"weight": 0,
				"room_id": 0,
				"up_id": 0,
				"parent_area_id": 0,
				"area_id": 0,
				"live_status": 0,
				"av_id": 0
			  },
			  {
				"id": 138257,
				"title": "这样玩扭蛋，iPhone13抢不停",
				"location": "room_advertisement",
				"position": 4,
				"pic": "https://i0.hdslb.com/bfs/live/2ed3e410a8b76ca8348ff60ebd0a3823c6224627.jpg",
				"link": "https://live.bilibili.com/activity/live-activity-full/full-next/pc.html?is_live_full_webview=1&app_name=bls_spring_2022&is_live_webview=1&visit_id=1jgojq64vu3k#/small",
				"weight": 0,
				"room_id": 0,
				"up_id": 0,
				"parent_area_id": 0,
				"area_id": 0,
				"live_status": 0,
				"av_id": 0
			  },
			  {
				"id": 138532,
				"title": "命运2账号绑定上线",
				"location": "room_advertisement",
				"position": 5,
				"pic": "https://i0.hdslb.com/bfs/live/9e4269e99d40a0e16e6e0fec1be9baa7bfe1f148.jpg",
				"link": "https://www.bilibili.com/blackboard/activity-S6PDb7v2Vk.html",
				"weight": 0,
				"room_id": 0,
				"up_id": 0,
				"parent_area_id": 0,
				"area_id": 0,
				"live_status": 0,
				"av_id": 0
			  }
			]
		  },
		  "skin_info": {
			"id": 338,
			"skin_name": "春之召唤",
			"skin_config": "{\"zip\":\"http://i0.hdslb.com/bfs/live/5d055b3f4743e6a3ae0eee31ab89da09205360f2.zip\",\"md5\":\"FF8FD277C37EB2B27398DB71D9EFB9CB\",\"platform\":\"web\",\"version\":\"1\",\"headInfoBgPic\":\"http://i0.hdslb.com/bfs/live/d048b155cd305b7604ebc334b77eb1e15b7f5719.jpg\",\"giftControlBgPic\":\"http://i0.hdslb.com/bfs/live/ee9efffa5086694c0add50b32a426c92f811763d.jpg\",\"rankListBgPic\":\"http://i0.hdslb.com/bfs/live/0fb1d57b9d6656eb1eab2c682314007e505c1c49.jpg\",\"mainText\":\"#FFffffff\",\"normalText\":\"#FFffe2b2\",\"highlightContent\":\"#FFd99d1b\",\"border\":\"#FF050e45\",\"infoCardBgPic\":\"\"}",
			"show_text": "资格赛晋级主播奖励",
			"skin_url": "https://i0.hdslb.com/bfs/live/0c881097219ea41d7ee45307df2d4a79fa74c884.png",
			"start_time": 1646910907,
			"end_time": 1647515707,
			"current_time": 1647267407
		  },
		  "web_banner_info": {
			"id": 0,
			"title": "",
			"left": "",
			"right": "",
			"jump_url": "",
			"bg_color": "",
			"hover_color": "",
			"text_bg_color": "",
			"text_hover_color": "",
			"link_text": "",
			"link_color": "",
			"input_color": "",
			"input_text_color": "",
			"input_hover_color": "",
			"input_border_color": "",
			"input_search_color": ""
		  },
		  "lol_info": {
			"lol_activity": {
			  "status": 0,
			  "guess_cover": "http://i0.hdslb.com/bfs/live/61d1c4bcce470080a5408d6c03b7b48e0a0fa8d7.png",
			  "vote_cover": "https://i0.hdslb.com/bfs/activity-plat/static/20190930/4ae8d4def1bbff9483154866490975c2/oWyasOpox.png",
			  "vote_h5_url": "https://live.bilibili.com/p/html/live-app-wishhelp/index.html?is_live_half_webview=1&hybrid_biz=live-app-wishhelp&hybrid_rotate_d=1&hybrid_half_ui=1,3,100p,360,0c1333,0,30,100;2,2,375,100p,0c1333,0,30,100;3,3,100p,360,0c1333,0,30,100;4,2,375,100p,0c1333,0,30,100;5,3,100p,360,0c1333,0,30,100;6,3,100p,360,0c1333,0,30,100;7,3,100p,360,0c1333,0,30,100;8,3,100p,360,0c1333,0,30,100;",
			  "vote_use_h5": true
			}
		  },
		  "pk_info": null,
		  "battle_info": null,
		  "silent_room_info": {
			"type": "level",
			"level": 1,
			"second": -1,
			"expire_time": 2145888000
		  },
		  "switch_info": {
			"close_guard": false,
			"close_gift": false,
			"close_online": false,
			"close_danmaku": false
		  },
		  "record_switch_info": {
			"record_tab": false
		  },
		  "room_config_info": {
			"dm_text": "发个弹幕呗~"
		  },
		  "gift_memory_info": {
			"list": null
		  },
		  "new_switch_info": {
			"room-socket": 1,
			"room-prop-send": 1,
			"room-sailing": 1,
			"room-info-popularity": 1,
			"room-danmaku-editor": 1,
			"room-effect": 1,
			"room-fans_medal": 1,
			"room-report": 1,
			"room-feedback": 1,
			"room-player-watermark": 1,
			"room-recommend-live_off": 1,
			"room-activity": 1,
			"room-web_banner": 1,
			"room-silver_seeds-box": 1,
			"room-wishing_bottle": 1,
			"room-board": 1,
			"room-supplication": 1,
			"room-hour_rank": 1,
			"room-week_rank": 1,
			"room-anchor_rank": 1,
			"room-info-integral": 1,
			"room-super-chat": 1,
			"room-tab": 1,
			"room-hot-rank": 1,
			"fans-medal-progress": 1,
			"gift-bay-screen": 1,
			"room-enter": 1,
			"room-my-idol": 1,
			"room-topic": 1,
			"fans-club": 1
		  },
		  "super_chat_info": {
			"status": 1,
			"jump_url": "https://live.bilibili.com/p/html/live-app-superchat2/index.html?is_live_half_webview=1&hybrid_half_ui=1,3,100p,70p,ffffff,0,30,100;2,2,375,100p,ffffff,0,30,100;3,3,100p,70p,ffffff,0,30,100;4,2,375,100p,ffffff,0,30,100;5,3,100p,60p,ffffff,0,30,100;6,3,100p,60p,ffffff,0,30,100;7,3,100p,60p,ffffff,0,30,100",
			"icon": "https://i0.hdslb.com/bfs/live/0a9ebd72c76e9cbede9547386dd453475d4af6fe.png",
			"ranked_mark": 0,
			"message_list": []
		  },
		  "online_gold_rank_info_v2": {
			"list": [
			  {
				"uid": 39914863,
				"face": "http://i1.hdslb.com/bfs/face/b69eb70673c2fd78f73df18cee12fd69f4b6e6be.jpg",
				"uname": "要低调灬",
				"score": "2000",
				"rank": 1,
				"guard_level": 3
			  },
			  {
				"uid": 2437811,
				"face": "http://i2.hdslb.com/bfs/face/80c65810f06c4b5b814c2f3557f1434632b476b0.jpg",
				"uname": "菈妮情夫褪色者",
				"score": "601",
				"rank": 2,
				"guard_level": 0
			  },
			  {
				"uid": 4669315,
				"face": "http://i1.hdslb.com/bfs/face/b3b4442d36a776f39bd9260e129a48687aa0e7fc.jpg",
				"uname": "苧晓淅",
				"score": "385",
				"rank": 3,
				"guard_level": 0
			  },
			  {
				"uid": 10450805,
				"face": "http://i2.hdslb.com/bfs/face/72a88e345142fbaa0869e69b27781de8de8d509a.jpg",
				"uname": "緋赤perio",
				"score": "305",
				"rank": 4,
				"guard_level": 0
			  },
			  {
				"uid": 320843,
				"face": "http://i2.hdslb.com/bfs/face/83fba7ab6c6f4aa62dea84595ce9858d6fc30426.jpg",
				"uname": "岚影SC",
				"score": "305",
				"rank": 5,
				"guard_level": 3
			  },
			  {
				"uid": 7998634,
				"face": "http://i1.hdslb.com/bfs/face/b8b347a64bf943c119397f8feedca6d7042f5980.jpg",
				"uname": "乌拉拉拉不拉拉布拉多",
				"score": "302",
				"rank": 6,
				"guard_level": 3
			  },
			  {
				"uid": 28197336,
				"face": "http://i0.hdslb.com/bfs/face/360fbf5a897d9a209d93b28cd23c6a7898b22456.jpg",
				"uname": "智慧满满的琪露诺亲",
				"score": "300",
				"rank": 7,
				"guard_level": 0
			  }
			]
		  },
		  "dm_emoticon_info": {
			"is_open_emoticon": 1,
			"is_shield_emoticon": 0
		  },
		  "dm_tag_info": {
			"dm_tag": 0,
			"platform": [],
			"extra": "",
			"dm_chronos_extra": "",
			"dm_mode": [],
			"dm_setting_switch": 0
		  },
		  "topic_info": {
			"topic_id": 0,
			"topic_name": ""
		  },
		  "game_info": {
			"game_status": 0
		  },
		  "watched_show": {
			"switch": true,
			"num": 20933,
			"text_small": "2.0万",
			"text_large": "2.0万人看过",
			"icon": "",
			"icon_location": 0,
			"icon_web": ""
		  },
		  "video_connection_info": null,
		  "player_throttle_info": {
			"status": 1,
			"normal_sleep_time": 1800,
			"fullscreen_sleep_time": 3600,
			"tab_sleep_time": 1800,
			"prompt_time": 30
		  },
		  "guard_info": {
			"count": 608,
			"anchor_guard_achieve_level": 100
		  },
		  "hot_rank_info": {
			"rank": 0,
			"trend": 0,
			"countdown": 793,
			"timestamp": 1647267407,
			"url": "https://live.bilibili.com/p/html/live-app-hotrank/index.html?clientType=2&area_id=0&parent_area_id=0&second_area_id=0",
			"icon": "",
			"area_name": "",
			"new_data": {
			  "rank": 0,
			  "trend": 0,
			  "countdown": 0,
			  "timestamp": 1647267407,
			  "url": "https://live.bilibili.com/p/html/live-app-hotrank/index.html?clientType=2&area_id=0&parent_area_id=0&second_area_id=0",
			  "icon": "",
			  "area_name": "",
			  "rank_desc": "未上榜"
			}
		  }
		}
	  }`
	s := statusResp{}
	err := json.Unmarshal([]byte(data), &s)
	if err != nil {
		t.Log(err)
	}
	t.Log(s)
}
