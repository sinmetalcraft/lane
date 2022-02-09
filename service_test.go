package lane_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/sinmetalcraft/lane"
)

func TestService_LaneUpAndDone(t *testing.T) {
	ctx := context.Background()

	s, err := lane.NewService(ctx)
	if err != nil {
		t.Fatal(err)
	}

	key := "hoge"
	buzzer, goSign := s.LaneUp(ctx, key, 10*time.Second)
	if buzzer == nil {
		t.Error("buzzer is nil")
	}
	if e, g := true, goSign; e != g {
		t.Errorf("want goSign %t but got %t", e, g)
	}

	receiveCounterCh := make(chan int)
	wg := &sync.WaitGroup{}
	const waitLineCount = 10
	for i := 0; i < waitLineCount; i++ {
		wg.Add(1)
		go func(ctx context.Context, i int) {
			buzzer, goSign := s.LaneUp(ctx, key, 10*time.Second)
			if buzzer == nil {
				t.Error("buzzer is nil")
			}
			if e, g := false, goSign; e != g {
				t.Errorf("want goSign %t but got %t\n", e, g)
			}
			wg.Done() // 待ち行列に並んだことを通達

			select {
			case <-buzzer.GoSign:
				receiveCounterCh <- i
			case <-ctx.Done():
				t.Error("context canceled")
			}
		}(ctx, i)
	}

	wg.Wait() // みんな待ち行列に並んで、待っている状態になった

	// 処理が完了したら、待っているみんなに伝える
	if err := s.Done(ctx, key); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < waitLineCount; i++ {
		v := <-receiveCounterCh
		t.Logf("done: %d\n", v)
	}
}
