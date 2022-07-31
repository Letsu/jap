package jap

import (
	"regexp"
	"strconv"
	"strings"
)

type RunningConfig struct {
	Hostname        string
	FullConfig      string
	FullConfigNoNew string
	Interfaces      []CiscoInterface
	Vlans           []Vlan
}

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

		// Get all interfaces
		re, _ = regexp.Compile("interface ([\\w\\/\\.\\-\\:]+)")
		if re.MatchString(firstLine) {
			inter, err := ParseInterface(part)
			if err != nil {
				return RunningConfig{}, err
			}
			running.Interfaces = append(running.Interfaces, inter)
			continue
		}

		// Get vlans
		re = regexp.MustCompile(`^vlan ([0-9]+)`)
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

		// Router OSPF

		// Router BGP

		// Get lines

		//log.Println(firstLine)
	}

	return running, nil
}
