package jap

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func Generate(parsed any) (string, error) {
	var config string

	t := reflect.TypeOf(parsed)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("cmd")
		if tag != "" {
			cmd, err := GenerateCMD(reflect.ValueOf(&parsed).Elem().Elem().Field(i), tag)
			if err != nil {
				return "", err
			}
			if cmd == "" {
				continue
			}
			cmd = "  " + cmd + "\n"
			config = config + cmd
		}
	}
	config = config + "!"
	return config, nil
}

func GenerateCMD(field reflect.Value, tag string) (string, error) {
	var cmd string

	switch field.Type().Kind() {
	case reflect.String:
		value := field.String()
		if value == "" {
			return "", nil
		}
		cmd = fmt.Sprintf(tag, value)
	case reflect.Int:
		value := field.Int()
		if value == 0 {
			return "", nil
		}
		cmd = fmt.Sprintf(tag, value)
	case reflect.Bool:
		value := field.Bool()
		if !value {
			return "", nil
		}
		cmd = tag
	case reflect.Float64:
		value := field.Float()
		if value == 0.0 {
			return "", nil
		}
		cmd = fmt.Sprintf(tag, value)
	case reflect.Slice:
		switch field.Type().String() {
		case "[]string":
			slice, _ := field.Interface().([]string)
			if len(slice) == 0 {
				return "", nil
			}
			cmds := ""
			for i2, s := range slice {
				cmd1 := fmt.Sprintf(tag, s)
				if i2 == 0 {
					cmds = cmds + cmd1
				} else {
					cmds = cmds + "\n" + "  " + cmd1
				}
			}
			cmd = cmd + cmds
		case "[]int":
			slice, _ := field.Interface().([]int)
			if len(slice) == 0 {
				return "", nil
			}
			var sliceStr []string
			for _, s := range slice {
				text := strconv.Itoa(s)
				sliceStr = append(sliceStr, text)
			}
			cmd = fmt.Sprintf(tag, strings.Join(sliceStr, ","))
		default:
			if field.Type().Kind() == reflect.Slice && tag != "" {

				for i2 := 0; i2 < field.Len(); i2++ {
					cmds := tag
					for i1 := 0; i1 < field.Type().Elem().NumField(); i1++ {
						gcmd, _ := GenerateCMD(field.Index(i2).Field(i1), field.Index(0).Type().Field(i1).Tag.Get("cmd"))
						cmds = cmds + gcmd
						//cmds = fmt.Sprintf(field1.Index(0).Type().Field(i1).Tag.Get("cmd"), )
					}
					if i2 == 0 {
						cmd = cmd + cmds
					} else {
						cmd = cmd + "\n" + "  " + cmds
					}
				}
			}
		}
	default:
		panic(field.Type().Kind().String() + " not implemented")
	}

	return cmd, nil
}
