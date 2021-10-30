package untold

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	ErrNotAPointer = errors.New("expected pointer to a Struct")
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

		return fmt.Errorf("tag \"untold\" works only with strings. used on %q", reflectionField.Type())
	}

	return nil
}

func tagName(tag reflect.StructTag, target string) string {
	tags := strings.Split(string(tag), " ")
	if len(tags) == 0 {
		return ""
	}

	for i := range tags {
		tagParts := strings.Split(tags[i], ":")
		if len(tagParts) != 2 {
			return ""
		}

		if tagParts[0] == target {
			return strings.Trim(tagParts[1], "\"")
		}
	}

	return ""
}