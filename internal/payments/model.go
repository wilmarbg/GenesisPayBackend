package payments

import (
	"time"

	"github.com/google/uuid"
)

type Card struct {
	ID			uuid.UUID	`gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CardToken		string		`gorm:"column:card_token;uniqueIndex;not null" json:"token"`
	ClientID		uuid.UUID	`gorm:"type:uuid;not null" json:"client_id"`
	CardNumberEncrypted	[]byte		`gorm:"column:card_number_encrypted;not null" json:"-"`
	CardNumberMasked	string		`gorm:"column:card_number_masked;not null" json:"masked_number"`
	ExpirationMonth		int		`gorm:"column:expiration_month;not null" json:"expiration_month"`
	ExpirationYear		int		`gorm:"column:expiration_year;not null" json:"expiration_year"`
	Status			string		`gorm:"column:status;not null;default:'ACTIVA'" json:"status"`
	AvailableBalance	float64		`gorm:"column:available_balance;not null;default:0" json:"balance"`
	CreatedAt		time.Time	`gorm:"column:created_at" json:"created_at"`
	UpdatedAt		time.Time	`gorm:"column:updated_at" json:"updated_at"`
}

func (Card) TableName() string	{ return `"pag".cards` }

type Transaction struct {
	ID		uuid.UUID	`gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CardID		uuid.UUID	`gorm:"type:uuid;not null" json:"card_id"`
	MerchantID	uuid.UUID	`gorm:"type:uuid;not null" json:"merchant_id"`
	ProductID	*uuid.UUID	`gorm:"type:uuid" json:"product_id,omitempty"`
	Amount		float64		`gorm:"column:amount;not null" json:"amount"`
	Type		string		`gorm:"column:type;not null" json:"type"`
	Status		string		`gorm:"column:status;not null" json:"status"`
	Description	string		`gorm:"column:description" json:"description"`
	CreatedAt	time.Time	`gorm:"column:created_at" json:"created_at"`
}

func (Transaction) TableName() string	{ return `"pag".transactions` }

type AuditLog struct {
	ID		int64		`gorm:"primaryKey;autoIncrement" json:"id"`
	UserID		*uuid.UUID	`gorm:"type:uuid" json:"user_id"`
	Action		string		`gorm:"column:action;not null" json:"action"`
	EntityType	string		`gorm:"column:entity_type;not null" json:"entity_type"`
	EntityID	*uuid.UUID	`gorm:"type:uuid" json:"entity_id"`
	Details		string		`gorm:"column:details" json:"details"`
	IPAddress	*string		`gorm:"column:ip_address" json:"ip_address"`
	CreatedAt	time.Time	`gorm:"column:created_at" json:"created_at"`
}

func (AuditLog) TableName() string	{ return `"pag".audit_logs` }

type IssueCardRequest struct {
	ClientID	uuid.UUID	`json:"client_id" binding:"required"`
	InitialLimit	float64		`json:"initial_limit"`
}

type ProcessPaymentRequest struct {
	CardToken	string		`json:"card_token"`
	MerchantID	uuid.UUID	`json:"merchant_id" binding:"required"`
	ProductID	*uuid.UUID	`json:"product_id"`
	Amount		float64		`json:"amount" binding:"required"`
	Description	string		`json:"description"`
}

type CardResponse struct {
	ID		uuid.UUID	`json:"id"`
	Token		string		`json:"token"`
	ClientID	uuid.UUID	`json:"client_id"`
	MaskedNumber	string		`json:"masked_number"`
	ExpirationMonth	int		`json:"expiration_month"`
	ExpirationYear	int		`json:"expiration_year"`
	Status		string		`json:"status"`
	Balance		float64		`json:"balance"`
	CreatedAt	time.Time	`json:"created_at"`
}

type TransactionResponse struct {
	ID		uuid.UUID	`json:"id"`
	CardID		uuid.UUID	`json:"card_id"`
	MerchantID	uuid.UUID	`json:"merchant_id"`
	MerchantName	string		`json:"merchant_name,omitempty"`
	Amount		float64		`json:"amount"`
	Type		string		`json:"type"`
	Status		string		`json:"status"`
	Description	string		`json:"description"`
	CreatedAt	time.Time	`json:"created_at"`
}
