package memorystorage

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/evr-gh/go-sys-monitor-deamon/internal/interfaces"
	"github.com/evr-gh/go-sys-monitor-deamon/internal/logger"
	"github.com/stretchr/testify/require"
)

func TestStorage(t *testing.T) {
	t.Parallel()

	outputInto := &bytes.Buffer{}
	logg := logger.New(interfaces.DEBUG, outputInto)

	AveragingPeriodLimit := int64(1000)

	t.Run("date at", func(t *testing.T) {
		ms := New(logg, "test", AveragingPeriodLimit)
		data := struct{ some string }{some: "some"}
		for i := 0; i < 200; i++ {
			ms.Push(data, time.Now())
		}

		actC := 0
		for range ms.GetElementsSince(time.Now().Add(1 * time.Microsecond)) {
			actC++
		}
		require.Equal(t, 0, actC)
	})
	t.Run("storage size limit", func(t *testing.T) {
		dStart := time.Now()
		tSize := 50

		ms := New(logg, "test", int64(tSize))

		data := struct{ some string }{some: "some"}
		for i := 0; i < 200; i++ {
			ms.Push(data, time.Now())
		}

		actC := 0
		for range ms.GetElementsSince(dStart) {
			actC++
		}

		require.Equal(t, tSize, actC)
	})

	t.Run("storage parallel", func(t *testing.T) {
		t.Parallel()
		dStart := time.Now()
		tSize := 500

		ms1 := New(logg, "test1", int64(tSize))

		ms2 := New(logg, "test2", int64(tSize))

		data := struct{ some string }{some: "some"}

		wg := &sync.WaitGroup{}
		wg.Add(10)

		for w := 0; w < 10; w++ {
			go func() {
				defer wg.Done()
				for i := 0; i < 50; i++ {
					if w == 0 {
						ms1.Push(data, time.Now())
					} else {
						ms2.Push(data, time.Now())
					}
				}
			}()
		}

		wg.Wait()

		actC1 := 0
		for range ms1.GetElementsSince(dStart) {
			actC1++
		}

		actC2 := 0
		for range ms2.GetElementsSince(dStart) {
			actC2++
		}
		require.Equal(t, 50, actC1)
		require.Equal(t, 450, actC2)
	})
}
