package qfield

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

type Ids []uint

// gorm 自定义结构需要实现 Value Scan 两个方法
// Value 实现方法
func (p Ids) Value() (driver.Value, error) {
	// Scan 方法存在bug，null的时候并不会触发，所以在存的时候设置默认值
	if p == nil {
		p = []uint{}
	}
	return json.Marshal(p)
}

// Scan 实现方法
func (p *Ids) Scan(data interface{}) error {
	json.Unmarshal(data.([]byte), p)
	return nil
}

// 转回原始类型
func (p *Ids) GetSlice() []uint {
	return []uint(*p)
}

// 转回原始类型
func (p *Ids) Str() string {
	strArr := []string{}
	for _, v := range *p {
		strArr = append(strArr, fmt.Sprintf("%d", v))
	}
	return strings.Join(strArr, ",")
}
