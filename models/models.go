package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	UserRoleAdmin    = "admin"
	UserRoleCustomer = "customer"
)

type User struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	UserID         string             `bson:"user_id"`
	Username       string             `bson:"username"`
	Email          string             `bson:"email"`
	FirstName      *string            `bson:"first_name"`
	LastName       *string            `bson:"last_name"`
	Phone          *string            `bson:"phone"`
	Password       string             `bson:"password"`
	Token          *string            `bson:"token"`
	RefreshToken   *string            `bson:"refresh_token"`
	CreatedAt      time.Time          `bson:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at"`
	AddressDetails []Address          `bson:"address"`
	UserCart       []ProductUser      `bson:"user_cart"`
	OrderStatus    string             `bson:"order_status"`
	Role           string             `bson:"role"`
}

type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"` // MongoDB's unique identifier for the product
	ProductName string             `bson:"product_name"`
	ProductID   string             `bson:"product_id"`
	Description string             `bson:"description"`
	Price       uint64             `bson:"price"`
	Quantity    int                `bson:"quantity"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
}

type ProductUser struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	ProductName string             `bson:"product_name"`
	Description *string            `bson:"description"`
	Price       uint64             `bson:"price"`
	Rating      uint8              `bson:"rating"`
	ImageURL    *string            `bson:"image_url"`
}

type Address struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"` // MongoDB's unique identifier for the address
	HouseNo    *string            `bson:"house_no"`
	Street     *string            `bson:"street"`
	City       *string            `bson:"city"`
	State      *string            `bson:"state"`
	PostalCode *string            `bson:"postal_code"`
	Country    *string            `bson:"country"`
}

type Order struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"` // MongoDB's unique identifier for the order
	OrderCart     []ProductUser      `bson:"order_cart"`
	OrderedAt     time.Time          `bson:"ordered_at"`
	Price         uint64             `bson:"price"`
	PaymentMethod Payment            `bson:"payment_method"`
	Discount      uint64             `bson:"discount"`
	UserID        string             `bson:"user_id"`
	Products      []Product          `bson:"products"`
}

type Payment struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"` // MongoDB's unique identifier for the payment
	Digital bool               `bson:"digital_payment"`
	COD     bool               `bson:"cod"`
}
