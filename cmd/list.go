package cmd

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

type PortStatus struct {
	Port   uint
	Status string
}

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all port forwards",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		forwardedPorts := []PortStatus{{Port: 8000, Status: "forwarded"}} // Example data
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetStyle(table.StyleColoredDark)
		t.AppendHeader(table.Row{"Port", "Status"})
		for _, port := range forwardedPorts {
			t.AppendRow(table.Row{port.Port, port.Status})
		}
		t.Render()
		os.Exit(0)
	},
}
