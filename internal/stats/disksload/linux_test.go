//go:build linux

package disksload

import (
	"context"
	"testing"

	"github.com/evr-gh/go-sys-monitor-deamon/internal/models"
	"github.com/stretchr/testify/require"
)

func TestGetDisksLoad(t *testing.T) {
	t.Run("should parse iostat output correctly", func(t *testing.T) {
		exec := func(_ context.Context, command string, args []string) (string, error) {
			require.Equal(t, "iostat", command)
			require.Equal(t, []string{""}, args)
			return `Linux 6.1.166-1-generic (alm01)         29.09.2026      _x86_64_        (12 CPU)

Device             tps    kB_read/s    kB_wrtn/s    kB_dscd/s    kB_read    kB_wrtn    kB_dscd
dm-0              0,16         2,99         8,32       574,77     386197    1075256   74317272
dm-1              0,24         2,96         8,32       574,77     382697    1075256   74317272
sda               7,69        38,18       158,01      3633,31    4936362   20430893  469782964
sdb               0,19         3,07         8,32       574,77     396428    1075256   74317272`, nil
		}

		result, err := GetDisksLoad(context.Background(), exec)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.DisksLoad, 4)

		require.Equal(t, "sda", result.DisksLoad[2].Name)
		require.Equal(t, 7.69, result.DisksLoad[2].Tps)
		require.Equal(t, 38.18, result.DisksLoad[2].KpsRead)
		require.Equal(t, 158.01, result.DisksLoad[2].KpsWrite)

		require.Equal(t, "sdb", result.DisksLoad[3].Name)
		require.Equal(t, 0.19, result.DisksLoad[3].Tps)
		require.Equal(t, 3.07, result.DisksLoad[3].KpsRead)
		require.Equal(t, 8.32, result.DisksLoad[3].KpsWrite)
	})
}

func TestGetStats(t *testing.T) {
	t.Run("should return disk stats", func(t *testing.T) {
		result, err := GetStats(context.Background())
		require.NoError(t, err)
		require.NotNil(t, result)
		require.IsType(t, &models.DisksLoad{}, result)
	})
}
