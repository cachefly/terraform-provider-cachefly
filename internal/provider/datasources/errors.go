package datasources

import "strings"

// isNotFoundError reports whether err is an API 404 response. The SDK formats
// API errors as "API error <status>: <body>".
func isNotFoundError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "API error 404")
}
