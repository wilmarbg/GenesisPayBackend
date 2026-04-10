package merchants

import (
	"errors"

	"github.com/google/uuid"
)

type MerchantService struct {
	Repo *MerchantRepository
}

func NewMerchantService(repo *MerchantRepository) *MerchantService {
	return &MerchantService{Repo: repo}
}

func (s *MerchantService) Create(req CreateMerchantRequest, role string, authUserID uuid.UUID) (*Merchant, error) {

	if role != "ADMINISTRADOR" && role != "COMERCIO" {
		return nil, errors.New("No tiene permiso para crear un perfil de comercio")
	}

	existing, _ := s.Repo.FindByNIT(req.NIT)
	if existing != nil {
		return nil, errors.New("El NIT ya está registrado")
	}

	merchant := &Merchant{
		UserID:		authUserID,
		BusinessName:	req.BusinessName,
		NIT:		req.NIT,
		Category:	req.Category,
		Address:	req.Address,
		Description:	req.Description,
		IsActive:	true,
	}

	if err := s.Repo.Create(merchant); err != nil {
		return nil, errors.New("Error al crear el comercio")
	}

	return merchant, nil
}

func (s *MerchantService) FindAll(page, pageSize int, category, search string) ([]Merchant, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	return s.Repo.FindAll(page, pageSize, category, search)
}

func (s *MerchantService) FindByID(id uuid.UUID) (*Merchant, error) {
	merchant, err := s.Repo.FindByID(id)
	if err != nil {
		return nil, errors.New("Comercio no encontrado")
	}
	return merchant, nil
}

func (s *MerchantService) FindByUserID(userID uuid.UUID) (*Merchant, error) {
	merchant, err := s.Repo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("Perfil de comercio no encontrado")
	}
	return merchant, nil
}

func (s *MerchantService) Update(id uuid.UUID, req UpdateMerchantRequest, userID uuid.UUID, role string) (*Merchant, error) {
	merchant, err := s.Repo.FindByID(id)
	if err != nil {
		return nil, errors.New("Comercio no encontrado")
	}

	if merchant.UserID != userID && role != "ADMINISTRADOR" {
		return nil, errors.New("No tienes permiso para actualizar este comercio")
	}

	if req.BusinessName != "" {
		merchant.BusinessName = req.BusinessName
	}
	if req.Category != "" {
		merchant.Category = req.Category
	}
	if req.Address != "" {
		merchant.Address = req.Address
	}
	if req.Description != "" {
		merchant.Description = req.Description
	}
	if req.IsActive != nil {
		merchant.IsActive = *req.IsActive
	}

	if err := s.Repo.Update(merchant); err != nil {
		return nil, errors.New("Error al actualizar el comercio")
	}

	return merchant, nil
}

func (s *MerchantService) Delete(id uuid.UUID, role string) error {
	if role != "ADMINISTRADOR" {
		return errors.New("Solo administradores pueden eliminar comercios")
	}

	_, err := s.Repo.FindByID(id)
	if err != nil {
		return errors.New("Comercio no encontrado")
	}

	if err := s.Repo.SoftDelete(id); err != nil {
		return errors.New("Error al eliminar el comercio")
	}
	return nil
}

type ProductService struct {
	Repo		*ProductRepository
	MerchantRepo	*MerchantRepository
}

func NewProductService(repo *ProductRepository, merchantRepo *MerchantRepository) *ProductService {
	return &ProductService{Repo: repo, MerchantRepo: merchantRepo}
}

func (s *ProductService) Create(merchantID uuid.UUID, req CreateProductRequest, userID uuid.UUID, role string) (*Product, error) {
	merchant, err := s.MerchantRepo.FindByID(merchantID)
	if err != nil {
		return nil, errors.New("Comercio no encontrado")
	}

	if merchant.UserID != userID && role != "ADMINISTRADOR" {
		return nil, errors.New("No tienes permiso para agregar productos a este comercio")
	}

	product := &Product{
		MerchantID:	merchantID,
		Name:		req.Name,
		Description:	req.Description,
		Price:		req.Price,
		Stock:		req.Stock,
		ImageURL:	req.ImageURL,
		IsActive:	true,
	}

	if err := s.Repo.Create(product); err != nil {
		return nil, errors.New("Error al crear el producto")
	}

	return product, nil
}

func (s *ProductService) FindByMerchantID(merchantID uuid.UUID, page, pageSize int) ([]Product, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	return s.Repo.FindByMerchantID(merchantID, page, pageSize)
}

func (s *ProductService) Update(productID, merchantID uuid.UUID, req UpdateProductRequest, userID uuid.UUID, role string) (*Product, error) {
	merchant, err := s.MerchantRepo.FindByID(merchantID)
	if err != nil {
		return nil, errors.New("Comercio no encontrado")
	}

	if merchant.UserID != userID && role != "ADMINISTRADOR" {
		return nil, errors.New("No tienes permiso para actualizar productos de este comercio")
	}

	product, err := s.Repo.FindByID(productID)
	if err != nil {
		return nil, errors.New("Producto no encontrado")
	}

	if product.MerchantID != merchantID {
		return nil, errors.New("El producto no pertenece al comercio especificado")
	}

	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
	}
	if req.ImageURL != "" {
		product.ImageURL = req.ImageURL
	}
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	if err := s.Repo.Update(product); err != nil {
		return nil, errors.New("Error al actualizar el producto")
	}

	return product, nil
}

func (s *ProductService) Delete(productID, merchantID uuid.UUID, userID uuid.UUID, role string) error {
	merchant, err := s.MerchantRepo.FindByID(merchantID)
	if err != nil {
		return errors.New("Comercio no encontrado")
	}

	if merchant.UserID != userID && role != "ADMINISTRADOR" {
		return errors.New("No tienes permiso para eliminar productos de este comercio")
	}

	product, err := s.Repo.FindByID(productID)
	if err != nil {
		return errors.New("Producto no encontrado")
	}

	if product.MerchantID != merchantID {
		return errors.New("El producto no pertenece al comercio especificado")
	}

	if err := s.Repo.Delete(productID); err != nil {
		return errors.New("Error al eliminar el producto")
	}
	return nil
}
