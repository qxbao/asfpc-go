package routes

import (
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/server/modules/routes/services/category"
)

func InitCategoryRoutes(s *infras.Server) {
	e := s.Echo
	services := category.CategoryRoutingService{
		Server: s,
	}
	
	e.GET("/category/list", services.GetCategories)
	e.POST("/category/add", services.AddCategory)
	e.DELETE("/category/delete/:id", services.DeleteCategory)
	e.PUT("/category/assign", services.UpdateCategory)
}
