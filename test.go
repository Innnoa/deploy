package main

import "recovery-unit-deploy/service/deploy"

func test() {
	(*deploy.Deploy).GetAllPackages(nil, "C982743")
	// (*deploy.Deploy).DoInstall(nil)
}
