package services

import (
	"context"
	"time"

	contractorm "github.com/goravel/framework/contracts/database/orm"
	contractevent "github.com/goravel/framework/contracts/event"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/mail"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/filesystem"

	"goravel/app/events"
	"goravel/app/facades"
	"goravel/app/jobs"
	"goravel/app/models"
)

func AppCurrentLocale() string {
	return facades.App().CurrentLocale(context.Background())
}

func ArtisanCall() error {
	return facades.Artisan().Call("list")
}

func Auth(ctx contractshttp.Context) error {
	return facades.Auth(ctx).Logout()
}

func Cache() string {
	if err := facades.Cache().Put("name", "goravel", 1*time.Minute); err != nil {
		facades.Log().Error("cache.put.error", err)
	}

	return facades.Cache().Get("name", "test").(string)
}

func Config() string {
	return facades.Config().GetString("app.name", "test")
}

func Crypt(str string) (string, error) {
	res, err := facades.Crypt().EncryptString(str)
	if err != nil {
		return "", err
	}

	return facades.Crypt().DecryptString(res)
}

func Event() error {
	return facades.Event().Job(&events.OrderShipped{}, []contractevent.Arg{
		{Type: "string", Value: "test"},
		{Type: "int", Value: 1234},
	}).Dispatch()
}

func Gate() bool {
	return facades.Gate().Allows("update-post", map[string]any{
		"post": "test",
	})
}

func Grpc() error {
	_, err := facades.Grpc().Client(context.Background(), "user")

	return err
}

func Hash() (string, error) {
	return facades.Hash().Make("Goravel")
}

func Lang(ctx context.Context) string {
	return facades.Lang(ctx).Get("name")
}

func Log() {
	facades.Log().Debug("test")
}

func Mail() error {
	return facades.Mail().From(mail.Address{Address: "example@example.com", Name: "example"}).
		To([]string{"example@example.com"}).
		Subject("Subject").
		Content(mail.Content{Html: "<h1>Hello Goravel</h1>"}).
		Send()
}

func Orm() error {
	if err := facades.Orm().Query().Create(&models.User{Name: "Goravel"}); err != nil {
		return err
	}

	var user models.User

	return facades.Orm().Query().Where("id = ?", 1).Find(&user)
}

func OrmTransaction() error {
	return facades.Orm().Transaction(func(tx contractorm.Query) error {
		var test models.User
		if err := tx.Create(&test); err != nil {
			return err
		}

		var test1 models.User

		return tx.Where("id = ?", test.ID).Find(&test1)
	})
}

func Queue() error {
	return facades.Queue().Job(&jobs.Test{}, []queue.Arg{}).Dispatch()
}

func Storage() (string, error) {
	file, _ := filesystem.NewFile("1.txt")

	return facades.Storage().WithContext(context.Background()).PutFile("file", file)
}

func Validation() string {
	validator, _ := facades.Validation().Make(
		context.Background(),
		map[string]any{"a": "b"},
		map[string]any{"a": "required"},
	)
	errors := validator.Errors()

	return errors.One("a")
}

func View() bool {
	return facades.View().Exists("welcome.tmpl")
}
