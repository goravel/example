package models

import "github.com/goravel/framework/database/orm"

type Person struct {
	orm.Model
	Type       int    `json:"type"`
	Education  int    `json:"education"`
	Graduate   int    `json:"graduate"`
	StreetCode string `json:"street_code"`
	Sex        int    `json:"sex"`
}

func (Person *Person) TableName() string {
	return "persons"
}
