package libpdb

import (
	"strconv"
	"testing"
)

func BenchmarkNewTopicHandle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewTopicHandle(strconv.Itoa(i))
	}
}
