package untold

import (
	"github.com/fatih/structtag"
	"github.com/pkg/errors"
	"reflect"
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

		tags, err := structtag.Parse(string(reflectionType.Field(i).Tag))
		if err != nil {
			return err
		}

		tag, err := tags.Get("untold")
		if err != nil {
			continue
		}

		if reflectionField.Kind() == reflect.String && reflectionField.CanSet() {
			value, resolveErr := resolve(tag.Name)
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
