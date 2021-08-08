package q

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
)

func GetIdFormReq(r *ghttp.Request) (id uint) {
	if id = r.GetUint("id"); id == 0 {
		JsonExit(r, 1, "未获得id参数")
	}
	return
}

func AssignParamFormReq(r *ghttp.Request, param interface{}) {
	// 先从入参中获取
	if err := r.Parse(param); err != nil {
		if reflect.TypeOf(err).Elem().Name() == "validationError" {
			JsonExit(r, 400, err.Error())
		} else {
			g.Log("exception").Error(fmt.Sprintf("%+v", err))
			JsonExit(r, 500, err.Error())
		}
	}

	// 解析标签
	dtoType := reflect.TypeOf(param).Elem()
	dtoValue := reflect.ValueOf(param).Elem()
	parseParamByTag(r, dtoType, dtoValue)
}

func parseParamByTag(r *ghttp.Request, dtoType reflect.Type, dtoValue reflect.Value) {

	for i := 0; i < dtoType.NumField(); i++ {
		itemType := dtoType.Field(i)
		itemValue := dtoValue.Field(i)

		// 是匿名结构体，递归字段
		if itemType.Anonymous && itemType.Tag.Get("select") != "_" {
			if itemType.Type.Kind() == reflect.Ptr {
				parseParamByTag(r, itemType.Type.Elem(), itemValue.Elem())
			} else {
				parseParamByTag(r, itemType.Type, itemValue)
			}
			continue
		}
		// 判断和默认值是否相同
		eqDefaultValue := false
		defaultTag, isTagExisted := itemType.Tag.Lookup("default")
		if isTagExisted {
			defaultValue := getDefaultValue(r, itemType.Type, defaultTag)
			eqDefaultValue = defaultValue == itemValue.Elem().Interface()
		}

		// swaggerignore:true
		ignoreTag := itemType.Tag.Get("swaggerignore")
		// 只允许传默认值或者不传
		if ignoreTag == "true" && !itemValue.IsNil() && !eqDefaultValue {
			err := fmt.Errorf("不允许传递%s参数", itemType.Name)
			g.Log("exception").Error(err)
			JsonExit(r, 400, err.Error())
		}

		// 通过ctx标签获取
		ctxTag := itemType.Tag.Get("ctx")
		if ctxTag != "" {
			arr := strings.Split(ctxTag, ".")
			ctxTagName := arr[0]
			ctxVar := r.GetCtxVar(ctxTagName).Interface()
			if ctxVar == nil {
				err := fmt.Errorf("获取不到%s的值", ctxTagName)
				g.Log("exception").Error(err)
				JsonExit(r, 500, err.Error())
			}
			ctxVarRef := reflect.ValueOf(ctxVar)
			if len(arr) == 1 {
				itemValue.Set(ctxVarRef)
			} else {
				ctxFieldName := arr[1]
				// reflect.Value 的零值是 reflect.Invalid类型
				ctxFieldValue := ctxVarRef.Elem().FieldByName(ctxFieldName)
				if ctxFieldValue.Kind() == reflect.Invalid || ctxFieldValue.IsNil() {
					err := fmt.Errorf("获取不到%s的值", ctxTag)
					g.Log("exception").Error(err)
					JsonExit(r, 500, err.Error())
				}
				if len(arr) == 2 {
					itemValue.Set(ctxFieldValue)
				}
				if len(arr) == 3 {
					ctxFieldDeepName := arr[2]
					ctxFieldValue = ctxFieldValue.Elem().FieldByName(ctxFieldDeepName)
					if ctxFieldValue.Kind() == reflect.Invalid || ctxFieldValue.IsNil() {
						err := fmt.Errorf("获取不到%s的值", ctxTag)
						g.Log("exception").Error(err)
						JsonExit(r, 500, err.Error())
					}
					itemValue.Set(ctxFieldValue)
				}

			}
		}
	}
}

func getDefaultValue(r *ghttp.Request, typ reflect.Type, defaultTag string) (defaultValue interface{}) {
	switch typ.Elem().Kind() {
	case reflect.String:
		defaultValue = defaultTag
	case reflect.Bool:
		if value, err := strconv.ParseBool(defaultTag); err != nil {
			g.Log("exception").Error(err)
			JsonExit(r, 500, err.Error())
		} else {
			defaultValue = value
		}
	case reflect.Int:
		if value, err := strconv.ParseInt(defaultTag, 0, 64); err != nil {
			g.Log("exception").Error(err)
			JsonExit(r, 500, err.Error())
		} else {
			defaultValue = int(value)
		}
	case reflect.Uint:
		if value, err := strconv.ParseUint(defaultTag, 0, 64); err != nil {
			g.Log("exception").Error(err)
			JsonExit(r, 500, err.Error())
		} else {
			defaultValue = uint(value)
		}
	case reflect.Float32:
		if value, err := strconv.ParseFloat(defaultTag, 32); err != nil {
			g.Log("exception").Error(err)
			JsonExit(r, 500, err.Error())
		} else {
			defaultValue = float32(value)
		}
	case reflect.Float64:
		if value, err := strconv.ParseFloat(defaultTag, 64); err != nil {
			g.Log("exception").Error(err)
			JsonExit(r, 500, err.Error())
		} else {
			defaultValue = value
		}
	default:
		err := errors.New("不支持的默认值类型" + typ.String())
		g.Log("exception").Error(err)
		JsonExit(r, 500, err.Error())
	}
	return
}
