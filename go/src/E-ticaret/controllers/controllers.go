package controllers

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/Burak-Atas/ecommerce/database"
	"github.com/Burak-Atas/ecommerce/models"
	generate "github.com/Burak-Atas/ecommerce/tokens"

	"gopkg.in/gomail.v2"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var UserCollection *mongo.Collection = database.UserData(database.Client, "Users")
var CategoryCollection *mongo.Collection = database.CategoryData(database.Client, "Category")
var ProductCollection *mongo.Collection = database.ProductData(database.Client, "Products")
var IsContoCollection *mongo.Collection = database.IsContoData(database.Client, "Isconto")
var MySparisCollection *mongo.Collection = database.MySparis(database.Client, "MySparis")

var Validate = validator.New()

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userpassword string, givenpassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(givenpassword), []byte(userpassword))
	valid := true
	msg := ""
	if err != nil {
		msg = "Login Or Passowrd is Incorerct"
		valid = false
	}
	return valid, msg
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validationErr := Validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr})
			return
		}

		count, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
		}
		/*
				count, err = UserCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
				defer cancel()
				if err != nil {
					log.Panic(err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": err})
					return
				}
			if count > 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Phone is already in use"})
				return
			}
		*/

		password := HashPassword(*user.Password)
		user.Password = &password

		user.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_ID = user.ID.Hex()
		token, refreshtoken, _ := generate.TokenGenerator(*user.Email, *user.First_Name, *user.Last_Name, user.User_ID)
		user.Token = &token
		user.Refresh_Token = &refreshtoken
		user.UserCart = make([]models.ProductUser, 0)
		user.Address_Details = make([]models.Address, 0)
		user.Order_Status = make([]models.Order, 0)
		_, inserterr := UserCollection.InsertOne(ctx, user)
		if inserterr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "not created"})
			return
		}
		defer cancel()
		c.JSON(http.StatusCreated, "Successfully Signed Up!!")
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User
		var founduser models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
		err := UserCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&founduser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login or password incorrect"})
			return
		}
		PasswordIsValid, msg := VerifyPassword(*user.Password, *founduser.Password)
		defer cancel()
		if !PasswordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			fmt.Println(msg)
			return
		}
		token, refreshToken, _ := generate.TokenGenerator(*founduser.Email, *founduser.First_Name, *founduser.Last_Name, founduser.User_ID)
		defer cancel()
		generate.UpdateAllTokens(token, refreshToken, founduser.User_ID)
		c.JSON(http.StatusOK, founduser)
	}
}

func ProductViewerAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var products models.Product
		defer cancel()
		if err := c.BindJSON(&products); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		products.Product_ID = primitive.NewObjectID()
		_, anyerr := ProductCollection.InsertOne(ctx, products)
		if anyerr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Not Created"})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, "Successfully added our Product Admin!!")
	}
}

func SearchProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		var productlist []models.Product
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		cursor, err := ProductCollection.Find(ctx, bson.D{{}})
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "Someting Went Wrong Please Try After Some Time")
			return
		}
		err = cursor.All(ctx, &productlist)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		defer cursor.Close(ctx)
		if err := cursor.Err(); err != nil {
			// Don't forget to log errors. I log them really simple here just
			// to get the point across.
			log.Println(err)
			c.IndentedJSON(400, "invalid")
			return
		}
		defer cancel()
		c.IndentedJSON(200, productlist)

	}
}

func SearchProductByQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		var searchproducts []models.Product
		queryParam := c.Query("name")
		if queryParam == "" {
			log.Println("query is empty")
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Error": "Invalid Search Index"})
			c.Abort()
			return
		}
		id, _ := primitive.ObjectIDFromHex(queryParam)

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		searchquerydb, err := ProductCollection.Find(ctx, bson.M{"_id": id})
		if err != nil {
			c.IndentedJSON(404, "something went wrong in fetching the dbquery")
			return
		}
		err = searchquerydb.All(ctx, &searchproducts)
		if err != nil {
			log.Println(err)
			c.IndentedJSON(400, "invalid")
			return
		}
		defer searchquerydb.Close(ctx)
		if err := searchquerydb.Err(); err != nil {
			log.Println(err)
			c.IndentedJSON(400, "invalid request")
			return
		}
		defer cancel()
		fmt.Println(searchproducts)
		c.IndentedJSON(200, searchproducts[0])
	}
}

func AddCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		var category models.Category

		if err := c.BindJSON(&category); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validationErr := Validate.Struct(category)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr})
			return
		}

		id := primitive.NewObjectID()

		category.Category_ID = id

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, inserterr := CategoryCollection.InsertOne(ctx, category)
		if inserterr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "not created"})
			return
		}
		defer cancel()
		c.JSON(http.StatusCreated, "Successfully categroy!!")
	}
}

func GetCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		var categories []models.Category

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cursor, err := CategoryCollection.Find(ctx, bson.D{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching categories"})
			return
		}

		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var category models.Category
			if err := cursor.Decode(&category); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while decoding category"})
				return
			}
			categories = append(categories, category)
		}

		if err := cursor.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cursor error"})
			return
		}

		c.JSON(http.StatusOK, categories)
	}
}

func SaveImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Dosya uzantısını al
		ext := filepath.Ext(file.Filename)
		// Yeni dosya adı oluştur
		newFileName := strings.ReplaceAll(file.Filename, ext, "") + "_" + GenerateRandomString(10) + ext
		// Dosyayı kaydetmek için hedef dizini oluştur
		dstPath := "static" + "/" + newFileName

		if err := c.SaveUploadedFile(file, dstPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Dosya kaydedilemedi"})
			return
		}

		// Kaydedilen dosyanın URL'sini oluştur
		url := c.Request.Host + "/" + dstPath

		// Sonuç olarak URL'yi geri döndür
		c.JSON(http.StatusOK, gin.H{"url": url})
	}
}

// Rastgele bir dize oluşturmak için
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func SaveCategroyImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ext := filepath.Ext(file.Filename)
		newFileName := strings.ReplaceAll(file.Filename, ext, "") + "_" + GenerateRandomString(10) + ext
		dstPath := "static" + "/" + newFileName

		if err := c.SaveUploadedFile(file, dstPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Dosya kaydedilemedi"})
			return
		}

		// Kaydedilen dosyanın URL'sini oluştur
		url := c.Request.Host + "/" + dstPath

		// Sonuç olarak URL'yi geri döndür
		c.JSON(http.StatusOK, gin.H{"url": url})
	}
}

func RemoveTheCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		categoryQueryId := c.Query("id")
		fmt.Println("categroy id", categoryQueryId)

		if categoryQueryId == "" {
			c.JSON(400, gin.H{"error": "id not found"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		categoryPrimitiveId, err := primitive.ObjectIDFromHex(categoryQueryId)
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid id format"})
			return
		}

		_, err = CategoryCollection.DeleteOne(ctx, bson.M{"_id": categoryPrimitiveId})

		if err != nil {
			c.JSON(400, gin.H{"error": "category not deleted"})
			return
		}

		c.JSON(200, gin.H{"message": "Category deleted successfully"})
	}
}

func SendEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		email := c.Query("email")
		if email == "" {
			c.JSON(400, gin.H{"error": "email not found"})
			return
		}

		sendEmail(email, "deneme maili", "deneme")
		c.JSON(200, "başarılı")

	}
}

func sendEmail(to string, subject string, body string) {
	m := gomail.NewMessage()
	m.SetHeader("From", "firatshop5@gmail.com")
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer("smtp.gmail.com", 587, "firatshop5@gmail.com", "pcdo ooqr oprq xkoq")

	if err := d.DialAndSend(m); err != nil {
		log.Println("Could not send email: ", err)
		return
	}

	log.Println("Email sent successfully!")
}

func GetCategoryWithQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Query("id")
		if id == "" {
			c.JSON(400, gin.H{"error": "ID is required"})
			return
		}
		queryId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid ID format"})
			return
		}
		var m models.Category

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = CategoryCollection.FindOne(ctx, bson.M{"_id": queryId}).Decode(&m)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(404, gin.H{"error": "Category not found"})
			} else {
				c.JSON(500, gin.H{"error": "Internal server error"})
			}
			return
		}

		c.JSON(200, m)
	}
}

func DeleteProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		productId := c.Query("id")
		deleteId, err := primitive.ObjectIDFromHex(productId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err = ProductCollection.DeleteOne(ctx, bson.M{"_id": deleteId})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, "Ürün Başarılı şekilde silindi!")
	}
}
