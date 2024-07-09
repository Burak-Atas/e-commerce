package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Burak-Atas/ecommerce/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrCantFindProduct    = errors.New("can't find product")
	ErrCantDecodeProducts = errors.New("can't find product")
	ErrUserIDIsNotValid   = errors.New("user is not valid")
	ErrCantUpdateUser     = errors.New("cannot add product to cart")
	ErrCantRemoveItem     = errors.New("cannot remove item from cart")
	ErrCantGetItem        = errors.New("cannot get item from cart ")
	ErrCantBuyCartItem    = errors.New("cannot update the purchase")
)

func AddProductToCart(ctx context.Context, prodCollection, userCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	searchfromdb, err := prodCollection.Find(ctx, bson.M{"_id": productID})
	if err != nil {
		log.Println(err)
		return ErrCantFindProduct
	}
	var productcart []models.ProductUser
	err = searchfromdb.All(ctx, &productcart)
	if err != nil {
		log.Println(err)
		return ErrCantDecodeProducts
	}
	fmt.Println(productID)
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrUserIDIsNotValid
	}

	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "usercart", Value: bson.D{{Key: "$each", Value: productcart}}}}}}
	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return ErrCantUpdateUser
	}
	return nil
}

func RemoveCartItem(ctx context.Context, prodCollection, userCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrUserIDIsNotValid
	}
	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.M{"$pull": bson.M{"usercart": bson.M{"_id": productID}}}
	_, err = userCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		return ErrCantRemoveItem
	}
	return nil
}
func RemoveCartItemOne(ctx context.Context, prodCollection, userCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	id, err := primitive.ObjectIDFromHex(userID)
	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	var user *models.User

	if err != nil {
		log.Println(err)
		return ErrUserIDIsNotValid
	}

	err = userCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		log.Println(err)
		return nil
	}

	// Sepet içindeki ürünlerin sayısını kontrol et
	count := 0
	for _, item := range user.UserCart {
		if item.Product_ID == productID {
			count++
		}
	}

	// Eğer aynı ID'li birden fazla ürün varsa, yalnızca birini sil
	if count > 1 {
		// Aynı ID'li ürünlerin indeksini bulun
		var productIndex int
		for i, item := range user.UserCart {
			if item.Product_ID == productID {
				productIndex = i
				break
			}
		}

		// Yeni sepet listesini oluşturun
		newCart := append(user.UserCart[:productIndex], user.UserCart[productIndex+1:]...)

		// Kullanıcıyı güncelleyin
		update := bson.M{"$set": bson.M{"usercart": newCart}}
		_, err = userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			return ErrCantRemoveItem
		}
	} else {
		// Yalnızca bir ürün varsa, normal $pull işlemini kullan
		update := bson.M{"$pull": bson.M{"usercart": bson.M{"_id": productID}}}
		_, err = userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			return ErrCantRemoveItem
		}
	}

	return nil
}
func BuyItemFromCart(ctx context.Context, userCollection *mongo.Collection, MySparisCollection *mongo.Collection, userID string) error {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrUserIDIsNotValid
	}
	var getcartitems models.User
	var ordercart models.Order
	ordercart.Order_ID = primitive.NewObjectID() // Her sipariş için benzersiz bir ID oluşturun
	ordercart.Orderered_At = time.Now()
	ordercart.Order_Cart = make([]models.ProductUser, 0)
	ordercart.Payment_Method.COD = true

	// Sepetteki ürünleri topla ve siparişin toplam fiyatını hesapla
	unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$usercart"}}}}
	grouping := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$_id"}, {Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: "$usercart.price"}}}}}}
	currentresults, err := userCollection.Aggregate(ctx, mongo.Pipeline{unwind, grouping})
	ctx.Done()
	if err != nil {
		panic(err)
	}
	var getusercart []bson.M
	if err = currentresults.All(ctx, &getusercart); err != nil {
		panic(err)
	}
	var total_price int32
	for _, user_item := range getusercart {
		price := user_item["total"]
		total_price = price.(int32)
	}
	ordercart.Price = int(total_price)

	// Kullanıcının sepetindeki ürünleri al
	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	err = userCollection.FindOne(ctx, filter).Decode(&getcartitems)
	if err != nil {
		log.Println(err)
		return err
	}

	// Sepette ürün varsa
	if len(getcartitems.UserCart) > 0 {
		// Sepetteki ürünleri yeni siparişe ekle
		ordercart.Order_Cart = getcartitems.UserCart

		// Yeni siparişi kullanıcının sipariş listesine ekle
		update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "orders", Value: ordercart}}}}
		_, err = userCollection.UpdateMany(ctx, filter, update)
		if err != nil {
			log.Println(err)
			return err
		}
	} else {
		// Sepet boşsa, hata mesajı döndür
		return nil
	}

	// Sepet boşalt
	usercart_empty := make([]models.ProductUser, 0)
	filtered := bson.D{primitive.E{Key: "_id", Value: id}}
	updated := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "usercart", Value: usercart_empty}}}}
	_, err = userCollection.UpdateOne(ctx, filtered, updated)
	if err != nil {
		return ErrCantBuyCartItem
	}

	return nil
}

func InstantBuyer(ctx context.Context, prodCollection, userCollection *mongo.Collection, productID primitive.ObjectID, UserID string) error {
	id, err := primitive.ObjectIDFromHex(UserID)
	if err != nil {
		log.Println(err)
		return ErrUserIDIsNotValid
	}
	var product_details models.ProductUser
	var orders_detail models.Order
	orders_detail.Order_ID = primitive.NewObjectID()
	orders_detail.Orderered_At = time.Now()
	orders_detail.Order_Cart = make([]models.ProductUser, 0)
	orders_detail.Payment_Method.COD = true
	err = prodCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: productID}}).Decode(&product_details)
	if err != nil {
		log.Println(err)
	}
	orders_detail.Price = product_details.Price
	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "orders", Value: orders_detail}}}}
	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println(err)
	}
	filter2 := bson.D{primitive.E{Key: "_id", Value: id}}
	update2 := bson.M{"$push": bson.M{"orders.$[].order_list": product_details}}
	_, err = userCollection.UpdateOne(ctx, filter2, update2)
	if err != nil {
		log.Println(err)
	}
	return nil
}
