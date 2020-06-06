package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/j3ssie/goverview/core"
	"github.com/spf13/cobra"
)

var options core.Options
var commands = &cobra.Command{
	Use:  "goverview",
	Long: fmt.Sprintf("goverview - Get overview about list of URLs - %v by %v", core.VERSION, core.AUTHOR),
	Run:  run,
}

func main() {
	commands.Flags().IntVarP(&options.Concurrency, "concurrency", "c", 30, "Set the concurrency level")
	commands.Flags().IntVarP(&options.Threads, "threads", "t", 10, "Set the threads level for do screenshot")
	commands.Flags().IntVarP(&options.Level, "level", "l", 0, "Set level to calculate CheckSum")
	// output
	commands.Flags().StringVarP(&options.Output, "output", "o", "out", "Output Directory")
	commands.Flags().StringVarP(&options.ScreenShotFile, "screenshot", "S", "", "Summary File for Screenshot (default 'out/content-summary.txt')")
	commands.Flags().StringVarP(&options.ContentFile, "content", "C", "", "Summary File for Content (default 'out/screenshot-summary.txt')")
	commands.Flags().StringVarP(&options.WordList, "wordlist", "W", "", "Wordlists File build from HTTP Content (default 'out/wordlists.txt')")
	// mics options
	commands.Flags().BoolVar(&options.SortTag, "sortTag", false, "Sort HTML tag before do checksum")
	commands.Flags().BoolVar(&options.SkipWords, "skip-words", false, "Skip wordlist builder")
	commands.Flags().BoolVarP(&options.SkipScreen, "skip-screen", "Q", false, "Skip screenshot")
	commands.Flags().BoolVar(&options.SkipProbe, "skip-probe", false, "Skip probing for checksum")
	commands.Flags().BoolVarP(&options.SaveReponse, "save-response", "M", false, "Save HTTP response")
	commands.Flags().BoolVarP(&options.InputAsBurp, "burp", "B", false, "Receive input as base64 burp request")
	// screen options
	commands.Flags().BoolVar(&options.AbsPath, "a", false, "Use Absolute path in summary")
	commands.Flags().BoolVarP(&options.Redirect, "redirect", "R", false, "Allow redirect")
	commands.Flags().IntVar(&options.Timeout, "timeout", 15, "screenshot timeout")
	commands.Flags().IntVar(&options.Retry, "retry", 2, "Number of retry")
	commands.Flags().IntVar(&options.ImgHeight, "height", 0, "Height screenshot")
	commands.Flags().IntVar(&options.ImgWidth, "width", 0, "Width screenshot")
	commands.Flags().BoolVarP(&options.Verbose, "verbose", "v", false, "Verbose output")
	commands.Flags().BoolVar(&options.Debug, "debug", false, "Debug output")
	commands.Flags().BoolP("version", "V", false, "Check version")
	commands.SetHelpFunc(HelpMessage)
	commands.Flags().SortFlags = false
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, _ []string) {
	core.InitLog(&options)
	version, _ := cmd.Flags().GetBool("version")
	if version {
		fmt.Printf("Version: %s\n", core.VERSION)
		os.Exit(0)
	}

	// prepare output
	prepareOutput()

	var wg sync.WaitGroup
	jobs := make(chan string, options.Concurrency)

	if !options.SkipProbe {
		// do probing
		core.GoodF("Probing HTTP")
		for i := 0; i < options.Concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for job := range jobs {
					// parsing Burp base64 to a URL
					if options.InputAsBurp {
						job = core.ParseBurpRequest(job)
					}
					if job == "" {
						continue
					}
					core.InforF("[probing] %v", job)
					checksum := core.CalcCheckSum(options, job)
					if checksum != "" {
						core.InforF("[checksum] %v - %v", job, checksum)
						core.AppendTo(options.ContentFile, checksum)
					}

				}
			}()
		}
	}

	if !options.SkipScreen {
		core.GoodF("Do Screenshot")
		// do screenshot
		for i := 0; i < options.Threads; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for job := range jobs {
					core.InforF("[screenshot] %v", job)
					imgScreen := core.DoScreenshot(options, job)
					if imgScreen != "" {
						core.InforF("Store image: %v %v", job, imgScreen)
						sum := fmt.Sprintf("%v ;; %v", job, imgScreen)
						core.AppendTo(options.ScreenShotFile, sum)
					}
				}
			}()
		}
	}

	// parse input from stdin
	sc := bufio.NewScanner(os.Stdin)
	go func() {
		for sc.Scan() {
			url := strings.TrimSpace(sc.Text())
			if err := sc.Err(); err == nil && url != "" {
				jobs <- url
			}
		}
		close(jobs)
	}()
	wg.Wait()

	printOutput()
}

