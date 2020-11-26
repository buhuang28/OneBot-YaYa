package cqhttp

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	xml "encoding/xml"
	"github.com/tidwall/gjson"

	"yaya/core"
)

type Result struct {
	Status  string      `json:"status"`
	Retcode int64       `json:"retcode"`
	Data    interface{} `json:"data"`
	Echo    interface{} `json:"echo"`
}

type Reply []map[string]interface{}

type XQGroupMemberList struct {
	List []*XQGroupMembers `json:"list"`
}
type XQGroupMembers struct {
	QQ  int64 `json:"QQ"`
	Lv  int64 `json:"lv"`
	Val int64 `json:"val"`
}
type CQGroupMember struct {
	GroupID         int64  `json:"group_id"`
	UserID          int64  `json:"user_id"`
	Nickname        string `json:"nickname"`
	Card            string `json:"card"`
	Sex             string `json:"sex"`
	Age             int64  `json:"age"`
	Area            string `json:"area"`
	JoinTime        int64  `json:"join_time"`
	LastSentTime    int64  `json:"last_sent_time"`
	Level           string `json:"level"`
	Role            string `json:"role"`
	Unfriendly      bool   `json:"unfriendly"`
	Title           string `json:"title"`
	TitleExpireTime int64  `json:"title_expire_time"`
	CardChangeable  bool   `json:"card_changeable"`
}
type CQGroupInfo struct {
	GroupID        int64  `json:"group_id"`
	GroupName      string `json:"group_name"`
	MemberCount    int64  `json:"member_count"`
	MaxMemberCount int64  `json:"max_member_count"`
}
type CQFriendInfo struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Remark   string `json:"remark"`
}

func (c *WSCYaml) wscApi(api []byte) {
	defer func() {
		if err := recover(); err != nil {
			ERROR("[响应服务] Bot %v 响应 %v 发生错误，忽略本次响应...... %v", c.BotID, c.Url, err)
		}
	}()

	req := gjson.ParseBytes(api)
	action := strings.ReplaceAll(req.Get("action").Str, "_async", "")
	params := req.Get("params")
	DEBUG("[响应服务] Bot %v 接收 %v 到API调用: %v 参数: %v", c.BotID, c.Url, req.Get("action").Str, string(api))

	if f, ok := wsApi[action]; ok {
		ret := f(c.BotID, params)
		if req.Get("echo").Int() != 0 {
			ret.Echo = req.Get("echo").Int()
		} else if req.Get("echo").Str != "" {
			ret.Echo = req.Get("echo").Str
		} else {
			ret.Echo, _ = req.Get("echo").Value().(map[string]interface{})
		}
		send, _ := json.Marshal(ret)
		c.Event <- send
	} else {
		ret := resultFail("no such api")
		if req.Get("echo").Int() != 0 {
			ret.Echo = req.Get("echo").Int()
		} else if req.Get("echo").Str != "" {
			ret.Echo = req.Get("echo").Str
		} else {
			ret.Echo, _ = req.Get("echo").Value().(map[string]interface{})
		}
		send, _ := json.Marshal(ret)
		c.Event <- send
	}
}

func (s *WSSYaml) wscApi(api []byte) {
	defer func() {
		if err := recover(); err != nil {
			ERROR("[响应服务] Bot %v 响应 %v:%v 发生错误，忽略本次响应...... %v", s.BotID, s.Host, s.Port, err)
		}
	}()

	req := gjson.ParseBytes(api)
	action := strings.ReplaceAll(req.Get("action").Str, "_async", "")
	params := req.Get("params")
	DEBUG("[响应服务] Bot %v 接收 %v:%v 到API调用: %v 参数: %v", s.BotID, s.Host, s.Port, req.Get("action").Str, string(api))

	if f, ok := wsApi[action]; ok {
		ret := f(s.BotID, params)
		if req.Get("echo").Int() != 0 {
			ret.Echo = req.Get("echo").Int()
		} else if req.Get("echo").Str != "" {
			ret.Echo = req.Get("echo").Str
		} else {
			ret.Echo, _ = req.Get("echo").Value().(map[string]interface{})
		}
		send, _ := json.Marshal(ret)
		s.Event <- send
	} else {
		ret := resultFail("no such api")
		if req.Get("echo").Int() != 0 {
			ret.Echo = req.Get("echo").Int()
		} else if req.Get("echo").Str != "" {
			ret.Echo = req.Get("echo").Str
		} else {
			ret.Echo, _ = req.Get("echo").Value().(map[string]interface{})
		}
		send, _ := json.Marshal(ret)
		s.Event <- send
	}
}

