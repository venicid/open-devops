package mem_index

import (
	"encoding/json"
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	ii "github.com/ning1875/inverted-index"
	"github.com/ning1875/inverted-index/labels"
	"open-devops/src/common"
	"open-devops/src/models"
	"strconv"
	"strings"
	"time"
)

// 准备具体的结构体，实现功能
type HostIndex struct {
	Ir *ii.HeadIndexReader
	Logger log.Logger
	Modulus int // 静态分片的模， 总数
	Num int		// 第几个
}

func (hi *HostIndex) FlushIndex()  {
	start := time.Now()

	// 计数
	r := new(models.ResourceHost)
	total := int(r.Count())
	ids := ""
	mine := 0
	for i := 1; i < total+1; i++ {
		// 先写单点逻辑
		if hi.Modulus == 0 {
			ids += fmt.Sprintf("%d,", i)
			mine++
			continue
		}
		// 分片匹配中了, keep的逻辑
		if i%hi.Modulus == hi.Num{
			ids += fmt.Sprintf("%d,", i)
			mine++
			continue
		}

	}
	ids = strings.TrimRight(ids, ",")
	inSql := fmt.Sprintf("id in (%s)", ids)
	level.Info(hi.Logger).Log("msg", "mem-index.FlushIndex.shard",
		"total", total,
		"mine", mine,
		"ids", ids,
		)

	objs,err :=models.ResourceHostGetMany(inSql)
	if err != nil{
		return
	}
	thisH := ii.NewHeadReader()

	// 自动刷node path
	thisGPAs := map[string]struct{}{}

	for _, item := range objs {
		m := make(map[string]string)
		m["hash"] = item.Hash
		tags := make(map[string]string)
		// 数组型，内网ips、公网ips、安全组
		prIps := []string{}
		puIps := []string{}


		// 单个kv
		m["uid"] = item.Uid
		m["name"] = item.Name
		m["cloud_provider"] = item.CloudProvider
		m["charging_mode"] = item.ChargingMode
		m["region"] = item.Region
		m["instance_type"] = item.InstanceType
		m["availability_zone"] = item.AvailabilityZone
		m["vpc_id"] = item.VpcId
		m["subnet_id"] = item.SubnetId
		m["status"] = item.Status
		m["account_id"] = strconv.FormatInt(item.AccountId, 10)

		// cpu mem
		m["cpu"] = item.CPU
		m["mem"] = item.Mem
		m["disk"] = item.Disk

		// g.p.a
		m["stree_group"] = item.StreeGroup
		m["stree_product"] = item.StreeProduct
		m["stree_app"] = item.StreeApp

		thisGPAs[fmt.Sprintf("%s.%s.%s", item.StreeGroup, item.StreeProduct, item.StreeApp)] = struct{}{}
		

		// json列表型
		json.Unmarshal([]byte(item.PrivateIps), &prIps)
		json.Unmarshal([]byte(item.PublicIps), &puIps)

		// jsonmap形式
		json.Unmarshal([]byte(item.Tags), &tags)
		

		// 调用倒排索引库，刷新索引
		thisH.GetOrCreateWithID(uint64(item.Id), item.Hash, mapTolsets(m))
		thisH.GetOrCreateWithID(uint64(item.Id), item.Hash, mapTolsets(tags))

		// 数组型
		for _, i := range prIps {
			mp := map[string]string{
				"private_ips": i,
			}
			thisH.GetOrCreateWithID(uint64(item.Id), item.Hash, mapTolsets(mp))
		}
		for _, i := range puIps {
			mp := map[string]string{
				"public_ips": i,
			}
			thisH.GetOrCreateWithID(uint64(item.Id), item.Hash, mapTolsets(mp))
		}


	}

	hi.Ir.Reset(thisH)

	level.Info(hi.Logger).Log("msg", "mem-index.FlushIndex.time.took",
		"took", time.Since(start).Seconds(),
	)


	// 自动将g.p.a添加到node_path
	go func() {
		level.Info(hi.Logger).Log("msg", "mem-index.FlushIndex.Add.GPA.to.PATH",
			"num", len(thisGPAs),
		)
		for node := range thisGPAs{
			inputs := common.NodeCommonReq{
				Node:        node,
			}
			models.StreePathAddOne(&inputs, hi.Logger)
		}
	}()
	
}

func mapTolsets(m map[string]string) labels.Labels  {
	var lset labels.Labels

	for k, v := range m {
		l := labels.Label{
			Name: k,
			Value: v,
		}
		lset = append(lset, l)
	}
	return lset
}


func (hi *HostIndex) GetIndexReader() *ii.HeadIndexReader  {
	return hi.Ir
}

func (hi *HostIndex) GetLogger() log.Logger {
	return hi.Logger
}


