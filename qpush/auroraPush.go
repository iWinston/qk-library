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
	if len(registrationIds) == 0 {
		return nil
	}
	result := &AuroraPushResult{}
	g.Client().
		SetHeader("Authorization", "Basic "+getAuroraToken()).ContentJson().
		PostVar("https://api.jpush.cn/v3/push", g.Map{
			"platform": "all",
			"audience": g.Map{
				"registration_id": registrationIds,
			},
			"notification": g.Map{
				"alert": title,
			},
		}).
		Scan(result)
	if result.SendNo != "0" {
		panic(fmt.Errorf("极光推送失败，错误码：%v，标题：%v，用户标识：%v", result.Error, title, registrationIds))
	} else {
		return qtype.Str(result.MsgId)
	}

}

type AuroraPushResult struct {
	SendNo string `json:"sendno"`
	MsgId  string `json:"msg_id"`
	Error  g.Map  `json:"error"`
}
