package statistics

import (
	"context"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"open-devops/src/common"
	"open-devops/src/models"
	mem_index "open-devops/src/modules/server/mem-index"
	"open-devops/src/modules/server/metric"
	"strconv"
	"strings"
	"time"
)

func TreeNodeStatisticsManager(ctx context.Context, logger log.Logger) error  {
	level.Info(logger).Log("msg", "TreeNodeStatisticsManager.start")

	ticker := time.NewTicker(15*time.Second)
	defer ticker.Stop()
	for {
		select {
		case <- ctx.Done():
			level.Info(logger).Log("msg", "TreeNodeStatisticsManager.exit.receive_quit_signal")
			return nil
		case <- ticker.C:
			level.Debug(logger).Log("msg", "CloudSyncManger.cron")
			statisticsWork(logger)
		}
	}

}

func statisticsWork(logger log.Logger)  {

	irs := mem_index.GetAllResourceIndexReader()
	level.Info(logger).Log("msg", "statisticsWork.start.num", "num", len(irs))

	// 获取所有g.p.a列表
	qReq := &common.NodeCommonReq{
		QueryType:   5,
	}
	allGPAS := models.StreePathQuery(qReq, logger)

	for resourceType, ir := range irs {
		resourceType := resourceType

		ir := ir
		go func() {
			// 全局的
			// 按region的分布
			s := ir.GetIndexReader().GetGroupByLabel(common.LABEL_REGION)
			for _, i := range s.Group {
				metric.ResourceNumRegionCount.With(
					prometheus.Labels{
						common.LABEL_RESOURCE_TYPE:resourceType,
						common.LABEL_REGION: i.Name}).Set(float64(i.Value))
			}

			// 按照CloudProvider的分布
			p:= ir.GetIndexReader().GetGroupByLabel(common.LABEL_CLOUD_PROVIDER)
			for _, i := range p.Group {
				metric.ResourceNumCloudProviderCount.With(
					prometheus.Labels{
						common.LABEL_RESOURCE_TYPE:resourceType,
						common.LABEL_CLOUD_PROVIDER: i.Name}).Set(float64(i.Value))
			}

			// 单个gpa
			for _, gpa := range allGPAS{
				ss := strings.Split(gpa, ".")
				if len(ss) != 3{
					continue
				}
				g := ss[0]
				p := ss[1]
				a := ss[2]
				csG := &common.SingleTagReq{
					Key:   common.LABEL_STREE_G,
					Value: g,
					Type:  1,
				}
				csP := &common.SingleTagReq{
					Key:   common.LABEL_STREE_P,
					Value: p,
					Type:  1,
				}
				csA := &common.SingleTagReq{
					Key:   common.LABEL_STREE_A,
					Value: a,
					Type:  1,
				}

				matcherG :=[]*common.SingleTagReq{
					csG,
				}

				matcherGP :=[]*common.SingleTagReq{
					csG,
					csP,
				}
				matcherGPA :=[]*common.SingleTagReq{
					csG,
					csP,
					csA,
				}

				gpaNumbWork(resourceType, g, matcherG, metric.GPAAllNumCount)
				gpaNumbWork(resourceType, g + "." + p, matcherGP, metric.GPAAllNumCount)
				gpaNumbWork(resourceType, g + "." + p +"." +a, matcherGPA, metric.GPAAllNumCount)


				// 这是g的，按照不同标签的分布
				gpaLabelNumWork(resourceType, common.LABEL_REGION, g, matcherG, ir, metric.GPAAllNumRegionCount)
				gpaLabelNumWork(resourceType, common.LABEL_CLOUD_PROVIDER, g, matcherG, ir, metric.GPAAllNumCloudProviderCount)
				gpaLabelNumWork(resourceType, common.LABEL_CLUSTER, g, matcherG, ir, metric.GPAAllNumClusterCount)
				gpaLabelNumWork(resourceType, common.LABEL_INSTANCE_TYPE, g, matcherG, ir, metric.GPAAllNumInstanceTypeCount)

				// 这是g.p的，按照不同标签的分布
				gpaLabelNumWork(resourceType, common.LABEL_REGION, g + "." + p, matcherGP, ir, metric.GPAAllNumRegionCount)
				gpaLabelNumWork(resourceType, common.LABEL_CLOUD_PROVIDER, g + "." + p, matcherGP, ir, metric.GPAAllNumCloudProviderCount)
				gpaLabelNumWork(resourceType, common.LABEL_CLUSTER, g + "." + p, matcherGP, ir, metric.GPAAllNumClusterCount)
				gpaLabelNumWork(resourceType, common.LABEL_INSTANCE_TYPE, g + "." + p, matcherGP, ir, metric.GPAAllNumInstanceTypeCount)

				// 这是g.p.a的，按照不同标签的分布
				gpaLabelNumWork(resourceType, common.LABEL_REGION, g + "." + p +"." +a, matcherGPA, ir, metric.GPAAllNumRegionCount)
				gpaLabelNumWork(resourceType, common.LABEL_CLOUD_PROVIDER, g + "." + p +"." +a, matcherGPA, ir, metric.GPAAllNumCloudProviderCount)
				gpaLabelNumWork(resourceType, common.LABEL_CLUSTER, g + "." + p +"." +a, matcherGPA, ir, metric.GPAAllNumClusterCount)
				gpaLabelNumWork(resourceType, common.LABEL_INSTANCE_TYPE, g + "." + p +"." +a, matcherGPA, ir, metric.GPAAllNumInstanceTypeCount)


				if resourceType == common.RESOURCE_HOST {
					// 这是g的
					hostSpecial(resourceType, common.LABEL_CPU, g, matcherG, ir, metric.GPAHostCpuCores)
					hostSpecial(resourceType, common.LABEL_MEM, g, matcherG, ir, metric.GPAHostMemGbs)
					hostSpecial(resourceType, common.LABEL_DISK, g, matcherG, ir, metric.GPAHostDiskGbs)

					// 这是g.p的
					hostSpecial(resourceType, common.LABEL_CPU, g+"."+p, matcherGP, ir, metric.GPAHostCpuCores)
					hostSpecial(resourceType, common.LABEL_MEM, g+"."+p, matcherGP, ir, metric.GPAHostMemGbs)
					hostSpecial(resourceType, common.LABEL_DISK, g+"."+p, matcherGP, ir, metric.GPAHostDiskGbs)

					// 这是g.p.a的
					hostSpecial(resourceType, common.LABEL_CPU, g+"."+p+"."+a, matcherGPA, ir, metric.GPAHostCpuCores)
					hostSpecial(resourceType, common.LABEL_MEM, g+"."+p+"."+a, matcherGPA, ir, metric.GPAHostMemGbs)
					hostSpecial(resourceType, common.LABEL_DISK, g+"."+p+"."+a, matcherGPA, ir, metric.GPAHostDiskGbs)

				}
			}


		}()
	}

}

