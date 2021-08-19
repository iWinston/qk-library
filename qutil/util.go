package qutil

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strings"

	"github.com/gogf/gf/frame/g"
)

func IsZeroOfUnderlyingType(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

// Hash 用Hmac加密
func Hash(key string) string {
	salt := g.Cfg().Get("server.Salt").(string)
	h := hmac.New(sha256.New, []byte(salt))
	h.Write([]byte(key))
	return fmt.Sprintf("%02x", h.Sum(nil))
}

func StructToMap(s interface{}) (m g.MapStrAny) {
	j, _ := json.Marshal(&s)
	_ = json.Unmarshal(j, &m)
	return
}

func SliceToMap(s interface{}) (m []interface{}) {
	j, _ := json.Marshal(&s)
	_ = json.Unmarshal(j, &m)
	return
}

func SliceToMaps(s interface{}) (m g.List) {
	j, _ := json.Marshal(&s)
	_ = json.Unmarshal(j, &m)
	return
}

func GetDeepType(typ reflect.Type) reflect.Type {
	resKind := typ.Kind()
	if resKind == reflect.Array || resKind == reflect.Slice || resKind == reflect.Ptr {
		return GetDeepType(typ.Elem())
	} else {
		return typ
	}
}

func RandomStr(len int) string {
	result := []byte{}
	for i := 0; i < len; i++ {
		// result.WriteString(fmt.Sprintf("%d", 65+rand.Intn(25)))
		result = append(result, byte(65+rand.Intn(25)))
	}
	return string(result[:])
}

func JoinNotEmptyStr(strArr []string, sep string) string {
	newArr := []string{}
	for _, str := range strArr {
		if str != "" {
			newArr = append(newArr, str)
		}
	}
	return strings.Join(newArr, sep)
}

func IsInUint(a []uint, item uint) bool {
	hasId := false
	for i := range a {
		if a[i] == item {
			hasId = true
			break
		}
	}
	return hasId
}
