package formparse

import (
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
	setAllValues(req, val, tt)

	return *val
}

func setAllValues(req *http.Request, instance interface{}, tt reflect.Type) {
	tv := reflect.ValueOf(instance)
	if tv.Kind() == reflect.Pointer {
		tv = tv.Elem()
	}

	for i := 0; i < tt.NumField(); i++ {
		fld := tv.Field(i)
		sfld := tt.Field(i)

		if !fld.CanSet() {
			continue
		}

		if fld.Kind() == reflect.Struct {
			x := reflect.New(fld.Type()).Interface()
			setAllValues(req, x, fld.Type())
			fld.Set(reflect.ValueOf(x).Elem())
		}

		ctag := sfld.Tag.Get("cookie")
		qtag := sfld.Tag.Get("query")
		ftag := sfld.Tag.Get("form")
		if ctag != "" {
			c, err := req.Cookie(ctag)
			if err == nil {
				setValue(fld, c.Value)
			}
		} else if qtag != "" {
			c := req.URL.Query().Get(qtag)
			setValue(fld, c)
		} else if ftag != "" {
			c := req.Form[ftag]
			setValues(fld, c)
		}
	}
}

func setValue(fld reflect.Value, val string) {
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
