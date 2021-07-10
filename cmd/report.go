package cmd

import (
	"github.com/j3ssie/goverview/core"
	"github.com/spf13/cobra"
	"path"
)

func init() {
	var reportCmd = &cobra.Command{
		Use:   "report",
		Short: "Generate HTML Report based on screenshot output",
		RunE:  runReport,
	}
	reportCmd.Flags().StringVar(&options.ReportFile, "report", "report.html", "Report name")
	RootCmd.AddCommand(reportCmd)
}

func runReport(_ *cobra.Command, _ []string) error {
	options.ReportFile = path.Join(options.Output, "report.html")
	core.RenderReport(options)
	return nil
}
