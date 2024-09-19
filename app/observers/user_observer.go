package observers

import (
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/support/color"
)

type UserObserver struct{}

func (u *UserObserver) Retrieved(event orm.Event) error {
	color.Red().Println("event:Retrieved")
	return nil
}

func (u *UserObserver) Creating(event orm.Event) error {
	color.Red().Println("event:Creating")
	return nil
}

func (u *UserObserver) Created(event orm.Event) error {
	color.Red().Println("event:Created")
	return nil
}

func (u *UserObserver) Updating(event orm.Event) error {
	color.Red().Println("event:Updating")
	return nil
}

func (u *UserObserver) Updated(event orm.Event) error {
	color.Red().Println("event:Updated")
	return nil
}

func (u *UserObserver) Saving(event orm.Event) error {
	color.Red().Println("event:Saving")
	return nil
}

func (u *UserObserver) Saved(event orm.Event) error {
	color.Red().Println("event:Saved")
	return nil
}

func (u *UserObserver) Deleting(event orm.Event) error {
	color.Red().Println("event:Deleting")
	return nil
}

func (u *UserObserver) Deleted(event orm.Event) error {
	color.Red().Println("event:Deleted")
	return nil
}

func (u *UserObserver) ForceDeleting(event orm.Event) error {
	color.Red().Println("event:ForceDeleting")
	return nil
}

func (u *UserObserver) ForceDeleted(event orm.Event) error {
	color.Red().Println("event:ForceDeleted")
	return nil
}