// 通过索引的 GetGroupDistributionByLabel接口获取个数分布
//  每个g.p.a在每种资源上 目标标签分布情况
func gpaLabelNumWork(resourceType string, targetLabel string, gpaName string,
	matcher []*common.SingleTagReq, ir mem_index.ResourceIndexer,
	ms *prometheus.GaugeVec)  {

	req := common.ResourceQueryReq{
		ResourceType: resourceType,
		Labels:       matcher,
		TargetLabel:  targetLabel,
	}
	matchIds := mem_index.GetMatchIdsByIndex(req)
	statsRs := ir.GetIndexReader().GetGroupDistributionByLabel(req.TargetLabel, matchIds)
	for _, x := range statsRs.Group {
		ms.With(prometheus.Labels{
			common.LABEL_GPA_NAME: gpaName,
			common.LABEL_RESOURCE_TYPE: resourceType,
			targetLabel: x.Name,
		}).Set(float64(x.Value))

	}


}

// 通过索引的 GetGroupByLabel接口获取个数分布
// 每个g.p.a在每种资源上的计数统计
func gpaNumbWork(resourceType string,  gpaName string,
	matcher []*common.SingleTagReq,
	ms *prometheus.GaugeVec)  {

	req:=common.ResourceQueryReq{
		ResourceType: resourceType,
		Labels:       matcher,
	}

	matchIds := mem_index.GetMatchIdsByIndex(req)
	if len(matchIds) >0 {
		ms.With(prometheus.Labels{
			common.LABEL_GPA_NAME: gpaName,
			common.LABEL_RESOURCE_TYPE: resourceType,
		}).Set(float64(len(matchIds)))
	}

	}


	// host特殊的
