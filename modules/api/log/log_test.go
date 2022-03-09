package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimeFormatWithTimezone(t *testing.T) {
	t1 := "2022-01-10T10:33:10+08:00"
	t2 := "2022-01-10T02:33:10Z"
	assert.Equal(t, timeFormat(t1), timeFormat(t2))
}
