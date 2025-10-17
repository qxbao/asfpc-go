package infras

type TraceRequestDTO struct {
	RequestID int32 `param:"request_id" validate:"required"`
}

type ChartRequestDTO struct {
	CategoryID int32 `query:"category_id" validate:"required"`
}