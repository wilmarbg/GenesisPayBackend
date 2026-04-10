package clients

import (
	"errors"
	"fmt"
	"log"
	"time"

	"genesis-pay-backend/internal/payments"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClientService struct {
	Repo		*ClientRepository
	EncryptKey	string
}

func NewClientService(repo *ClientRepository, encryptKey string) *ClientService {
	return &ClientService{
		Repo:		repo,
		EncryptKey:	encryptKey,
	}
}

func (s *ClientService) Create(userID uuid.UUID, req CreateClientRequest) (*Client, error) {

	existingClient, _ := s.Repo.FindByDPI(req.DPI)
	if existingClient != nil {
		return nil, errors.New("El DPI ya está registrado")
	}

	client := &Client{
		UserID:			userID,
		FirstName:		req.FirstName,
		LastName:		req.LastName,
		DPI:			req.DPI,
		Phone:			req.Phone,
		Address:		req.Address,
		AffiliationStatus:	"PENDIENTE",
		CreditLimit:		0,
	}

	if err := s.Repo.Create(client); err != nil {
		return nil, errors.New("Error al crear el cliente")
	}

	return client, nil
}

func (s *ClientService) FindAll(page, pageSize int, status, search string) ([]Client, int64, error) {

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	return s.Repo.FindAll(page, pageSize, status, search)
}

func (s *ClientService) FindByID(id uuid.UUID) (*Client, error) {
	client, err := s.Repo.FindByID(id)
	if err != nil {
		return nil, errors.New("Cliente no encontrado")
	}
	return client, nil
}

func (s *ClientService) FindByUserID(userID uuid.UUID) (*Client, error) {
	client, err := s.Repo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("Perfil de cliente no encontrado")
	}
	return client, nil
}

func (s *ClientService) Update(id uuid.UUID, req UpdateClientRequest, userID uuid.UUID, role string) (*Client, error) {
	client, err := s.Repo.FindByID(id)
	if err != nil {
		return nil, errors.New("Cliente no encontrado")
	}

	if client.UserID != userID && role != "ADMINISTRADOR" {
		return nil, errors.New("No tienes permiso para actualizar este cliente")
	}

	if req.FirstName != "" {
		client.FirstName = req.FirstName
	}
	if req.LastName != "" {
		client.LastName = req.LastName
	}
	if req.Phone != "" {
		client.Phone = req.Phone
	}
	if req.Address != "" {
		client.Address = req.Address
	}

	if err := s.Repo.Update(client); err != nil {
		return nil, errors.New("Error al actualizar el cliente")
	}

	return client, nil
}

func (s *ClientService) Delete(id uuid.UUID, role string) error {

	if role != "ADMINISTRADOR" {
		return errors.New("No tienes permiso para eliminar clientes")
	}

	_, err := s.Repo.FindByID(id)
	if err != nil {
		return errors.New("Cliente no encontrado")
	}

	if err := s.Repo.SoftDelete(id); err != nil {
		return errors.New("Error al desactivar el cliente")
	}

	return nil
}

func (s *ClientService) UpdateStatus(id uuid.UUID, req UpdateStatusRequest, role string, adminID uuid.UUID) error {

	if role != "ADMINISTRADOR" {
		return errors.New("No tienes permiso para cambiar el estado de afiliación")
	}

	client, err := s.Repo.FindByID(id)
	if err != nil {
		return errors.New("Cliente no encontrado")
	}

	validTransition := false
	switch client.AffiliationStatus {
	case "PENDIENTE":
		if req.Status == "APROBADO" {
			if req.CreditLimit == nil || *req.CreditLimit <= 0 {
				return errors.New("Se requiere un límite de crédito mayor a 0 para aprobar")
			}
			validTransition = true
		} else if req.Status == "RECHAZADO" {
			validTransition = true
		}
	case "APROBADO":
		if req.Status == "ACTIVO" {
			return errors.New("Use el endpoint /activate para activar al cliente y emitir su tarjeta")
		}
	case "ACTIVO":
		if req.Status == "SUSPENDIDO" {
			validTransition = true
		}
	case "SUSPENDIDO":
		if req.Status == "ACTIVO" {
			validTransition = true
		}
	}

	if !validTransition {
		return errors.New("Transición de estado no permitida")
	}

	err = s.Repo.DB.Transaction(func(tx *gorm.DB) error {

		if errUpdate := s.Repo.UpdateStatus(id, req.Status, req.CreditLimit); errUpdate != nil {
			return errors.New("Error al actualizar el estado del cliente")
		}

		if client.AffiliationStatus == "ACTIVO" && req.Status == "SUSPENDIDO" {

			if res := tx.Exec(`UPDATE "pag".cards SET status = 'CONGELADA' WHERE client_id = ?`, id); res.Error != nil {
				return errors.New("Error al congelar la tarjeta asociada")
			}
		} else if client.AffiliationStatus == "SUSPENDIDO" && req.Status == "ACTIVO" {

			if res := tx.Exec(`UPDATE "pag".cards SET status = 'ACTIVA' WHERE client_id = ? AND status = 'CONGELADA'`, id); res.Error != nil {
				return errors.New("Error al reactivar la tarjeta asociada")
			}
		}

		var creditLimit float64
		if req.CreditLimit != nil {
			creditLimit = *req.CreditLimit
		}

		auditDetails := fmt.Sprintf(`{"accion": "CAMBIO_ESTADO_CLIENTE", "nuevo_estado": "%s", "limite_credito": %.2f, "admin_id": "%s"}`, req.Status, creditLimit, adminID.String())

		queryAudit := `INSERT INTO "pag".audit_logs (user_id, action, entity_type, entity_id, details, created_at) 
			VALUES (?, 'UPDATE_CLIENT_STATUS', 'client', ?, ?::jsonb, ?)`

		if err := tx.Exec(queryAudit, adminID, id, auditDetails, time.Now()).Error; err != nil {
			log.Printf("Error registrando auditoría JSONB: %v", err)
			return errors.New("Error al registrar auditoría")
		}

		return nil
	})

	return err
}

