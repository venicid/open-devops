package common

import "github.com/ning1875/inverted-index/labels"

type ResourceQueryReq struct {
	ResourceType string `json:"resource_type" binding:"required"`  // 资源的类型
	Labels []*SingleTagReq  `json:"labels" binding:"required"`  // 查询的标签组
	TargetLabel string `json:"target_label"`  // 目标 g.p.a

}

type SingleTagReq struct {
	Key string `json:"key" binding:"required"`  // 标签的名字
	Value string `json:"value" binding:"required"`  // 标签的值，可以试正则表达式
	Type int `json:"type" binding:"required"`  // 类型 1-4 = != ~= ~!
}

type QueryResponse struct {
	Code int `json:"code"`
	CurrentPage int `json:"current_page"`
	PageSize int `json:"page_size"`
	PageCount int `json:"page_count"`
	TotalCount int `json:"total_count"`
	Result interface{} `json:"result"`
}


// 将前端传入的数据，转为后端的数据类型
func FormatLabelMatcher(ls []*SingleTagReq) []*labels.Matcher  {

	matcher := make([]*labels.Matcher, 0)
	for _, i := range  ls{
		mType, ok := labels.MatchMap[i.Type]
		if ! ok {
			continue
		}
		matcher = append(matcher,
			labels.MustNewMatcher(mType, i.Key, i.Value))

	}
	return matcher
}