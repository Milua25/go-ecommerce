package controllers

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// User Response Model
type UserResponse struct {
	ID             bson.ObjectID    `json:"id"`
	Username       string                `json:"username"`
	Email          string                `json:"email"`
	FirstName      *string               `json:"first_name"`
	LastName       *string               `json:"last_name"`
	Phone          *string               `json:"phone"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
	UserID         string                `json:"user_id"`
	AddressDetails []AddressResponse     `json:"address_details"`
	UserCart       []ProductUserResponse `json:"user_cart"`
	OrderStatus    string                `json:"order_status"`
	Token          *string               `json:"token,omitempty"`
	RefreshToken   *string               `json:"refresh_token,omitempty"`
}

// Product Response Model
type ProductResponse struct {
	ID          bson.ObjectID `json:"id"`
	ProductName string             `json:"product_name"`
	ProductID   string             `json:"product_id"`
	Description string             `json:"description"`
	Price       uint64             `json:"price"`
	Quantity    int                `json:"quantity"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// ProductUser Response Model
type ProductUserResponse struct {
	ID          string `json:"id"`
	ProductName string `json:"product_name"`
	Description string `json:"description"`
	Price       uint64 `json:"price"`
	Rating      uint8  `json:"rating"`
	ImageURL    string `json:"image_url"`
}

// Address Response Model
type AddressResponse struct {
	ID         bson.ObjectID `json:"id"`
	HouseNo    string             `json:"house_no"`
	Street     string             `json:"street"`
	City       string             `json:"city"`
	State      string             `json:"state"`
	PostalCode string             `json:"postal_code"`
	Country    string             `json:"country"`
}

// Order Response Model
type OrderResponse struct {
	ID            bson.ObjectID    `json:"id"`
	Order_Cart    []ProductUserResponse `json:"order_cart"`
	Ordered_At    time.Time             `json:"ordered_at"`
	Price         uint64                `json:"price"`
	PaymentMethod string                `json:"payment_method"`
	Discount      uint64                `json:"discount"`
	UserID        string                `json:"user_id"`
	Products      []ProductResponse     `json:"products"`
}

// Payment Response Model
type PaymentResponse struct {
	ID      bson.ObjectID `json:"id"`
	Digital bool               `json:"digital_payment"`
	COD     bool               `json:"cod"`
}

// Generic responses
type MessageResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
