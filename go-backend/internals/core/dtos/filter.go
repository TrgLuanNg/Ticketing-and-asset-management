package dtos

type ModelAdvancedFilter struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type Filter struct {
	PageIndex       int                   `json:"pageIndex"`
	PageSize        int                   `json:"pageSize"`
	OrderBy         string                `json:"orderBy"`
	SearchText      string                `json:"searchText"`
	UserSession     string                `json:"userSession"`
	SkipCount       bool                  `json:"skipCount"`
	AdvancedFilters []ModelAdvancedFilter `json:"advancedFilters"`
}

type ResponseAPI struct {
	Status       bool   `json:"status"`
	Message      string `json:"message,omitempty"`
	Data         any    `json:"data,omitempty"`
	TotalRecords int64  `json:"totalRecords,omitempty"`
}

func SuccessResponse(data any, message ...string) ResponseAPI {
	msg := "Success"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return ResponseAPI{
		Status:  true,
		Data:    data,
		Message: msg,
	}
}

func PagedResponse(data any, total int64, message ...string) ResponseAPI {
	msg := "Success"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return ResponseAPI{
		Status:       true,
		Data:         data,
		TotalRecords: total,
		Message:      msg,
	}
}

func ErrorResponse(message string) ResponseAPI {
	return ResponseAPI{
		Status:  false,
		Message: message,
	}
}
