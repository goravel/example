package models

import (
	"fmt"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/orm"
)

type User struct {
	orm.Model
	Name   string
	Avatar string
	orm.SoftDeletes
}

func (u *User) DispatchesEvents() map[contractsorm.EventType]func(contractsorm.Event) error {
	return map[contractsorm.EventType]func(contractsorm.Event) error{
		contractsorm.EventCreating: func(event contractsorm.Event) error {
			fmt.Println("creating")

			return nil
		},
		contractsorm.EventCreated: func(event contractsorm.Event) error {
			fmt.Println("created")

			return nil
		},
		contractsorm.EventSaving: func(event contractsorm.Event) error {
			fmt.Println("saving")

			return nil
		},
		contractsorm.EventSaved: func(event contractsorm.Event) error {
			fmt.Println("saved")

			return nil
		},
	}
}
