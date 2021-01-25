package onebot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/tidwall/gjson"

	"yaya/core"
)

var apiMap ApiMap

type Result struct {
	Status  string      `json:"status"`
	Retcode int64       `json:"retcode"`
	Data    interface{} `json:"data"`
	Echo    interface{} `json:"echo"`
}

// Routers XQApi路由
type Routers struct {
}

// ApiMap XQApi的name与method对应表
type ApiMap struct {
	this     Routers
	name     []string
	function []func(this *Routers, bot *BotYaml, params gjson.Result) Result
}

// Get 二分法获得字符串对应的method
func (apiMap *ApiMap) Get(name string) func(this *Routers, bot *BotYaml, params gjson.Result) Result {
	length := len(apiMap.name)
	low := 0
	high := length - 1
	var fun func(this *Routers, bot *BotYaml, params gjson.Result) Result
	for low <= high {
		mid := (low + high) / 2
		switch {
		default:
			fun = apiMap.function[mid]
			return fun
		case apiMap.name[mid] > name:
			high = mid - 1
		case apiMap.name[mid] < name:
			low = mid + 1
		}
	}
	return nil
}

// Register 反射所有Routers的method并注册
func (apiMap *ApiMap) Register(this *Routers) {
	obj := reflect.TypeOf(this)
	var i int = 0
	for i < obj.NumMethod() {
		apiMap.name = append(apiMap.name, obj.Method(i).Name)
		apiMap.function = append(apiMap.function, obj.Method(i).Func.Interface().(func(this *Routers, bot *BotYaml, params gjson.Result) Result))
		i++
	}
	sort.Sort(apiMap)
}

func (apiMap *ApiMap) Len() int { return len(apiMap.name) }

func (apiMap *ApiMap) Less(i, j int) bool { return apiMap.name[i] < apiMap.name[j] }

func (apiMap *ApiMap) Swap(i, j int) {
	apiMap.name[i], apiMap.name[j] = apiMap.name[j], apiMap.name[i]
	apiMap.function[i], apiMap.function[j] = apiMap.function[j], apiMap.function[i]
}

// CallApi 调用XQApi
func (apiMap *ApiMap) CallApi(action string, bot int64, params gjson.Result) Result {
	name := action2fname(action)
	if apiMap.Get(name) == nil {
		return makeError("no such api")
	}
	botConfig := Conf.getBotConfig(bot)
	return apiMap.Get(name)(&apiMap.this, botConfig, params)
}

// action2func OneBot的action转驼峰命名
func action2fname(action string) string {
	up := true
	name := ""
	for _, r := range action {
		if up {
			name += strings.ToUpper(string(r))
			up = false
			continue
		}
		if string(r) == "_" {
			up = true
		} else {
			name += string(r)
		}
	}
	return name
}

func makeError(err string) Result {
	return Result{
		Status:  "failed",
		Retcode: 100,
		Data:    map[string]interface{}{"error": err},
		Echo:    nil,
	}
}

func makeOk(data interface{}) Result {
	return Result{
		Status:  "ok",
		Retcode: 0,
		Data:    data,
		Echo:    nil,
	}
}

func (this *Routers) SendGroupMsg(bot *BotYaml, params gjson.Result) Result {
	return this.SendMsg(bot, params)
}

func (this *Routers) SendPrivateMsg(bot *BotYaml, params gjson.Result) Result {
	return this.SendMsg(bot, params)
}

func (this *Routers) DeleteMsg(bot *BotYaml, params gjson.Result) Result {
	var id int64 = params.Get("message_id").Int()
	if id == 0 {
		return makeError("无效'message_id'")
	}
	var xe XEvent
	if bot.DB != nil {
		bot.dbSelect(&xe, "id="+core.Int2Str(id))
	}
	if xe.ID == 0 {
		return makeError("查询无此消息")
	}
	core.WithdrawMsgEX(
		xe.SelfID,
		xe.MseeageType,
		xe.GroupID,
		xe.UserID,
		xe.MessageNum,
		xe.MessageID,
		xe.Time,
	)
	return makeOk(nil)
}

