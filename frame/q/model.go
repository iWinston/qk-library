package q

import (
	"errors"
	"fmt"
	"reflect"

	"strings"

	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/util/gconv"
	"github.com/iWinston/qk-library/frame/qmodel"
	"github.com/iWinston/qk-library/qutil"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type meta struct {
	Joins    []string
	Preloads []string
	Selects  []string
	Orders   []string
	Wheres   []string
}

func GenSqlByRes(tx *gorm.DB, res interface{}) *gorm.DB {
	var resMeta meta
	resType := reflect.TypeOf(res).Elem() //通过反射获取type定义
	scanRes(&resMeta, resType, tx)
	for _, order := range resMeta.Orders {
		tx.Order(order)
	}

	for _, preload := range resMeta.Preloads {
		tx.Preload(preload)
	}
	for _, join := range resMeta.Joins {
		JoinByRelation(tx, join)
	}
	tx.Statement.Selects = resMeta.Selects

	return tx
}

// 默认是相等的条件，_代表不筛选此字段
func GenSqlByParam(tx *gorm.DB, param interface{}) *gorm.DB {
	var (
		paramType  = reflect.TypeOf(param).Elem() //通过反射获取type定义
		paramValue = reflect.ValueOf(param).Elem()
	)
	var paramMeta meta
	scanParam(&paramMeta, paramType, paramValue, tx)

	for _, order := range paramMeta.Orders {
		tx.Order(order)
	}
	for _, join := range paramMeta.Joins {
		JoinByRelation(tx, join)
	}
	return tx
}

func scanRes(resMeta *meta, resType reflect.Type, tx *gorm.DB) {
	// 传的是数组则再获取数组里的struct
	if resType.Kind() == reflect.Slice {
		resType = resType.Elem()
	}

	for i := 0; i < resType.NumField(); i++ {
		item := resType.Field(i)
		// 是匿名结构体，递归字段
		if item.Anonymous && item.Tag.Get("select") != "_" {
			if item.Type.Kind() == reflect.Ptr {
				scanRes(resMeta, item.Type.Elem(), tx)
			} else {
				scanRes(resMeta, item.Type, tx)
			}
			continue
		}

		isPreloadTagExisted := setPreloadMeta(resMeta, item)
		setOrderMeta(resMeta, item, tx)

		if !strings.Contains(item.Type.String(), "model.") && !isPreloadTagExisted {
			setSelectMeta(resMeta, item, tx)
		}
	}
}

func scanParam(meta *meta, paramType reflect.Type, paramValue reflect.Value, tx *gorm.DB) {
	// 传的是数组则再获取数组里的struct
	if paramType.Kind() == reflect.Slice {
		paramType = paramType.Elem()
	}

	for i := 0; i < paramType.NumField(); i++ {
		itemType := paramType.Field(i)
		itemValue := paramValue.Field(i)
		// 是匿名结构体，递归字段
		if itemType.Anonymous && itemType.Tag.Get("select") != "_" {
			if itemType.Type.Kind() == reflect.Ptr {
				scanParam(meta, itemType.Type.Elem(), itemValue.Elem(), tx)
			} else {
				scanParam(meta, itemType.Type, itemValue, tx)
			}
			continue
		}

		setOrderMeta(meta, itemType, tx)
		setWhereMeta(meta, itemType, itemValue, tx)
	}
}

func setWhereMeta(meta *meta, itemType reflect.StructField, itemValue reflect.Value, tx *gorm.DB) {
	// 此处是默认所有的Dto都是指针类型,结构体或者数组
	if qutil.IsZeroOfUnderlyingType(itemValue.Interface()) {
		return
	}
	// if itemType.Type.Kind() == reflect.Ptr {
	// 	if itemValue.Elem().Interface() == "" {
	// 		return
	// 	}
	// }

	var tag, isTagExisted = itemType.Tag.Lookup("where")
	var (
		operator             = "="
		columnName, relation = getColumnNameAndRelation(tx, itemType.Name, "")
	)
	if isTagExisted {
		if tag != "" {
			tagArr := strings.Split(tag, ";")
			operator = tagArr[0]
			if len(tagArr) == 2 {
				columnName, relation = getColumnNameAndRelation(tx, itemType.Name, tagArr[1])
				if relation != "" {
					meta.Joins = append(meta.Joins, relation)
				}
			}
		}
		genCondition(tx, columnName, operator, itemValue.Interface())
	}
}

func setOrderMeta(resMeta *meta, item reflect.StructField, tx *gorm.DB) {
	var orderTag, isOrderTagExisted = item.Tag.Lookup("order")
	if isOrderTagExisted {
		if orderTag == "" {
			orderTag = item.Name + " " + "desc"
		} else {
			orderTagArr := strings.Split(orderTag, ";")
			if len(orderTagArr) == 1 {
				columnName, _ := getColumnNameAndRelation(tx, item.Name, "")
				orderTag = columnName + " " + orderTag
			}
			if len(orderTagArr) == 2 {
				columnName, relation := getColumnNameAndRelation(tx, item.Name, orderTagArr[1])
				orderTag = columnName + " " + orderTagArr[0]
				if relation != "" {
					resMeta.Joins = append(resMeta.Joins, relation)
				}
			}
		}
		resMeta.Orders = append(resMeta.Orders, orderTag)
	}
}

func setPreloadMeta(resMeta *meta, item reflect.StructField) bool {
	var preloadTag, isPreloadTagExisted = item.Tag.Lookup("preload")
	if isPreloadTagExisted {
		if preloadTag == "" {
			preloadTag = item.Name
		}
		resMeta.Preloads = append(resMeta.Preloads, preloadTag)
	}
	return isPreloadTagExisted
}

func setSelectMeta(resMeta *meta, item reflect.StructField, tx *gorm.DB) {
	selectTag := item.Tag.Get("select")
	if selectTag != "_" {
		columnName, relation := getColumnNameAndRelation(tx, item.Name, selectTag)
		selectItem := columnName + " as `" + item.Name + "`"
		resMeta.Selects = append(resMeta.Selects, selectItem)
		if relation != "" {
			resMeta.Joins = append(resMeta.Joins, relation)
		}
	}

}

func genCondition(sql *gorm.DB, name, operator string, itemValue interface{}) {
	switch operator {
	case "_":
	case "!=":
		sql.Where(name+" != ?", itemValue)
	case "in":
		sql.Where(name+" in ?", itemValue)
	case "=":
		sql.Where(name+" = ?", itemValue)
	case ">":
		sql.Where(name+" > ?", itemValue)
	case "<":
		sql.Where(name+" < ?", itemValue)
	case "like":
		sql.Where(name+" LIKE ?", "%"+*itemValue.(*string)+"%")
	}
}

func getColumnNameAndRelation(tx *gorm.DB, fieldName string, tag string) (columnName string, relation string) {
	var (
		tableName = getTableName(tx)

		arr = strings.Split(tag, ".")
		len = len(arr)
	)
	// 为空代表没有tag，默认值是结构体的字段名
	if tag == "" {
		columnName = "`" + tableName + "`.`" + tx.NamingStrategy.ColumnName("", fieldName) + "`"
		return
	}
	// 长度1代表是自定义字段名
	if len == 1 {
		columnName = "`" + tableName + "`.`" + tx.NamingStrategy.ColumnName("", arr[0]) + "`"
	}
	// 长度2代表是连表字段
	if len == 2 {
		columnName = "`" + arr[0] + "`.`" + tx.NamingStrategy.ColumnName(arr[0], arr[1]) + "`"
		relation = arr[0]
	}
	return
}

func JoinByRelation(tx *gorm.DB, relation string) {
	isContains := false
	for _, join := range tx.Statement.Joins {
		if join.Name == relation || strings.Contains(join.Name, relation+"` on") || strings.Contains(join.Name, relation+"` ON") || strings.Contains(join.Name, relation+" on") || strings.Contains(join.Name, relation+" ON") {
			isContains = true
			break
		}
	}
	if isContains {
		return
	}

	joinName := relation
	tableName := getTableName(tx)
	modelType := reflect.TypeOf(tx.Statement.Model).Elem()
	relationField, ok := modelType.FieldByName(relation)
	if !ok {
		panic(fmt.Sprintf("%s 中没有 %s 关联字段", tableName, relation))
	}
	if relationField.Type.Kind() == reflect.Slice || relationField.Type.Elem().Kind() == reflect.Slice {
		relationType := qutil.GetDeepType(relationField.Type).String()
		arr := strings.Split(relationType, ".")
		relationType = arr[len(arr)-1]
		relationTableName := tx.NamingStrategy.TableName(relationType)

		gormTag := relationField.Tag.Get("gorm")
		tagMap := schema.ParseTagSetting(gormTag, ";")
		many2manyTableName := tx.NamingStrategy.TableName(tagMap["MANY2MANY"])

		if many2manyTableName == "" {
			// columnName := tx.NamingStrategy.ColumnName("", relation)
			joinName = fmt.Sprintf("LEFT JOIN `%s` `%s` ON `%s`.`%s_id` = `%s`.`id`", relationTableName, relation, relation, tableName, tableName)
			// joinName = fmt.Sprintf("LEFT JOIN `%s` `%s` ON `%s`.`%s_id` = `%s`.`id`", relationTableName, relation, tableName, columnName, relation)
		} else {
			joinName = fmt.Sprintf("LEFT JOIN `%s` `%s` ON `%s`.`%s_id`=`%s`.`id` LEFT JOIN `%s` `%s` ON `%s`.`id`=`%s`.`%s_id`",
				many2manyTableName, many2manyTableName, many2manyTableName, tableName, tableName, relationTableName, relation, relation, many2manyTableName, relationTableName)
		}
		tx.Distinct().Joins(joinName)
	} else {
		tx.Joins(joinName)
	}
}

func getTableName(tx *gorm.DB) (tableName string) {
	tableName = tx.Statement.Table
	if tableName == "" {
		tableName = tx.NamingStrategy.TableName(reflect.TypeOf(tx.Statement.Model).Elem().Name())
	}
	if tabler, ok := tx.Statement.Model.(schema.Tabler); ok {
		tableName = tabler.TableName()
	}
	return
}
func Paginate(req interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		var page Page
		gconv.Struct(req, &page)
		if page.Current > 0 && page.PageSize > 0 {
			offset := (page.Current - 1) * page.PageSize
			return db.Offset(offset).Limit(page.PageSize)
		}
		return db
	}
}

