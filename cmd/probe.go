package cmd

import (
	"fmt"
	"github.com/j3ssie/goverview/core"
	"github.com/j3ssie/goverview/utils"
	"github.com/panjf2000/ants"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

func init() {
	var probeCmd = &cobra.Command{
		Use:   "probe",
		Short: "Do Probing on target",
		Long:  "Scan config",
		RunE:  runProbe,
	}
	probeCmd.Flags().BoolVarP(&options.SaveReponse, "save-response", "M", false, "Save HTTP response")
	probeCmd.Flags().BoolVarP(&options.Probe.OnlySummary,"no-output", "N", false, "Only store summary file")
	probeCmd.Flags().BoolVar(&options.Probe.WordsSummary,"words",  false, "Get words from ")
	RootCmd.AddCommand(probeCmd)
}

func runProbe(cmd *cobra.Command, _ []string) error {
	// prepare output
	prepareOutput()
	var wg sync.WaitGroup
	client := core.BuildClient(options)
	p, _ := ants.NewPoolWithFunc(options.Concurrency, func(i interface{}) {
		defer wg.Done()
		job := i.(string)

		if strings.TrimSpace(job) == "" {
			return
		}
		if options.InputAsBurp {
			job = core.ParseBurpRequest(job)
		}

		utils.InforF("[probing] %v", job)
		out := core.CalcCheckSum(options, job, client)
		if out != "" {
			fmt.Println(out)
			if options.Probe.OnlySummary || !options.NoOutput{
				core.AppendTo(options.ContentFile, out)
			}
			// only print output but not store it into a file
			//utils.InforF("[checksum] %v - %v", job, out)
		}

	}, ants.WithPreAlloc(true))
	defer p.Release()

	//
	//jobs := make(chan string, options.Concurrency)
	//
	//client := core.BuildClient(options)
	//if !options.SkipProbe {
	//	// do probing
	//	for i := 0; i < options.Concurrency; i++ {
	//		wg.Add(1)
	//		go func() {
	//			defer wg.Done()
	//			for job := range jobs {
	//				// parsing Burp base64 to a URL
	//				if options.InputAsBurp {
	//					job = core.ParseBurpRequest(job)
	//				}
	//				if job == "" {
	//					continue
	//				}
	//				utils.InforF("[probing] %v", job)
	//				out := core.CalcCheckSum(options, job, client)
	//				if out != "" {
	//					// only print output but not store it into a file
	//					if options.NoOutput {
	//						fmt.Println(out)
	//						continue
	//					}
	//
	//					utils.InforF("[checksum] %v - %v", job, out)
	//					core.AppendTo(options.ContentFile, out)
	//				}
	//			}
	//		}()
	//	}
	//}

	for _, raw := range inputs {
		wg.Add(1)
		err := p.Invoke(raw)
		if err != nil {
			utils.ErrorF("Invoke error: %s", err)
		}
	}
	wg.Wait()
	printOutput()
	return nil
}

func prepareOutput() {
	if options.NoOutput && !options.SkipScreen {
		utils.ErrorF("Can't disable output without skip screenshot")
		fmt.Fprintf(os.Stderr, "Command should be: goverview -N -Q ...\n")
		os.Exit(-1)
	}

	if options.NoOutput {
		options.SaveRedirectURL = true
		options.Output = ""
		return
	}

	// prepare output
	err := os.MkdirAll(options.Output, 0750)
	if err != nil {
		utils.ErrorF("Can't create output directory")
	}
	options.ContentOutput = path.Join(options.Output, "contents")
	options.Screen.ScreenOutput = path.Join(options.Output, "screenshots")
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
		utils.GoodF("Checksum summary in: %v", options.ContentFile)
	}
	if core.FileExists(options.WordList) {
		utils.GoodF("Wordlists summary in: %v", options.WordList)
	}
	if utils.EmptyDir(options.ScreenOutput) {
		os.RemoveAll(options.ScreenOutput)
	}
	if core.FileExists(options.ScreenShotFile) {
		utils.GoodF("Screenshot summary in: %v", options.ScreenShotFile)
	}
}
