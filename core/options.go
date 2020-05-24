package core

// Options global options
type Options struct {
	Output         string
	ScreenOutput   string
	ContentOutput  string
	ScreenShotFile string
	CheckSumFile   string
	ContentFile    string
	WordList       string
	LogFile        string
	Concurrency    int
	Threads        int
	Retry          int
	Level          int
	Redirect       bool
	SkipWords      bool
	SkipScreen     bool
	SkipProbe      bool
	SaveReponse    bool
	InputAsBurp    bool
	SortTag        bool
	Verbose        bool
	Debug          bool
	AbsPath        bool
	Timeout        int
	ImgWidth       int
	ImgHeight      int
}

// Request all information about request
type Request struct {
	Timeout  int
	Repeat   int
	Scheme   string
	Host     string
	Port     string
	Path     string
	URL      string
	Proxy    string
	Method   string
	Redirect bool
	Headers  []map[string]string
	Body     string
	Beautify string
}

// Response all information about response
type Response struct {
	HasPopUp       bool
	StatusCode     int
	Status         string
	ContentType    string
	Headers        []map[string]string
	Body           string
	ResponseTime   float64
	Length         int
	Beautify       string
	BeautifyHeader string
}
