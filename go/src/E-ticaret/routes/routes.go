package routes

import (
	"github.com/Burak-Atas/ecommerce/controllers"
	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/users/signup", controllers.SignUp())
	incomingRoutes.POST("/users/login", controllers.Login())
	incomingRoutes.GET("/sendmail", controllers.SendEmail())

	incomingRoutes.POST("/admin/addproduct", controllers.ProductViewerAdmin())
	incomingRoutes.GET("/admin/product/delete", controllers.DeleteProduct())

	incomingRoutes.POST("/admin/saveimage", controllers.SaveImage())
	incomingRoutes.POST("/admin/category/saveimage", controllers.SaveCategroyImage())
	incomingRoutes.GET("/admin/category/delete", controllers.RemoveTheCategory())
	incomingRoutes.GET("/users/getcategoryqueryid", controllers.GetCategoryWithQuery())

	incomingRoutes.GET("/users/productview", controllers.SearchProduct())
	incomingRoutes.GET("/users/search", controllers.SearchProductByQuery())
}
