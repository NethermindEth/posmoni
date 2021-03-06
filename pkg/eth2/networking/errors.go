package networking

const (
	parseDataError     = "Could not parse event data: %v"
	RequestFailedError = "GET %s failed. Error: %v"
	ReadBodyError      = "read contents of response failed. Error: %v"
	BadResponseError   = "GET %s failed. Status code: %d. Body: %s"
)
