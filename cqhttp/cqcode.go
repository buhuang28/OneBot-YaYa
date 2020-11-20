package cqhttp

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/tidwall/gjson"
)

// cqCode2Array 字符串CQ码转数组
func cqCode2Array(text gjson.Result) gjson.Result {
	if len(text.Get("#.type").Array()) == 0 {
		message := text.Str

		cqcode := regexp.MustCompile(`\[CQ:(.*?)\]`)
		codeList := cqcode.FindAllStringSubmatch(message, -1)
		codeLen := len(codeList)
		messageElem := []string{}
		if codeLen == 0 {
			messageElem = append(messageElem, message)
		} else {
			sMSGe := "start<-" + message + "<-end"
			codeElem := ""
			preElem := ""
			endElem := ""
			for i, c := range codeList {
				codeElem = c[0]
				split := strings.Split(sMSGe, codeElem)
				preElem = split[0]
				endElem = "start<-" + split[1]
				if preElem != "start<-" {
					messageElem = append(messageElem, strings.ReplaceAll(preElem, "start<-", ""))
				}
				messageElem = append(messageElem, codeElem)
				if i+1 == codeLen {
					if endElem != "start<-<-end" {
						messageElem = append(messageElem, strings.ReplaceAll(strings.ReplaceAll(endElem, "start<-", ""), "<-end", ""))
					}
				}
				sMSGe = endElem
			}
		}

		paramsArray := []map[string]interface{}{}
		for _, e := range messageElem {
			if len(cqcode.FindAllStringSubmatch(e, -1)) == 0 {
				paramsArray = append(paramsArray, map[string]interface{}{"type": "text", "data": map[string]interface{}{"text": e}})
			} else {
				codeR1 := regexp.MustCompile(`\[CQ:(.*?),(.*)\]`)
				codeR2 := regexp.MustCompile(`\[CQ:(.*)\]`)
				code := codeR1.FindAllStringSubmatch(e, -1)
				codeType := ""
				codeParm := ""
				if len(code) != 0 {
					codeType = code[0][1]
					codeParm = code[0][2]
				} else {
					code = codeR2.FindAllStringSubmatch(e, -1)
					codeType = code[0][1]
				}

				switch codeType {
				case "face":
					faceR := regexp.MustCompile(`id=(.*)`)
					id := faceR.FindAllStringSubmatch(codeParm, -1)[0][1]
					paramsArray = append(paramsArray, map[string]interface{}{"type": "face", "data": map[string]interface{}{"id": id}})
				case "image":
					imageR := regexp.MustCompile(`file=(.*)`)
					file := imageR.FindAllStringSubmatch(codeParm, -1)[0][1]
					paramsArray = append(paramsArray, map[string]interface{}{"type": "image", "data": map[string]interface{}{"file": file}})
				case "record":
					recordR := regexp.MustCompile(`file=(.*)`)
					file := recordR.FindAllStringSubmatch(codeParm, -1)[0][1]
					paramsArray = append(paramsArray, map[string]interface{}{"type": "record", "data": map[string]interface{}{"file": file}})
				case "video":
					videoR := regexp.MustCompile(`file=(.*)`)
					file := videoR.FindAllStringSubmatch(codeParm, -1)[0][1]
					paramsArray = append(paramsArray, map[string]interface{}{"type": "viedo", "data": map[string]interface{}{"file": file}})
				case "at":
					atR := regexp.MustCompile(`qq=(.*)`)
					qq := atR.FindAllStringSubmatch(codeParm, -1)[0][1]
					paramsArray = append(paramsArray, map[string]interface{}{"type": "at", "data": map[string]interface{}{"qq": qq}})
				case "rps":
					paramsArray = append(paramsArray, map[string]interface{}{"type": "rps", "data": map[string]interface{}{}})
				case "dice":
					paramsArray = append(paramsArray, map[string]interface{}{"type": "dice", "data": map[string]interface{}{}})
				case "shake":
					paramsArray = append(paramsArray, map[string]interface{}{"type": "shake", "data": map[string]interface{}{}})
				case "poke":
					pokeR := regexp.MustCompile(`type=(.*?),id=(.*)`)
					typ := pokeR.FindAllStringSubmatch(codeParm, -1)[0][1]
					id := pokeR.FindAllStringSubmatch(codeParm, -1)[0][2]
					paramsArray = append(paramsArray, map[string]interface{}{"type": "poke", "data": map[string]interface{}{"type": typ, "id": id}})
				case "anonymous":
					paramsArray = append(paramsArray, map[string]interface{}{"type": "anonymous", "data": map[string]interface{}{}})
				case "share":
					shareR1 := regexp.MustCompile(`url=(.*?),title=(.*?),content=(.*?),image=(.*)`)
					shareR2 := regexp.MustCompile(`url=(.*?),title=(.*?),content=(.*)`)
					shareR3 := regexp.MustCompile(`url=(.*?),title=(.*?),image=(.*)`)
					shareR4 := regexp.MustCompile(`url=(.*?),title=(.*)`)
					if len(shareR1.FindAllStringSubmatch(codeParm, -1)) != 0 {
						url := shareR1.FindAllStringSubmatch(codeParm, -1)[0][1]
						title := shareR1.FindAllStringSubmatch(codeParm, -1)[0][2]
						content := shareR1.FindAllStringSubmatch(codeParm, -1)[0][3]
						image := shareR1.FindAllStringSubmatch(codeParm, -1)[0][4]
						paramsArray = append(paramsArray, map[string]interface{}{"type": "share", "data": map[string]interface{}{"url": url, "title": title, "content": content, "image": image}})
					} else if len(shareR2.FindAllStringSubmatch(codeParm, -1)) != 0 {
						url := shareR2.FindAllStringSubmatch(codeParm, -1)[0][1]
						title := shareR2.FindAllStringSubmatch(codeParm, -1)[0][2]
						content := shareR2.FindAllStringSubmatch(codeParm, -1)[0][3]
						paramsArray = append(paramsArray, map[string]interface{}{"type": "share", "data": map[string]interface{}{"url": url, "title": title, "content": content, "image": ""}})
					} else if len(shareR3.FindAllStringSubmatch(codeParm, -1)) != 0 {
						url := shareR3.FindAllStringSubmatch(codeParm, -1)[0][1]
						title := shareR3.FindAllStringSubmatch(codeParm, -1)[0][2]
						image := shareR3.FindAllStringSubmatch(codeParm, -1)[0][3]
						paramsArray = append(paramsArray, map[string]interface{}{"type": "share", "data": map[string]interface{}{"url": url, "title": title, "content": "", "image": image}})
					} else if len(shareR4.FindAllStringSubmatch(codeParm, -1)) != 0 {
						url := shareR4.FindAllStringSubmatch(codeParm, -1)[0][1]
						title := shareR4.FindAllStringSubmatch(codeParm, -1)[0][2]
						paramsArray = append(paramsArray, map[string]interface{}{"type": "share", "data": map[string]interface{}{"url": url, "title": title, "content": "", "image": ""}})
					}
				case "contact":
					contactR := regexp.MustCompile(`type=(.*?),id=(.*)`)
					typ := contactR.FindAllStringSubmatch(codeParm, -1)[0][1]
					id := contactR.FindAllStringSubmatch(codeParm, -1)[0][2]
					paramsArray = append(paramsArray, map[string]interface{}{"type": "contact", "data": map[string]interface{}{"type": typ, "id": id}})
				case "location":
					locationR1 := regexp.MustCompile(`lat=(.*?),lon=(.*?),title=(.*?),content=(.*)`)
					locationR2 := regexp.MustCompile(`lat=(.*?),lon=(.*?),title=(.*)`)
					locationR3 := regexp.MustCompile(`lat=(.*?),lon=(.*?),content=(.*)`)
					locationR4 := regexp.MustCompile(`lat=(.*?),lon=(.*)`)
					if len(locationR1.FindAllStringSubmatch(codeParm, -1)) != 0 {
						lat := locationR1.FindAllStringSubmatch(codeParm, -1)[0][1]
						lon := locationR1.FindAllStringSubmatch(codeParm, -1)[0][2]
						title := locationR1.FindAllStringSubmatch(codeParm, -1)[0][3]
						content := locationR1.FindAllStringSubmatch(codeParm, -1)[0][4]
						paramsArray = append(paramsArray, map[string]interface{}{"type": "location", "data": map[string]interface{}{"lat": lat, "lon": lon, "title": title, "content": content}})
					} else if len(locationR2.FindAllStringSubmatch(codeParm, -1)) != 0 {
						lat := locationR2.FindAllStringSubmatch(codeParm, -1)[0][1]
						lon := locationR2.FindAllStringSubmatch(codeParm, -1)[0][2]
						title := locationR2.FindAllStringSubmatch(codeParm, -1)[0][3]
						paramsArray = append(paramsArray, map[string]interface{}{"type": "location", "data": map[string]interface{}{"lat": lat, "lon": lon, "title": title, "content": ""}})
					} else if len(locationR3.FindAllStringSubmatch(codeParm, -1)) != 0 {
						lat := locationR3.FindAllStringSubmatch(codeParm, -1)[0][1]
						lon := locationR3.FindAllStringSubmatch(codeParm, -1)[0][2]
						content := locationR3.FindAllStringSubmatch(codeParm, -1)[0][3]
						paramsArray = append(paramsArray, map[string]interface{}{"type": "location", "data": map[string]interface{}{"lat": lat, "lon": lon, "title": "", "content": content}})
					} else if len(locationR4.FindAllStringSubmatch(codeParm, -1)) != 0 {
						lat := locationR4.FindAllStringSubmatch(codeParm, -1)[0][1]
						lon := locationR4.FindAllStringSubmatch(codeParm, -1)[0][2]
						paramsArray = append(paramsArray, map[string]interface{}{"type": "location", "data": map[string]interface{}{"lat": lat, "lon": lon, "title": "", "content": ""}})
					}
				case "music":
					musicR1 := regexp.MustCompile(`type=(.*?),url=(.*?),audio=(.*?),title=(.*?),content=(.*?),image=(.*)`)
					musicR2 := regexp.MustCompile(`type=(.*?),id=(.*)`)
					if len(musicR1.FindAllStringSubmatch(codeParm, -1)) != 0 {
						typ := musicR1.FindAllStringSubmatch(codeParm, -1)[0][1]
						url := musicR1.FindAllStringSubmatch(codeParm, -1)[0][2]
						audio := musicR1.FindAllStringSubmatch(codeParm, -1)[0][3]
						title := musicR1.FindAllStringSubmatch(codeParm, -1)[0][4]
						content := musicR1.FindAllStringSubmatch(codeParm, -1)[0][5]
						image := musicR1.FindAllStringSubmatch(codeParm, -1)[0][6]
						paramsArray = append(paramsArray, map[string]interface{}{"type": "music", "data": map[string]interface{}{"type": typ, "url": url, "audio": audio, "title": title, "content": content, "image": image}})
					} else if len(musicR2.FindAllStringSubmatch(codeParm, -1)) != 0 {
						typ := musicR2.FindAllStringSubmatch(codeParm, -1)[0][1]
						id := musicR2.FindAllStringSubmatch(codeParm, -1)[0][2]
						paramsArray = append(paramsArray, map[string]interface{}{"type": "music", "data": map[string]interface{}{"type": typ, "id": id}})
					}
				case "reply":
					paramsArray = append(paramsArray, map[string]interface{}{"type": text, "data": map[string]interface{}{"text": codeParm}})
				case "forward":
					paramsArray = append(paramsArray, map[string]interface{}{"type": text, "data": map[string]interface{}{"text": codeParm}})
				case "node":
					paramsArray = append(paramsArray, map[string]interface{}{"type": text, "data": map[string]interface{}{"text": codeParm}})
				case "xml":
					xmlR := regexp.MustCompile(`data=(.*)`)
					data := xmlR.FindAllStringSubmatch(codeParm, -1)[0][1]
					paramsArray = append(paramsArray, map[string]interface{}{"type": "xml", "data": map[string]interface{}{"data": data}})
				case "json":
					jsonR := regexp.MustCompile(`data=(.*)`)
					data := jsonR.FindAllStringSubmatch(codeParm, -1)[0][1]
					paramsArray = append(paramsArray, map[string]interface{}{"type": "json", "data": map[string]interface{}{"data": data}})
				case "emoji":
					faceR := regexp.MustCompile(`id=(.*)`)
					id := faceR.FindAllStringSubmatch(codeParm, -1)[0][1]
					paramsArray = append(paramsArray, map[string]interface{}{"type": "emoji", "data": map[string]interface{}{"id": id}})
				default:
					paramsArray = append(paramsArray, map[string]interface{}{"type": "error", "data": map[string]interface{}{"error": e}})
				}
			}
		}
		data, _ := json.Marshal(paramsArray)
		return gjson.Parse(string(data))
	}
	return text
}

