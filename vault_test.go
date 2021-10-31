package untold

import (
	"embed"
	"testing"
)

//go:embed test
var fs embed.FS

func TestFindSecret(t *testing.T) {
	v := (NewVault(fs, Environment("test"), PathPrefix("test"))).(*vault)

	if err := v.loadKeys(); err != nil {
		t.Fatal(err)
	}

	type test struct {
		input  string
		output string
		err    string
	}

	tests := []test{
		{input: "test", output: "test", err: ""},
		{input: "doesnt_exist", output: "", err: "secret \"doesnt_exist\" for \"test\" environment not found"},
	}

	for i := range tests {
		value, err := v.findSecret(tests[i].input)
		if err != nil && err.Error() != tests[i].err {
			t.Errorf("expected error to be %v, got %v", tests[i].input, err)
		}

		if value != tests[i].output {
			t.Errorf("expected to get %q, got %q", tests[i].output, value)
		}
	}
}

func TestLoadNotExistingKeys(t *testing.T) {
	v := (NewVault(fs, Environment("not_existing"), PathPrefix("test"))).(*vault)

	err := v.loadKeys()
	if err.Error() != "read public key file for \"not_existing\" environment: open test/not_existing.public: file does not exist" {
		t.Errorf("unexpected error %q", err.Error())
	}
}

func TestLoadKeysBadPrefix(t *testing.T) {
	v := (NewVault(fs, Environment("test"), PathPrefix("doesnt_exist"))).(*vault)

	err := v.loadKeys()
	if err.Error() != "read public key file for \"test\" environment: open doesnt_exist/test.public: file does not exist" {
		t.Errorf("unexpected error %q", err.Error())
	}
}
