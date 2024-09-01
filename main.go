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

	// Start HTTP server by facades.Route().
	//go func() {
	//	if err := facades.Route().Run(); err != nil {
	//		facades.Log().Errorf("Route run error: %v", err)
	//	}
	//}()
	//
	//// Start GRPC server
	//go func() {
	//	if err := facades.Grpc().Run(); err != nil {
	//		facades.Log().Errorf("Run grpc error: %+v", err)
	//	}
	//}()

	//select {}

	var restaurant models.Restaurant
	query := facades.Orm().Query().(*gorm.QueryImpl)
	//if err := query.Instance().SetupJoinTable(&models.Restaurant{}, "Categories", &models.RestaurantCategory{}); err != nil {
	//	fmt.Printf("SetupJoinTable error: %+v\n", err)
	//}
	if err := query.With("Categories").Find(&restaurant, 1); err != nil {
		fmt.Printf("Find error: %+v\n", err)
	}

	fmt.Printf("Restaurant: %+v\n", restaurant)
}
