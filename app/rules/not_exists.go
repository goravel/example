package rules

import (
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/facades"
)

/**
 * not_exists 验证一个值在某个表中的字段中不存在，相较于Laravel，支持同时判断多个字段
 * not_exists verify a value does not exist in a table field, compared to Laravel, support judging multiple fields at the same time
 * 用法：not_exists:表名称,字段名称,字段名称,字段名称
 * Usage: not_exists:table_name,field_name,field_name,field_name
 * 例子：not_exists:users,phone,email
 * Example: not_exists:users,phone,email
 */

type NotExists struct {
}

// Signature The name of the rule.
func (receiver *NotExists) Signature() string {
	return "not_exists"
}

// Passes Determine if the validation rule passes.
func (receiver *NotExists) Passes(_ validation.Data, val any, options ...any) bool {

	tableName := options[0].(string)
	fieldName := options[1].(string)
	requestValue := val.(string)

	if len(requestValue) == 0 {
		return false
	}

	var count int64
	query := facades.Orm().Query().Table(tableName).Where(fieldName, requestValue)
	if len(options) > 2 {
		for i := 2; i < len(options); i++ {
			query = query.OrWhere(options[i].(string), requestValue)
		}
	}
	err := query.Count(&count)
	if err != nil {
		return false
	}

	return count == 0
}

// Message Get the validation error message.
func (receiver *NotExists) Message() string {
	return "record already exists"
}
