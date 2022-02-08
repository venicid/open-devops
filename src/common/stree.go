package common

// 操作数结构的通用对象: 公用的意思是，增删改查都用这个
// 各个字段在每个verb是不一样的
// 增删改查

type NodeCommonReq struct {
	Node string `json:"node"`   // 服务节点名称：可以一段式，也可以是两段式 inf  inf.mon
	QueryType int `json:"query_type"`  // 查询模式 1,2,3
	ForceDelete bool `json:"force_delete"`  // 子节点强制删除
}