// xq2cqCode 普通XQ码转CQ码
func xq2cqCode(message string) string {
	// 转艾特
	message = strings.ReplaceAll(message, "[@", "[CQ:at,qq=")
	// 转emoji
	message = strings.ReplaceAll(message, "[emoji", "[CQ:emoji,id=")

	// 转face
	face := regexp.MustCompile(`\[Face(.*?)\.gif]`)
	for _, f := range face.FindAllStringSubmatch(message, -1) {
		oldpic := f[0]
		newpic := fmt.Sprintf("[CQ:face,id=%s]", f[1])
		message = strings.ReplaceAll(message, oldpic, newpic)
	}

	// 转图片
	pic := regexp.MustCompile(`\[pic={(.*?)-(.*?)-(.*?)-(.*?)-(.*?)}(\..*?),(.*?)\]`)
	for _, p := range pic.FindAllStringSubmatch(message, -1) {
		oldpic := p[0]
		newpic := fmt.Sprintf("[CQ:image,file=%s%s%s%s%s%s]", p[1], p[2], p[3], p[4], p[5], p[6])
		message = strings.ReplaceAll(message, oldpic, newpic)
	}

	// 转语音
	voi := regexp.MustCompile(`\[Voi={(.*?)-(.*?)-(.*?)-(.*?)-(.*?)}(\..*?),(.*?)\]`)
	for _, v := range voi.FindAllStringSubmatch(message, -1) {
		oldpic := v[0]
		newpic := fmt.Sprintf("[CQ:record,file=%s%s%s%s%s]", v[1], v[2], v[3], v[4], v[5])
		message = strings.ReplaceAll(message, oldpic, newpic)
	}

	return message
}

