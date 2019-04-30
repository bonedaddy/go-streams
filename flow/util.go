package flow

import (
	"sync"

	"github.com/reugn/go-streams"
)

//stream from inlet to outlet
func DoStream(outlet streams.Outlet, inlet streams.Inlet) {
	go func() {
		for elem := range outlet.Out() {
			inlet.In() <- elem
		}
		close(inlet.In())
	}()
}

//splits stream to two flows
//first  - satisfies the condition
//second - doesn't satisfy the condition
func Split(outlet streams.Outlet, cond func(interface{}) bool) [2]streams.Flow {
	condTrue := NewPassThrough()
	condFalse := NewPassThrough()
	go func() {
		for elem := range outlet.Out() {
			if cond(elem) {
				condTrue.In() <- elem
			} else {
				condFalse.In() <- elem
			}
		}
		close(condTrue.In())
		close(condFalse.In())
	}()
	return [...]streams.Flow{condTrue, condFalse}
}

//fans out the stream to magntude number of Flows
func FanOut(outlet streams.Outlet, magnitude int) []streams.Flow {
	out := make([]streams.Flow, magnitude)
	for i := 0; i < magnitude; i++ {
		out = append(out, NewPassThrough())
	}
	go func() {
		for elem := range outlet.Out() {
			for i := 0; i < magnitude; i++ {
				out[i].In() <- elem
			}
		}
		for i := 0; i < magnitude; i++ {
			close(out[i].In())
		}
	}()
	return out
}

//merges outlets to single Flow
func Merge(outlets ...streams.Outlet) streams.Flow {
	merged := NewPassThrough()
	var wg sync.WaitGroup
	wg.Add(len(outlets))
	for _, out := range outlets {
		go func(outlet streams.Outlet) {
			for elem := range outlet.Out() {
				merged.In() <- elem
			}
			wg.Done()
		}(out)
	}
	//close merged.In() on last outlet close
	go func(wg *sync.WaitGroup) {
		wg.Wait()
		close(merged.In())
	}(&wg)
	return merged
}
