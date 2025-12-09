package middleware

import "fmt"

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

func colorMethod(method string) string {
	switch method {
	case "GET":
		return colorBlue + method + colorReset
	case "POST":
		return colorGreen + method + colorReset
	case "PUT":
		return colorYellow + method + colorReset
	case "DELETE":
		return colorRed + method + colorReset
	case "PATCH":
		return colorCyan + method + colorReset
	case "OPTIONS":
		return colorCyan + method + colorReset
	default:
		return method
	}
}

func colorStatus(status int) string {
	statusStr := fmt.Sprintf("%v", status)
	switch {
	case status >= 200 && status < 300:
		return colorGreen + statusStr + colorReset
	case status >= 300 && status < 400:
		return colorYellow + statusStr + colorReset
	case status >= 400 && status < 500:
		return colorRed + statusStr + colorReset
	case status >= 500:
		return colorRed + statusStr + colorReset
	default:
		return statusStr
	}
}
