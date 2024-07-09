package controllers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Burak-Atas/ecommerce/database"
	"github.com/Burak-Atas/ecommerce/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Application struct {
	prodCollection *mongo.Collection
	userCollection *mongo.Collection
}

func NewApplication(prodCollection, userCollection *mongo.Collection) *Application {
	return &Application{
		prodCollection: prodCollection,
		userCollection: userCollection,
	}
}

func (app *Application) AddToCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryID := c.Query("id")
		fmt.Println("product id", productQueryID)
		if productQueryID == "" {
			log.Println("product id is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}
		userQueryID := c.GetString("uid")

		if userQueryID == "" {
			log.Println("user id is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}
		productID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = database.AddProductToCart(ctx, app.prodCollection, app.userCollection, productID, userQueryID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
		}
		c.IndentedJSON(200, "Successfully Added to the cart")
	}
}

func (app *Application) RemoveItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryID := c.Query("id")
		if productQueryID == "" {
			log.Println("product id is inavalid")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}

		userQueryID := c.GetString("uid")

		if userQueryID == "" {
			log.Println("user id is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("UserID is empty"))
		}

		ProductID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = database.RemoveCartItem(ctx, app.prodCollection, app.userCollection, ProductID, userQueryID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}
		c.IndentedJSON(200, "Successfully removed from cart")
	}
}

func (app *Application) RemoveItemOne() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryID := c.Query("id")
		if productQueryID == "" {
			log.Println("product id is inavalid")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}

		userQueryID := c.GetString("uid")

		if userQueryID == "" {
			log.Println("user id is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("UserID is empty"))
		}

		ProductID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = database.RemoveCartItemOne(ctx, app.prodCollection, app.userCollection, ProductID, userQueryID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}
		c.IndentedJSON(200, "Successfully removed from cart")
	}
}

func GetItemFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		user_id := c.GetString("uid")

		if user_id == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "invalid id"})
			c.Abort()
			return
		}

		usert_id, _ := primitive.ObjectIDFromHex(user_id)

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var filledcart models.User
		err := UserCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: usert_id}}).Decode(&filledcart)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "id not found"})
			return
		}

		filter_match := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: usert_id}}}}
		unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$usercart"}}}}
		grouping := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$_id"}, {Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: "$usercart.price"}}}}}}
		pointcursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{filter_match, unwind, grouping})
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "aggregation error"})
			return
		}

		var listing []bson.M
		if err = pointcursor.All(ctx, &listing); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cursor error"})
			return
		}

		if len(listing) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no data found"})
			return
		}

		response := gin.H{
			"total":    listing[0]["total"],
			"usercart": filledcart.UserCart,
		}

		c.JSON(http.StatusOK, response)
	}
}

func (app *Application) BuyFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		userQueryID := c.GetString("uid")

		if userQueryID == "" {
			log.Panicln("user id is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("UserID is empty"))
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		err := database.BuyItemFromCart(ctx, app.userCollection, MySparisCollection, userQueryID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
		}
		fmt.Println("mail gönderme işlemi", c.GetString("email"))
		sendEmail(c.GetString("email"), "SATIN_ALMA", "")

		c.IndentedJSON(200, "Successfully Placed the order")
	}
}

func (app *Application) InstantBuy() gin.HandlerFunc {
	return func(c *gin.Context) {
		UserQueryID := c.GetString("uid")

		if UserQueryID == "" {
			log.Println("UserID is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("UserID is empty"))
		}
		ProductQueryID := c.Query("pid")
		if ProductQueryID == "" {
			log.Println("Product_ID id is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("product_id is empty"))
		}
		productID, err := primitive.ObjectIDFromHex(ProductQueryID)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = database.InstantBuyer(ctx, app.prodCollection, app.userCollection, productID, UserQueryID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
		}
		c.IndentedJSON(200, "Successully placed the order")
	}
}

func (app *Application) Buy() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User

		UserQueryID := c.GetString("uid")
		if UserQueryID == "" {
			log.Println("UserID is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("UserID is empty"))
		}

		query := bson.M{"user_id": UserQueryID}
		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		app.userCollection.FindOne(ctx, query).Decode(&user)

		if len(user.UserCart) == 0 {
			c.JSON(400, gin.H{"error": "lütfen sepete ürün ekleyin"})
			return
		}

		var card models.Order
		if err := c.BindJSON(&card); err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("UserID is empty"))
			return
		}

		filter_match := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: UserQueryID}}}}
		unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$usercart"}}}}
		grouping := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$_id"}, {Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: "$usercart.price"}}}}}}
		pointcursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{filter_match, unwind, grouping})
		var listing []bson.M
		if err = pointcursor.All(ctx, &listing); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cursor error"})
			return
		}

		if len(listing) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no data found"})
			return
		}
		total := listing[0]["total"]

		c.JSON(200, gin.H{"message": "başarıyla satın aldınız", "miktar": total})

	}
}

func (app Application) IsConto() gin.HandlerFunc {
	return func(c *gin.Context) {
		var isConto models.Code

		if err := c.BindJSON(&isConto); err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("UserID is empty"))
			return
		}
		fmt.Println("isconto name :", isConto.Name)

		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := IsContoCollection.FindOne(ctx, bson.M{"name": isConto.Name}).Decode(&isConto)

		if err != nil {
			c.JSON(400, "isconto kullanılmaktadır")
			return
		}

		c.JSON(200, gin.H{
			"isconto": isConto.IsConto,
		})
	}
}
