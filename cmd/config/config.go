package config

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"

	f "github.com/danielunderwood/portctl/forwarding"
)

type Config map[string]interface{}

var C Config

func Init() {
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
	}
	err = viper.Unmarshal(&C)
	if err != nil {
		panic(err)
	}

	updatedRules := make(map[string]interface{})
	// First pass
	for k, v := range C {
		var decoded f.ForwardRuleBase
		err := mapstructure.Decode(v, &decoded)
		if err != nil {
			panic(err)
		}
		fmt.Println(k, decoded)
		switch decoded.Type {
		case f.ForwardECS:
			fmt.Println("Handling ECS Forward")
			var ecsForward f.ECSForwardRule
			err = mapstructure.Decode(v, &ecsForward)
			if err != nil {
				panic(err)
			}
			fmt.Printf("ECS Forward Rule: %v\n", ecsForward)
			updatedRules[k] = ecsForward
		default:
			panic(fmt.Errorf("unknown forward type %s", decoded.Type))
		}
	}
	C = updatedRules
	fmt.Printf("Updated Rules: %v\n", updatedRules)
}
