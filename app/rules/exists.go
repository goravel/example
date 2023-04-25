package rules

import (
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/facades"
)

/**
 * exists 验证一个值在某个表中的字段中存在，相较于Laravel，支持同时判断多个字段
 * exists verify a value exists in a table field, compared to Laravel, support judging multiple fields at the same time
 * 用法：exists:表名称,字段名称,字段名称,字段名称
 * Usage: exists:table_name,field_name,field_name,field_name
 * 例子：exists:users,phone,email
 * Example: exists:users,phone,email
 */

type Exists struct {
}

// Signature The name of the rule.
func (receiver *Exists) Signature() string {
	return "exists"
}

// Passes Determine if the validation rule passes.
func (receiver *Exists) Passes(_ validation.Data, val any, options ...any) bool {

	tableName := options[0].(string)
	fieldName := options[1].(string)
	requestValue := val.(string)

	// Check if it is empty
	if len(requestValue) == 0 {
		return false
	}

	var count int64
	query := facades.Orm.Query().Table(tableName).Where(fieldName, requestValue)
	if len(options) > 2 {
		for i := 2; i < len(options); i++ {
			query = query.OrWhere(options[i].(string), requestValue)
		}
	}
	err := query.Count(&count)
	if err != nil {
		return false
	}

	return count != 0
}

// Message Get the validation error message.
func (receiver *Exists) Message() string {
	return "记录不存在"
}
