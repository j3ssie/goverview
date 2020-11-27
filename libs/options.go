package libs

// Options global options
type Options struct {
	Output          string
	ScreenOutput    string
	ContentOutput   string
	ScreenShotFile  string
	CheckSumFile    string
	ContentFile     string
	WordList        string
	LogFile         string
	TmpDir          string
	Concurrency     int
	Threads         int
	Headers         []string
	Inputs          []string
	InputFile       string
	Timeout         int
	Retry           int
	Level           int
	NoOutput        bool
	Redirect        bool
	SkipWords       bool
	SkipScreen      bool
	SkipProbe       bool
	SaveReponse     bool
	SaveRedirectURL bool
	InputAsBurp     bool
	SortTag         bool
	JsonOutput      bool
	Verbose         bool
	Debug           bool
	AbsPath         bool
	ScreenTimeout   int
	ImgWidth        int
	ImgHeight       int
	Probe           ProbeOpt
	Screen          ScreenOpt
	Fin             FinOpt
}

// ProbeOpt options for probing
type ProbeOpt struct {
	OnlySummary   bool
	WordsSummary  bool
	ContentOutput string
}

type ScreenOpt struct {
	ScreenOutput  string
	ScreenTimeout int
	ImgWidth      int
	ImgHeight     int
}

type FinOpt struct {
	TechFile string
	Depth    int
	Loaded   bool
}