func (this *Routers) GetMsg(bot *BotYaml, params gjson.Result) Result {
	var id int64 = params.Get("message_id").Int()
	if id == 0 {
		return makeError("无效'message_id'")
	}
	var xe XEvent
	if bot.DB != nil {
		bot.dbSelect(&xe, "id="+core.Int2Str(id))
	}
	if xe.ID == 0 {
		return makeError("查询无此消息")
	}
	return makeOk(map[string]interface{}{
		"time":         xe.Time,
		"message_type": xq2cqMsgType(xe.MseeageType),
		"message_id":   xe.ID,
		"real_id":      xe.MessageID,
		"sender": Event{
			"user_id":  xe.UserID,
			"nickname": "unknown",
			"sex":      "unknown",
			"age":      0,
			"area":     "",
			"card":     "",
			"level":    "",
			"role":     "unknown",
			"title":    "unknown",
		},
		"message": xe.Message,
	})
}

func (this *Routers) GetForwardMsg(bot *BotYaml, params gjson.Result) Result {
	return makeError("先驱不支持")
}

func (this *Routers) SendLike(bot *BotYaml, params gjson.Result) Result {
	var userID int64 = params.Get("user_id").Int()
	if userID == 0 {
		return makeError("无效'user_id'")
	}
	core.UpVote(
		bot.Bot,
		userID,
	)
	return makeOk(nil)
}

func (this *Routers) SetGroupKick(bot *BotYaml, params gjson.Result) Result {
	var groupID int64 = params.Get("group_id").Int()
	var userID int64 = params.Get("user_id").Int()
	var rejectAddRequest bool = params.Get("reject_add_request").Bool()
	if groupID == 0 {
		return makeError("无效'group_id'")
	}
	if userID == 0 {
		return makeError("无效'user_id'")
	}
	core.KickGroupMBR(
		bot.Bot,
		groupID,
		userID,
		rejectAddRequest,
	)
	return makeOk(nil)
}

func (this *Routers) SetGroupBan(bot *BotYaml, params gjson.Result) Result {
	var groupID int64 = params.Get("group_id").Int()
	var userID int64 = params.Get("user_id").Int()
	var duration int64 = params.Get("duration").Int()
	if groupID == 0 {
		return makeError("无效'group_id'")
	}
	if userID == 0 {
		return makeError("无效'user_id'")
	}
	core.ShutUP(
		bot.Bot,
		groupID,
		userID,
		duration,
	)
	return makeOk(nil)
}

func (this *Routers) SetGroupAnonymousBan(bot *BotYaml, params gjson.Result) Result {
	return makeError("先驱不支持")
}

func (this *Routers) SetGroupWholeBan(bot *BotYaml, params gjson.Result) Result {
	var groupID int64 = params.Get("group_id").Int()
	var enable bool = params.Get("enable").Bool()
	if groupID == 0 {
		return makeError("无效'group_id'")
	}
	if enable {
		core.ShutUP(
			bot.Bot,
			groupID,
			0,
			1,
		)
	} else {
		core.ShutUP(
			bot.Bot,
			groupID,
			0,
			0,
		)
	}
	return makeOk(nil)
}

func (this *Routers) SetGroupAdmin(bot *BotYaml, params gjson.Result) Result {
	return makeError("先驱不支持")
}

func (this *Routers) SetGroupAnonymous(bot *BotYaml, params gjson.Result) Result {
	var groupID int64 = params.Get("group_id").Int()
	var enable bool = params.Get("enable").Bool()
	if groupID == 0 {
		return makeError("无效'group_id'")
	}
	core.SetAnon(
		bot.Bot,
		groupID,
		enable,
	)
	return makeOk(nil)
}

func (this *Routers) SetGroupCard(bot *BotYaml, params gjson.Result) Result {
	var groupID int64 = params.Get("group_id").Int()
	var userID int64 = params.Get("user_id").Int()
	var card string = params.Get("enable").Str
	if groupID == 0 {
		return makeError("无效'group_id'")
	}
	if userID == 0 {
		return makeError("无效'user_id'")
	}
	core.SetGroupCard(
		bot.Bot,
		groupID,
		userID,
		card,
	)
	return makeOk(nil)
}

func (this *Routers) SetGroupName(bot *BotYaml, params gjson.Result) Result {
	return makeError("先驱不支持")
}

func (this *Routers) SetGroupLeave(bot *BotYaml, params gjson.Result) Result {
	var groupID int64 = params.Get("group_id").Int()
	if groupID == 0 {
		return makeError("无效'group_id'")
	}
	core.QuitGroup(
		bot.Bot,
		groupID,
	)
	return makeOk(nil)
}