func hostSpecial(resourceType string, targetLabel string, gpaName string, matcher []*common.SingleTagReq, ir mem_index.ResourceIndexer, ms *prometheus.GaugeVec) {
	req := common.ResourceQueryReq{
		ResourceType: resourceType,
		Labels:       matcher,
		TargetLabel:  targetLabel,
	}

	matchIds := mem_index.GetMatchIdsByIndex(req)
	statsRe := ir.GetIndexReader().GetGroupDistributionByLabel(targetLabel, matchIds)
	var all uint64
	for _, x := range statsRe.Group {
		num, _ := strconv.Atoi(x.Name)
		all += uint64(num) * x.Value
	}
	if all > 0 {
		ms.With(prometheus.Labels{common.LABEL_GPA_NAME: gpaName}).Set(float64(all))
	}

}


/**

gpa_all_region_num_count{gpa_name="ads",region="beijing",resource_type="resource_host"} 5
gpa_all_region_num_count{gpa_name="ads",region="hangzhou",resource_type="resource_host"} 4
gpa_all_region_num_count{gpa_name="ads",region="shanghai",resource_type="resource_host"} 5
gpa_all_region_num_count{gpa_name="ads.cicd",region="beijing",resource_type="resource_host"} 1
gpa_all_region_num_count{gpa_name="ads.cicd",region="hangzhou",resource_type="resource_host"} 1
gpa_all_region_num_count{gpa_name="ads.cicd",region="shanghai",resource_type="resource_host"} 2
gpa_all_region_num_count{gpa_name="ads.cicd.kafaka",region="shanghai",resource_type="resource_host"} 2
gpa_all_region_num_count{gpa_name="ads.cicd.zookeeper",region="beijing",resource_type="resource_host"} 1
gpa_all_region_num_count{gpa_name="ads.cicd.zookeeper",region="hangzhou",resource_type="resource_host"} 1
gpa_all_region_num_count{gpa_name="ads.k8s",region="beijing",resource_type="resource_host"} 4
gpa_all_region_num_count{gpa_name="ads.k8s.kafaka",region="beijing",resource_type="resource_host"} 1
gpa_all_region_num_count{gpa_name="ads.k8s.prometheus",region="beijing",resource_type="resource_host"} 1
gpa_all_region_num_count{gpa_name="ads.k8s.zookeeper",region="beijing",resource_type="resource_host"} 2
gpa_all_region_num_count{gpa_name="ads.monitor",region="hangzhou",resource_type="resource_host"} 3
gpa_all_region_num_count{gpa_name="ads.monitor",region="shanghai",resource_type="resource_host"} 3

 */


/*
gpa_host_cpu_cores{gpa_name="ads"} 356
gpa_host_cpu_cores{gpa_name="ads.cicd"} 92
gpa_host_cpu_cores{gpa_name="ads.cicd.kafaka"} 24
gpa_host_cpu_cores{gpa_name="ads.cicd.zookeeper"} 68
gpa_host_cpu_cores{gpa_name="ads.k8s"} 108
gpa_host_cpu_cores{gpa_name="ads.k8s.kafaka"} 32
gpa_host_cpu_cores{gpa_name="ads.k8s.prometheus"} 8
gpa_host_cpu_cores{gpa_name="ads.k8s.zookeeper"} 68
gpa_host_cpu_cores{gpa_name="ads.monitor"} 156
gpa_host_cpu_cores{gpa_name="ads.monitor.kafaka"} 96
gpa_host_cpu_cores{gpa_name="ads.monitor.prometheus"} 4
gpa_host_cpu_cores{gpa_name="ads.monitor.zookeeper"} 56
gpa_host_cpu_cores{gpa_name="inf"} 472
gpa_host_cpu_cores{gpa_name="inf.cicd"} 164
gpa_host_cpu_cores{gpa_name="inf.cicd.kafaka"} 32
 */