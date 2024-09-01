package main

import (
	"fmt"

	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
	"goravel/bootstrap"
)

func main() {
	// This bootstraps the framework and gets it ready for use.
	bootstrap.Boot()

	var restaurant models.Restaurant
	query := facades.Orm().Query().(*gorm.QueryImpl)
	if err := query.With("Categories").Find(&restaurant, 1); err != nil {
		fmt.Printf("Find error: %+v\n", err)
	}

	fmt.Printf("Restaurant: %+v\n", restaurant)
}
