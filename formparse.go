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
	setAllValues(req, val, tt, "")

	return *val
}

func ParseFormArray[T any](req *http.Request) []T {
	var list []T

	req.ParseForm()
	i := 0
	more := true

	for more {
		suffix := fmt.Sprintf("_%d", i)
		var t *T = new(T)
		tt := reflect.TypeOf(t).Elem()
		more = setAllValues(req, t, tt, suffix)
		if more {
			list = append(list, *t)
			i++
		}
	}

	return list
}

func setAllValues(req *http.Request, instance interface{}, tt reflect.Type, suffix string) bool {
	hasMore := true
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
			setAllValues(req, x, fld.Type(), suffix)
			fld.Set(reflect.ValueOf(x).Elem())
			continue
		}

		ctag := sfld.Tag.Get("cookie")
		qtag := sfld.Tag.Get("query")
		ftag := sfld.Tag.Get("form")
		if ctag != "" {
			c, err := req.Cookie(ctag + suffix)
			if err == nil {
				setValue(fld, c.Value)
				hasMore = c != nil && len(c.Value) > 0
			}
		} else if qtag != "" {
			c := req.URL.Query().Get(qtag + suffix)
			setValue(fld, c)
			hasMore = len(c) > 0
		} else if ftag != "" {
			c := req.Form[ftag+suffix]
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
