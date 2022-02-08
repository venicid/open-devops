package models

import (
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"open-devops/src/common"
	"strings"
)

type StreePath struct {
	Id int64 `json:"id"`
	Level int64 `json:"level"`
	Path string `json:"path"`
	NodeName string `json:"node_name"`
}

// 插入一条记录
func (sp *StreePath) AddOne() (int64, error) {
	rowAffect, err := DB["stree"].InsertOne(sp)
	return rowAffect, err
}



// 带部分条件查询一条记录函数
func (sp *StreePath) GetOne() (*StreePath, error) {
	exist, err := DB["stree"].Get(sp)
	if err !=nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}
	return sp, nil
}


// 检查一个记录是否存在
func (sp *StreePath) CheckExist() (bool, error) {
	exist, err := DB["stree"].Exist(sp)
	return exist, err
}

/*
函数区域
*/

// 带参数查询一条记录函数 level=3 and path=/0
func StreePathGet(where string, args ...interface{}) (*StreePath, error)  {
	var obj StreePath
	has, err := DB["stree"].Where(where, args...).Get(&obj)
	if err !=nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return &obj, nil
}

func StreePathAddOne(req *common.NodeCommonReq, logger log.Logger)  {
	// 要求新增是三段式 g.p.a
	res := strings.Split(req.Node, ".")
	if len(res) != 3{
		level.Info(logger).Log("msg", "add.path.invalidate", "path", req.Node)
		return
	}
	// g p a
	g, p, a := res[0], res[1], res[2]

	// 先查g
	nodeG := &StreePath{
		Level:    1,
		Path:     "0",
		NodeName: g,
	}
	dbG, err := nodeG.GetOne()
	if err != nil {
		level.Error(logger).Log("msg", "check.g.failed", "path", req.Node, "err", err)
		return
	}
	// 根据g的查询结果在判断
	switch dbG {
	case nil:
		// 说明g不存在，依次插入g.p.a
		level.Info(logger).Log("msg", "g_not_exist", "path", req.Node)

		// 插入g
		_,err := nodeG.AddOne()
		if err != nil{
			level.Error(logger).Log("msg", "g_not_exist_add_g_failed", "path", req.Node, "err", err)
			return
		}
		level.Info(logger).Log("msg", "g_not_exist_add_g_success", "path", req.Node, "err", err)

		// 插入p
		pathP := fmt.Sprintf("/%d", nodeG.Id)
		nodeP := &StreePath{
			Level:    2,
			Path:     pathP,
			NodeName: p,
		}
		_,err  = nodeP.AddOne()
		if err != nil{
			level.Error(logger).Log("msg", "g_not_exist_add_p_failed", "path", req.Node, "err", err)
			return
		}
		level.Info(logger).Log("msg", "g_not_exist_add_p_success", "path", req.Node, "err", err)


		// 插入a
		pathA := fmt.Sprintf("%s/%d", pathP, nodeP.Id)
		nodeA := &StreePath{
			Level:    3,
			Path:     pathA,
			NodeName: a,
		}
		_,err  = nodeA.AddOne()
		if err != nil{
			level.Error(logger).Log("msg", "g_not_exist_add_a_failed", "path", req.Node, "err", err)
			return
		}
		level.Info(logger).Log("msg", "g_not_exist_add_a_success", "path", req.Node, "err", err)


	default:
		// 说明g存在，再查p
		level.Info(logger).Log("msg", "g_exist", "path", req.Node)

	}


}


// 编写新增node的测试函数
func StreePathAddTest(logger log.Logger)  {
	ns := []string{
		"inf.monitor.thanos",
		"inf.monitor.kafka",
		"waimai.qiangdan.queue",
	}
	for _, n := range ns{
		req := &common.NodeCommonReq{
			Node: n,
		}
		StreePathAddOne(req, logger)
	}
}