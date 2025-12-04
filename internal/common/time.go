//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-12-04

package common

import (
	"fmt"
	"time"
)

func CostString(d time.Duration) string {
	sec := d.Seconds()
	if sec >= 1 {
		return fmt.Sprintf("%.2fs", sec)
	}
	ms := d.Milliseconds()
	return fmt.Sprintf("%dms", ms)
}
