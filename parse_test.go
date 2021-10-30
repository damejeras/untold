package untold

import (
	"testing"
)

func TestSimpleParse(t *testing.T) {
	type holder struct {
		Value string `untold:"value"`
	}

	var h holder
	err := parse(&h, func(name string) (string, error) {
		return name, nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if h.Value != "value" {
		t.Errorf("expected %q, got %q", "value", h.Value)
	}
}

func TestNestedParse(t *testing.T) {
	type holder struct {
		Child struct {
			Value string `untold:"value"`
		}
	}

	var h holder
	err := parse(&h, func(name string) (string, error) { return name, nil })

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if h.Child.Value != "value" {
		t.Errorf("expected %q, got %q", "value", h.Child.Value)
	}
}

func TestNonPointerParse(t *testing.T) {
	type holder struct {
		Child struct {
			Value string `untold:"value"`
		}
	}

	var h holder
	err := parse(h, func(name string) (string, error) { return name, nil })

	if err == nil {
		t.Errorf("expected to get a error")
	}
}

func TestPointerChildParse(t *testing.T) {
	type Child struct {
		Value string `untold:"value"`
	}

	h := struct {
		Child *Child
	}{
		Child: &Child{},
	}

	err := parse(&h, func(name string) (string, error) { return name, nil })
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if h.Child.Value != "value" {
		t.Errorf("expected Child.Value to be %q, got %q", "value", h.Child.Value)
	}
}

func TestUnexportedValue(t *testing.T) {
	type holder struct {
		value string `untold:"value"`
	}

	var h holder
	err := parse(&h, func(name string) (string, error) { return name, nil })

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if h.value != "" {
		t.Errorf("expected %q, got %q", "", h.value)
	}
}
