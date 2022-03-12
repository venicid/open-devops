package main


// 使用prometheus log 和version注入
import (
	"context"
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	_ "github.com/go-sql-driver/mysql"
	"github.com/oklog/run"
	"github.com/prometheus/common/promlog"
	promlogflag "github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"open-devops/src/models"
	"open-devops/src/modules/server/config"
	"open-devops/src/modules/server/rpc"
	"open-devops/src/modules/server/web"
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
	configFile = app.Flag("config.file", "open-devops-server configuration file path").Short('c').Default("open-devops-server.yml").String()

)

func main() {
	// 版本信息
	app.Version(version.Print("open-devops-server"))
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
	level.Info(logger).Log("msg", "config.LoadFile.success", "file.path", *configFile, "content.mysql.num", len(sConfig.MysqlS))

	// 打印mysql配置
	fmt.Println(sConfig.MysqlS[0])

	// 初始化mysql
	/*
	 Error: unknown driver "mysql" (forgotten import?)
	 解决: _ "github.com/go-sql-driver/mysql"
	*/
	models.InitMysql(sConfig.MysqlS)
	level.Info(logger).Log("msg", "load.mysql.success", "db.num",  len(models.DB))


	/*
	测试函数
	*/
	models.StreePathAddTest(logger)
	//models.StreePathQueryTest1(logger)
	//models.StreePathQueryTest2(logger)
	//models.StreePathQueryTest3(logger)
	//models.StreePathDeleteTest(logger)
	//models.StreePathForeceDeleteTest(logger)

	// 测试server资源
	//models.AddResourceHostTest()


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
				for{
					ticker := time.NewTicker(5*time.Second)
					select {
					case <- ctxAll.Done():
						//level.Warn(logger).Log("msg", "我是模块01退出，接收到了cancelAll")
						return nil
					case <- ticker.C:
						//level.Warn(logger).Log("msg", "我是模块01")

					}
				}

			},

			func(err error) {

			},
		)
	}

	{
		// rpc server
		g.Add(
			func() error {
				errChan := make(chan error, 1)
				go func() {
					errChan <- rpc.Start(sConfig.RpcAddr, logger)
					//errChan <- rpc.Start(":8080", logger)
					//errChan <- rpc.Start(":a8080", logger)  // 测试rpc server error
				}()
				select {
				case err := <- errChan:
					level.Error(logger).Log("msg", "rpc server error", "err", err)
					return err

				// 若注释掉，点击stop，进程不响应Done, 未退出
				case <- ctxAll.Done():
					level.Info(logger).Log("msg", "receive_quit_signal_rpc_server_exit")
					return nil
				}
			},
			func(err error) {
				cancelAll()
			})
	}

	{
		// http
		g.Add(
			func() error {
				errChan := make(chan error, 1)
				go func() {
					errChan <- web.StartGin(sConfig.HttpAddr, logger)
				}()
				select {
				case err := <- errChan:
					level.Error(logger).Log("msg", "http server error", "err", err)
					return err

				// 若注释掉，点击stop，进程不响应Done, 未退出
				case <- ctxAll.Done():
					level.Info(logger).Log("msg", "receive_quit_signal_web_server_exit")
					return nil
				}
			},
			func(err error) {
				cancelAll()
			})
	}

	g.Run()

/*
测试：
	启动...
   context.Background.WithCancel

	退出信号...
   level=warn ts=2022-03-05T11:30:28.343+08:00 caller=server.go:126 msg="Receive SIGTERM, exiting gracefully..."
*/

/*
测试：
	启动...
   level=info ts=2022-03-05T11:59:10.755+08:00 caller=server.go:97 msg=load.mysql.success db.num=1
   context.Background.WithCancel

	level=warn ts=2022-03-05T11:59:15.756+08:00 caller=server.go:151 msg=我是模块01
   level=warn ts=2022-03-05T11:59:20.761+08:00 caller=server.go:151 msg=我是模块01

	退出信号...
   level=warn ts=2022-03-05T11:59:23.592+08:00 caller=server.go:126 msg="Receive SIGTERM, exiting gracefully..."
   level=warn ts=2022-03-05T11:59:23.592+08:00 caller=server.go:148 msg=我是模块01退出，接收到了cancelAll


*/

}



/*
1. 交叉编译
 » go build -o serverx_osx

2. 运行
 » ./serverx_osx
open-devops-server.yml
 » ./serverx_osx -c a.yml
a.yml
» ./serverx_osx -c b.yml
b.yml
 » ./serverx_osx --version
1.0
» ./serverx_osx -h
usage: serverx_osx [<flags>]

The open-devops-server

Flags:
  -h, --help     Show context-sensitive help (also try --help-long and --help-man).
  -c, --config.file="open-devops-server.yml"
                 open-devops-server configuration file path
      --version  Show application version.
*/