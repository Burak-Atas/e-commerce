package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Burak-Atas/ecommerce/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (app *Application) GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.GetString("uid")
		if userId == "" {
			c.JSON(400, gin.H{"error": "User ID not provided"})
			return
		}

		var foundUser models.User
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := UserCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&foundUser)
		if err != nil {
			c.JSON(500, gin.H{"error": "Error finding user"})
			return
		}

		c.JSON(200, gin.H{"email": foundUser.Email, "first_name": foundUser.First_Name, "last_name": foundUser.Last_Name})
	}
}

func (app *Application) UpdateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.GetString("uid")

		if userId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user id not found"})
			return
		}

		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		password := user.Password
		newPassword := HashPassword(*password)

		update := bson.M{
			"$set": bson.M{
				"password":   newPassword,
				"first_name": user.First_Name,
				"last_name":  user.Last_Name,
				"email":      user.Email,
			},
		}

		query := bson.M{"user_id": userId}

		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, err := app.userCollection.UpdateOne(ctx, query, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "not updated"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "update successful",
		})
	}
}

func (app Application) GetSparis() gin.HandlerFunc {
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
		unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$orders"}}}}
		pointcursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{filter_match, unwind})
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
			"sparis": filledcart.Order_Status,
		}

		c.JSON(http.StatusOK, response)
	}
}
