// main
package main

import (
	"github.com/breezedup/goserver.v2/core"
	"github.com/breezedup/goserver.v2/core/module"
)

func main() {
	defer core.ClosePackages()
	core.LoadPackages("config.json")

	waiter := module.Start()
	waiter.Wait()
}
