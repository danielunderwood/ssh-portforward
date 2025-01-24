package forwarding

import (
	"errors"
	"strings"
)

type ForwardType string

const (
	ForwardSSH = "SSH"
	ForwardECS = "ECS"
)

var AllForwardTypes = []ForwardType{ForwardSSH, ForwardECS}

func (f *ForwardType) String() string {
	return string(*f)
}

func (f *ForwardType) Set(value string) error {
	for _, v := range AllForwardTypes {
		if strings.EqualFold(string(v), value) {
			*f = v
			return nil
		}
	}
	return errors.New("invalid forward type")
}

func (f *ForwardType) Type() string {
	return "ForwardType"
}

type ForwardRuleBase struct {
	Port uint64 `mapstructure:"port"`
	Type string `mapstructure:"type"`
}

type ECSForwardRule struct {
	Base      ForwardRuleBase `mapstructure:",squash"`
	Profile   string
	Cluster   string
	Container string
}

type SSHForwardRule struct {
	Base     ForwardRuleBase `mapstructure:",squash"`
	Host     string
	User     string
	Identity string
}