func (h *HTTPYaml) wscApi(path string, api []byte) []byte {
	defer func() {
		if err := recover(); err != nil {
			ERROR("[响应服务] Bot %v 响应 %v:%v 发生错误，忽略本次响应...... %v", h.BotID, h.Host, h.Port, err)
		}
	}()

	action := strings.ReplaceAll(path, "/", "")
	TEST("%v", action)
	req := gjson.ParseBytes(api)
	params := req.Get("params")
	DEBUG("[响应服务] Bot %v 接收 %v:%v 到API调用: %v 参数: %v", h.BotID, h.Host, h.Port, req.Get("action").Str, string(api))

	if f, ok := wsApi[action]; ok {
		ret := f(h.BotID, params)
		if req.Get("echo").Int() != 0 {
			ret.Echo = req.Get("echo").Int()
		} else if req.Get("echo").Str != "" {
			ret.Echo = req.Get("echo").Str
		} else {
			ret.Echo, _ = req.Get("echo").Value().(map[string]interface{})
		}
		send, _ := json.Marshal(ret)
		return send
	} else {
		ret := resultFail("no such api")
		if req.Get("echo").Int() != 0 {
			ret.Echo = req.Get("echo").Int()
		} else if req.Get("echo").Str != "" {
			ret.Echo = req.Get("echo").Str
		} else {
			ret.Echo, _ = req.Get("echo").Value().(map[string]interface{})
		}
		send, _ := json.Marshal(ret)
		return send
	}
}

func (h *HTTPYaml) reply(send []byte, reply []byte) {
	defer func() {
		if err := recover(); err != nil {
			ERROR("[HTTP POST][快速回复] Bot %v 响应 %v 发生错误，忽略本次响应...... %v", h.PostUrl, h.Port, err)
		}
	}()
	DEBUG("[HTTP POST][快速回复] Bot %v 接收 %v 到API调用: %v", h.BotID, h.Host, h.Port, string(reply))
	senddata := gjson.ParseBytes(send)
	replydata := gjson.ParseBytes(reply)
	messageType := senddata.Get("message_type").Str
	userID := senddata.Get("user_id").Int()
	groupID := senddata.Get("group_id").Int()
	atSender := replydata.Get("at_sender").Bool()
	messages := replydata.Get("reply")
	msg := ""
	if atSender {
		msg = fmt.Sprintf("[@%v]", userID)
	}
	switch messageType {
	case "group":
		SendMessage(h.BotID, 2, groupID, 0, messages, msg)
	case "private":
		SendMessage(h.BotID, 1, 0, userID, messages, msg)
	default:
		if groupID != 0 {
			SendMessage(h.BotID, 2, groupID, 0, messages, msg)
		} else {
			SendMessage(h.BotID, 1, 0, userID, messages, msg)
		}
	}
}

