// Simple persistent port forwarding to remote host

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/rgzr/sshtun"
)

type ForwardRule struct {
	LocalPort  int
	RemotePort int
}

func parseInt(data interface{}) (int, error) {
	switch t := data.(type) {
	case int:
		return t, nil
	case int64:
		return int(t), nil
	default:
		return 0, fmt.Errorf("unable to parse %T", data)
	}
}

func (f *ForwardRule) UnmarshalTOML(data interface{}) error {
	// m, ok := data.(map[string]interface{})

	// Unable to parse as map, try to parse as int
	// if !ok {
	// 	port, ok := data.(int64)
	// 	if !ok {
	// 		return fmt.Errorf("unable to parse %T", data)
	// 	}
	// 	f.LocalPort = int(port)
	// 	f.RemotePort = int(port)
	// 	return nil
	// }
	// if v, ok := m["local_port"]; ok {
	// 	f.LocalPort = v.(int)
	// }
	// if v, ok := m["remote_port"]; ok {
	// 	f.RemotePort = v.(int)
	// }
	switch data := data.(type) {
	case int, int64:
		port, err := parseInt(data)
		if err != nil {
			return fmt.Errorf("unable to parse %T", data)
		}
		f.LocalPort = port
		f.RemotePort = port
	case map[string]interface{}:
		fmt.Println("map")
		if v, ok := data["local_port"]; ok {
			f.LocalPort, _ = parseInt(v)
		}
		if v, ok := data["remote_port"]; ok {
			f.RemotePort, _ = parseInt(v)
		}
	default:
		return fmt.Errorf("unable to parse %T", data)
	}
	return nil
}

type Config struct {
	Host  string
	Key   string
	Ports []ForwardRule
	User  string
}

func main() {
	config := flag.String("c", "config.toml", "Path to config file")
	flag.Parse()

	// Read config file
	data, err := os.ReadFile(*config)
	if err != nil {
		log.Fatalf("Error reading config file: %+v", err)
		panic(err)
	}
	var conf Config
	_, err = toml.Decode(string(data), &conf)
	if err != nil {
		log.Fatalf("Error decoding config file: %+v", err)
		panic(err)
	}

	log.Printf("Config: %v", conf)

	ctx, cancel := context.WithCancel(context.Background())

	// Interrupt handling -- when receiving SIGINT, cancel the context
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		for sig := range sig {
			log.Printf("Signal received: %+v", sig)
			cancel()
		}
	}()

	for _, rule := range conf.Ports {
		log.Printf("Starting tunnel for %+v", rule)
		tun := sshtun.New(rule.LocalPort, conf.Host, rule.RemotePort)
		if conf.Key != "" {
			tun.SetKeyFile(conf.Key)
		} else {
			tun.SetSSHAgent()
		}
		tun.SetUser(conf.User)
		tun.SetTunneledConnState(func(tun *sshtun.SSHTun, state *sshtun.TunneledConnState) {
			log.Printf("%+v", state)
		})
		tun.SetConnState(func(conn *sshtun.SSHTun, state sshtun.ConnState) {
			switch state {
			case sshtun.StateStarting:
				log.Printf("STATE of %+v is Starting", rule)
			case sshtun.StateStarted:
				log.Printf("STATE of %+v  is Started", rule)
			case sshtun.StateStopped:
				log.Printf("STATE of %+v  is Stopped", rule)
			}
		})

		go func() {
			for {
				if err := tun.Start(ctx); err != nil {
					log.Printf("Error starting tunnel: %v", err)
				}
				time.Sleep(5 * time.Second)
			}
		}()
	}

	// Wait for the context to be cancelled
	<-ctx.Done()
}
