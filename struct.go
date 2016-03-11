// Copyright 2016 polaris. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author：polaris	polaris@studygolang.com

package jsonutils

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/polaris1119/logger"
)

// ParseJsonSliceStruct 解析 json 到 slice struct
func ParseJsonSliceStruct(body []byte, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &json.InvalidUnmarshalError{Type: reflect.TypeOf(v)}
	}

	sjson, err := simplejson.NewJson(body)
	if err != nil {
		return err
	}

	rv = rv.Elem()
	if rv.Kind() == reflect.Slice {
		for i, length := 0, rv.Len(); i < length; i++ {
			modelVal := rv.Index(i)
			ParseJsonOneStruct(modelVal, sjson.GetIndex(i))
		}
		return nil
	} else {
		return ParseJsonOneStruct(rv, sjson)
	}
}

// ParseJsonOneStruct 解析 json 到 struct
func ParseJsonOneStruct(rv reflect.Value, sjson *simplejson.Json) error {
	if rv.Kind() != reflect.Struct {
		return errors.New("v must be pointer of struct")
	}

	fieldType := rv.Type()

	for i, fieldCount := 0, rv.NumField(); i < fieldCount; i++ {
		fieldVal := rv.Field(i)
		if !fieldVal.CanSet() {
			continue
		}

		structField := fieldType.Field(i)
		structTag := structField.Tag
		name := structTag.Get("json")

		var (
			tmpJson *simplejson.Json
			ok      bool
		)
		if tmpJson, ok = sjson.CheckGet(name); !ok {
			name = structField.Name
			if tmpJson, ok = sjson.CheckGet(name); !ok {
				continue
			}
		}

		switch structField.Type.Kind() {
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint, reflect.Uintptr:
			val := tmpJson.MustUint64()
			if val == 0 {
				val, _ = strconv.ParseUint(tmpJson.MustString("0"), 10, 64)
			}
			fieldVal.SetUint(val)
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			val := tmpJson.MustInt64()
			if val == 0 {
				val, _ = strconv.ParseInt(tmpJson.MustString("0"), 10, 64)
			}
			fieldVal.SetInt(val)
		case reflect.String:
			val, err := tmpJson.String()
			if err != nil {
				v, err := tmpJson.Uint64()
				if err != nil {
					logger.Errorln("[string]name=", name, ";field_type=string;value=", tmpJson)
				} else {
					val = strconv.FormatUint(v, 10)
				}
			}
			fieldVal.SetString(val)
		case reflect.Bool:
			val := tmpJson.MustBool()
			if !val {
				val, _ = strconv.ParseBool(tmpJson.MustString("false"))
			}
			fieldVal.SetBool(val)
		case reflect.Float32, reflect.Float64:
			val := tmpJson.MustFloat64()
			if val == 0.0 {
				val, _ = strconv.ParseFloat(tmpJson.MustString("0.0"), 64)
			}
			fieldVal.SetFloat(val)
		case reflect.Struct:
			if structField.Type.Name() == "Time" {
				local, _ := time.LoadLocation("Local")
				val, err := time.ParseInLocation("2006-01-02 15:04:05", tmpJson.MustString(), local)
				if err == nil {
					fieldVal.Set(reflect.ValueOf(val))
				}
			}
		default:
			logger.Errorln("name=", name, ";field_type=", structField.Type.Kind(), ";value=", tmpJson)
		}
	}

	return nil
}
