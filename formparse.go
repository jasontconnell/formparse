package formparse

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

func ParseForm[T any](req *http.Request) T {
	var val *T = new(T)

	req.ParseForm()
	tt := reflect.TypeOf(val).Elem()
	setAllValues(req, val, tt, "", "")

	return *val
}

func ParseFormArray[T any](req *http.Request) []T {
	var item T
	etype := reflect.TypeOf(item).Elem()
	atype := reflect.ArrayOf(0, etype)

	list := reflect.New(atype).Elem()

	req.ParseForm()
	i := 0
	more := true

	for more {
		suffix := fmt.Sprintf("_%d", i)
		var t *T = new(T)
		item := reflect.ValueOf(t).Elem()

		more = setAllValues(req, t, etype, "", suffix)
		if more {
			list = reflect.Append(list, item)
			i++
		}
	}

	array := list.Interface().([]T)
	return array
}

// instance should be a pointer to slice
func setArrayValues(req *http.Request, instance interface{}, tt reflect.Type, prefix string) bool {
	more := true
	i := 0
	valueInstance := reflect.ValueOf(instance).Elem()

	log.Println("setting array values for type", tt.Name())

	for more {
		suffix := fmt.Sprintf("_%d", i)
		t := reflect.New(tt)
		more = setAllValues(req, t, tt, prefix, suffix)

		if more {
			valueInstance = reflect.Append(valueInstance, t)
		}
	}

	log.Println("value", instance, valueInstance)

	return false
}

func setAllValues(req *http.Request, instance interface{}, tt reflect.Type, prefix, suffix string) bool {
	hasMore := true
	tv := reflect.ValueOf(instance)
	if tv.Kind() == reflect.Pointer {
		tv = tv.Elem()
	}

	for i := 0; i < tt.NumField(); i++ {
		fld := tv.Field(i)
		sfld := tt.Field(i)

		log.Println(sfld.Name)

		if !fld.CanSet() {
			continue
		}

		ftag := sfld.Tag.Get("form")
		ctag := sfld.Tag.Get("cookie")
		qtag := sfld.Tag.Get("query")

		if fld.Kind() == reflect.Struct {
			log.Println("setting struct", sfld.Name)
			x := reflect.New(fld.Type()).Interface()
			setAllValues(req, x, fld.Type(), prefix, suffix)
			fld.Set(reflect.ValueOf(x).Elem())
			continue
		} else if fld.Kind() == reflect.Slice && fld.Type().Elem().Kind() == reflect.Struct {
			log.Println(sfld.Name, fld.Type().Kind(), fld.Type().Elem().Kind())
			list := reflect.New(reflect.ArrayOf(0, fld.Type())).Interface()
			setArrayValues(req, list, fld.Type().Elem(), ftag+"_")
			log.Println("set array values, got", list, "setting value on", sfld.Name)
			fld.Set(reflect.ValueOf(list).Elem())
			continue
		}

		if ctag != "" {
			c, err := req.Cookie(prefix + ctag + suffix)
			if err == nil {
				setValue(fld, c.Value)
				hasMore = c != nil && len(c.Value) > 0
			}
		} else if qtag != "" {
			c := req.URL.Query().Get(prefix + qtag + suffix)
			setValue(fld, c)
			hasMore = len(c) > 0
		} else if ftag != "" {
			c := req.Form[prefix+ftag+suffix]
			setValues(fld, c)
			hasMore = len(c) > 0
		}
	}
	return hasMore
}

func setValue(fld reflect.Value, val string) {
	if len(val) == 0 {
		return
	}

	if fld.Type().Name() == "string" {
		fld.SetString(val)
	} else if fld.Type().Name() == "int" {
		x, _ := strconv.Atoi(val)
		fld.SetInt(int64(x))
	} else if fld.Type().Name() == "bool" {
		b := boolVal(val)
		fld.SetBool(b)
	} else {
		log.Println(fld.Type().Name())
	}
}

func setValues(fld reflect.Value, vals []string) {
	if len(vals) == 0 {
		return
	}
	if fld.Type().Kind() != reflect.Slice {
		setValue(fld, vals[0])
	} else {
		stype := fld.Type().Elem().Name()
		if stype == "string" {
			cp := make([]string, len(vals))
			copy(cp, vals)
			vval := reflect.ValueOf(cp)
			fld.Set(vval)
		} else if stype == "int" {
			v := []int{}
			for _, val := range vals {
				x, err := strconv.Atoi(val)
				if err == nil {
					v = append(v, x)
				}
			}
			vval := reflect.ValueOf(v)
			fld.Set(vval)
		}
	}
}

func boolVal(val string) bool {
	return val == "on" || val == "1" || strings.ToLower(val) == "true" || val == "yes"
}