// cq2xqCode 普通CQ码转XQ码
func cq2xqCode(message string) string {
	// 转艾特
	message = strings.ReplaceAll(message, "[CQ:at,qq=", "[@")
	// 转emoji
	message = strings.ReplaceAll(message, "[CQ:emoji,id=", "[emoji")

	// 转face
	face := regexp.MustCompile(`\[CQ:face,id=(.*?)\]`)
	for _, f := range face.FindAllStringSubmatch(message, -1) {
		oldpic := f[0]
		newpic := fmt.Sprintf("[Face%s.gif]", f[1])
		message = strings.ReplaceAll(message, oldpic, newpic)
	}

	// 转图片
	pic := regexp.MustCompile(`\[CQ:image,file=(.*?)\]`)
	for _, p := range pic.FindAllStringSubmatch(message, -1) {
		oldpic := p[0]
		newpic := fmt.Sprintf("[pic=%s]", p[1])
		message = strings.ReplaceAll(message, oldpic, newpic)
	}

	// 转语音
	voi := regexp.MustCompile(`\[CQ:record,file=(.*?)\]`)
	for _, v := range voi.FindAllStringSubmatch(message, -1) {
		oldpic := v[0]
		newpic := fmt.Sprintf("[Voi=%s]", v[1])
		message = strings.ReplaceAll(message, oldpic, newpic)
	}
	return message
}