func (this *Routers) SetGroupSpecialTitle(bot *BotYaml, params gjson.Result) Result {
	return makeError("先驱不支持")
}

func (this *Routers) SetFriendAddRequest(bot *BotYaml, params gjson.Result) Result {
	var flag int64 = params.Get("flag").Int()
	var approve bool = params.Get("approve").Bool()
	var remark string = params.Get("remark").Str
	if flag == 0 {
		return makeError("无效'flag'")
	}
	if approve {
		core.HandleFriendEvent(
			bot.Bot,
			flag,
			10,
			remark,
		)
	} else {
		core.HandleFriendEvent(
			bot.Bot,
			flag,
			20,
			remark,
		)
	}
	return makeOk(nil)
}

func (this *Routers) SetGroupAddRequest(bot *BotYaml, params gjson.Result) Result {
	flag  := params.Get("flag").Str
	if flag == "" {
		return makeError("无效'flag'")
	}
	var approve int64
	if params.Get("approve").Bool() {
		approve = 10
	}else {
		approve = 20
	}
	reason := params.Get("reason").Str

	split := strings.Split(flag, "|")
	core.HandleGroupEvent(bot.Bot,
		213,
		params.Get("user_id").Int(),
		core.Str2Int(split[1]),
		core.Str2Int(split[2]),
		approve,
		reason,
	)
	return makeOk(nil)
}

func (this *Routers) GetLoginInfo(bot *BotYaml, params gjson.Result) Result {
	nickname := strings.Split(core.GetNick(
		bot.Bot,
		bot.Bot,
	), "\n")[0]
	return makeOk(map[string]interface{}{
		"user_id":  bot.Bot,
		"nickname": nickname,
	})
}

func (this *Routers) GetStrangerInfo(bot *BotYaml, params gjson.Result) Result {
	var userID int64 = params.Get("user_id").Int()
	if userID == 0 {
		return makeError("无效'user_id'")
	}
	var nickname string = core.GetNick(
		bot.Bot,
		userID,
	)
	var sex string = xq2cqSex(
		core.GetGender(
			bot.Bot,
			userID,
		),
	)
	var age int64 = core.GetAge(
		bot.Bot,
		userID,
	)
	return makeOk(map[string]interface{}{
		"user_id":  userID,
		"nickname": nickname,
		"sex":      sex,
		"age":      age,
	})
}

func (this *Routers) GetFriendList(bot *BotYaml, params gjson.Result) Result {
	var list string = core.GetFriendList(bot.Bot)
	if list == "" {
		return makeError("获取好友列表失败")
	}
	g := gjson.Parse(list)
	friendList := []map[string]interface{}{}
	for _, o := range g.Get("result.0.mems").Array() {
		info := map[string]interface{}{
			"user_id":  o.Get("uin").Int(),
			"nickname": unicode2chinese(o.Get("name").Str),
			"remark":   "unknown",
		}
		friendList = append(friendList, info)
	}
	return makeOk(friendList)
}

func (this *Routers) GetGroupInfo(bot *BotYaml, params gjson.Result) Result {
	var groupID int64 = params.Get("group_id").Int()
	if groupID == 0 {
		return makeError("无效'group_id'")
	}
	var name string = core.GetGroupName(
		bot.Bot,
		groupID,
	)
	members := strings.Split(core.GetGroupMemberNum(
		bot.Bot,
		groupID,
	), "\n")
	var (
		count int64
		max   int64
	)
	if len(members) != 2 {
		count = -1
		max = -1
	} else {
		count = core.Str2Int(members[0])
		max = core.Str2Int(members[1])
	}
	return makeOk(map[string]interface{}{
		"group_id":         groupID,
		"group_name":       name,
		"member_count":     count,
		"max_member_count": max,
	})
}

