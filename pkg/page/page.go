package page

// Request represents a pagination request.
type Request struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// Normalize ensures page request values are valid.
func (p *Request) Normalize() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}

// Offset returns the offset for database query.
func (p *Request) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Limit returns the limit for database query.
func (p *Request) Limit() int {
	return p.PageSize
}

// Response represents a pagination response.
type Response struct {
	Page      int   `json:"page"`
	PageSize  int   `json:"page_size"`
	Total     int64 `json:"total"`
	TotalPage int   `json:"total_page"`
}

// NewResponse creates a new pagination response.
func NewResponse(page, pageSize int, total int64) *Response {
	totalPage := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPage++
	}
	return &Response{
		Page:      page,
		PageSize:  pageSize,
		Total:     total,
		TotalPage: totalPage,
	}
}

// FromRequest creates a response from a request with total count.
func FromRequest(req *Request, total int64) *Response {
	return NewResponse(req.Page, req.PageSize, total)
}

// HasNext checks if there is a next page.
func (p *Response) HasNext() bool {
	return p.Page < p.TotalPage
}

// HasPrev checks if there is a previous page.
func (p *Response) HasPrev() bool {
	return p.Page > 1
}

// IsEmpty checks if the response has no data.
func (p *Response) IsEmpty() bool {
	return p.Total == 0
}
