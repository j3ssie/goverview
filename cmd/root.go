package cmd

import (
	"bufio"
	"fmt"
	"github.com/j3ssie/goverview/core"
	"github.com/j3ssie/goverview/libs"
	"github.com/j3ssie/goverview/utils"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var options = libs.Options{}
var inputs []string
var RootCmd = &cobra.Command{
	Use:   "goverview",
	Short: "goverview",
	Long:  fmt.Sprintf("goverview - Get overview about list of URLs - %v by %v", libs.VERSION, libs.AUTHOR),
}

// Execute main function
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().IntVarP(&options.Concurrency, "concurrency", "c", 10, "Set the concurrency level")
	RootCmd.PersistentFlags().IntVarP(&options.Threads, "threads", "t", 5, "Set the threads level for do screenshot")
	RootCmd.PersistentFlags().IntVarP(&options.Level, "level", "l", 0, "Set level to calculate CheckSum")
	// inputs
	RootCmd.PersistentFlags().StringSliceVarP(&options.Inputs, "inputs", "i", []string{}, "Custom headers (e.g: -H 'Referer: {{.BaseURL}}') (Multiple -H flags are accepted)")
	RootCmd.PersistentFlags().StringVarP(&options.InputFile, "inputFile", "I", "", "Custom headers (e.g: -H 'Referer: {{.BaseURL}}') (Multiple -H flags are accepted)")
	// output
	RootCmd.PersistentFlags().BoolVarP(&options.JsonOutput, "json", "j", false, "Output as JSON")
	RootCmd.PersistentFlags().BoolVarP(&options.NoOutput, "no-output", "N", false, "No output")
	RootCmd.PersistentFlags().StringVarP(&options.Output, "output", "o", "out", "Output Directory")
	RootCmd.PersistentFlags().StringVarP(&options.ScreenShotFile, "screenshot", "S", "", "Summary File for Screenshot (default 'out/screenshot-summary.txt')")
	RootCmd.PersistentFlags().StringVarP(&options.ContentFile, "content", "C", "", "Summary File for Content (default 'out/content-summary.txt')")
	RootCmd.PersistentFlags().StringVarP(&options.WordList, "wordlist", "W", "", "Wordlists File build from HTTP Content (default 'out/wordlists.txt')")
	// mics options
	RootCmd.PersistentFlags().BoolVarP(&options.InputAsBurp, "burp", "B", false, "Receive input as base64 burp request")
	RootCmd.PersistentFlags().BoolVar(&options.SortTag, "sortTag", false, "Sort HTML tag before do checksum")
	// HTTP options
	RootCmd.PersistentFlags().BoolVarP(&options.Redirect, "redirect", "L", false, "Allow redirect")
	RootCmd.PersistentFlags().BoolVarP(&options.SaveRedirectURL, "save-redirect", "R", false, "Save redirect URL to overview file too")
	RootCmd.PersistentFlags().IntVar(&options.Timeout, "timeout", 15, "HTTP timeout")
	RootCmd.PersistentFlags().IntVar(&options.Retry, "retry", 0, "Number of retry")
	RootCmd.PersistentFlags().StringSliceVarP(&options.Headers, "headers", "H", []string{}, "Custom headers (e.g: -H 'Referer: {{.BaseURL}}') (Multiple -H flags are accepted)")

	RootCmd.PersistentFlags().BoolVarP(&options.Verbose, "verbose", "v", false, "Verbose output")
	RootCmd.PersistentFlags().BoolVar(&options.Debug, "debug", false, "Debug output")
	RootCmd.PersistentFlags().BoolP("version", "V", false, "Print version")
	RootCmd.SetHelpFunc(HelpMessage)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	fmt.Fprintf(os.Stderr, "goverview %v by %v\n", libs.VERSION, libs.AUTHOR)
	core.InitConfig(&options)
	utils.InitLog(&options)
	var urls []string
	if len(options.Inputs) > 0 {
		urls = append(urls, options.Inputs...)
	}
	if options.InputFile != "" && utils.FileExists(options.InputFile) {
		urls = append(urls, utils.ReadingLines(options.InputFile)...)
	}

	// input as stdin
	if len(urls) == 0 {
		stat, _ := os.Stdin.Stat()
		// detect if anything came from std
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			utils.DebugF("Reading input from stdin")
			sc := bufio.NewScanner(os.Stdin)
			for sc.Scan() {
				url := strings.TrimSpace(sc.Text())
				if err := sc.Err(); err == nil && url != "" {
					urls = append(urls, url)
				}
			}
			//
			//// store stdin as a temp file
			//urlFile := path.Join(options.TmpDir, fmt.Sprintf("raw-%s", utils.RandomString(8)))
			//os.MkdirAll(options.TmpDir, 0755)
			//utils.DebugF("Write stdin data to: %v", urlFile)
			//utils.WriteToFile(urlFile, strings.Join(urls, "\n"))
		}
	}
	inputs = urls
}

// HelpMessage print help message
func HelpMessage(cmd *cobra.Command, _ []string) {
	h := fmt.Sprintf("goverview - Overview about list of URLs - %v by %v\n\n", libs.VERSION, libs.AUTHOR)
	h += cmd.UsageString()

	h += "\n\nChecksum Content Level:\n"
	h += "  0 - Only check for src in <script> tag\n"
	h += "  1 - Check for all structure of HTML tag + src in <script> tag\n"
	h += "  2 - Check for all structure of HTML tag + src in <script> <img> <a> tag\n"
	h += "  5 - Entire HTTP response"

	h += "\n\nExamples:\n"
	h += "  # Only get summary \n"
	h += "  cat http_lists.txt | goverview probe -N -c 50 | tee only-overview.txt\n\n"
	h += "  # Get summary content and store raw response without screenshot \n"
	h += "  cat http_lists.txt | goverview probe -c 20 -M --json\n\n"
	h += "  # Only do screenshot \n"
	h += "  cat list_of_urls.txt | goverview --skip-probe \n\n"
	h += "  # Do screnshot \n"
	h += "  cat http_lists.txt | goverview screen -c 5 --json\n\n"
	h += "  # Do screnshot based on success HTTP site \n"
	h += "  cat overview/target.com-http-overview.txt | jq -r '. | select(.status==\"200\") | .url' | goverview screen -c 5 -o overview -S overview/target.com-screen.txt\n\n"
	fmt.Printf(h)
}