func (this *Routers) GetGroupList(bot *BotYaml, params gjson.Result) Result {
	list := core.GetGroupList(bot.Bot)
	if list == "" {
		return makeError("获取群列表失败")
	}
	g := gjson.Parse(list)
	groupList := []map[string]interface{}{}
	for _, o := range g.Get("create").Array() {
		info := map[string]interface{}{
			"group_id":         o.Get("gc").Int(),
			"group_name":       unicode2chinese(o.Get("gn").Str),
			"member_count":     0,
			"max_member_count": 0,
		}
		groupList = append(groupList, info)
	}
	for _, o := range g.Get("manage").Array() {
		info := map[string]interface{}{
			"group_id":         o.Get("gc").Int(),
			"group_name":       unicode2chinese(o.Get("gn").Str),
			"member_count":     0,
			"max_member_count": 0,
		}
		groupList = append(groupList, info)
	}
	for _, o := range g.Get("join").Array() {
		info := map[string]interface{}{
			"group_id":         o.Get("gc").Int(),
			"group_name":       unicode2chinese(o.Get("gn").Str),
			"member_count":     0,
			"max_member_count": 0,
		}
		groupList = append(groupList, info)
	}
	return makeOk(groupList)
}

func (this *Routers) GetGroupMemberInfo(bot *BotYaml, params gjson.Result) Result {
	var groupID int64 = params.Get("group_id").Int()
	var userID int64 = params.Get("user_id").Int()
	if groupID == 0 {
		return makeError("无效'group_id'")
	}
	if userID == 0 {
		return makeError("无效'user_id'")
	}
	return makeOk(map[string]interface{}{
		"group_id":          groupID,
		"user_id":           userID,
		"nickname":          core.GetNick(bot.Bot, userID),
		"card":              core.GetNick(bot.Bot, userID),
		"sex":               []string{"unknown", "male", "female"}[core.GetGender(bot.Bot, userID)],
		"age":               core.GetAge(bot.Bot, userID),
		"area":              "unknown",
		"join_time":         0,
		"last_sent_time":    0,
		"level":             "unknown",
		"role":              "unknown",
		"unfriendly":        false,
		"title":             "unknown",
		"title_expire_time": 0,
		"card_changeable":   true,
	})
}

func (this *Routers) GetGroupMemberList(bot *BotYaml, params gjson.Result) Result {
	var groupID int64 = params.Get("group_id").Int()
	if groupID == 0 {
		return makeError("无效'group_id'")
	}
	list := core.GetGroupMemberList_C(
		bot.Bot,
		groupID,
	)
	if list == "" {
		return makeError("获取群员列表失败")
	}
	g := gjson.Parse(list)
	memberList := []map[string]interface{}{}
	for _, o := range g.Get("list").Array() {
		member := map[string]interface{}{
			"group_id":          groupID,
			"user_id":           o.Get("QQ").Int(),
			"nickname":          "unknown",
			"card":              "unknown",
			"sex":               "unknown",
			"age":               0,
			"area":              "unknown",
			"join_time":         0,
			"last_sent_time":    0,
			"level":             o.Get("lv").Int(),
			"role":              "unknown",
			"unfriendly":        false,
			"title":             "unknown",
			"title_expire_time": 0,
			"card_changeable":   true,
		}
		memberList = append(memberList, member)
	}
	return makeOk(memberList)
}

func (this *Routers) GetGroupHonorInfo(bot *BotYaml, params gjson.Result) Result {
	var groupID int64 = params.Get("group_id").Int()
	var type_ string = params.Get("message_type").Str
	if groupID == 0 {
		return makeError("无效'group_id'")
	}
	cookie := fmt.Sprintf("%s%s", core.GetCookies(bot.Bot), core.GetGroupPsKey(bot.Bot))
	var honorType int64 = 1
	switch type_ {
	case "talkative":
		honorType = 1
	case "performer":
		honorType = 2
	case "legend":
		honorType = 3
	case "strong_newbie":
		honorType = 5
	case "emotion":
		honorType = 6
	}
	data := groupHonor(groupID, honorType, cookie)
	if data != nil {
		data = data[bytes.Index(data, []byte(`window.__INITIAL_STATE__=`))+25:]
		data = data[:bytes.Index(data, []byte("</script>"))]
		ret := GroupHonorInfo{}
		json.Unmarshal(data, &ret)
		return makeOk(ret)
	} else {
		return makeError("error")
	}
}

func (this *Routers) GetCookies(bot *BotYaml, params gjson.Result) Result {
	var domain string = params.Get("domain").Str
	switch domain {
	case "qun.qq.com":
		return makeOk(map[string]interface{}{"cookies": core.GetCookies(bot.Bot) + core.GetGroupPsKey(bot.Bot)})
	case "qzone.qq.com":
		return makeOk(map[string]interface{}{"cookies": core.GetCookies(bot.Bot) + core.GetZonePsKey(bot.Bot)})
	default:
		return makeOk(map[string]interface{}{"cookies": core.GetCookies(bot.Bot)})
	}
}

