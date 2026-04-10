package clients

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClientRepository struct {
	DB *gorm.DB
}

func NewClientRepository(db *gorm.DB) *ClientRepository {
	return &ClientRepository{DB: db}
}

func (r *ClientRepository) Create(client *Client) error {
	return r.DB.Create(client).Error
}

func (r *ClientRepository) FindByID(id uuid.UUID) (*Client, error) {
	var client Client
	err := r.DB.Where("id = ? AND deleted_at IS NULL", id).First(&client).Error
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *ClientRepository) FindByUserID(userID uuid.UUID) (*Client, error) {
	var client Client
	err := r.DB.Where("user_id = ? AND deleted_at IS NULL", userID).First(&client).Error
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *ClientRepository) FindByDPI(dpi string) (*Client, error) {
	var client Client
	err := r.DB.Where("dpi = ? AND deleted_at IS NULL", dpi).First(&client).Error
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *ClientRepository) FindAll(page, pageSize int, status, search string) ([]Client, int64, error) {
	var clients []Client
	var totalRows int64

	query := r.DB.Model(&Client{}).Where("deleted_at IS NULL")

	if status != "" {
		query = query.Where("affiliation_status = ?", status)
	}

	if search != "" {
		searchQuery := "%" + search + "%"
		query = query.Where("first_name ILIKE ? OR last_name ILIKE ? OR dpi ILIKE ?", searchQuery, searchQuery, searchQuery)
	}

	if err := query.Count(&totalRows).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Find(&clients).Error
	if err != nil {
		return nil, 0, err
	}

	return clients, totalRows, nil
}

func (r *ClientRepository) Update(client *Client) error {
	return r.DB.Save(client).Error
}

func (r *ClientRepository) SoftDelete(id uuid.UUID) error {
	now := time.Now()

	return r.DB.Model(&Client{}).Where("id = ?", id).Update("deleted_at", &now).Error
}

func (r *ClientRepository) UpdateStatus(id uuid.UUID, status string, creditLimit *float64) error {
	updates := map[string]interface{}{
		"affiliation_status":	status,
		"updated_at":		time.Now(),
	}
	if creditLimit != nil {
		updates["credit_limit"] = *creditLimit
	}

	return r.DB.Model(&Client{}).Where("id = ?", id).Updates(updates).Error
}
