package infras

type AddCategoryRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type UpdateCategoryRequest struct {
	Id          int32  `json:"id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type AddGroupCategoryRequest struct {
	GroupId    int32  `json:"group_id" validate:"required"`
	CategoryId int32  `json:"category_id" validate:"required"`
}

type DeleteGroupCategoryRequest = AddGroupCategoryRequest