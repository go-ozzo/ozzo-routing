package routing

import (
	"errors"
	"reflect"
)

const formTag = "form"

func ReadForm(form map[string][]string, data interface{}) error {
	rv := reflect.ValueOf(data)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("data must be a pointer")
	}
	rv = indirect(rv)
	if rv.Kind() != reflect.Struct {
		return errors.New("data must be a pointer to a struct")
	}

	return readForm(form, "", rv)
}

func readForm(form map[string][]string, prefix string, rv reflect.Value) error {
	n := rv.NumField()
	rt := rv.Type()
	for i := 0; i < n; i++ {
		field := rt.Field(i)
		tag := field.Tag.Get(formTag)

		// only handle anonymous or exported fields
		if !field.Anonymous && field.PkgPath != "" || tag == "-" {
			continue
		}

		ft := field.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}

		name := tag
		if name == "" && !field.Anonymous {
			name = field.Name
		}
		if name != "" && prefix != "" {
			name = prefix + "." + name
		}

		if ft.Kind() != reflect.Struct {
			if err := setField(rv.Field(i), form, name); err != nil {
				return err
			}
			continue
		}

		if name == "" {
			name = prefix
		}
		if err := readForm(form, name, rv.Field(i)); err != nil {
			return err
		}
	}
	return nil
}

func setField(rv reflect.Value, form map[string][]string, name string) error {
	value, ok := form[name]
	if !ok {
		return nil
	}
	rv = indirect(rv)
	if rv.Kind() != reflect.Slice {
		return setValue(rv, value[0])
	}

	n := len(value)
	slice := reflect.MakeSlice(rv.Type(), n, n)
	for i := 0; i < n; i++ {
		if err := setValue(slice.Index(i), value[i]); err != nil {
			return err
		}
	}
	rv.Set(slice)
	return nil
}

func setValue(rv reflect.Value, value string) error {
	return nil
}

// indirect dereferences pointers and returns the actual value it points to.
// If a pointer is nil, it will be initialized with a new value.
func indirect(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}
