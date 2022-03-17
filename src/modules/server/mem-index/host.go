package mem_index

import (
	"encoding/json"
	"fmt"
	"github.com/go-kit/log"
	ii "github.com/ning1875/inverted-index"
	"github.com/ning1875/inverted-index/labels"
	"open-devops/src/models"
	"strconv"
	"strings"
)

// 准备具体的结构体，实现功能
type HostIndex struct {
	Ir *ii.HeadIndexReader
	Logger log.Logger
	Modules int // 静态分片的模， 总数
	Num int		// 第几个
}

func (hi *HostIndex) FlushIndex()  {
	// 计数
	r := new(models.ResourceHost)
	total := int(r.Count())
	ids := ""
	for i := 0; i < total; i++ {
		// 先写单点逻辑
		if hi.Modules == 0 {
			ids += fmt.Sprintf("%d", i)
			continue
		}
		// 分片匹配中了, keep的逻辑
		if i%hi.Modules == hi.Num{
			ids += fmt.Sprintf("%d", i)
			continue
		}
	}
	ids = strings.TrimRight(ids, ",")
	inSql := fmt.Sprintf("id in (%s)", ids)
	objs,err :=models.ResourceHostGetMany(inSql)
	if err != nil{
		return
	}
	thisH := ii.NewHeadReader()
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

		// json列表型
		json.Unmarshal([]byte(item.PrivateIps), &prIps)
		json.Unmarshal([]byte(item.PublicIps), &puIps)

		// jsonmap形式
		json.Unmarshal([]byte(item.Tags), &tags)

		// g.p.a
		m["stree_group"] = item.StreeGroup
		m["stree_product"] = item.StreeProduct
		m["stree_app"] = item.StreeApp

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


