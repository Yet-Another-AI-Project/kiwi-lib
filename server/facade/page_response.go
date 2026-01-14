package facade

// @Description 分页响应
type PageResponse[T any] struct {
	// list
	List []T `json:"list" swaggertype:"object"`
	// 当前页
	PageNum int `json:"page_num"`
	// 页大小
	PageSize int `json:"page_size"`
	// 总条数
	Total int `json:"total"`
}
