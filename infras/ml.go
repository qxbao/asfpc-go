package infras

type MLTrainDTO struct {
	ModelName *string `json:"model_name"`
	AutoTune  *bool   `json:"auto_tune"`
	Trials    *int    `json:"trials,omitempty" validate:"omitempty,min=1"`
}

type WithModelNameDTO struct {
	ModelName string `json:"model_name" query:"model_name" validate:"required"`
}
