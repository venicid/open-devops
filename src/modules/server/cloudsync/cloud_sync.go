package cloudsync

import (
	"context"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"open-devops/src/common"
	"sync"
	"time"
)

type CloudResource interface {
	sync()
}

type CloudAliBaba struct {

}

type CloudTencent struct {

}


// 接口容器,承载的是多个资源的同步接口
var (
	cloudResourceContainer = make(map[string]CloudResource)
)

// 资源注册
func cRegister(name string, cr CloudResource)  {
	cloudResourceContainer[name] = cr
}

func Init(logger log.Logger)  {
	hs := &HostSync{
		TableName: common.RESOURCE_HOST,
		Logger: logger,
	}
	cRegister(common.RESOURCE_HOST, hs)
}

// 资源管理接口容器的管理端
func CloudSyncManager(ctx context.Context, logger log.Logger) error  {
	level.Info(logger).Log("msg", "CloudSyncManger.start", "resource_num", len(cloudResourceContainer))

	ticker := time.NewTicker(10*time.Second)
	doCloudSync(logger)
	defer ticker.Stop()
	for {
		select {
		case <- ctx.Done():
			level.Info(logger).Log("msg", "CloudSyncManger.exit.receive_quit_signal", "resource_num", len(cloudResourceContainer))
			return nil
		case <- ticker.C:
			level.Info(logger).Log("msg", "CloudSyncManger.cron", "resource_num", len(cloudResourceContainer))
			doCloudSync(logger)
		}
	}
}

func doCloudSync(logger log.Logger)  {
	var wg sync.WaitGroup
	wg.Add(len(cloudResourceContainer))
	for _, sy := range cloudResourceContainer{
		sy := sy
		go func() {
			defer wg.Done()
			sy.sync()
		}()
	}

	wg.Wait()

}