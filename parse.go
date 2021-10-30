package untold

import (
	"fmt"
	"reflect"
)

type resolveFn func(name string) (string, error)

func parse(dst interface{}, resolve resolveFn) error {
	pointerReflection := reflect.ValueOf(dst)
	if pointerReflection.Kind() != reflect.Ptr || pointerReflection.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	return parseRecursively(pointerReflection, resolve)
}

func parseRecursively(reflection reflect.Value, resolve resolveFn) error {
	if reflection.Kind() == reflect.Ptr {
		reflection = reflection.Elem()
	}

	for i := 0; i < reflection.Type().NumField(); i++ {
		reflectionField := reflection.Field(i)
		if !reflectionField.CanSet() {
			continue
		}

		switch {
		case reflectionField.Kind() == reflect.Ptr && !reflectionField.Addr().IsNil() && reflectionField.CanAddr():
			if err := parseRecursively(reflectionField, resolve); err != nil {
				return fmt.Errorf("%q: %v", reflection.Type().Field(i).Name, err)
			}
		case reflectionField.Kind() == reflect.Struct:
			if err := parseRecursively(reflectionField, resolve); err != nil {
				return fmt.Errorf("%q: %v", reflection.Type().Field(i).Name, err)
			}
		case reflectionField.Kind() == reflect.String:
			tagValue := reflection.Type().Field(i).Tag.Get("untold")
			value, resolveErr := resolve(tagValue)
			if resolveErr != nil {
				return fmt.Errorf("%q: resolve %q: %v", reflection.Type().Field(i).Name, tagValue, resolveErr)
			}

			if value != "" {
				reflectionField.SetString(value)
			}
		}
	}

	return nil
}
