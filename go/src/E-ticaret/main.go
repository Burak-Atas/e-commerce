package main

import (
	"log"
	"os"

	"github.com/akhil/ecommerce-yt/controllers"
	"github.com/akhil/ecommerce-yt/database"
	"github.com/akhil/ecommerce-yt/middleware"
	"github.com/akhil/ecommerce-yt/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	app := controllers.NewApplication(database.ProductData(database.Client, "Products"), database.UserData(database.Client, "Users"))

	router := gin.New()
	router.Use(gin.Logger())

	// CORS middleware'ini ekleyin
	// Özelleştirilmiş CORS ayarları
	corsConfig := cors.Config{
		AllowOrigins:     []string{"http://localhost:4200"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "token"},
		AllowCredentials: true,
	}

	router.Use(cors.New(corsConfig))
	routes.UserRoutes(router)
	router.POST("/addcategory", controllers.AddCategory())
	router.GET("/getcotegory", controllers.GetCategory())

	router.Use(middleware.Authentication())
	router.GET("/addtocart", app.AddToCart())
	router.GET("/removeitem", app.RemoveItem())
	router.GET("/removeitemone", app.RemoveItemOne())
	router.GET("/listcart", controllers.GetItemFromCart())
	router.POST("/addaddress", controllers.AddAddress())
	router.GET("/getaddress", controllers.GetAddress())

	router.PUT("/edithomeaddress", controllers.EditHomeAddress())
	router.PUT("/editworkaddress", controllers.EditWorkAddress())
	router.GET("/deleteaddresses", controllers.DeleteAddress())
	router.GET("/cartcheckout", app.BuyFromCart())
	router.GET("/instantbuy", app.InstantBuy())

	log.Fatal(router.Run(":" + port))
}
