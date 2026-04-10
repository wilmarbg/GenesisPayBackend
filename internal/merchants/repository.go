package merchants

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MerchantRepository struct {
	DB *gorm.DB
}

func NewMerchantRepository(db *gorm.DB) *MerchantRepository {
	return &MerchantRepository{DB: db}
}

func (r *MerchantRepository) Create(merchant *Merchant) error {
	return r.DB.Create(merchant).Error
}

func (r *MerchantRepository) FindByID(id uuid.UUID) (*Merchant, error) {
	var merchant Merchant
	err := r.DB.Where("id = ? AND deleted_at IS NULL", id).First(&merchant).Error
	if err != nil {
		return nil, err
	}
	return &merchant, nil
}

func (r *MerchantRepository) FindByNIT(nit string) (*Merchant, error) {
	var merchant Merchant
	err := r.DB.Where("nit = ? AND deleted_at IS NULL", nit).First(&merchant).Error
	if err != nil {
		return nil, err
	}
	return &merchant, nil
}

func (r *MerchantRepository) FindByUserID(userID uuid.UUID) (*Merchant, error) {
	var merchant Merchant
	err := r.DB.Where("user_id = ? AND deleted_at IS NULL", userID).First(&merchant).Error
	if err != nil {
		return nil, err
	}
	return &merchant, nil
}

func (r *MerchantRepository) FindAll(page, pageSize int, category, search string) ([]Merchant, int64, error) {
	var merchants []Merchant
	var totalRows int64

	query := r.DB.Model(&Merchant{}).Where("deleted_at IS NULL AND is_active = ?", true)

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if search != "" {
		searchQuery := "%" + search + "%"
		query = query.Where("business_name ILIKE ? OR nit ILIKE ?", searchQuery, searchQuery)
	}

	if err := query.Count(&totalRows).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Find(&merchants).Error
	if err != nil {
		return nil, 0, err
	}

	return merchants, totalRows, nil
}

func (r *MerchantRepository) Update(merchant *Merchant) error {
	return r.DB.Save(merchant).Error
}

func (r *MerchantRepository) SoftDelete(id uuid.UUID) error {
	now := time.Now()
	return r.DB.Model(&Merchant{}).Where("id = ?", id).Update("deleted_at", &now).Error
}

func (r *MerchantRepository) UpdateBalance(id uuid.UUID, amount float64) error {
	return r.DB.Model(&Merchant{}).Where("id = ?", id).Update("balance", gorm.Expr("balance + ?", amount)).Error
}

type ProductRepository struct {
	DB *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{DB: db}
}

func (r *ProductRepository) Create(product *Product) error {
	return r.DB.Create(product).Error
}

func (r *ProductRepository) FindByID(id uuid.UUID) (*Product, error) {
	var product Product
	err := r.DB.Where("id = ?", id).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) FindByMerchantID(merchantID uuid.UUID, page, pageSize int) ([]Product, int64, error) {
	var products []Product
	var totalRows int64

	query := r.DB.Model(&Product{}).Where("merchant_id = ? AND is_active = ?", merchantID, true)

	if err := query.Count(&totalRows).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Find(&products).Error
	if err != nil {
		return nil, 0, err
	}

	return products, totalRows, nil
}

func (r *ProductRepository) Update(product *Product) error {
	return r.DB.Save(product).Error
}

func (r *ProductRepository) Delete(id uuid.UUID) error {
	return r.DB.Where("id = ?", id).Delete(&Product{}).Error
}

func (r *ProductRepository) UpdateStock(id uuid.UUID, quantity int) error {
	return r.DB.Model(&Product{}).Where("id = ?", id).Update("stock", gorm.Expr("stock + ?", quantity)).Error
}
