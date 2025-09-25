package infras

type MLTrainDTO struct {
	ModelName *string `json:"model_name"`
	AutoTune  *bool   `json:"auto_tune"`
}