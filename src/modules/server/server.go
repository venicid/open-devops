package main


// 使用prometheus log 和version注入
import (
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	_ "github.com/go-sql-driver/mysql"
	"github.com/prometheus/common/promlog"
	promlogflag "github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"open-devops/src/models"
	"open-devops/src/modules/server/config"
	"os"
	"path/filepath"
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
	//models.StreePathAddTest(logger)
	//models.StreePathQueryTest1(logger)
	//models.StreePathQueryTest2(logger)
	//models.StreePathQueryTest3(logger)
	models.StreePathDeleteTest(logger)

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