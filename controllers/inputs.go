package controllers

import (
	"time"

	"github.com/milua25/e-commerce-backend/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User Input Models
type CreateUserInput struct {
	ID           primitive.ObjectID `json:"-"`
	UserID       string             `json:"-"`
	Username     string             `json:"username" binding:"required" validate:"min=2,max=30"`
	Email        string             `json:"email" binding:"required,email"`
	FirstName    *string            `json:"first_name" validate:"omitempty,min=2,max=30"`
	LastName     *string            `json:"last_name" validate:"omitempty,min=2,max=30"`
	Phone        string             `json:"phone" binding:"required" validate:"startswith=+,e164"`
	Password     string             `json:"password" binding:"required,min=6" validate:"min=6"`
	CreatedAt    time.Time          `json:"-"`
	UpdatedAt    time.Time          `json:"-"`
	Token        *string            `json:"-"`
	RefreshToken *string            `json:"-"`
	Role         string             `json:"role"`
}

// UpdateUserInput is used for updating user details
type UpdateUserInput struct {
	FirstName *string `json:"first_name" validate:"omitempty,min=2,max=30"`
	LastName  *string `json:"last_name" validate:"omitempty,min=2,max=30"`
	Phone     *string `json:"phone" validate:"omitempty,e164"`
	Email     *string `json:"email" binding:"omitempty,email" validate:"omitempty,email"`
	Role      *string `json:"role" validate:"omitempty,oneof=admin customer"`
}

// LoginInput is used for user login
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Product Input Models
type CreateProductInput struct {
	ID          primitive.ObjectID `json:"-"`
	ProductName string             `json:"product_name" binding:"required"`
	ProductID   string             `json:"-"`
	Description string             `json:"description" binding:"required"`
	Price       uint64             `json:"price" binding:"required,gt=0"`
	Quantity    int                `json:"quantity" binding:"required,gt=0"`
	CreatedAt   time.Time          `json:"-"`
	UpdatedAt   time.Time          `json:"-"`
}

// UpdateProductInput is used for updating product details
type UpdateProductInput struct {
	ProductName *string `json:"product_name" validate:"omitempty,min=2,max=30"`
	Description *string `json:"description" validate:"omitempty,min=2,max=100"`
	Price       *uint64 `json:"price" binding:"omitempty,gt=0" validate:"omitempty,gt=0"`
	Quantity    *int    `json:"quantity" binding:"omitempty,gt=0"`
}

// ProductUser Input Model
type ProductUserInput struct {
	ProductName string `json:"product_name" binding:"required"`
	Description string `json:"description" binding:"required"`
	Price       uint64 `json:"price" binding:"required,gt=0"`
	Rating      uint8  `json:"rating" binding:"min=0,max=5"`
	ImageURL    string `json:"image_url" binding:"required,url"`
}

// Address Input Models
type CreateAddressInput struct {
	HouseNo    string `json:"house_no" binding:"required"`
	Street     string `json:"street" binding:"required"`
	City       string `json:"city" binding:"required"`
	State      string `json:"state" binding:"required"`
	PostalCode string `json:"postal_code" binding:"required"`
	Country    string `json:"country" binding:"required"`
}

type UpdateAddressInput struct {
	HouseNo    *string `json:"house_no"`
	Street     *string `json:"street"`
	City       *string `json:"city"`
	State      *string `json:"state"`
	PostalCode *string `json:"postal_code"`
	Country    *string `json:"country"`
}

// Order Input Models
type CreateOrderInput struct {
	PaymentMethod string               `json:"payment_method" binding:"required"`
	Discount      uint64               `json:"discount"`
	Products      []models.ProductUser `json:"products" binding:"required"`
}

// Payment Input Models
type PaymentInput struct {
	Digital bool `json:"digital_payment"`
	COD     bool `json:"cod" binding:"required"`
}

func (input CreateAddressInput) ToModel() models.Address {
	return models.Address{
		ID:         primitive.NewObjectID(),
		HouseNo:    &input.HouseNo,
		Street:     &input.Street,
		City:       &input.City,
		State:      &input.State,
		PostalCode: &input.PostalCode,
		Country:    &input.Country,
	}
}

func (input UpdateAddressInput) ToModel() models.Address {
	return models.Address{
		HouseNo:    input.HouseNo,
		Street:     input.Street,
		City:       input.City,
		State:      input.State,
		PostalCode: input.PostalCode,
		Country:    input.Country,
	}
}

func (input CreateUserInput) ToModel() models.User {
	return models.User{
		ID:           input.ID,
		UserID:       input.UserID,
		Username:     input.Username,
		Email:        input.Email,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Phone:        &input.Phone,
		Password:     input.Password,
		CreatedAt:    input.CreatedAt,
		UpdatedAt:    input.UpdatedAt,
		Token:        input.Token,
		RefreshToken: input.RefreshToken,
		Role:         input.Role,
	}
}

func (input CreateProductInput) ToModel() models.Product {
	return models.Product{
		ID:          primitive.NewObjectID(),
		ProductID:   primitive.NewObjectID().Hex(),
		ProductName: input.ProductName,
		Description: input.Description,
		Price:       input.Price,
		Quantity:    input.Quantity,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (input UpdateProductInput) ToModel() models.Product {
	return models.Product{
		ProductName: *input.ProductName,
		Description: *input.Description,
		Price:       *input.Price,
		Quantity:    *input.Quantity,
	}
}
