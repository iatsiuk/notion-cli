package output

import "io"

// Formatter writes data to w in a specific format.
type Formatter interface {
	Format(w io.Writer, data any) error
}
