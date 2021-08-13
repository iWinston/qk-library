package q

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/go-sql-driver/mysql"
	"github.com/iWinston/qk-library/frame/qservice"
	"github.com/iWinston/qk-library/qutil"
	"github.com/iWinston/qk-library/qxlsx"
	"gorm.io/gorm"

	"github.com/gogf/gf/net/ghttp"
)

func handleError(r *ghttp.Request, err error) {
	qservice.ReqContext.SetError(r.Context(), err)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		JsonExit(r, 404, "数据不存在")
	} else if err, ok := err.(*mysql.MySQLError); ok {
		switch err.Number {
		case 1062:
			m := regexp.MustCompile(`Duplicate entry '(.*)' for.*`).FindStringSubmatch(err.Message)
			JsonExit(r, 400, fmt.Sprintf("%s 已存在", m[1]))
		case 1054:
			JsonExit(r, 400, err.Error())
		}
	}
	JsonExit(r, 1, err.Error())
}

func ResponseWithMeta(r *ghttp.Request, err error, data interface{}, total int64) {
	if err != nil {
		handleError(r, err)
	} else {
		JsonExit(r, 0, "ok", data, total)
	}
}

func ResponseWithData(r *ghttp.Request, err error, data interface{}) {
	if err != nil {
		handleError(r, err)
	} else {
		JsonExit(r, 0, "ok", data)
	}
}

func Response(r *ghttp.Request, err error) {
	if err != nil {
		handleError(r, err)
	} else {
		JsonExit(r, 0, "ok")
	}
}

func dataToLowerCamelMap(data interface{}) interface{} {
	typ := reflect.TypeOf(data)
	if typ != nil && typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ != nil && typ.Kind() == reflect.Slice && typ.Elem().Kind() == reflect.Struct {
		m := qutil.SliceToMap(data)
		lowerCamelMapArr(m)
		return m
	}
	if typ != nil && typ.Kind() == reflect.Struct {
		m := qutil.StructToMap(data)
		lowerCamelMap(m)
		return m
	}
	return data
}

func lowerCamelMap(m map[string]interface{}) {
	for k, v := range m {
		typ := reflect.TypeOf(v)
		if typ != nil && typ.Kind() == reflect.Map {
			lowerCamelMap(v.(map[string]interface{}))
		}
		if typ != nil && typ.Kind() == reflect.Slice {
			lowerCamelMapArr(v.([]interface{}))
		}

		if unicode.IsUpper(rune(k[0])) {
			newKey := strings.ToLower(k[:1]) + k[1:]
			m[newKey] = v
			delete(m, k)
		}
	}
}

func lowerCamelMapArr(ms []interface{}) {
	for _, m := range ms {
		if reflect.TypeOf(m).Kind() == reflect.Map {
			lowerCamelMap(m.(map[string]interface{}))
		}
	}
}

// func scanDataToAddTag(data interface{}, resType reflect.Type) {
// 	resType = qutil.GetDeepType(resType)
// 	for i := 0; i < resType.NumField(); i++ {
// 		item := resType.Field(i)
// 		// 是匿名结构体，递归字段
// 		if item.Anonymous {
// 			scanDataToAddTag(data, item.Type)
// 			continue
// 		}
// 		tag := item.Tag.Get("json")
// 		fmt.Println(tag)
// 		if tag == "" {
// 			item.Tag = `json:"test"`
// 		}
// 	}
// }

// 数据返回通用JSON数据结构
type JsonResponse struct {
	Code    int    `json:"code"`    // 错误码((0:成功, 1:失败, >1:错误码))
	Message string `json:"message"` // 提示信息
}

type JsonResponseWithData struct {
	Code    int         `json:"code"`    // 错误码((0:成功, 1:失败, >1:错误码))
	Message string      `json:"message"` // 提示信息
	Data    interface{} `json:"data"`    // 返回数据(业务接口定义具体数据结构)
}

type ResMeta struct {
	Total    int64 `json:"total"`
	Current  int64 `json:"current"`
	PageSize int64 `json:"pageSize"`
}

// 数据返回通用JSON数据结构(带分页)
type JsonResponseWithMeta struct {
	Meta ResMeta `json:"meta"`
	JsonResponseWithData
}

// 标准返回结果数据结构封装。
func Json(r *ghttp.Request, code int, message string, data ...interface{}) {
	if len(data) == 0 {
		r.Response.WriteJson(JsonResponse{
			Code:    code,
			Message: message,
		})
	} else {
		jsonResponseWithData := JsonResponseWithData{
			Code:    code,
			Message: message,
			Data:    dataToLowerCamelMap(data[0]),
		}
		if len(data) == 1 {
			r.Response.WriteJson(jsonResponseWithData)
		}
		if len(data) > 1 {
			r.Response.WriteJson(JsonResponseWithMeta{
				JsonResponseWithData: jsonResponseWithData,
				Meta: ResMeta{
					Total:    data[1].(int64),
					Current:  r.GetQueryInt64("current"),
					PageSize: r.GetQueryInt64("pageSize"),
				},
			})
		}
	}

}

func HttpStatus(r *ghttp.Request, err int) {
	if err >= 100 && err < 600 {
		r.Response.WriteHeader(err)
	}
}

// 返回JSON数据并退出当前HTTP执行函数。
func JsonExit(r *ghttp.Request, code int, msg string, data ...interface{}) {
	HttpStatus(r, code)
	Json(r, code, msg, data...)
	r.Exit()
}

func WriteXlsxExit(r *ghttp.Request, f *excelize.File) {
	r.Response.Header().Set("Content-Disposition", "attachment; filename=file.xls")
	f.Write(r.Response.Writer)
	r.Exit()
}

func ResponseWithXlsx(r *ghttp.Request, err error, res interface{}) {
	if err != nil {
		Response(r, err)
	}
	f, err := qxlsx.Xlsx.Write(res)
	if err != nil {
		Response(r, err)
	}
	WriteXlsxExit(r, f)
}
