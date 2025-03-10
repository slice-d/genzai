package fastlane

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/tidwall/lotsa"
)

type MyType struct {
	hello int
}

type ChanMyType struct{ base ChanPointer }

func (ch *ChanMyType) Send(value *MyType) {
	ch.base.Send(unsafe.Pointer(value))
}
func (ch *ChanMyType) Recv() *MyType {
	return (*MyType)(ch.base.Recv())
}

func TestOrder(t *testing.T) {
	// testing that order is preserved
	type msgT struct{ i, thread int }
	var ch Chan

	N := 1000000
	T := 100
	go func() {
		lotsa.Ops(N, 100, func(i, thread int) {
			ch.Send(&msgT{i, thread})
		})
		ch.Send(nil)
	}()
	// create unique buckets per thread and store each message
	// sequentially in their respective bucket.
	m := make(map[int][]int)
	for {
		v := ch.Recv()
		if v == nil {
			break
		}
		msg := v.(*msgT)
		m[msg.thread] = append(m[msg.thread], msg.i)
	}
	// check that each bucket contains ordered data check for duplicates
	all := make(map[int]bool)
	for thread := 0; thread < T; thread++ {
		b, ok := m[thread]
		if !ok {
			t.Fatal("missing bucket")
		}
		if len(b) != N/T {
			t.Fatal("invalid bucket size")
		}
		h := -1
		for i := 0; i < len(b); i++ {
			if b[i] <= h {
				t.Fatal("out of order")
			}
			h = b[i]
			if all[h] {
				t.Fatal("duplicate value")
			}
			all[h] = true
		}
	}
}

func fixLeft(s string, n int) string {
	return (s + strings.Repeat(" ", n))[:n]
}
func fixRight(s string, n int) string {
	return (strings.Repeat(" ", n) + s)[len(s):]
}

func printResults(key string, N, P int, dur time.Duration) {
	s := fixLeft(key, 12) + " "
	s += fixLeft(fmt.Sprintf("%d ops in %dms", N, int(dur.Seconds()*1000)), 22) + " "
	s += fixRight(fmt.Sprintf("%d/sec", int(float64(N)/dur.Seconds())), 12) + " "
	s += fixRight(fmt.Sprintf("%dns/op", int(dur/time.Duration(N))), 10) + " "
	s += fixRight(fmt.Sprintf("%s %4d producer", (s + strings.Repeat(" ", 100))[:60], P), 14)
	fmt.Printf("%s\n", strings.TrimSpace(s))
}

func TestFastlaneChan(t *testing.T) {
	N := 1000000
	for P := 1; P < 1000; P *= 10 {
		start := time.Now()
		benchmarkFastlaneChan(N, P, false)
		printResults("fastlane", N, P, time.Since(start))
	}
}

func TestGoChanUnbuffered(t *testing.T) {
	N := 1000000
	var start time.Time
	for P := 1; P < 1000; P *= 10 {
		start = time.Now()
		benchmarkGoChan(N, 0, P, false)
		printResults("go-chan(0)", N, P, time.Since(start))
	}
}

func TestGoChan10(t *testing.T) {
	N := 1000000
	var start time.Time
	for P := 1; P < 1000; P *= 10 {
		start = time.Now()
		benchmarkGoChan(N, 10, P, false)
		printResults("go-chan(10)", N, P, time.Since(start))
	}
}

func TestGoChan100(t *testing.T) {
	N := 1000000
	var start time.Time
	for P := 1; P < 1000; P *= 10 {
		start = time.Now()
		benchmarkGoChan(N, 100, P, false)
		printResults("go-chan(100)", N, P, time.Since(start))
	}
}

func benchmarkFastlaneChan(N int, P int, validate bool) {
	var ch ChanUint64
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for i := 0; i < N; i++ {
			v := ch.Recv()
			if validate {
				if v != uint64(i) {
					panic("out of order")
				}
			}
		}
		wg.Done()
	}()
	lotsa.Ops(N, P, func(i, _ int) {
		ch.Send(uint64(i))
	})
	wg.Wait()
}

func benchmarkGoChan(N, buffered int, producers int, validate bool) {
	ch := make(chan uint64, buffered)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for i := 0; i < N; i++ {
			v := <-ch
			if validate {
				if v != uint64(i) {
					panic("out of order")
				}
			}
		}
		wg.Done()
	}()
	lotsa.Ops(N, producers, func(i, _ int) {
		ch <- uint64(i)
	})
	wg.Wait()
}

func Benchmark100ProducerFastlaneChan(b *testing.B) {
	b.ReportAllocs()
	benchmarkFastlaneChan(b.N, 100, false)
}

func Benchmark100ProducerGoChan100(b *testing.B) {
	//b.ReportAllocs()
	benchmarkGoChan(b.N, 100, 100, false)
}

func Benchmark100ProducerGoChan10(b *testing.B) {
	//b.ReportAllocs()
	benchmarkGoChan(b.N, 10, 100, false)
}

func Benchmark100ProducerGoChanUnbuffered(b *testing.B) {
	//b.ReportAllocs()
	benchmarkGoChan(b.N, 0, 100, false)
}