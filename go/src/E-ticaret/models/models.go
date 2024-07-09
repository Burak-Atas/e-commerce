package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID              primitive.ObjectID `json:"_id" bson:"_id"`
	First_Name      *string            `json:"first_name"`
	Last_Name       *string            `json:"last_name"`
	Password        *string            `json:"password"   validate:"required,min=6"`
	Email           *string            `json:"email"      validate:"email,required"`
	Phone           *string            `json:"phone"`
	Token           *string            `json:"token"`
	Refresh_Token   *string            `josn:"refresh_token"`
	Created_At      time.Time          `json:"created_at"`
	Updated_At      time.Time          `json:"updtaed_at"`
	User_ID         string             `json:"user_id"`
	UserCart        []ProductUser      `json:"usercart" bson:"usercart"`
	Address_Details []Address          `json:"address" bson:"address"`
	Order_Status    []Order            `json:"orders" bson:"orders"`
}

type Product struct {
	Product_ID   primitive.ObjectID `bson:"_id" json:"id"`
	Product_Name *string            `json:"name"`
	Price        *uint64            `json:"price"`
	Isactive     *bool              `json:"isActive"`
	Image        *string            `json:"imageUrl"`
	Description  *string            `json:"description"`
	CategoryId   *string            `json:"categoryId"`
}

type ProductUser struct {
	Product_ID   primitive.ObjectID `bson:"_id" json:"id"`
	Product_Name *string            `json:"name"`
	Price        int                `json:"price"  bson:"price"`
	Image        *string            `json:"imageUrl"`
	CategoryId   *string            `json:"categoryId"`
	Isactive     *bool              `json:"isActive"`
}

type Address struct {
	Address_id primitive.ObjectID `bson:"_id" json:"id"`
	House      *string            `json:"house_name"`
	Street     *string            `json:"street_name"`
	City       *string            `json:"city_name" `
	Pincode    *string            `json:"pin_code"`
}

type Order struct {
	Order_ID       primitive.ObjectID `bson:"_id" json:"id"`
	Order_Cart     []ProductUser      `json:"order_list"  bson:"order_list"`
	Orderered_At   time.Time          `json:"ordered_on"  bson:"ordered_on"`
	Price          int                `json:"total_price" bson:"total_price"`
	IsConto        *int               `json:"isConto"    bson:"isConto"`
	Payment_Method Payment            `json:"payment_method" bson:"payment_method"`
}

type Payment struct {
	Digital bool `json:"digital" bson:"digital"`
	COD     bool `json:"cod"     bson:"cod"`
}

type Category struct {
	Category_ID primitive.ObjectID `bson:"_id" json:"id"`
	Name        string             `json:"name"`
	Image       *string            `json:"imageUrl"  bson:"imageUrl"`
}

type Code struct {
	Code_ID primitive.ObjectID `bson:"_id" json:"id"`
	Name    string             `json:"name"`
	IsConto int                `json:"isConto"`
}
