// +build go1.8

package akita

import "net/url"

// PathUnescape is wraps `url.PathUnescape`
func PathUnescape(s string) (string, error) {
	return url.PathUnescape(s)
}