func (s *ClientService) Activate(id uuid.UUID, adminID uuid.UUID) (map[string]interface{}, error) {
	log.Println("Activate paso 1: buscando cliente", id)
	client, err := s.Repo.FindByID(id)
	if err != nil {
		return nil, errors.New("Cliente no encontrado")
	}

	log.Println("Activate paso 2: estado del cliente:", client.AffiliationStatus)
	if client.AffiliationStatus != "APROBADO" {
		return nil, errors.New("El cliente debe estar en estado APROBADO para activarlo")
	}

	log.Println("Activate paso 3: credit_limit:", client.CreditLimit)
	if client.CreditLimit <= 0 {
		return nil, errors.New("El cliente debe tener un límite de crédito mayor a 0")
	}

	var cardRes *payments.CardResponse

	err = s.Repo.DB.Transaction(func(tx *gorm.DB) error {

		log.Println("Activate paso 4: verificando tarjeta existente...")
		var existingCount int64
		tx.Table(`"pag".cards`).Where("client_id = ?", id).Count(&existingCount)
		if existingCount > 0 {
			return errors.New("El cliente ya tiene una tarjeta asignada")
		}

		log.Println("Activate paso 5: generando número de tarjeta...")
		number := payments.GenerateCardNumber()

		log.Println("Activate paso 6: cifrando número...")
		encryptedNumber, errEnc := payments.Encrypt(number, s.EncryptKey)
		if errEnc != nil {
			return errors.New("Error al cifrar el número de tarjeta")
		}
		maskedNumber := payments.MaskCardNumber(number)
		token := uuid.New().String()
		cardID := uuid.New()
		expMonth := int(time.Now().Month())
		expYear := time.Now().Year() + 5

		log.Println("Activate paso 7: creando tarjeta en DB...")
		queryCard := `INSERT INTO "pag".cards (id, card_token, client_id, card_number_encrypted, card_number_masked, expiration_month, expiration_year, status, available_balance, created_at, updated_at) 
			VALUES (?, ?, ?, ?, ?, ?, ?, 'ACTIVA', ?, ?, ?)`

		if err := tx.Exec(queryCard, cardID, token, id, encryptedNumber, maskedNumber, expMonth, expYear, client.CreditLimit, time.Now(), time.Now()).Error; err != nil {
			log.Printf("Error INSERT tarjeta: %v", err)
			return errors.New("Error al insertar tarjeta en base de datos")
		}

		log.Println("Activate paso 8: actualizando estado a ACTIVO...")
		if errUpdate := s.Repo.UpdateStatus(id, "ACTIVO", nil); errUpdate != nil {
			return errors.New("Error al cambiar estado del cliente a ACTIVO")
		}

		log.Println("Activate paso 9: registrando audit log...")
		details := fmt.Sprintf(`{"accion": "ACTIVAR_CLIENTE_Y_TARJETA", "balance_inicial": %.2f, "admin_id": "%s"}`, client.CreditLimit, adminID.String())

		queryAudit := `INSERT INTO "pag".audit_logs (user_id, action, entity_type, entity_id, details, created_at) 
			VALUES (?, 'ISSUE_CARD_ACTIVATE', 'client', ?, ?::jsonb, ?)`

		if err := tx.Exec(queryAudit, adminID, id, details, time.Now()).Error; err != nil {
			log.Printf("Error registrando auditoría JSONB: %v", err)
			return errors.New("Error al registrar auditoría")
		}

		cardRes = &payments.CardResponse{
			ID:			cardID,
			Token:			token,
			ClientID:		id,
			MaskedNumber:		maskedNumber,
			ExpirationMonth:	expMonth,
			ExpirationYear:		expYear,
			Status:			"ACTIVA",
			Balance:		client.CreditLimit,
			CreatedAt:		time.Now(),
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"card": map[string]interface{}{
			"card_number_masked":	cardRes.MaskedNumber,
			"available_balance":	cardRes.Balance,
			"status":		cardRes.Status,
			"expiration_month":	cardRes.ExpirationMonth,
			"expiration_year":	cardRes.ExpirationYear,
		},
		"client": map[string]interface{}{
			"first_name":		client.FirstName,
			"last_name":		client.LastName,
			"affiliation_status":	"ACTIVO",
		},
	}, nil
}
