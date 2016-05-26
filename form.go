package routing

import (
	"errors"
	"reflect"
	"strconv"
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
	rv = indirect(rv)
	rt := rv.Type()
	n := rt.NumField()
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
			if err := readFormField(form, name, rv.Field(i)); err != nil {
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

func readFormField(form map[string][]string, name string, rv reflect.Value) error {
	value, ok := form[name]
	if !ok {
		return nil
	}
	rv = indirect(rv)
	if rv.Kind() != reflect.Slice {
		return setFormFieldValue(rv, value[0])
	}

	n := len(value)
	slice := reflect.MakeSlice(rv.Type(), n, n)
	for i := 0; i < n; i++ {
		if err := setFormFieldValue(slice.Index(i), value[i]); err != nil {
			return err
		}
	}
	rv.Set(slice)
	return nil
}

func setFormFieldValue(rv reflect.Value, value string) error {
	switch rv.Kind() {
	case reflect.Bool:
		if value == "" {
			value = "false"
		}
		v, err := strconv.ParseBool(value)
		if err == nil {
			rv.SetBool(v)
		}
		return err
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value == "" {
			value = "0"
		}
		v, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			rv.SetInt(v)
		}
		return err
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if value == "" {
			value = "0"
		}
		v, err := strconv.ParseUint(value, 10, 64)
		if err == nil {
			rv.SetUint(v)
		}
		return err
	case reflect.Float32, reflect.Float64:
		if value == "" {
			value = "0"
		}
		v, err := strconv.ParseFloat(value, 64)
		if err == nil {
			rv.SetFloat(v)
		}
		return err
	case reflect.String:
		rv.SetString(value)
		return nil
	default:
		return errors.New("Unknown type: " + rv.Kind().String())
	}
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
