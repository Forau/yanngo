package swagger

import (
	"time"
)

// Since swagger generates 'Date', we need to define it, and thus do not need to modify generated files
type Date time.Time
