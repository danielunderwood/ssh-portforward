package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/danielunderwood/portctl/cmd/config"
	f "github.com/danielunderwood/portctl/forwarding"
	"github.com/danielunderwood/portctl/forwarding/ecs"
	"github.com/danielunderwood/portctl/forwarding/ssh"
)

var startForwardCmd = &cobra.Command{
	Use:   "start-forward",
	Short: "Forward ports from config file",
	Long:  "Forward ports from config file",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Config: %v", config.C)
		for k, v := range config.C {
			fmt.Printf("Key: %s, Value: %v\n", k, v)
			switch v := v.(type) {
			case f.ECSForwardRule:
				// ecs.ForwardECS(v)
				ecs.ForwardECS(v)
			default:
				fmt.Printf("Unknown type %T for key %s\n", v, k)
			}
		}
	},
}

var forwardCmd = &cobra.Command{
	Use:   "forward",
	Short: "Forward a single port based on CLI arguments",
	Run: func(cmd *cobra.Command, args []string) {
		switch forwardType {
		case f.ForwardECS:
			if ecsForwardCluster == "" || ecsForwardContainer == "" || forwardRemotePort == 0 {
				fmt.Println("Cluster, container, and remote port must be specified for ECS forwarding")
				os.Exit(1)
			}
			if forwardLocalPort == 0 {
				forwardLocalPort = forwardRemotePort
			}
			rule := f.ECSForwardRule{
				Base:      f.ForwardRuleBase{Type: f.ForwardECS, Port: forwardRemotePort},
				Cluster:   ecsForwardCluster,
				Container: ecsForwardContainer,
			}
			ecs.ForwardECS(rule)
		case f.ForwardSSH:
			if sshHost == "" || forwardRemotePort == 0 {
				fmt.Println("Host and remote port must be specified for SSH forwarding")
				os.Exit(1)
			}
			if forwardLocalPort == 0 {
				forwardLocalPort = forwardRemotePort
			}
			rule := f.SSHForwardRule{
				Base:     f.ForwardRuleBase{Type: f.ForwardSSH, Port: forwardRemotePort},
				Host:     sshHost,
				User:     sshUser,
				Identity: sshIdentityKey,
			}
			ssh.ForwardSSH(rule)
		default:
			panic("Unreachable")
		}
	},
}

var forwardType f.ForwardType
var ecsForwardCluster string
var ecsForwardContainer string
var forwardLocalPort uint64
var forwardRemotePort uint64
var sshHost string
var sshIdentityKey string
var sshUser string

func init() {
	forwardCmd.Flags().VarP(&forwardType, "type", "t", "Type of forwarding (e.g., ecs, ssh)")

	// TODO: Existence of this should depend on forwardType
	forwardCmd.Flags().StringVar(&ecsForwardCluster, "cluster", "", "ECS Cluster")
	forwardCmd.Flags().StringVar(&ecsForwardContainer, "container", "", "ECS Container")
	forwardCmd.Flags().Uint64VarP(&forwardLocalPort, "local-port", "l", 0, "Local port to forward")
	forwardCmd.Flags().Uint64VarP(&forwardRemotePort, "remote-port", "r", 0, "Remote port to forward")

	// TODO: Same, but for ssh
	forwardCmd.Flags().StringVar(&sshHost, "host", "", "SSH Host")
	forwardCmd.Flags().StringVarP(&sshIdentityKey, "identity", "i", "", "SSH Identity Key")
	forwardCmd.Flags().StringVarP(&sshUser, "user", "u", os.Getenv("USER"), "SSH User")
}