var wsApi = map[string]func(int64, gjson.Result) Result{
	"send_msg": func(bot int64, p gjson.Result) Result {
		message_type := p.Get("message_type").Str
		group_id := p.Get("group_id").Int()
		user_id := p.Get("user_id").Int()
		messages := p.Get("message")
		switch message_type {
		case "group":
			return SendMessage(bot, 2, group_id, 0, messages, "")
		case "private":
			return SendMessage(bot, 1, 0, user_id, messages, "")
		default:
			if group_id != 0 {
				return SendMessage(bot, 2, group_id, 0, messages, "")
			} else {
				return SendMessage(bot, 1, 0, user_id, messages, "")
			}
		}
	},
	"send_private_msg": func(bot int64, p gjson.Result) Result {
		user_id := p.Get("user_id").Int()
		messages := p.Get("message")
		return SendMessage(bot, 1, 0, user_id, messages, "")
	},
	"send_group_msg": func(bot int64, p gjson.Result) Result {
		group_id := p.Get("group_id").Int()
		messages := p.Get("message")
		return SendMessage(bot, 2, group_id, 0, messages, "")
	},
	"delete_msg": func(bot int64, p gjson.Result) Result {
		return resultFail(map[string]interface{}{"data": "还没写，催更去GitHub提issue"})
	},
	"get_msg": func(bot int64, p gjson.Result) Result {
		return resultFail(map[string]interface{}{"data": "还没写，催更去GitHub提issue"})
	},
	"get_forward_msg": func(bot int64, p gjson.Result) Result {
		return resultFail(map[string]interface{}{"data": "先驱好像不支持"})
	},
	"send_like": func(bot int64, p gjson.Result) Result {
		user_id := p.Get("user_id").Int()
		core.UpVote(bot, user_id)
		return resultOK(map[string]interface{}{"message_id": 0})
	},
	"set_group_kick": func(bot int64, p gjson.Result) Result {
		group_id := p.Get("group_id").Int()
		user_id := p.Get("user_id").Int()
		reject_add_request := p.Get("reject_add_request").Bool()
		core.KickGroupMBR(bot, group_id, user_id, reject_add_request)
		return resultOK(map[string]interface{}{"message_id": 0})
	},
	"set_group_ban": func(bot int64, p gjson.Result) Result {
		group_id := p.Get("group_id").Int()
		user_id := p.Get("user_id").Int()
		duration := p.Get("duration").Int()
		core.ShutUP(bot, group_id, user_id, duration)
		return resultOK(map[string]interface{}{"message_id": 0})
	},
	"set_group_anonymous_ban": func(bot int64, p gjson.Result) Result {
		return resultFail(map[string]interface{}{"data": "还没写，催更去GitHub提issue"})
	},
	"set_group_whole_ban": func(bot int64, p gjson.Result) Result {
		group_id := p.Get("group_id").Int()
		enable := p.Get("enable").Bool()
		if enable {
			core.ShutUP(bot, group_id, 0, 1)
		} else {
			core.ShutUP(bot, group_id, 0, 0)
		}
		return resultOK(map[string]interface{}{"message_id": 0})
	},
	"set_group_admin": func(bot int64, p gjson.Result) Result {
		return resultFail(map[string]interface{}{"data": "先驱好像不支持"})
	},
	"set_group_anonymous": func(bot int64, p gjson.Result) Result {
		group_id := p.Get("group_id").Int()
		enable := p.Get("enable").Bool()
		core.SetAnon(bot, group_id, enable)
		return resultOK(map[string]interface{}{"message_id": 0})
	},
	"set_group_card": func(bot int64, p gjson.Result) Result {
		group_id := p.Get("group_id").Int()
		user_id := p.Get("user_id").Int()
		card := p.Get("card").Str
		core.SetGroupCard(bot, group_id, user_id, card)
		return resultOK(map[string]interface{}{"message_id": 0})
	},
	"set_group_name": func(bot int64, p gjson.Result) Result {
		return resultFail(map[string]interface{}{"data": "先驱好像不支持"})
	},
	"set_group_leave": func(bot int64, p gjson.Result) Result {
		group_id := p.Get("group_id").Int()
		core.QuitGroup(bot, group_id)
		return resultOK(map[string]interface{}{"message_id": 0})
	},
	"set_group_special_title": func(bot int64, p gjson.Result) Result {
		return resultFail(map[string]interface{}{"data": "先驱好像不支持"})
	},
	"set_friend_add_request": func(bot int64, p gjson.Result) Result {
		flag := p.Get("flag").Str
		remark := p.Get("remark").Str
		approve := p.Get("approve").Bool()
		if approve {
			core.HandleFriendEvent(bot, core.Str2Int(flag), 10, remark)
		} else {
			core.HandleFriendEvent(bot, core.Str2Int(flag), 20, remark)
		}
		return resultOK(map[string]interface{}{})
	},
	"set_group_add_request": func(bot int64, p gjson.Result) Result {
		flagdata := strings.Split(p.Get("flag").Str, "|")
		subType := flagdata[0]
		groupID := flagdata[1]
		flag := flagdata[2]
		approve := p.Get("approve").Bool()
		reason := p.Get("reason").Str
		if approve {
			core.HandleGroupEvent(bot, core.Str2Int(subType), 0, core.Str2Int(groupID), core.Str2Int(flag), 10, reason)
		} else {
			core.HandleGroupEvent(bot, core.Str2Int(subType), 0, core.Str2Int(groupID), core.Str2Int(flag), 20, reason)
		}
		return resultOK(map[string]interface{}{})
	},
	"get_login_info": func(bot int64, p gjson.Result) Result {
		nickname := core.GetNick(bot, bot)
		return resultOK(map[string]interface{}{"user_id": bot, "nickname": nickname})
	},
	"get_stranger_info": func(bot int64, p gjson.Result) Result {
		return resultFail(map[string]interface{}{"data": "还没写，催更去GitHub提issue"})
	},
	"get_friend_list": func(bot int64, p gjson.Result) Result {
		list := core.GetFriendList_B(bot)
		if list == "" {
			return resultFail(map[string]interface{}{"data": "ERROR"})
		}
		cqFriendList := []CQFriendInfo{}
		for _, xqFriend := range strings.Split(list, "/n") {
			cqFriendInfo := CQFriendInfo{
				UserID:   core.Str2Int(xqFriend),
				Nickname: "unknow",
				Remark:   "unknow",
			}
			cqFriendList = append(cqFriendList, cqFriendInfo)
		}
		return resultOK(cqFriendList)
	},
	"get_group_info": func(bot int64, p gjson.Result) Result {
		return resultFail(map[string]interface{}{"data": "还没写，催更去GitHub提issue"})
	},
	"get_group_list": func(bot int64, p gjson.Result) Result {
		list := core.GetGroupList_B(bot)
		if list == "" {
			return resultFail(map[string]interface{}{"data": "ERROR"})
		}
		cqGroupList := []CQGroupInfo{}
		for _, xqGroup := range strings.Split(list, "/n") {
			cqGroupInfo := CQGroupInfo{
				GroupID:        core.Str2Int(xqGroup),
				GroupName:      "unknow",
				MemberCount:    0,
				MaxMemberCount: 0,
			}
			cqGroupList = append(cqGroupList, cqGroupInfo)
		}
		return resultOK(cqGroupList)
	},
	"get_group_member_info": func(bot int64, p gjson.Result) Result {
		return resultFail(map[string]interface{}{"data": "还没写，催更去GitHub提issue"})
	},
	"get_group_member_list": func(bot int64, p gjson.Result) Result {
		group_id := p.Get("group_id").Int()
		list := core.GetGroupMemberList_C(bot, group_id)
		if list == "" {
			return resultFail(map[string]interface{}{"data": "ERROR"})
		}
		xqGroupMemberList := XQGroupMemberList{}
		cqGroupMemberList := []CQGroupMember{}
		json.Unmarshal([]byte(list), &xqGroupMemberList)
		for _, xqMember := range xqGroupMemberList.List {
			cqMember := CQGroupMember{
				GroupID:         group_id,
				UserID:          xqMember.QQ,
				Nickname:        "unknow",
				Card:            "unknow",
				Sex:             "unknow",
				Age:             0,
				Area:            "unknow",
				JoinTime:        0,
				LastSentTime:    0,
				Level:           core.Int2Str(xqMember.Lv),
				Role:            "unknow",
				Unfriendly:      false,
				Title:           "unknow",
				TitleExpireTime: 0,
				CardChangeable:  true,
			}
			cqGroupMemberList = append(cqGroupMemberList, cqMember)
		}
		return resultOK(cqGroupMemberList)
	},
	"get_group_honor_info": func(bot int64, p gjson.Result) Result {
		return resultFail(map[string]interface{}{"data": "还没写，催更去GitHub提issue"})
	},
	"get_cookies": func(bot int64, p gjson.Result) Result {
		return resultFail(map[string]interface{}{"data": "还没写，催更去GitHub提issue"})
	},
	"get_csrf_token": func(bot int64, p gjson.Result) Result {
		return resultFail(map[string]interface{}{"data": "还没写，催更去GitHub提issue"})
	},
	"get_record": func(bot int64, p gjson.Result) Result {
		return resultFail(map[string]interface{}{"data": "还没写，催更去GitHub提issue"})
	},
	"get_image": func(bot int64, p gjson.Result) Result {
		return resultFail(map[string]interface{}{"data": "还没写，催更去GitHub提issue"})
	},
	"can_send_image": func(bot int64, p gjson.Result) Result {
		return resultFail(map[string]interface{}{"yes": true})
	},
	"can_send_record": func(bot int64, p gjson.Result) Result {
		return resultFail(map[string]interface{}{"yes": true})
	},
	"get_status": func(bot int64, p gjson.Result) Result {
		online := core.IsOnline(bot, bot)
		return resultFail(map[string]interface{}{"online": online, "good": true})
	},
	"get_version_info": func(bot int64, p gjson.Result) Result {
		app_info := gjson.Parse(AppInfoJson)
		app_version := app_info.Get("pver")
		return resultFail(map[string]interface{}{"app_name": "OneBot-YaYa", "app_version": app_version, "protocol_version": "v11"})
	},
	"set_restart": func(bot int64, p gjson.Result) Result {
		return resultFail(map[string]interface{}{"data": "还没写，催更去GitHub提issue"})
	},
	"clean_cache": func(bot int64, p gjson.Result) Result {
		return resultFail(map[string]interface{}{"data": "还没写，催更去GitHub提issue"})
	},
	// 先驱新增
	"out_put_log": func(bot int64, p gjson.Result) Result {
		text := p.Get("text").Str
		core.OutPutLog(text)
		return resultOK(map[string]interface{}{})
	},
}

