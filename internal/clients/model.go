package clients

import (
	"time"

	"github.com/google/uuid"
)

type Client struct {
	ID			uuid.UUID	`gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID			uuid.UUID	`gorm:"type:uuid;not null" json:"user_id"`
	FirstName		string		`gorm:"column:first_name;not null" json:"first_name"`
	LastName		string		`gorm:"column:last_name;not null" json:"last_name"`
	DPI			string		`gorm:"column:dpi;uniqueIndex;not null" json:"dpi"`
	Phone			string		`gorm:"column:phone;not null" json:"phone"`
	Address			string		`gorm:"column:address;not null" json:"address"`
	AffiliationStatus	string		`gorm:"column:affiliation_status;not null;default:'PENDIENTE'" json:"affiliation_status"`
	CreditLimit		float64		`gorm:"column:credit_limit;not null;default:0" json:"credit_limit"`
	CreatedAt		time.Time	`gorm:"column:created_at" json:"created_at"`
	UpdatedAt		time.Time	`gorm:"column:updated_at" json:"updated_at"`
	DeletedAt		*time.Time	`gorm:"column:deleted_at" json:"deleted_at,omitempty"`
}

func (Client) TableName() string {
	return `"cli".clients`
}

type CreateClientRequest struct {
	FirstName	string	`json:"first_name" binding:"required"`
	LastName	string	`json:"last_name" binding:"required"`
	DPI		string	`json:"dpi" binding:"required"`
	Phone		string	`json:"phone" binding:"required"`
	Address		string	`json:"address" binding:"required"`
}

type UpdateClientRequest struct {
	FirstName	string	`json:"first_name"`
	LastName	string	`json:"last_name"`
	Phone		string	`json:"phone"`
	Address		string	`json:"address"`
}

type UpdateStatusRequest struct {
	Status		string		`json:"status" binding:"required"`
	CreditLimit	*float64	`json:"credit_limit"`
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
