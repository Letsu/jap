package jap

import (
	"fmt"
	"regexp"
	"strconv"
)

type Ospf struct {
	ProcessID               int
	ProcessVRF              string
	LogAdjacencyChange      bool `reg:"log-adjacency-changes detail" cmd:"log-adjacency-changes detail"`
	PassiveInterfaceDefault bool `reg:"passive-interface default" cmd:"passive-interface default"`
	PassiveInterface        []int
	Network                 []OspfNetwork `reg:"network.*" cmd:"network"`
}

type OspfNetwork struct {
	NetworkNumber string `reg:"network (\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}) (?:\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}) area (?:\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}|\\d+)" cmd:" %s"`
	WildCard      string `reg:"network (?:\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}) (\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}) area (?:\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}|\\d+)" cmd:" %s"`
	Area          string `reg:"network (?:\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}) (?:\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}) area (\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}|\\d+)" cmd:" area %s"`
}

func (o *Ospf) Parse(part string) error {
	re := regexp.MustCompile(`router ospf (\d+)(?: vrf ([[:print:]]+))?`)
	head := re.FindStringSubmatch(part)
	o.ProcessID, _ = strconv.Atoi(head[1])
	if len(head) > 2 {
		o.ProcessVRF = head[2]
	}

	err := processParse(part, o)
	if err != nil {
		return err
	}

	return nil
}

func (o Ospf) Generate() (string, error) {
	var config string

	if o.ProcessVRF != "" {
		config = fmt.Sprintf("router ospf %d vrf %s\n", o.ProcessID, o.ProcessVRF)
	} else {
		config = fmt.Sprintf("router ospf %d\n", o.ProcessID)
	}

	generated, err := Generate(o)
	if err != nil {
		return "", err
	}
	config = config + generated

	return config, nil
}