func prepareOutput() {
	// prepare output
	err := os.MkdirAll(options.Output, 0750)
	if err != nil {
		core.ErrorF("Can't create output directory")
	}
	options.ContentOutput = path.Join(options.Output, "contents")
	options.ScreenOutput = path.Join(options.Output, "screenshots")
	os.MkdirAll(options.ContentOutput, 0750)
	if !options.SkipScreen {
		os.MkdirAll(options.ScreenOutput, 0750)
	}

	if options.AbsPath {
		options.Output, _ = filepath.Abs(options.Output)
	}
	if options.ScreenShotFile == "" {
		options.ScreenShotFile = path.Join(options.Output, "screenshot-summary.txt")
	}
	if options.ContentFile == "" {
		options.ContentFile = path.Join(options.Output, "content-summary.txt")
	}
	if options.WordList == "" {
		options.WordList = path.Join(options.Output, "wordlists.txt")
	}
}

func printOutput() {
	// unique the content of wordlist file
	core.CleanWords(options.WordList)

	// print output
	if core.FileExists(options.ContentFile) {
		core.GoodF("Checksum summary in: %v", options.ContentFile)
	}
	if core.FileExists(options.WordList) {
		core.GoodF("Wordlists summary in: %v", options.WordList)
	}
	if core.FileExists(options.ScreenShotFile) {
		core.GoodF("Screenshot summary in: %v", options.ScreenShotFile)
	}
}

// HelpMessage print help message
func HelpMessage(_ *cobra.Command, _ []string) {
	h := fmt.Sprintf("goverview - Overview about list of URLs - %v by %v", core.VERSION, core.AUTHOR)
	h += "\n\nUsage:\n"
	h += "cat <input_file> | goverview [options]\n\n"
	h += `Flags:
  -c, --concurrency int     Set th/e concurrency level (default 30)
  -t, --threads int         Set the threads level for do screenshot (default 10)
  -l, --level int           Set level to calculate CheckSum
  -o, --output string       Output Directory (default "out")
  -S, --screenshot string   Summary File for Screenshot (default 'out/content-summary.txt')
  -C, --content string      Summary File for Content (default 'out/screenshot-summary.txt')
  -W, --wordlist string     Wordlists File build from HTTP Content (default 'out/wordlists.txt')
  -Q, --skip-screen         Skip screenshot
      --skip-probe          Skip probing for checksum
  -M, --save-response       Save HTTP response
      --a                   Use Absolute path in summary
  -R, --redirect            Allow redirect
  -B, --burp	            Receive input as base64 burp request
      --timeout int         screenshot timeout (default 10)
      --retry int           Number of retry
      --height int          Height screenshot
      --width int           Width screenshot
  -v, --verbose             Verbose output
      --debug               Debug output
  -V, --version             Check version
  -h, --help                help for goverview`
	h += "\n\nChecksum Content Level:\n"
	h += "  0 - Only check for src in <script> tag\n"
	h += "  1 - Check for all structure of HTML tag + src in <script> tag\n"
	h += "  2 - Check for all structure of HTML tag + src in <script> <img> <a> tag\n"
	h += "  5 - Entire HTTP response"

	h += "\n\nExamples:\n"
	h += "  cat list_of_urls.txt | goverview -l 1\n"
	h += "  cat list_of_urls.txt | goverview -v -Q -l 1\n"
	h += "  cat list_of_urls.txt | goverview -v -Q -M -l 2\n"
	h += "  cat list_of_urls.txt | goverview --skip-probe -o overview -S overview/target-screen.txt\n"
	h += "\n"
	fmt.Printf(h)
}
