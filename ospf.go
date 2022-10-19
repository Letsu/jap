package jap

type Ospf struct {
	LogAdjacencyChange      bool `reg:"log-adjacency-changes detail" cmd:"log-adjacency-changes detail"`
	PassiveInterfaceDefault bool `reg:"passive-interface default" cmd:"passive-interface default"`
	PassiveInterface        []int
	Network                 []OspfNetwork
}

type OspfNetwork struct {
	NetworkNumber string `reg:"network (\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}) (?:\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}) area (?:\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}|\d+)"`
	WildCard      string `reg:"network (?:\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}) (\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}) area (?:\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}|\d+)"`
	Area          string `reg:"network (?:\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}) (?:\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}) area (\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}|\d+)"`
}

func ParseOSPF(part string) (Ospf, error) {

	return Ospf{}, nil
}