// MustFirstExit 如果查找失败，退出请求
func MustFirstExit(r *ghttp.Request, tx *gorm.DB, errorTip string, dest interface{}, conds ...interface{}) *gorm.DB {
	result := tx.First(dest, conds...)
	if err := result.Error; err != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			Response(r, gerror.New(errorTip))
		}
	}
	return result
}

// MustFirst 如果查找失败，panic
func MustFirst(tx *gorm.DB, tipError error, dest interface{}, conds ...interface{}) *gorm.DB {
	result := tx.First(dest, conds...)
	if err := result.Error; err != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			panic(tipError.Error())
		}
	}
	return result
}

// MustCreateExit 如果cond查找结果为0，则保存dest，否则退出请求
func MustCreateExit(ctx *qmodel.ReqContext, errorTip string, dest interface{}, cond interface{}) *gorm.DB {
	var oldNum int64 = 0
	ctx.TX.Model(dest).Where(cond).Count(&oldNum)
	if oldNum != 0 {
		Response(ctx.Request, gerror.New(errorTip))
	}
	return ctx.OrgTX.Create(dest)
}

// MustCreate 如果cond查找结果为0，则保存dest，否则panic
func MustCreate(tx *gorm.DB, err error, dest interface{}, cond interface{}) *gorm.DB {
	var oldNum int64 = 0
	tx.Model(dest).Where(cond).Count(&oldNum)
	if oldNum != 0 {
		panic(err.Error())
	}
	return tx.Create(dest)
}
