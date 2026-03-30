package seeders

import (
	"goravel/app/facades"
	"goravel/app/models"
)

type DatabaseSeeder struct {
}

// Signature The name and signature of the seeder.
func (s *DatabaseSeeder) Signature() string {
	return "DatabaseSeeder"
}

// Run executes the seeder logic.
func (s *DatabaseSeeder) Run() error {
	return facades.Orm().Query().Create(&models.User{
		Name: "migration",
		Mail: "migration@goravel.dev",
	})
}
