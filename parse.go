package jap

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type RunningConfig struct {
	Hostname        string
	FullConfig      string
	FullConfigNoNew string
	Vlans           []Vlan
	Interfaces      []CiscoInterface
	OSPFProcess     []Ospf
}

// Parse parses a standard cisco running config.
func Parse(config string) (RunningConfig, error) {
	var running RunningConfig

	//Remove all empty spaces from the config and split for parts
	running.FullConfig = config
	re := regexp.MustCompile(`(?m)^\s*$[\r\n]*|[\r\n]+\s+\z`) // Regex to search all newline on Windows and Linux with no content in the line
	running.FullConfigNoNew = re.ReplaceAllString(running.FullConfig, "")

	re = regexp.MustCompile(`(?m)^!`)
	splitRun := re.Split(running.FullConfigNoNew, -1)

	//Go through and parse all parts of the config
	for _, part := range splitRun {
		fistLineArr := strings.Split(part, "\n")
		if len(fistLineArr) == 1 {
			continue
		}
		firstLine := fistLineArr[1]

		// Get hostname
		re = regexp.MustCompile(`(?m)^hostname ([[:print:]]+)`)
		fullHostname := re.FindStringSubmatch(part)
		if len(fullHostname) > 0 {
			running.Hostname = fullHostname[1]
			continue
		}

		// Get vlans
		re = regexp.MustCompile(`^\s*vlan (\d+)`)
		vlanPart := re.FindStringSubmatch(firstLine)
		if len(vlanPart) > 1 {
			vlanId, _ := strconv.Atoi(vlanPart[1])
			vlan, err := ParseVlan(part, vlanId)
			if err != nil {
				return RunningConfig{}, err
			}

			running.Vlans = append(running.Vlans, vlan)
			continue
		}

		// Get all interfaces
		re, _ = regexp.Compile(`^\s*interface ([\w\/\.\-\:]+)`)
		if re.MatchString(firstLine) {
			var inter CiscoInterface
			err := inter.Parse(part)
			if err != nil {
				return RunningConfig{}, err
			}
			running.Interfaces = append(running.Interfaces, inter)
			continue
		}

		// Router OSPF
		re, _ = regexp.Compile(`^\s*router ospf (\d+)( vrf ([[:print:]]+))?`)
		if re.MatchString(firstLine) {
			var ospf Ospf
			err := ospf.Parse(part)
			if err != nil {
				return RunningConfig{}, err
			}
			running.OSPFProcess = append(running.OSPFProcess, ospf)
			continue
		}

		// Router BGP

		// Get lines

		//log.Println(firstLine)
	}

	return running, nil
}

// processParse takes a part of the config and a pointer to a struct with tags on how to parse the config part.
// processParse can also take a reflect.Value to call itself.
func processParse(part string, parsed any) error {
	// Check if type is already a "reflect.Value", to let the function call itself in case of a struct in a struct
	var tmp, rv reflect.Value
	// In the normal case the function takes a struct.
	//In case of a struct as parsed parameter it creates a copy of the struct to write to the copy of its reflection
	// When a reflect.Value is already used as parsed parameter this is not needed and cause an error.
	// The Value still needs to be a reflect of a struct value to function. This is needed to work with recursion e.g.
	// structs in structs.
	// I don't really now how I got all this reflect stuff together and working so note to further self:
	// Make this more stable as soon I understand this shit...
	if reflect.TypeOf(parsed).String() != "reflect.Value" {
		// Generate a copy of the "parsed" interface to fill it with values
		v := reflect.Indirect(reflect.ValueOf(&parsed)).Elem()
		// tmp is the copy and can be written to.
		// parsed, rv and rt are read only. parsed can be completely overwritten but no values can be set on by one
		tmp = reflect.New(v.Elem().Type()).Elem()
		tmp.Set(v.Elem())

		rv = reflect.Indirect(reflect.ValueOf(parsed))
	} else {
		// write any value as reflect.Value to all the copied and real value
		// in this case parsed is writeable and no copy needed
		tmp, rv = parsed.(reflect.Value), parsed.(reflect.Value)
	}
	rt := rv.Type()

	// for through all field of the struct, get the regex tag and fill it with the found data
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		tag := field.Tag.Get("reg")
		if tag != "" {
			re := regexp.MustCompile(tag)
			// @todo check if no is with the command!
			data := re.FindAllStringSubmatch(part, -1)
			if len(data) == 0 {
				continue
			}

			// Check the Line of the struct field
			switch field.Type.Kind() {
			case reflect.String:
				tmp.Field(i).SetString(data[0][1])
				break
			case reflect.Int:
				value, err := strconv.ParseInt(data[0][1], 10, 64)
				if err != nil {
					return nil
				}
				tmp.Field(i).SetInt(value)
				break
			case reflect.Bool:
				tmp.Field(i).SetBool(true)
				break
			case reflect.Float64:
				float, err := strconv.ParseFloat(data[0][1], 64)
				if err != nil {
					return nil
				}
				tmp.Field(i).SetFloat(float)
			default:
				panic(field.Type.String() + " not implemented!")
			case reflect.Slice:
				switch field.Type.String() {
				case "[]string":
					for _, d := range data {
						value := tmp.Field(i)
						value.Set(reflect.Append(value, reflect.ValueOf(d[1])))
					}
				case "[]int":
					value := tmp.Field(i)
					for _, d := range data {
						seperated := strings.Split(d[2], ",")
						for _, number := range seperated {
							if strings.Contains(number, "-") {
								vlanSplit := strings.Split(number, "-")
								from, _ := strconv.Atoi(vlanSplit[0])
								to, _ := strconv.Atoi(vlanSplit[1])
								for j := from; j <= to; j++ {
									value.Set(reflect.Append(value, reflect.ValueOf(j)))
								}
								continue
							}
							vlanI, _ := strconv.Atoi(number)
							value.Set(reflect.Append(value, reflect.ValueOf(vlanI)))
						}
					}
				default:
					values := tmp.Field(i)
					if values.Type().Kind() == reflect.Slice {
						for _, t := range data {
							tmp2 := reflect.New(values.Type().Elem()).Elem()
							err := processParse(strings.Join(t, " "), tmp2)
							if err != nil {
								return err
							}
							values.Set(reflect.Append(values, tmp2))
						}
						continue
					}
					panic(field.Type.String() + " not implemented!")
				}
			}
		}
	}

	// Overwrite parsed with tmp to get the values back to the caller.
	// In case parsed is a reflect.Value it needs to write into the element and not the Value itself
	if reflect.TypeOf(parsed).String() != "reflect.Value" {
		reflect.ValueOf(parsed).Elem().Set(tmp)
		return nil
	}

	parsed.(reflect.Value).Set(tmp)
	return nil
}
