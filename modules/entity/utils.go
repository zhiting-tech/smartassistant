package entity

import (
	errors2 "errors"
	"fmt"
	"gorm.io/gorm"
	"reflect"
	"regexp"
	"strings"
)


const meiHuiZhiJuBrandName = "meihuizhiju"

// CheckIllegalRepeatDate 判断生效时间格式是否合法
func CheckIllegalRepeatDate(repeatDate string) bool {
	pattern := `^[1-7]{1,7}$`
	reg := regexp.MustCompile(pattern)
	if !reg.MatchString(repeatDate) {
		return false
	}

	// "1122" 视为不合法字符串
	if !checkRepeatStr(repeatDate) {
		return false
	}
	return true
}

// checkRepeatStr 判断重复字符串
func checkRepeatStr(str string) bool {
	if str == "" {
		return false
	}

	var strMap = map[rune]bool{
		'1': false,
		'2': false,
		'3': false,
		'4': false,
		'5': false,
		'6': false,
		'7': false,
	}

	// 根据ASCII码判断字符是否重复
	for _, v := range str {
		if strMap[v] == true {
			return false
		}

		if val, ok := strMap[v]; ok && !val {
			strMap[v] = true
		}

	}
	return true
}

func IsAreaType(areaType AreaType) bool {
	if areaType == AreaOfCompany || areaType == AreaOfHome {
		return true
	}
	return false
}

func IsHome(areaType AreaType) bool {
	return areaType == AreaOfHome
}

func IsCompany(areaType AreaType) bool {
	return areaType == AreaOfCompany
}


// CopyTable 复制一个数据库中的表到另一个数据库中的表上
func CopyTable(src *gorm.DB, dst *gorm.DB, table interface{}, delete bool) (err error) {

	var (
		tableSlice reflect.Value
	)

	// 构建table的指针类型
	tableType := reflect.ValueOf(table).Type()
	modelType := tableType
	if tableType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	modelValue := reflect.New(modelType)
	modelInterface := modelValue.Interface()
	if modelInterface == nil {
		return fmt.Errorf("Create table pointer error")
	}

	// 生成table类型的Slice
	tableSlice = reflect.MakeSlice(reflect.SliceOf(modelType), 0, 0)

	// 生成Slice的指针
	addr := reflect.New(tableSlice.Type())
	addr.Elem().Set(tableSlice)
	valuesAddr := addr.Elem().Addr().Interface()

	// 查找src数据库中该表所有的数据
	err = src.Model(modelInterface).Find(valuesAddr).Error
	if err != nil {
		return err
	}
	if delete {
		// 删除dst数据库该表所有的数据
		err = dst.Model(modelInterface).Unscoped().Where("true").Delete(nil).Error
		if err != nil {
			if !errors2.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}
	}
	// 判断Slice是否大于0,大于0则创建
	if addr.Elem().Len() > 0 {
		values := addr.Elem().Interface()
		err = dst.Model(modelInterface).Create(values).Error
		if err != nil {
			if !errors2.Is(err, gorm.ErrEmptySlice) {
				return err
			}
		}
	}

	return nil
}

func IsMeiHuiZhiJuBrand(pluginID string) bool {
	pluginName := strings.Split(pluginID, ".")
	if len(pluginName) > 0 {
		return pluginName[0] == meiHuiZhiJuBrandName
	}
	return false
}