func resultOK(data interface{}) Result {
	return Result{
		Status:  "ok",
		Retcode: 200,
		Data:    data,
		Echo:    nil,
	}
}

func resultFail(data interface{}) Result {
	return Result{
		Status:  "failed",
		Retcode: 100,
		Data:    data,
		Echo:    nil,
	}
}

func SendMessage(selfID int64, messageType int64, groupID int64, userID int64, messages gjson.Result, out string) Result {
	messages = cqCode2Array(messages)
	for _, message := range messages.Array() {
		switch message.Get("type").Str {
		case "text":
			out += message.Get("data.*").Str
		case "face":
			out += fmt.Sprintf("[Face%s.gif]", message.Get("data.*").Str)
		case "image":
			image := message.Get("data.file").Str
			image = strings.ReplaceAll(image, `\/`, `/`)
			if strings.Contains(image, "base64://") {
				path := Base64SaveImage(strings.ReplaceAll(image, "base64://", ""))
				out += fmt.Sprintf("[pic=%s]", path)
			} else if strings.Contains(image, "file:///") {
				out += fmt.Sprintf("[pic=%s]", strings.ReplaceAll(image, "file:///", ""))
			} else if strings.Contains(image, "http://") {
				out += fmt.Sprintf("[pic=%s]", image)
			} else if strings.Contains(image, "https://") {
				out += fmt.Sprintf("[pic=%s]", image)
			} else {
				out += fmt.Sprintf("[pic=%s]", "error")
			}
		case "record":
			record := message.Get("data.*").Str
			if strings.Contains(record, "base64://") {
				path := Base64SaveRecord(strings.ReplaceAll(record, "base64://", ""))
				out += fmt.Sprintf("[Voi=%s]", path)
			} else if strings.Contains(record, "file:///") {
				out += fmt.Sprintf("[Voi=%s]", strings.ReplaceAll(record, "file:///", ""))
			} else if strings.Contains(record, "http://") {
				out += fmt.Sprintf("[Voi=%s]", record)
			} else {
				out += fmt.Sprintf("[Voi=%s]", "error")
			}
		case "video":
			video := message.Get("data.*").Str
			if strings.Contains(video, "base64://") {
				path := Base64SaveVideo(strings.ReplaceAll(video, "base64://", ""))
				out += fmt.Sprintf("[Voi=%s]", path)
			} else if strings.Contains(video, "file:///") {
				out += fmt.Sprintf("[Voi=%s]", strings.ReplaceAll(video, "file:///", ""))
			} else if strings.Contains(video, "http://") {
				out += fmt.Sprintf("[Voi=%s]", video)
			} else {
				out += fmt.Sprintf("[Voi=%s]", "error")
			}
		case "at":
			out += fmt.Sprintf("[@%s]", message.Get("data.*").Str)
		case "rps":
			out += "[no such element]"
		case "dice":
			out += "[no such element]"
		case "shake":
			core.ShakeWindow(selfID, userID)
		case "poke":
			out += "[no such element]"
		case "anonymous":
			out += "[no such element]"
		case "share":
			out += "[no such element]"
		case "contact":
			out += "[no such element]"
		case "location":
			out += "[no such element]"
		case "music":
			typ := message.Get("data.type").Str
			if typ == "custom" {
				url := message.Get("data.url").Str
				audio := message.Get("data.audio").Str
				title := message.Get("data.title").Str
				content := message.Get("data.content").Str
				image := message.Get("data.image").Str
				music := SendCustomMusic(url, audio, title, content, image)
				TEST("json格式为%v", music)
				core.SendXML(selfID, 1, messageType, groupID, userID, music, 0)
			} else {
				out += "[no such element]"
			}
		case "reply":
			out += "[no such element]"
		case "forward":
			out += "[no such element]"
		case "node":
			out += "[no such element]"
		case "xml":
			xml := message.Get("data.*").Str
			core.SendJSON(selfID, 1, 2, groupID, userID, xml)
		case "json":
			json := message.Get("data.*").Str
			core.SendJSON(selfID, 1, 2, groupID, userID, json)
		case "emoji":
			out += fmt.Sprintf("[emoji=%s]", message.Get("data.*").Str)
		default:
			WARN("CQ码解析失败，将以原格式返回：%v", message.Str)
			out += message.Str
		}
	}
	messageID := "0"
	if out != "" {
		messageID = core.SendMsgEX_V2(selfID, messageType, groupID, userID, out, 0, false, " ")
	}
	return resultOK(map[string]interface{}{"message_id": messageID})
}

