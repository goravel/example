package bootstrap

import "goravel/app/facades"

func CommandsFilter() []string {
	if facades.Config().GetString("app.env") == "production" {
		return []string{
			"up",
			"down",
			"make:*",
			"vendor:publish",
		}
	}

	return nil
}
