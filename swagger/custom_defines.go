// Contains generated structs from https://api.test.nordnet.se/next/2/api-docs/swagger
package swagger

import (
	"time"
)

const DateFormat = "2006-01-02"

// Since swagger generates 'Date', we need to define it, and thus do not need to modify generated files
type Date struct {
	time.Time
}

// MarshalJSON implements json.Marshaler interface.
func (d Date) MarshalJSON() ([]byte, error) {
	res := []byte{'"'}
	res = append(res, []byte(d.Format(DateFormat))...)
	return append(res, '"'), nil
}

// UnmarshalJSON implements json.Unmarshaler inferface.
func (d *Date) UnmarshalJSON(buf []byte) (err error) {
	if buf[0] == '"' {
		buf = buf[1:]
	}
	if buf[len(buf)-1] == '"' {
		buf = buf[:len(buf)-1]
	}
	d.Time, err = time.Parse(DateFormat, string(buf))
	return
}
