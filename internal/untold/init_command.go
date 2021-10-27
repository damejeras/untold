package untold

import (
	"context"
	"crypto/rand"
	"embed"
	"flag"
	"github.com/damejeras/untold"
	"github.com/damejeras/untold/internal/cli"
	"github.com/google/subcommands"
	"golang.org/x/crypto/nacl/box"
	"os"
	"path/filepath"
)

//go:embed templates/*
var templates embed.FS

type initCmd struct {
	environment string
}

func NewInitCommand() subcommands.Command {
	return &initCmd{
		environment: untold.DefaultEnvironment,
	}
}

func (i *initCmd) Name() string {
	return "init"
}

func (i *initCmd) Synopsis() string {
	return "initialize secrets vault."
}

func (i *initCmd) Usage() string {
	return `untold init [-env={environment}] [directory_name]:
  Initialize secrets vault.
`
}

func (i *initCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&i.environment, "env", i.environment, "set environment")
}

func (i *initCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	directory := f.Arg(0)
	if directory == "" {
		directory = untold.DefaultPathPrefix
		cli.Warnf("Directory name not provided, using default - \"%s\"", directory)
	}

	environment := i.environment
	if environment == "" || environment == untold.DefaultEnvironment {
		environment = untold.DefaultEnvironment
		cli.Warnf("No environment provided, using default - \"%s\"", environment)
	}

	if _, err := os.Stat(directory); !os.IsNotExist(err) {
		cli.Errorf("directory \"%s\" already exists", directory)

		return subcommands.ExitUsageError
	}

	if err := os.Mkdir(directory, 0755); err != nil {
		cli.Wrapf(err, "create directory \"%s\"", directory)

		return subcommands.ExitFailure
	}

	gitignoreContent, err := templates.ReadFile("templates/.gitignore")
	if err != nil {
		cli.Wrapf(err, "read .gitignore template")

		return subcommands.ExitFailure
	}

	if err := os.WriteFile(filepath.Join(directory, ".gitignore"), gitignoreContent, 0644); err != nil {


		cli.Wrapf(err, "create .gitignore file")

		return subcommands.ExitFailure
	}

	readmeContent, err := templates.ReadFile("templates/README.md")
	if err != nil {
		cli.Wrapf(err, "read README.md template")

		return subcommands.ExitFailure
	}

	if err := os.WriteFile(filepath.Join(directory, "README.md"), readmeContent, 0644); err != nil {
		cli.Wrapf(err, "create README.md file")

		return subcommands.ExitFailure
	}

	if err := os.Mkdir(filepath.Join(directory, environment), 0755); err != nil {
		cli.Wrapf(err, "create environment \"%s\" directory", environment)

		return subcommands.ExitFailure
	}

	if err := os.WriteFile(filepath.Join(directory, environment, ".gitkeep"), []byte("*"), 0644); err != nil {
		cli.Wrapf(err, "create .gitkeep file for \"%s\" environment", environment)

		return subcommands.ExitFailure
	}

	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		cli.Wrapf(err, "generate keypair")

		return subcommands.ExitFailure
	}

	if err := os.WriteFile(filepath.Join(directory, environment+".public"), untold.Base64Encode(publicKey[:]), 0644); err != nil {
		cli.Wrapf(err, "write public key for environment \"%s\"", environment)

		return subcommands.ExitFailure
	}

	if err := os.WriteFile(filepath.Join(directory, environment+".private"), untold.Base64Encode(privateKey[:]), 0644); err != nil {
		cli.Wrapf(err, "write private key for environment \"%s\"", environment)

		return subcommands.ExitFailure
	}

	cli.Successf("Vault structure with \"%s\" environment initialized.\n", environment)

	return subcommands.ExitSuccess
}
