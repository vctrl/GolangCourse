package main

import (
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
	ptr := reflect.ValueOf(out)
	if ptr.Kind() != reflect.Ptr {
		return TypeMismatchError(reflect.Ptr, ptr.Kind())
	}
	ov := ptr.Elem()
	if !ov.CanSet() {
		return fmt.Errorf("cannot set the value of %s", ov)
	}
	switch v := reflect.ValueOf(data); v.Kind() {
	case reflect.Map:
		if ov.Kind() != reflect.Struct {
			return TypeMismatchError(reflect.Struct, ov.Kind())
		}
		for _, e := range v.MapKeys() {
			field := ov.FieldByName(e.String())
			if !field.IsValid() {
				return fmt.Errorf("cannot find field %s", e.String())
			}
			if !field.CanAddr() {
				return fmt.Errorf("field %s is not addressable", field)
			}

			if err := i2s(v.MapIndex(e).Interface(), field.Addr().Interface()); err != nil {
				return err
			}
		}
	case reflect.Slice:
		if ov.Kind() != reflect.Slice {
			return TypeMismatchError(reflect.Slice, ov.Kind())
		}
		slice := reflect.MakeSlice(reflect.TypeOf(ov.Interface()), v.Len(), v.Len())
		ov.Set(slice)
		for i := 0; i < v.Len(); i++ {
			if err := i2s(v.Index(i).Interface(), slice.Index(i).Addr().Interface()); err != nil {
				return err
			}
		}
	case reflect.Float64:
		if ov.Kind() != reflect.Int {
			return TypeMismatchError(reflect.Int, ov.Kind())
		}

		f := v.Float()
		ov.SetInt(int64(f))
	case reflect.String:
		if ov.Kind() != reflect.String {
			return TypeMismatchError(reflect.String, ov.Kind())
		}
		ov.SetString(v.String())
	case reflect.Bool:
		if ov.Kind() != reflect.Bool {
			return TypeMismatchError(reflect.Bool, ov.Kind())
		}
		ov.SetBool(v.Bool())
	default:
		return fmt.Errorf("unhandled kind %s", v.Kind())
	}

	return nil
}

func TypeMismatchError(ek reflect.Kind, fk reflect.Kind) error {
	return fmt.Errorf("type mismatch, expected %s but was %s", ek, fk)
}
