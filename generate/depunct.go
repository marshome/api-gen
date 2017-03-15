package generate

import "bytes"
import "strings"
import "unicode"

// depunct removes '-', '.', '$', '/', '_' from identifers, making the
// following character uppercase. Multiple '_' are preserved.
func Depunct(ident string, needCap bool) string {
	var buf bytes.Buffer
	preserve_ := false
	for i, c := range ident {
		if c == '_' {
			if preserve_ || strings.HasPrefix(ident[i:], "__") {
				preserve_ = true
			} else {
				needCap = true
				continue
			}
		} else {
			preserve_ = false
		}
		if c == '-' || c == '.' || c == '$' || c == '/' {
			needCap = true
			continue
		}
		if needCap {
			c = unicode.ToUpper(c)
			needCap = false
		}
		buf.WriteByte(byte(c))
	}
	return buf.String()

}
