package qxlsx

import (
	"reflect"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gogf/gf/util/gconv"
	"github.com/iWinston/qk-library/qutil"
)

var Xlsx = xlsx{}

type xlsx struct{}

// 创建xlsx文件并写入数据
func (s *xlsx) Write(data interface{}, sheetNameAndOthers ...string) (f *excelize.File, err error) {
	// 默认名
	sheetName := "Sheet1"
	if len(sheetNameAndOthers) > 0 {
		sheetName = sheetNameAndOthers[0]
	}
	maps := qutil.SliceToMaps(data)

	f = excelize.NewFile()
	f.NewSheet(sheetName)

	typ := reflect.TypeOf(data)
	typ = qutil.GetDeepType(typ)

	keyArr := []string{}
	headerArr := []string{}

	for i := 0; i < typ.NumField(); i++ {
		itemType := typ.Field(i)
		tag := itemType.Tag.Get("comment")
		if tag != "" {
			headerArr = append(headerArr, tag)
			keyArr = append(keyArr, itemType.Name)
		}
	}
	f.SetSheetRow(sheetName, "A1", &headerArr)

	for index, v := range maps {
		axis := "A" + gconv.String(index+2)
		dataArr := []interface{}{}
		for _, key := range keyArr {
			dataArr = append(dataArr, v[key])
		}
		f.SetSheetRow(sheetName, axis, &dataArr)
	}

	return f, err
}
