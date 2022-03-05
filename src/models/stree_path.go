package models

import (
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"open-devops/src/common"
	"sort"
	"strings"
)

type StreePath struct {
	Id       int64  `json:"id"`
	Level    int64  `json:"level"`
	Path     string `json:"path"`
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
	if err != nil {
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

// 删除一条记录
func (sp *StreePath) DelOne() (int64, error){
	delNum, err := DB["stree"].Delete(sp)
	return delNum, err
}


// 函数区域
// 带参数查询一条记录函数 level=3 and path=/0
func StreePathGet(where string, args ...interface{}) (*StreePath, error) {
	var obj StreePath
	has, err := DB["stree"].Where(where, args...).Get(&obj)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return &obj, nil
}

// 带参数查询多条记录函数
func StreePathGetMany(where string, args ...interface{}) ([]StreePath, error) {
	var objs []StreePath
	err := DB["stree"].Where(where, args...).Find(&objs)
	if err != nil {
		return objs, err
	}

	return objs, nil
}

// 带参数删除多条记录函数
func StreePathDelMany(where string) (int64, error)  {
	rawSql := fmt.Sprintf(`delete from stree_path where %s`, where)
	res, err := DB["stree"].Exec(rawSql)
	if err != nil{
		return 0, err
	}
	rowsAffected, err := res.RowsAffected()
	return rowsAffected, err

}

// 查询数据的主函数
func StreePathQuery(req *common.NodeCommonReq, logger log.Logger) (res []string) {
	switch req.QueryType {
	case 1:
		// 根据g, 查询所有p列表 node=g query_type=1
		nodeG := &StreePath{
			Level:    1,
			Path:     "0",
			NodeName: req.Node,
		}
		dbG, err := nodeG.GetOne()
		if err != nil {
			level.Error(logger).Log("msg", "check.g.failed", "path", req.Node, "err", err)
			return
		}
		// 说明要查询的g不存在
		if dbG == nil{
			return
		}
		pathP := fmt.Sprintf("/%d", dbG.Id)
		whereStr := "level=? and path=?"
		ps, err := StreePathGetMany(whereStr, 2, pathP)
		if err != nil {
			level.Error(logger).Log("msg", "query_ps_failed", "path", req.Node, "err", err)
			return
		}
		for _,i :=range ps{
			res = append(res, i.NodeName)
		}
		sort.Strings(res)
		return
	case 2:
		/*
		编写query_type=2的查询，根据g查询，所有g.p.a的列表
		先查g，再查p，最后查a，中间一步都没有返回空
		*/
		// 根据g, 查询所有p列表 node=g query_type=1
		nodeG := &StreePath{
			Level:    1,
			Path:     "0",
			NodeName: req.Node,
		}
		dbG, err := nodeG.GetOne()
		if err != nil {
			level.Error(logger).Log("msg", "check.g.failed", "path", req.Node, "err", err)
			return
		}
		// 说明要查询的g不存在
		if dbG == nil{
			return
		}
		pathP := fmt.Sprintf("/%d", dbG.Id)
		whereStr := "level=? and path=?"
		ps, err := StreePathGetMany(whereStr, 2, pathP)
		if err != nil {
			level.Error(logger).Log("msg", "query_ps_failed", "path", req.Node, "err", err)
			return
		}

		if len(ps) == 0{
			// 说明p下没有a
			return
		}

		for _, p := range ps {
			pathA := fmt.Sprintf("%s/%d", p.Path, p.Id)
			as, err := StreePathGetMany(whereStr, 3, pathA)
			if err != nil{
				level.Error(logger).Log("msg", "query_as_failed", "path", req.Node, "err", err)
				continue
			}
			if len(as) == 0{
				continue
			}
			for _, a := range as {
				fullPath := fmt.Sprintf("%s.%s.%s", dbG.NodeName, p.NodeName, a.NodeName)
				res = append(res, fullPath)
			}
		}
		return
	case 3:
		/*
			编写query_type=3的查询 根据g.p查询 所有g.p.a的列表 node=g.p query_type=3
			先查询 g 和p，不存在直接返回空
			查p时需要带上p.name查询
		*/
		gps := strings.Split(req.Node, ".")
		g,p := gps[0], gps[1]
		// 根据g, 查询所有p列表 node=g query_type=1
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
		// 说明要查询的g不存在
		if dbG == nil{
			return
		}

		// g存在，不需要查询全量的p，只查匹配这个node_name的p
		pathP := fmt.Sprintf("/%d", dbG.Id)
		whereStr := "level=? and path=? and node_name=?"
		dbP, err := StreePathGet(whereStr, 2, pathP, p)
		if err != nil {
			level.Error(logger).Log("msg", "query_p_failed", "path", req.Node, "err", err)
			return
		}
		if dbP == nil{
			// 说明p不存在
			return
		}
		pathA := fmt.Sprintf("%s/%d", pathP, dbP.Id)
		whereStr  = "level=? and path=?"
		as, err := StreePathGetMany(whereStr, 3, pathA)
		if err != nil{
			level.Error(logger).Log("msg", "query_as_failed", "path", req.Node, "err", err)
			return
		}

		for _, a := range as {
			fullPath := fmt.Sprintf("%s.%s.%s", dbG.NodeName, dbP.NodeName, a.NodeName)
			res = append(res, fullPath)
		}
		sort.Strings(res)
		return
	}
	return
}

/*
Node操作
*/


func StreePathAddOne(req *common.NodeCommonReq, logger log.Logger) {
	// 要求新增是三段式 g.p.a
	res := strings.Split(req.Node, ".")
	if len(res) != 3 {
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
		_, err := nodeG.AddOne()
		if err != nil {
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
		_, err = nodeP.AddOne()
		if err != nil {
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
		_, err = nodeA.AddOne()
		if err != nil {
			level.Error(logger).Log("msg", "g_not_exist_add_a_failed", "path", req.Node, "err", err)
			return
		}
		level.Info(logger).Log("msg", "g_not_exist_add_a_success", "path", req.Node, "err", err)

	default:
		level.Info(logger).Log("msg", "g_exist", "path", req.Node)

		// 1 - 说明g存在，再查p
		pathP := fmt.Sprintf("/%d", dbG.Id)
		nodeP := &StreePath{
			Level:    2,
			Path:     pathP,
			NodeName: p,
		}
		dbP, err := nodeP.GetOne()
		if err != nil {
			level.Error(logger).Log("msg", "g_exist_check_p_failed", "path", req.Node, "err", err)
			return
		}
		level.Info(logger).Log("msg", "g_exist_check_p_success", "path", req.Node, "err", err)

		// 1.1 - 说明p存在，继续查a
		if dbP != nil {
			pathA := fmt.Sprintf("%s/%d", pathP, dbP.Id)
			nodeA := &StreePath{
				Level:    3,
				Path:     pathA,
				NodeName: a,
			}
			dbA, err := nodeA.GetOne()
			if err != nil {
				level.Error(logger).Log("msg", "g_p_exist_check_a_failed", "path", req.Node, "err", err)
				return
			}
			level.Info(logger).Log("msg", "g_p_exist_check_a_success", "path", req.Node)

			// 1.1.1 - 说明a，不存在，插入a
			if dbA == nil {
				_, err := nodeA.AddOne()
				if err != nil {
					level.Error(logger).Log("msg", "g_p_exist_add_a_failed", "path", req.Node, "err", err)
					return
				}
				level.Info(logger).Log("msg", "g_p_exist_add_a_success", "path", req.Node, "err", err)
				return
			}
			level.Info(logger).Log("msg", "g_p_a_exist", "path", req.Node)
			return
		}

		// 1.2 - p不存在，插入p和a
		level.Info(logger).Log("msg", "g_exist_p_a_not", "path", req.Node)
		_,err = nodeP.AddOne()
		if err != nil{
			level.Error(logger).Log("msg", "g_exist_add_p_failed", "path", req.Node, "err", err)
			return
		}
		level.Info(logger).Log("msg", "g_exist_add_p_success", "path", req.Node)

		// 1.2.1 - 插入a
		pathA := fmt.Sprintf("%s/%d", pathP, nodeP.Id)
		nodeA := &StreePath{
			Level:    3,
			Path:     pathA,
			NodeName: a,
		}
		_, err = nodeA.AddOne()
		if err != nil{
			level.Error(logger).Log("msg", "g_exist_add_a_failed", "path", req.Node, "err", err)
			return
		}
		level.Info(logger).Log("msg", "g_exist_add_a_success", "path", req.Node)
	}

}

func StreePathDelete(req *common.NodeCommonReq, logger log.Logger) (delNum int64){
	path := strings.Split(req.Node, ".") // inf.cicd
	pLevel := len(path)

	// 每次都要先校验g是否存在
	nodeG := &StreePath{
		Level:    1,
		Path:     "0",
		NodeName: path[0],
	}
	dbG, err := nodeG.GetOne()
	if err != nil {
		level.Error(logger).Log("msg", "query_g_failed", "path", req.Node, "err", err)
		return
	}
	if dbG == nil {
		return
	}
	pathP := fmt.Sprintf("/%d", dbG.Id)

	switch pLevel {

	// 传入g，如果g下有p就不让删g
	case 1:

		// 通过sql，强制删除g下的所有p和a
		if req.ForceDelete{
			delAWhereStr := fmt.Sprintf("path like '/%d/%%' and level = 3 ", dbG.Id)
			delANum, err := StreePathDelMany(delAWhereStr)
			if err != nil{
				level.Error(logger).Log("msg", "delete_pa_failed", "path", req.Node, "err", err)
				return
			}
			level.Info(logger).Log("msg", "delete_pa_success", "path", req.Node, "num", delNum, "del_where", delAWhereStr)
			delNum += delANum

			delPWhereStr := fmt.Sprintf("path like '/%d' and level = 2 ", dbG.Id)
			delPNum, err := StreePathDelMany(delPWhereStr)
			if err != nil{
				level.Error(logger).Log("msg", "delete_pa_failed", "path", req.Node, "err", err)
				return
			}
			level.Info(logger).Log("msg", "delete_pa_success", "path", req.Node, "num", delNum, "del_where", delPWhereStr)
			delNum += delPNum

			_, err = dbG.DelOne()
			if err != nil{
				level.Error(logger).Log("msg", "delete_g_failed", "path", req.Node, "err", err)
				return
			}
			level.Info(logger).Log("msg", "delete_g_success", "path", req.Node)
			delNum += 1
			return
		}

		// 查询p，
		whereStr := "level=? and path=?"
		ps, err := StreePathGetMany(whereStr, 2, pathP)
		if err != nil {
			level.Error(logger).Log("msg", "query_ps_failed", "path", req.Node, "err", err)
			return
		}

		// if p存在，不让删除
		if len(ps) > 0{
			level.Warn(logger).Log("msg", "del_g_reject", "path", req.Node, "reason", "g_has_ps", "ps_num", len(ps))
			return
		}
		// else 删除g
		delNum, err = dbG.DelOne()
		if err != nil {
			level.Error(logger).Log("msg", "del_g_failed", "path", req.Node, "err", err)
			return
		}
		level.Info(logger).Log("msg", "del_g_success", "path", req.Node)
		return

	// 传入g.p，如果p下有a就不让删除p
	case 2:
		// 强制删除p下的a,和p
		// inf.cicd
		if req.ForceDelete{
			// 查询p是否存在
			nodeP := &StreePath{
				Level:    2,
				Path:     pathP,  //  /90
				NodeName: path[1], // cicd
			}
			dbP, err := nodeP.GetOne()
			if err != nil{
				level.Error(logger).Log("msg", "query_p_failed", "path", req.Node, "err", err)
				return
			}
			if dbP == nil{
				return
			}

			// 删除p下的所有a
			delAWhereStr := fmt.Sprintf("path = '%v/%d' and level = 3 ", dbP.Path, dbP.Id )
			delANum, err := StreePathDelMany(delAWhereStr)
			if err != nil{
				level.Error(logger).Log("msg", "delete_a_failed", "path", req.Node, "err", err)
				return
			}
			level.Info(logger).Log("msg", "delete_a_success", "path", req.Node, "num", delANum, "del_where", delAWhereStr)
			delNum += delANum

			// 删除p
			delPNum, err := dbP.DelOne()
			if err != nil{
				level.Error(logger).Log("msg", "del_p_failed", "path", req.Node, "err", err)
				return
			}
			level.Info(logger).Log("msg", "del_p_success", "path", req.Node)
			delNum += delPNum

			return
		}

		// 查询p是否存在
		nodeP := &StreePath{
			Level:    2,
			Path:     pathP,
			NodeName: path[1],
		}
		dbP, err := nodeP.GetOne()
		if err != nil{
			level.Error(logger).Log("msg", "query_p_failed", "path", req.Node, "err", err)
			return
		}
		if dbP == nil{
			return
		}



		// 查询p下的a
		pathA :=fmt.Sprintf("%s/%d", dbP.Path, dbP.Id)
		whereStr := "level=? and path=?"
		as, err := StreePathGetMany(whereStr, 3, pathA)
		if err != nil {
			level.Error(logger).Log("msg", "query_as_failed", "path", req.Node, "err", err)
			return
		}

		// if g.p下存在很多a，拒绝删除
		if len(as) > 0{
			level.Warn(logger).Log("msg", "del_g_p_reject", "path", req.Node, "reason", "p_has_a", "as_num", len(as))
			return
		}
		// else 删除p成功
		delNum, err = dbP.DelOne()
		if err != nil{
			level.Error(logger).Log("msg", "del_p_failed", "path", req.Node, "err", err)
			return
		}
		level.Info(logger).Log("msg", "del_p_success", "path", req.Node)
		return

	// 传入g.p.a删除
	case 3:
		// 查询p是否存在
		nodeP := &StreePath{
			Level:    2,
			Path:     pathP,
			NodeName: path[1],
		}
		dbP, err := nodeP.GetOne()
		if err != nil{
			level.Error(logger).Log("msg", "query_p_failed", "path", req.Node, "err", err)
			return
		}
		if dbP == nil{
			return
		}

		// 查询a是否存在
		pathA :=fmt.Sprintf("%s/%d", dbP.Path, dbP.Id)
		whereStr := "level=? and path=? and node_name=?"
		dbA, err := StreePathGet(whereStr, 3, pathA, path[2])
		if err != nil {
			level.Error(logger).Log("msg", "query_a_failed", "path", req.Node, "err", err)
			return
		}
		if dbA == nil{
			return
		}

		// 删除a
		delNum, err = dbA.DelOne()
		if err != nil{
			level.Error(logger).Log("msg", "del_a_failed", "path", req.Node, "err", err)
			return
		}
		level.Info(logger).Log("msg", "del_a_success", "path", req.Node)
		return

	}
	return
}


/*
函数区域
*/

// 编写新增node的测试函数
func StreePathAddTest(logger log.Logger) {
	ns := []string{
		"inf.monitor.thanos",
		"inf.monitor.kafka",
		"inf.monitor.prometheus",
		"inf.monitor.m3db",
		"inf.cicd.gray",
		"inf.cicd.deploy",
		"inf.cicd.jenkins",
		"waimai.qiangdan.queue",
		"waimai.qiangdan.worker",
		"waimai.ditu.kafka",
		"waimai.ditu.es",
	}
	for _, n := range ns {
		req := &common.NodeCommonReq{
			Node: n,
		}
		StreePathAddOne(req, logger)
	}
}

// 编写查询node的测试函数
func StreePathQueryTest1(logger log.Logger) {
	ns := []string{
		"inf",
		"waimai",
		"a",
		"b",
		"c",
	}
	for _, n := range ns {
		req := &common.NodeCommonReq{
			Node: n,
			QueryType: 1,
		}
		res:= StreePathQuery(req, logger)
		level.Info(logger).Log("msg", "StreePathQuery.res","req.node", n, "num", len(res), "details", strings.Join(res, ","))
	}
}

// queryType = 2 查询
func StreePathQueryTest2(logger log.Logger) {
	ns := []string{
		"inf",
		"waimai",
		"a",
		"b",
		"c",
	}
	for _, n := range ns {
		req := &common.NodeCommonReq{
			Node: n,
			QueryType: 2,
		}
		res:= StreePathQuery(req, logger)
		level.Info(logger).Log("msg", "StreePathQuery.res","req.node", n, "num", len(res), "details", strings.Join(res, ","))
	}
}

// queryType = 3 查询
func StreePathQueryTest3(logger log.Logger) {
	ns := []string{
		"a.b",
		"b.a",
		"c.d",
		"inf",
		"inf.cicd",
		"inf.cicd.jenkins",
	}
	for _, n := range ns {
		req := &common.NodeCommonReq{
			Node: n,
			QueryType: 3,
		}
		res:= StreePathQuery(req, logger)
		level.Info(logger).Log("msg", "StreePathQuery.res","req.node", n, "num", len(res), "details", strings.Join(res, ","))
	}
}

// 编写删除node的测试函数
func StreePathDeleteTest(logger log.Logger) {
	ns := []string{
		"a.b",
		"b.a",
		"c.d",
		"inf.cicd.jenkins",
		"inf.cicd",
		"inf",
		"waimai",
	}
	for _, n := range ns {
		req := &common.NodeCommonReq{
			Node: n,
			QueryType: 3,
		}
		res:= StreePathDelete(req, logger)
		level.Info(logger).Log("msg", "StreePathDelete.res", "req.node", n, "del_num", res)
	}
}

// 编写强制删除node的测试函数
func StreePathForeceDeleteTest(logger log.Logger)  {
	ns := []string{
		"a.b",
		//"waimai",
		//"inf",
		//"inf.cicd",
		//"inf.monitor",
		"waimai.ditu",
	}
	for _, n := range ns {
		req := &common.NodeCommonReq{
			Node: n,
			ForceDelete: true,
		}
		res:= StreePathDelete(req, logger)
		level.Info(logger).Log("msg", "StreePathDelete.res", "req.node", n, "del_num", res)
	}
}