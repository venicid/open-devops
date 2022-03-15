package cloudsync

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"math/rand"
	"open-devops/src/common"
	"open-devops/src/models"
	"time"
)

type HostSync struct {
	CloudAliBaba
	CloudTencent
	TableName string
	Logger log.Logger
}

func (this *HostSync) sync()  {
	// 去调用公有云的sdk，取数据，我们使用一个mock方法

	start := time.Now()
	hs := genMockResourceHost()

	// 获取本地的uid对应的hashM
	uuidHashM, err := models.GetHostUidAndHash()
	if err != nil{
		level.Info(this.Logger).Log("msg", "models.GetHostUidAndHash", "err", err)
		return
	}

	// 准备toAddSet toModSet
	// - toAddSet , toModSet存放的都是 host对象
	//- toDelIds 放的是待删的uids
	toAddSet := make([]models.ResourceHost, 0)
	toModSet := make([]models.ResourceHost, 0)
	toDelIds := make([]string, 0)

	localUidSet := make(map[string]struct{})
	var toAddNum, toModNum, toDelNum int
	var suAddNum, suModNum, suDelNum int
	for _, h := range hs{
		localUidSet[h.Uid] = struct{}{}
		hash, ok := uuidHashM[h.Uid]
		if !ok{
			// 说明本地，么有，公有云有，要更新
			toAddSet = append(toAddSet, h)
			toAddNum ++
			continue
		}
		// 存在说明，还有判断hash
		if hash == h.Hash{
			continue
		}
		// 说明uid相同hash不同，某些字段变了，需要更新
		toModSet = append(toModSet, h)
		toModNum ++
	}

	for uid :=range uuidHashM{
		// 说明db中有个uid，远端公有云中没有
		if _,ok :=localUidSet[uid];ok{
			toDelIds = append(toDelIds, uid)
			toDelNum ++
		}
	}

	// 以上是我们的判断流程
	// 下面是执行
	// 新增
	for _,h :=range toAddSet{
		err := h.AddOne()
		if err != nil{
			level.Error(this.Logger).Log("msg", "ResourceHost.AddOne.error", "err", err, "name", h.Name)
			continue
		}
		suAddNum ++
	}

	// 修改
	for _,h :=range toModSet {
		isUpdate, err := h.UpdateByUid(h.Uid)
		if err !=nil{
			level.Error(this.Logger).Log("msg", "ResourceHost.UpdateByUid.error", "err", err, "name", h.Name)
			continue
		}
		if isUpdate{
			suModNum ++
		}
	}

	// 删除
	if len(toDelIds) >0{
		num, _ := models.BatchDeleteResource(common.RESOURCE_HOST, "uid", toDelIds)
		suDelNum= int(num)
	}

	timeTook := time.Since(start)
	level.Info(this.Logger).Log("msg", "ResourceHost.HostSync.res.print",
		"public.cloud.num", len(hs),
		"db.num", len(uuidHashM),

		"toAddNum", toAddNum,
		"toModNum", toModNum,
		"toDelNum", toDelNum,

		"suAddNum", suAddNum,
		"suModNum", suModNum,
		"suDelNum", suDelNum,
		"timeTook", timeTook.Seconds(),
		)



}

func genMockResourceHost() []models.ResourceHost {

	rand.Seed(time.Now().UnixNano())  // 代表随机种子数

	// g.p.a标签
	randGs := []string{"inf", "ads", "web", "sys"}
	randPs := []string{"monitor", "cicd", "k8s", "mq"}
	randAs := []string{"kafaka", "prometheus", "zookeeper", "es"}

	// cpu等资源随机
	randCpus := []string{"4","8", "16", "32", "64", "128"}
	randMems := []string{"8","16","32", "64", "128", "256", "512"}
	randDisks := []string{"128","256", "512", "1024", "2048","4096", "8192"}

	// 标签tags
	randMapKeys := []string{"arch", "idc", "os", "job"}
	randMapValues := []string{"linux", "beijing", "windows", "shanghai", "arm64", "darwin", "shijihulian"}

	// 公有云标签
	randRegions := []string{"beijing", "shanghai", "hangzhou", "guangzhou"}
	randCloudProviders := []string{"alibaba", "huawei", "tencent", "aws"}
	randClusters := []string{"bidata", "inf", "middleware", "business"}
	randInts := []string{"4c8g", "4c16g", "8c32g", "16c64g"}



	// 目的是，4选1，返回0-3的数组
	// 比如8-15， 15-8 +8
	//fr4 := func() int64{
	//	return rand.Int63n(3-0)+0
	//}
	//fr5 := func() int {
	//	return int(rand.Int63n(20-5) + 5)
	//}

	frn := func(n int) int {
		return rand.Intn(n)
	}

	frNum := func() int {
		// todo 随机数有问题
		rand.Seed(time.Now().UnixNano())
		return int(rand.Intn(60-25) +25)
	}

	hs := make([]models.ResourceHost, 0)
	for i := 0; i < frNum(); i++ {
		randN := i
		name := fmt.Sprintf("genMockResourceHost_host_%d", randN)
		ips := []string{fmt.Sprintf("8.8.8.%d", randN)}
		ipJ, _ := json.Marshal(ips)
		h := models.ResourceHost{
			Name:             name,
			PrivateIps:       ipJ,
			CPU:              randCpus[frn(len(randCpus)-1)],
			Mem:              randMems[frn(len(randMems)-1)],
			Disk:             randDisks[frn(len(randDisks)-1)],

			StreeGroup:       randGs[frn(len(randGs)-1)],
			StreeProduct:     randPs[frn(len(randPs)-1)],
			StreeApp:         randAs[frn(len(randAs)-1)],

			Region: randRegions[frn(len(randRegions) -1 )],
			CloudProvider: randCloudProviders[frn(len(randCloudProviders) -1)],
			InstanceType : randInts[frn(len(randInts) -1 )],
		}

		tagM := make(map[string]string)
		for _,i := range randMapKeys{
			tagM[i] = randMapValues[frn(len(randMapValues) -1 )]
		}
		tagM["cluster"] = randClusters[frn(len(randClusters)-1)]

		tarMJ, _ := json.Marshal(tagM)
		h.Tags = tarMJ

		hash := h.GetHash()
		h.Hash = hash

		md5o := md5.New()
		md5o.Write([]byte(name))
		h.Uid = hex.EncodeToString(md5o.Sum(nil))
		hs = append(hs, h)
	}
	return hs

}
