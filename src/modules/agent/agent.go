package main


// 使用prometheus log 和version注入
import (
	"context"
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/oklog/run"
	"github.com/prometheus/common/promlog"
	promlogflag "github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"open-devops/src/modules/agent/config"
	"open-devops/src/modules/agent/info"
	"open-devops/src/modules/agent/rpc"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var (
	// 命令行解析
	app = kingpin.New(filepath.Base(os.Args[0]), "The open-devops-server")
	// 指定配置文件
	configFile = app.Flag("config.file", "open-devops-agent configuration file path").Short('c').Default("open-devops-agent.yml").String()

)

func main() {
	// 版本信息
	app.Version(version.Print("open-devops-agent"))
	// 帮助信息
	app.HelpFlag.Short('h')

	promlogConfig := promlog.Config{}

	promlogflag.AddFlags(app, &promlogConfig)

	// 强制解析
	kingpin.MustParse(app.Parse(os.Args[1:]))
	fmt.Println(*configFile)

	// 设置logger
	var logger log.Logger
	logger = func(config *promlog.Config) log.Logger {
		var (
			l  log.Logger
			le level.Option
		)
		if config.Format.String() == "logfmt" {
			l = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
		} else {
			l = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
		}

		switch config.Level.String() {
		case "debug":
			le = level.AllowDebug()
		case "info":
			le = level.AllowInfo()
		case "warn":
			le = level.AllowWarn()
		case "error":
			le = level.AllowError()
		}
		l = level.NewFilter(l, le)
		l = log.With(l, "ts", log.TimestampFormat(
			func() time.Time { return time.Now().Local() },
			"2006-01-02T15:04:05.000Z07:00",
		), "caller", log.DefaultCaller)
		return l
	}(&promlogConfig)
	level.Info(logger).Log("msg", "using config.file", "file.path", *configFile)
	level.Debug(logger).Log("debug.msg", "using config.file", "file.path", *configFile)

	// 调用config
	sConfig, err := config.LoadFile(*configFile)
	if err != nil {
		level.Error(logger).Log("msg", "config.LoadFile Error,Exiting ...", "error", err)
		return
	}
	level.Info(logger).Log("msg", "config.LoadFile.success", "file.path", *configFile, "rpc_server_addr", sConfig.RpcServerAddr)


	// 初始化rpc client
	rpcCli := rpc.InitRpcCli(sConfig.RpcServerAddr, logger)
	rpcCli.Ping()

	/**
	编排开始
	*/
	var g run.Group
	ctxAll, cancelAll := context.WithCancel(context.Background())
	fmt.Println(ctxAll)
	{
		// 处理信号退出的handler
		term := make(chan os.Signal, 1)
		signal.Notify(term, os.Interrupt, syscall.SIGTERM)   // 处理ctrl+c , kill -15

		cancelC := make(chan struct{})
		g.Add(
			func() error {
				select {
				case <- term:
					level.Warn(logger).Log("msg", "Receive SIGTERM, exiting gracefully...")
					cancelAll()
					return nil
				case <- cancelC:
					level.Warn(logger).Log("msg", "other cancel exiting")
					return nil
				}
			},

			func(err error) {
				close(cancelC)
			},

		)
	}

	{
		g.Add(
			func() error {
				err := info.TickerInfoCollectAndReport(ctxAll, logger)
				if err != nil{
					level.Error(logger).Log("msg", "TickerInfoCollectAndReport.error", "err", err)
					return err
				}
				return err
		},
			func(err error) {
				cancelAll()
			},
		)
	}

	g.Run()



}



/*
# 启动server1，rpc1
level=info ts=2022-03-06T10:19:32.504+08:00 caller=server.go:98 msg=load.mysql.success db.num=1
context.Background.WithCancel
level=info ts=2022-03-06T10:19:32.504+08:00 caller=rpc.go:29 msg=rpc_server_aviabled_at rpcAddr=:8081
agent01


# 启动rpcClient1

open-devops-agent.yml
level=info ts=2022-03-06T10:19:37.159+08:00 caller=agent.go:76 msg="using config.file" file.path=open-devops-agent.yml
level=info ts=2022-03-06T10:19:37.160+08:00 caller=agent.go:85 msg=config.LoadFile.success file.path=open-devops-agent.yml rpc_server_addr=0.0.0.0:8081
level=info ts=2022-03-06T10:19:37.161+08:00 caller=ping.go:19 msg=Server.Ping.success serverAddr=0.0.0.0:8081 msg=收到了
context.Background.WithCancel
level=warn ts=2022-03-06T10:19:41.802+08:00 caller=agent.go:108 msg="Receive SIGTERM, exiting gracefully..."

*/


/*

linux下使用

打包上传到linux服务器，解压
# 编译
go build -o agent src/modules/agent.go

# 执行
./agent
*/