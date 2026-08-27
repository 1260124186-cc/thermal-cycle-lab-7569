package clock

import "time"

type Source interface{ Now() time.Time }
type System struct{}

func (System) Now() time.Time { return time.Now().UTC() }
