package qpush

import (
	"encoding/base64"
	"fmt"

	"github.com/gogf/gf/frame/g"
	"github.com/iWinston/qk-library/frame/qtype"
)

func getAuroraToken() string {
	appKey := g.Cfg().GetString("auroraPush.AppKey")
	masterSecret := g.Cfg().GetString("auroraPush.MasterSecret")
	data := []byte(appKey + ":" + masterSecret)
	str := base64.StdEncoding.EncodeToString(data)
	return str
}

func AuroraPushByRIds(title string, registrationIds []string) *string {
	result := &AuroraPushResult{}
	g.Client().
		SetHeader("Authorization", "Basic "+getAuroraToken()).ContentJson().
		PostVar("https://api.jpush.cn/v3/push", g.Map{"platform": "all", "audience": "all", "message": g.Map{
			"msg_content":  title,
			"content_type": "text",
			"title":        "msg",
		}}).Scan(result)
	if result.SendNo != "0" {
		panic(fmt.Errorf("极光推送失败，错误码：%v，标题：%v，用户标识：%v", result.SendNo, title, registrationIds))
	} else {
		return qtype.Str(result.MsgId)
	}

}

type AuroraPushResult struct {
	SendNo string
	MsgId  string `json:"msg_id"`
}
