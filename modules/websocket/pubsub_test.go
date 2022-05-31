package websocket

import (
	"testing"

	"github.com/zhiting-tech/smartassistant/pkg/rand"
)

// BenchmarkPubSub_ClientSubscribe
// BenchmarkPubSub_ClientSubscribe-12       899982       1945 ns/op     755 B/op      10 allocs/op
func BenchmarkPubSub_ClientSubscribe(b *testing.B) {
	tm := NewPubSub()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		topic := rand.String(10)
		tm.Subscribe(topic, func(msg interface{}) error {
			return nil
		})
	}
}

// BenchmarkPubSub_Publish
// BenchmarkPubSub_Publish-12      2190778      550.7 ns/op     24 B/op      3 allocs/op
func BenchmarkPubSub_Publish(b *testing.B) {
	tm := NewPubSub()

	for i := 0; i < 200000; i++ {
		topic := rand.String(10)
		tm.Subscribe(topic, func(msg interface{}) error {
			return nil
		})
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		topic := rand.String(4)
		tm.Publish(topic, []byte("hello"))
	}
}

// BenchmarkPubSub_Unsubscribe
// BenchmarkPubSub_Unsubscribe-12      2910027      391.5 ns/op      48 B/op       3 allocs/op
func BenchmarkPubSub_Unsubscribe(b *testing.B) {
	tm := NewPubSub()
	subscriber := tm.Subscribe(rand.String(10), func(msg interface{}) error {
		return nil
	})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		subscriber.Unsubscribe()
	}
}
