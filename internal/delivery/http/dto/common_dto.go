package dto

type Response struct {
	Success bool        `json:"success"`
	Status  string      `json:"status,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type PaginatedData struct {
	Items      interface{} `json:"items"`
	TotalCount int64       `json:"total_count"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
}

type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Status     string      `json:"status"`
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page,omitempty"`
	Limit      int         `json:"limit,omitempty"`
	TotalPages int         `json:"total_pages,omitempty"`
}

func Success(data interface{}) Response {
	return Response{
		Success: true,
		Status:  "success",
		Data:    data,
	}
}

func SuccessWithMsg(message string, data interface{}) Response {
	return Response{
		Success: true,
		Status:  "success",
		Message: message,
		Data:    data,
	}
}

func SuccessPaginated(data interface{}, total int64, page, limit, totalPages int) PaginatedResponse {
	return PaginatedResponse{
		Success:    true,
		Status:     "success",
		Data:       data,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}
}

func Fail(message string, err interface{}) Response {
	return Response{
		Success: false,
		Status:  "error",
		Message: message,
		Error:   err,
	}
}