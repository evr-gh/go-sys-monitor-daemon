package diskstats

import (
	"context"
	"testing"

	"github.com/evr-gh/go-sys-monitor-deamon/internal/models"
	"github.com/stretchr/testify/require"
)

func TestGetDisksLoad(t *testing.T) {
	t.Run("should parse iostat output correctly", func(t *testing.T) {
		count := 0

		exec := func(_ context.Context, command string, args []string) (string, error) {
			count++

			switch count {
			case 1:
				require.Equal(t, "df", command)
				require.Equal(t, []string{
					"-T", "-m", "--exclude-type=tmpfs",
					"--exclude-type=devtmpfs", "--exclude-type=udev",
				}, args)
				return `Файловая система         Тип     1M-блоков Использовано Доступно Использовано% Cмонтировано в
/dev/sda5                ext4       495246        34789   435229            8% /
/dev/sda2                ext2         1007          123      833           13% /boot
/dev/mapper/mpatha-part1 ext4       127435        54715    66202           46% /workspace
/dev/sda1                vfat          572            1      571            1% /boot/efi
/dev/loop1               iso9660      8189         8189        0          100% /workspace/install/installation
/dev/loop0               iso9660     38907        38907        0          100% /workspace/install/extended`, nil
			default:
				require.Equal(t, "df", command)
				require.Equal(t, []string{
					"-T", "-i", "--exclude-type=tmpfs",
					"--exclude-type=devtmpfs", "--exclude-type=udev",
				}, args)
				return `Файловая система         Тип       Iнодов IИспользовано IСвободно IИспользовано% Cмонтировано в
/dev/sda5                ext4    32276480        954681  31321799             3% /
/dev/sda2                ext2       65536           683     64853             2% /boot
/dev/mapper/mpatha-part1 ext4     8323072        184386   8138686             3% /workspace
/dev/sda1                vfat           0             0         0              - /boot/efi
/dev/loop1               iso9660        0             0         0              - /workspace/install/installation
/dev/loop0               iso9660        0             0         0              - /workspace/install/extended`, nil
			}
		}

		result, err := GetDiskStats(context.Background(), exec)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.DiskStats, 6)

		require.Equal(t, "/dev/sda5", result.DiskStats[0].Name)
		require.Equal(t, uint32(8), result.DiskStats[0].MBUsagePct)
		require.Equal(t, uint32(3), result.DiskStats[0].InUsagePct)

		require.Equal(t, "/dev/loop1", result.DiskStats[4].Name)
		require.Equal(t, uint32(100), result.DiskStats[4].MBUsagePct)
		require.Equal(t, uint32(0), result.DiskStats[4].InUsagePct)
	})
}

func TestGetStats(t *testing.T) {
	t.Run("should return disk stats", func(t *testing.T) {
		result, err := GetStats(context.Background())
		require.NoError(t, err)
		require.NotNil(t, result)
		require.IsType(t, &models.DiskStats{}, result)
	})
}
