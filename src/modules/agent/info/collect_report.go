package info

import (
	"context"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"open-devops/src/common"
	"time"
)

// 间隔器
func TickerInfoCollectAndReport(ctx context.Context, logger log.Logger) error {
	ticker := time.NewTicker(5*time.Second)
	level.Info(logger).Log("msg", "TickerInfoCollectAndReport.start")
	CollectBaseInfo(logger)
	defer ticker.Stop()
	for{
		select {
		case <- ctx.Done():
			level.Info(logger).Log("msg", "receive_quit_signal_and_quit")
			return nil
			case <- ticker.C:
				CollectBaseInfo(logger)
		}
	}

}


func CollectBaseInfo(logger log.Logger)  {
	var (
		sn string
	)

	snShellCloud := `curl -s http://169.254.169.254/a/meta-data/instance-id`

	// linux下使用
	snShellHost := `dmidecode -s system-serial-number | tail -n 1`

	sn ,err := common.ShellCommand(snShellCloud)
	if err != nil || sn == ""{
		sn, err = common.ShellCommand(snShellHost)
	}
	level.Info(logger).Log("msg", "CollectBaseInfo", "sn", sn)
}


