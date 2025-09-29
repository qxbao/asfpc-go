package infras

type TraceRequestDTO struct {
	RequestID int32 `param:"request_id" validate:"required"`
}