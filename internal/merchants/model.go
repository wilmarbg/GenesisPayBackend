package merchants

import (
	"time"

	"github.com/google/uuid"
)

type Merchant struct {
	ID		uuid.UUID	`gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID		uuid.UUID	`gorm:"type:uuid;not null" json:"user_id"`
	BusinessName	string		`gorm:"column:business_name;not null" json:"business_name"`
	NIT		string		`gorm:"column:nit;uniqueIndex;not null" json:"nit"`
	Category	string		`gorm:"column:category" json:"category"`
	Address		string		`gorm:"column:address" json:"address"`
	Description	string		`gorm:"column:description" json:"description"`
	Balance		float64		`gorm:"column:balance;not null;default:0" json:"balance"`
	IsActive	bool		`gorm:"column:is_active;not null;default:true" json:"is_active"`
	CreatedAt	time.Time	`gorm:"column:created_at" json:"created_at"`
	UpdatedAt	time.Time	`gorm:"column:updated_at" json:"updated_at"`
	DeletedAt	*time.Time	`gorm:"column:deleted_at" json:"deleted_at,omitempty"`
}

func (Merchant) TableName() string {
	return `"com".merchants`
}

type Product struct {
	ID		uuid.UUID	`gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	MerchantID	uuid.UUID	`gorm:"type:uuid;not null" json:"merchant_id"`
	Name		string		`gorm:"column:name;not null" json:"name"`
	Description	string		`gorm:"column:description" json:"description"`
	Price		float64		`gorm:"column:price;not null" json:"price"`
	Stock		int		`gorm:"column:stock;not null;default:0" json:"stock"`
	ImageURL	string		`gorm:"column:image_url" json:"image_url"`
	IsActive	bool		`gorm:"column:is_active;not null;default:true" json:"is_active"`
	CreatedAt	time.Time	`gorm:"column:created_at" json:"created_at"`
	UpdatedAt	time.Time	`gorm:"column:updated_at" json:"updated_at"`
}

func (Product) TableName() string {
	return `"com".products`
}

type CreateMerchantRequest struct {
	BusinessName	string	`json:"business_name" binding:"required"`
	NIT		string	`json:"nit" binding:"required"`
	Category	string	`json:"category"`
	Address		string	`json:"address"`
	Description	string	`json:"description"`
}

type UpdateMerchantRequest struct {
	BusinessName	string	`json:"business_name"`
	Category	string	`json:"category"`
	Address		string	`json:"address"`
	Description	string	`json:"description"`
	IsActive	*bool	`json:"is_active"`
}

type CreateProductRequest struct {
	Name		string	`json:"name" binding:"required"`
	Description	string	`json:"description"`
	Price		float64	`json:"price" binding:"required"`
	Stock		int	`json:"stock"`
	ImageURL	string	`json:"image_url"`
}

type UpdateProductRequest struct {
	Name		string		`json:"name"`
	Description	string		`json:"description"`
	Price		*float64	`json:"price"`
	Stock		*int		`json:"stock"`
	ImageURL	string		`json:"image_url"`
	IsActive	*bool		`json:"is_active"`
}

type PaginatedResponse struct {
	Success		bool		`json:"success"`
	Message		string		`json:"message"`
	Data		interface{}	`json:"data"`
	Page		int		`json:"page"`
	PageSize	int		`json:"page_size"`
	TotalRows	int64		`json:"total_rows"`
	TotalPages	int		`json:"total_pages"`
}
