package db

import (
	"reflect"
)

type LinkingWord int

const (
	AND LinkingWord = iota
	OR
)

func (lw LinkingWord) String() string {
	return []string{"AND", "OR"}[lw]
}

// ObjectToSQLCondition Does not support empty values!
func ObjectToSQLCondition(linkingWord LinkingWord, obj interface{}, permitDefault bool) (condition string, args []interface{}) {
	objValue := reflect.ValueOf(obj)
	objType := reflect.TypeOf(obj)
	condition = ""
	args = make([]interface{}, 0)

	isFirst := true
	for i := 0; i < objValue.NumField(); i++ {
		if !permitDefault && objValue.Field(i).IsZero() {
			continue
		}
		if !isFirst {
			condition += " " + linkingWord.String() + " "
		}
		condition += "`" + objType.Field(i).Name + "` = ?"
		args = append(args, objValue.Field(i).Interface())

		isFirst = false
	}
	return condition, args
}
