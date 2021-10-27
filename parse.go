package untold

import (
	"github.com/pkg/errors"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrNotAPointer = errors.Errorf("expected pointer to a Struct")
)

type resolveFn func(name string) (string, error)

func parse(dst interface{}, resolve resolveFn) error {
	pointerReflection := reflect.ValueOf(dst)
	if pointerReflection.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}

	reflection := pointerReflection.Elem()
	if reflection.Kind() != reflect.Struct {
		return ErrNotAPointer
	}

	return doParse(reflection, resolve)
}

func doParse(reflection reflect.Value, resolve resolveFn) error {
	reflectionType := reflection.Type()

	for i := 0; i < reflectionType.NumField(); i++ {
		reflectionField := reflection.Field(i)
		if !reflectionField.CanSet() {
			continue
		}

		if reflectionField.Kind() == reflect.Ptr && !reflectionField.IsNil() {
			err := parse(reflectionField.Interface(), resolve)
			if err != nil {
				return err
			}
			continue
		}

		if reflectionField.Kind() == reflect.Struct && reflectionField.CanAddr() && reflectionField.Type().Name() == "" {
			err := parse(reflectionField.Addr().Interface(), resolve)
			if err != nil {
				return err
			}
			continue
		}

		name := tagName(reflectionType.Field(i).Tag, "untold")
		if name == "" {
			continue
		}

		if reflectionField.Kind() == reflect.String && reflectionField.CanSet() {
			value, resolveErr := resolve(name)
			if resolveErr != nil {
				return resolveErr
			}

			reflectionField.SetString(value)

			continue
		}

		return errors.Errorf("tag \"untold\" works only with strings. used on %s", reflectionField.Type())
	}

	return nil
}

// https://github.com/fatih/structtag/blob/master/tags.go
func tagName(tag reflect.StructTag, target string) string {
	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// Scan to colon. A space, a quote or a control character is a syntax
		// error. Strictly speaking, control chars include the range [0x7f,
		// 0x9f], not just [0x00, 0x1f], but in practice, we ignore the
		// multi-byte control characters as it is simpler to inspect the tag's
		// bytes than the tag's runes.
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}

		if i == 0 {
			return ""
		}
		if i+1 >= len(tag) || tag[i] != ':' {
			return ""
		}
		if tag[i+1] != '"' {
			return ""
		}

		key := string(tag[:i])
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			return ""
		}

		qvalue := string(tag[:i+1])
		tag = tag[i+1:]

		value, err := strconv.Unquote(qvalue)
		if err != nil {
			return ""
		}

		res := strings.Split(value, ",")
		name := res[0]

		if key == target {
			return name
		}
	}

	return ""
}