func (this *Routers) GetCsrfToken(bot *BotYaml, params gjson.Result) Result {
	return makeError("暂未实现")
}

func (this *Routers) GetCredentials(bot *BotYaml, params gjson.Result) Result {
	var domain string = params.Get("domain").Str
	switch domain {
	case "qun.qq.com":
		return makeOk(map[string]interface{}{"cookies": core.GetCookies(bot.Bot) + core.GetGroupPsKey(bot.Bot)})
	case "qzone.qq.com":
		return makeOk(map[string]interface{}{"cookies": core.GetCookies(bot.Bot) + core.GetZonePsKey(bot.Bot)})
	default:
		return makeOk(map[string]interface{}{"cookies": core.GetCookies(bot.Bot)})
	}
}

func (this *Routers) GetRecord(bot *BotYaml, params gjson.Result) Result {
	return makeError("暂未实现")
}

func (this *Routers) GetImage(bot *BotYaml, params gjson.Result) Result {
	return makeError("暂未实现")
}

func (this *Routers) CanSendImage(bot *BotYaml, params gjson.Result) Result {
	return makeOk(map[string]interface{}{"yes": true})
}

func (this *Routers) CanSendRecord(bot *BotYaml, params gjson.Result) Result {
	return makeOk(map[string]interface{}{"yes": true})
}

func (this *Routers) GetStatus(bot *BotYaml, params gjson.Result) Result {
	return makeOk(map[string]interface{}{
		"online": core.IsOnline(bot.Bot, bot.Bot),
		"good":   true,
	})
}

func (this *Routers) GetVersionInfo(bot *BotYaml, params gjson.Result) Result {
	return makeOk(map[string]interface{}{
		"app_name":         "OneBot-YaYa",
		"app_version":      gjson.Parse(AppInfoJson).Get("pver"),
		"protocol_version": "v11",
	})
}

func (this *Routers) SetRestart(bot *BotYaml, params gjson.Result) Result {
	return makeError("暂未实现")
}

func (this *Routers) CleanCache(bot *BotYaml, params gjson.Result) Result {
	return makeError("暂未实现")
}

func (this *Routers) OutPutLog(bot *BotYaml, params gjson.Result) Result {
	var text string = params.Get("text").Str
	core.OutPutLog(text)
	return makeOk(nil)
}

func (this *Routers) SendXml(bot *BotYaml, params gjson.Result) Result {
	var groupID int64 = params.Get("group_id").Int()
	var userID int64 = params.Get("user_id").Int()
	var type_ string = params.Get("message_type").Str
	var data string = params.Get("data").Str
	if groupID == 0 {
		return makeError("无效'group_id'")
	}
	if userID == 0 {
		return makeError("无效'user_id'")
	}
	if groupID == 0 && userID == 0 {
		return makeError("无效'group_id'或'user_id'")
	}
	if type_ == "" {
		if groupID != 0 {
			type_ = "group"
		} else {
			type_ = "private"
		}
	}
	core.SendXML(
		bot.Bot,
		1,
		cq2xqMsgType(type_),
		groupID,
		userID,
		data,
		0,
	)
	return makeOk(map[string]interface{}{})
}

func (this *Routers) SendJson(bot *BotYaml, params gjson.Result) Result {
	var groupID int64 = params.Get("group_id").Int()
	var userID int64 = params.Get("user_id").Int()
	var type_ string = params.Get("message_type").Str
	var data string = params.Get("data").Str
	if groupID == 0 {
		return makeError("无效'group_id'")
	}
	if userID == 0 {
		return makeError("无效'user_id'")
	}
	if groupID == 0 && userID == 0 {
		return makeError("无效'group_id'或'user_id'")
	}
	if type_ == "" {
		if groupID != 0 {
			type_ = "group"
		} else {
			type_ = "private"
		}
	}
	core.SendJSON(
		bot.Bot,
		1,
		cq2xqMsgType(type_),
		groupID,
		userID,
		data,
	)
	return makeOk(map[string]interface{}{})
}
