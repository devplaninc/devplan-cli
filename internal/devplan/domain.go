package devplan

// GetBaseURL returns the base URL for API calls based on the domain flag
func GetBaseURL(domain string) string {
	switch domain {
	case "beta":
		return "https://beta.devplan.com"
	case "local":
		return "http://localhost:3000"
	default:
		return "https://app.devplan.com"
	}
}
