package jap

import (
	"fmt"
	"reflect"
)

func Generate(parsed any) (string, error) {
	var config string

	t := reflect.TypeOf(parsed)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("cmd")
		if tag != "" {
			var cmd string
			switch field.Type.Kind() {
			case reflect.String:
				value := reflect.ValueOf(&parsed).Elem().Elem().Field(i).String()
				if value == "" {
					continue
				}
				cmd = fmt.Sprintf(tag, value)
			case reflect.Int:
				value := reflect.ValueOf(&parsed).Elem().Elem().Field(i).Int()
				if value == 0 {
					continue
				}
				cmd = fmt.Sprintf(tag, value)
			case reflect.Bool:
				value := reflect.ValueOf(&parsed).Elem().Elem().Field(i).Bool()
				if !value {
					continue
				}
				cmd = tag
			case reflect.Float64:
				value := reflect.ValueOf(&parsed).Elem().Elem().Field(i).Float()
				if value == 0.0 {
					continue
				}
				cmd = fmt.Sprintf(tag, value)
			default:
				continue
			}
			cmd = "  " + cmd + "\n"
			config = config + cmd
		}
	}
	config = config + "!"
	return config, nil
}
