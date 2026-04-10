package payments

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CardRepository struct {
	DB *gorm.DB
}

func NewCardRepository(db *gorm.DB) *CardRepository {
	return &CardRepository{DB: db}
}

func (r *CardRepository) Create(card *Card) error {
	return r.DB.Create(card).Error
}

func (r *CardRepository) FindByID(id uuid.UUID) (*Card, error) {
	var card Card
	if err := r.DB.Where("id = ?", id).First(&card).Error; err != nil {
		return nil, err
	}
	return &card, nil
}

func (r *CardRepository) FindByToken(token string) (*Card, error) {
	var card Card
	if err := r.DB.Where("token = ?", token).First(&card).Error; err != nil {
		return nil, err
	}
	return &card, nil
}

func (r *CardRepository) FindByClientID(clientID uuid.UUID) ([]Card, error) {
	var cards []Card

	if err := r.DB.Where("client_id = ?", clientID).Order("created_at desc").Find(&cards).Error; err != nil {
		return nil, err
	}
	return cards, nil
}

func (r *CardRepository) UpdateStatus(id uuid.UUID, status string) error {
	return r.DB.Model(&Card{}).Where("id = ?", id).Update("status", status).Error
}

type TransactionRepository struct {
	DB *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{DB: db}
}

func (r *TransactionRepository) Create(tx *Transaction) error {
	return r.DB.Create(tx).Error
}

func (r *TransactionRepository) FindByID(id uuid.UUID) (*Transaction, error) {
	var tx Transaction
	if err := r.DB.Where("id = ?", id).First(&tx).Error; err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *TransactionRepository) FindByCardID(cardID uuid.UUID) ([]Transaction, error) {
	var txs []Transaction
	if err := r.DB.Where("card_id = ?", cardID).Order("created_at desc").Find(&txs).Error; err != nil {
		return nil, err
	}
	return txs, nil
}

func (r *TransactionRepository) FindByMerchantID(merchantID uuid.UUID) ([]Transaction, error) {
	var txs []Transaction
	if err := r.DB.Where("merchant_id = ?", merchantID).Order("created_at desc").Find(&txs).Error; err != nil {
		return nil, err
	}
	return txs, nil
}

func (r *TransactionRepository) FindAll() ([]Transaction, error) {
	var txs []Transaction
	if err := r.DB.Order("created_at desc").Find(&txs).Error; err != nil {
		return nil, err
	}
	return txs, nil
}

type AuditLogRepository struct {
	DB *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{DB: db}
}

func (r *AuditLogRepository) Create(log *AuditLog) error {
	return r.DB.Create(log).Error
}
