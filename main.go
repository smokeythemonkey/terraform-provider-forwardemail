package main

import (
	"context"
	"flag"
	"log"

	"github.com/forwardemail/terraform-provider-forwardemail/forwardemail"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Debug:   debug,
		Address: "search.opentofu.org/provider/forwardemail/forwardemail",
	}

	err := providerserver.Serve(context.Background(), forwardemail.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
