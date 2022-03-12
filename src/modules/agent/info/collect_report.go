package info

import (
	"context"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"open-devops/src/common"
	"open-devops/src/models"
	"open-devops/src/modules/agent/rpc"


	"time"
)

// 间隔器
func TickerInfoCollectAndReport(cli *rpc.RpcCli ,ctx context.Context, logger log.Logger) error {
	ticker := time.NewTicker(5*time.Second)
	level.Info(logger).Log("msg", "TickerInfoCollectAndReport.start")
	CollectBaseInfo(cli, logger)
	defer ticker.Stop()
	for{
		select {
		case <- ctx.Done():
			level.Info(logger).Log("msg", "receive_quit_signal_and_quit")
			return nil
			case <- ticker.C:
				CollectBaseInfo(cli, logger)
		}
	}

}



func CollectBaseInfo(cli *rpc.RpcCli , logger log.Logger)  {
	var (
		sn string
		cpu string
		mem string
		disk string
		err error
	)

	snShellCloud := `curl -s http://169.254.169.254/a/meta-data/instance-id`

	// linux下使用

	//snShellHost := `dmidecode -s system-serial-number | tail -n 1 |tr -d "\n"`
	//cpuShell := `cat /proc/cpuinfo |grep processor |wc -l |tr -d "\n"`
	//memShell := `cat /proc/meminfo | grep MemTotal | awk '{printf "%d", $2/1024/1024}'`

	cpuShell := `sysctl hw.physicalcpu | awk -F: '{printf $2}'`  // mac
	memShell := `sysctl hw.physicalcpu | awk -F: '{printf $2}'`  // mac
	snShellHost := `hostname |tr -d "\n"`

	diskShell := `df -m |grep '/dev/' | grep -v '/var/lib' |grep -v tmpfs |awk '{sum += $2};END{printf "%d", sum/1024}'`

	sn ,err = common.ShellCommand(snShellCloud)
	if err != nil || sn == ""{
		sn, err = common.ShellCommand(snShellHost)
	}
	level.Info(logger).Log("msg", "CollectBaseInfo", "sn", sn)

	cpu, err = common.ShellCommand(cpuShell)
	if err != nil || sn == ""{
		cpu, err = common.ShellCommand(cpuShell)
	}
	level.Info(logger).Log("msg", "CollectBaseInfo", "cpu", cpu)

	mem, err = common.ShellCommand(memShell)
	if err != nil || sn == ""{
		mem, err = common.ShellCommand(memShell)
	}
	level.Info(logger).Log("msg", "CollectBaseInfo", "mem", mem)

	disk, err = common.ShellCommand(diskShell)
	if err != nil || sn == ""{
		disk, err = common.ShellCommand(diskShell)
	}
	level.Info(logger).Log("msg", "CollectBaseInfo", "disk", disk)

	ipAddr := common.GetLocalIp()
	hostName := common.GetHostName()

	hostObj := models.AgentCollectInfo{
		SN:      sn,
		CPU:      cpu,
		Mem:      mem,
		Disk:     disk,
		IpAddr:   ipAddr,
		HostName: hostName,
	}

	cli.HostInfoReport(hostObj)


}


