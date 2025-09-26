package infras

type UpdateSettingsDTO struct {
	Settings map[string]string `json:"settings" validate:"required"`
}