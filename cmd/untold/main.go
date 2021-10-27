package main

import (
	"context"
	"flag"
	"github.com/damejeras/untold/internal/secret"
	"github.com/damejeras/untold/internal/untold"
	"github.com/damejeras/untold/internal/vault"
	"github.com/google/subcommands"
	"os"
)

func main() {
	subcommands.Register(untold.NewInitCommand(), "")
	subcommands.Register(subcommands.HelpCommand(), "")

	subcommands.Register(vault.NewCreateCommand(), "vault management")
	subcommands.Register(vault.NewRotateCommand(), "vault management")

	subcommands.Register(secret.NewAddCommand(), "secrets")
	subcommands.Register(secret.NewShowCommand(), "secrets")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
