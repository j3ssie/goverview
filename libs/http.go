package libs

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
	Location       string
	BeautifyHeader string
}