func Base64SaveImage(res string) string {
	data, err := base64.StdEncoding.DecodeString(res)
	if err != nil {
		ERROR("base64编码解码失败")
	}
	name := fmt.Sprintf("%x", md5.Sum(data))
	path := ImagePath + name + ".jpg"
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	defer f.Close()
	if err != nil {
		ERROR("base64编码保存图片失败")
	} else {
		_, err = f.Write(data)
		if err != nil {
			ERROR("base64编码写入图片失败")
		}
	}
	return path
}

func Base64SaveRecord(res string) string {
	data, err := base64.StdEncoding.DecodeString(res)
	if err != nil {
		ERROR("base64编码解码失败")
	}
	name := fmt.Sprintf("%x", md5.Sum(data))
	path := RecordPath + name + ".mp3"
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	defer f.Close()
	if err != nil {
		ERROR("base64编码保存语音失败")
	} else {
		_, err = f.Write(data)
		if err != nil {
			ERROR("base64编码写入语音失败")
		}
	}
	return path
}

func Base64SaveVideo(res string) string {
	data, err := base64.StdEncoding.DecodeString(res)
	if err != nil {
		ERROR("base64编码解码失败")
	}
	name := fmt.Sprintf("%x", md5.Sum(data))
	path := VideoPath + name + ".mp4"
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	defer f.Close()
	if err != nil {
		ERROR("base64编码保存视频失败")
	} else {
		_, err = f.Write(data)
		if err != nil {
			ERROR("base64编码写入视频失败")
		}
	}
	return path
}

func SendCustomMusic(url string, audio string, title string, content string, image string) string {
	music := fmt.Sprintf(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?><msg serviceID="2" templateID="1" action="web" brief="[分享] %s" sourceMsgId="0" url="%s" flag="0" adverSign="0" multiMsgFlag="0"><item layout="2"><audio cover="%s" src="%s"/><title>%s</title><summary>%s</summary></item><source name="音乐" icon="https://i.gtimg.cn/open/app_icon/01/07/98/56/1101079856_100_m.png" url="http://web.p.qq.com/qqmpmobile/aio/app.html?id=1101079856" action="app" a_actionData="com.tencent.qqmusic" i_actionData="tencent1101079856://" appid="1101079856" /></msg>`,
		XmlEscape(title), url, image, audio, XmlEscape(title), XmlEscape(content))
	return string(music)
}

func XmlEscape(c string) string {
	buf := new(bytes.Buffer)
	_ = xml.EscapeText(buf, []byte(c))
	return buf.String()
}
