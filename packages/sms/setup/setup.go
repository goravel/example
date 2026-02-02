package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
)

func main() {
	setup := packages.Setup(os.Args)
	moduleImport := setup.Paths().Module().Import()
	serviceProvider := "&sms.ServiceProvider{}"

	setup.Install(
		modify.RegisterProvider(moduleImport, serviceProvider),
	).Uninstall(
		modify.UnregisterProvider(moduleImport, serviceProvider),
	).Execute()
}
