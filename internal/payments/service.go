package payments

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentService struct {
	DB		*gorm.DB
	CardRepo	*CardRepository
	TxRepo		*TransactionRepository
	AuditRepo	*AuditLogRepository
	EncryptKey	string
}

func NewPaymentService(db *gorm.DB, cr *CardRepository, tr *TransactionRepository, ar *AuditLogRepository, key string) *PaymentService {
	return &PaymentService{
		DB:		db,
		CardRepo:	cr,
		TxRepo:		tr,
		AuditRepo:	ar,
		EncryptKey:	key,
	}
}

func (s *PaymentService) IssueCard(req IssueCardRequest, adminID uuid.UUID, role string) (*CardResponse, error) {
	if role != "ADMINISTRADOR" {
		return nil, errors.New("Solo el administrador puede emitir tarjetas")
	}

	number := GenerateCardNumber()

	encryptedNumber, err := Encrypt(number, s.EncryptKey)
	if err != nil {
		return nil, errors.New("Error al cifrar el número de tarjeta")
	}

	maskedNumber := MaskCardNumber(number)

	token := uuid.New().String()

	card := &Card{
		CardToken:		token,
		ClientID:		req.ClientID,
		CardNumberEncrypted:	encryptedNumber,
		CardNumberMasked:	maskedNumber,
		ExpirationMonth:	int(time.Now().Month()),
		ExpirationYear:		time.Now().Year() + 5,
		Status:			"ACTIVA",
		AvailableBalance:	req.InitialLimit,
	}

	if err := s.CardRepo.Create(card); err != nil {
		return nil, errors.New("Error al crear la tarjeta en la base de datos")
	}

	auditDetails := fmt.Sprintf(`{"accion": "ISSUE_CARD", "monto_inicial": %.2f, "admin_id": "%s"}`, req.InitialLimit, adminID.String())
	audit := &AuditLog{
		UserID:		&adminID,
		Action:		"ISSUE_CARD",
		EntityType:	"card",
		EntityID:	&card.ID,
		Details:	auditDetails,
	}
	_ = s.AuditRepo.Create(audit)

	return &CardResponse{
		ID:		card.ID,
		Token:		card.CardToken,
		ClientID:	card.ClientID,
		MaskedNumber:	card.CardNumberMasked,
		Status:		card.Status,
		Balance:	card.AvailableBalance,
		CreatedAt:	card.CreatedAt,
	}, nil
}

func (s *PaymentService) ProcessPayment(req ProcessPaymentRequest, clientID uuid.UUID) (*TransactionResponse, error) {
	var resultTx *Transaction

	err := s.DB.Transaction(func(tx *gorm.DB) error {

		var card Card
		if req.CardToken != "" {
			if err := tx.Where("card_token = ?", req.CardToken).First(&card).Error; err != nil {
				return errors.New("Tarjeta no encontrada")
			}
		} else {

			var clientIDFromCli uuid.UUID
			err := tx.Table(`"cli".clients`).Select("id").Where("user_id = ?", clientID).Row().Scan(&clientIDFromCli)
			if err != nil {
				return errors.New("No se encontró perfil de cliente para realizar el pago")
			}

			if err := tx.Where("client_id = ? AND status = 'ACTIVA'", clientIDFromCli).Order("created_at desc").First(&card).Error; err != nil {
				return errors.New("No tienes una tarjeta activa para realizar el pago")
			}
		}

		if req.CardToken != "" {
			var authClientID uuid.UUID
			tx.Table(`"cli".clients`).Select("id").Where("user_id = ?", clientID).Row().Scan(&authClientID)
			if card.ClientID != authClientID {
				return errors.New("Tarjeta no autorizada para este usuario")
			}
		}

		var clientStatus string
		err := tx.Table(`"cli".clients`).Select("affiliation_status").Where("id = ?", card.ClientID).Row().Scan(&clientStatus)
		if err != nil {
			return errors.New("No se pudo verificar el estado del cliente")
		}

		if clientStatus == "SUSPENDIDO" {
			return errors.New("Su cuenta está suspendida, no puede realizar compras")
		}
		if clientStatus != "ACTIVO" {
			return errors.New("El cliente no se encuentra en estado ACTIVO")
		}

		if card.Status == "CONGELADA" {
			return errors.New("La tarjeta se encuentra congelada")
		}
		if card.Status != "ACTIVA" {
			return errors.New("La tarjeta no se encuentra activa")
		}

		if card.AvailableBalance < req.Amount {
			return errors.New("Saldo insuficiente para realizar la compra")
		}

		var merchantIsActive bool
		err = tx.Table(`"com".merchants`).Select("is_active").Where("id = ?", req.MerchantID).Row().Scan(&merchantIsActive)
		if err != nil {
			return errors.New("Comercio no encontrado")
		}
		if !merchantIsActive {
			return errors.New("Este comercio no está disponible")
		}

		resCard := tx.Exec(`UPDATE "pag".cards SET available_balance = available_balance - ?, updated_at = ? WHERE id = ? AND available_balance >= ?`, req.Amount, time.Now(), card.ID, req.Amount)
		if resCard.Error != nil {
			return errors.New("Error al descontar saldo de la tarjeta")
		}
		if resCard.RowsAffected == 0 {
			return errors.New("Saldo insuficiente para realizar la compra")
		}

		resMerchant := tx.Exec(`UPDATE "com".merchants SET balance = balance + ?, updated_at = ? WHERE id = ?`, req.Amount, time.Now(), req.MerchantID)
		if resMerchant.Error != nil {
			return errors.New("Error al sumar saldo al comercio")
		}

		if req.ProductID != nil {
			resProduct := tx.Exec(`UPDATE "com".products SET stock = stock - 1, updated_at = ? WHERE id = ? AND stock >= 1 AND is_active = true`, time.Now(), *req.ProductID)
			if resProduct.Error != nil {
				return errors.New("Error al descontar stock del producto")
			}
			if resProduct.RowsAffected == 0 {

				var stock int
				tx.Table(`"com".products`).Select("stock").Where("id = ?", *req.ProductID).Row().Scan(&stock)
				if stock < 1 {
					return errors.New("Producto sin stock disponible")
				}
				return errors.New("El producto no está disponible")
			}
		}

		finalDesc := req.Description
		if finalDesc == "" {
			finalDesc = "Compra GenesisPay"
		}

		transaction := &Transaction{
			CardID:		card.ID,
			MerchantID:	req.MerchantID,
			ProductID:	req.ProductID,
			Amount:		req.Amount,
			Type:		"COMPRA",
			Status:		"COMPLETADA",
			Description:	finalDesc,
		}
		if err := tx.Create(transaction).Error; err != nil {
			return errors.New("Error al registrar la transacción")
		}

		resultTx = transaction

		auditDetails := fmt.Sprintf(`{"accion": "COMPRA_PROCESADA", "monto": %.2f, "merchant_id": "%s", "card_id": "%s"}`, req.Amount, req.MerchantID.String(), card.ID.String())
		audit := &AuditLog{
			UserID:		&clientID,
			Action:		"PROCESS_PAYMENT",
			EntityType:	"transaction",
			EntityID:	&transaction.ID,
			Details:	auditDetails,
		}
		if err := tx.Create(audit).Error; err != nil {
			return errors.New("Error al registrar auditoría")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &TransactionResponse{
		ID:		resultTx.ID,
		CardID:		resultTx.CardID,
		MerchantID:	resultTx.MerchantID,
		Amount:		resultTx.Amount,
		Type:		resultTx.Type,
		Status:		resultTx.Status,
		Description:	resultTx.Description,
		CreatedAt:	resultTx.CreatedAt,
	}, nil
}

func (s *PaymentService) GetTransactions(userID uuid.UUID, role string) ([]TransactionResponse, error) {
	var results []TransactionResponse

	query := s.DB.Table(`"pag".transactions t`).
		Select(`t.*, m.business_name as merchant_name`).
		Joins(`JOIN "com".merchants m ON t.merchant_id = m.id`)

	switch role {
	case "ADMINISTRADOR":
		err := query.Order("t.created_at desc").Scan(&results).Error
		return results, err

	case "CLIENTE":

		var clientID uuid.UUID
		err := s.DB.Table(`"cli".clients`).Select("id").Where("user_id = ?", userID).Row().Scan(&clientID)
		if err != nil {
			return nil, errors.New("No se encontró perfil de cliente para este usuario")
		}

		err = query.Joins(`JOIN "pag".cards c ON t.card_id = c.id`).
			Where("c.client_id = ?", clientID).
			Order("t.created_at desc").
			Scan(&results).Error
		return results, err

	case "COMERCIO":

		err := query.Where("m.user_id = ?", userID).
			Order("t.created_at desc").
			Scan(&results).Error
		return results, err

	default:
		return nil, errors.New("Rol no autorizado para consultar transacciones")
	}
}

func (s *PaymentService) RefundTransaction(txID uuid.UUID, adminID uuid.UUID, role string) (*TransactionResponse, error) {
	if role != "ADMINISTRADOR" {
		return nil, errors.New("Solo administradores pueden realizar reembolsos")
	}

	var resultTx *Transaction

	err := s.DB.Transaction(func(dbTx *gorm.DB) error {

		var originalTx Transaction
		if err := dbTx.Where("id = ?", txID).First(&originalTx).Error; err != nil {
			return errors.New("Transacción no encontrada")
		}

		if originalTx.Type != "COMPRA" || originalTx.Status != "COMPLETADA" {
			return errors.New("Solo se pueden reembolsar COMPRAS COMPLETADAS")
		}

		resCard := dbTx.Exec(`UPDATE "pag".cards SET available_balance = available_balance + ?, updated_at = ? WHERE id = ?`, originalTx.Amount, time.Now(), originalTx.CardID)
		if resCard.Error != nil {
			return errors.New("Error al devolver saldo a la tarjeta")
		}

		resMerchant := dbTx.Exec(`UPDATE "com".merchants SET balance = balance - ?, updated_at = ? WHERE id = ?`, originalTx.Amount, time.Now(), originalTx.MerchantID)
		if resMerchant.Error != nil {
			return errors.New("Error al descontar saldo del comercio")
		}

		if originalTx.ProductID != nil {
			dbTx.Exec(`UPDATE "com".products SET stock = stock + 1, updated_at = ? WHERE id = ?`, time.Now(), *originalTx.ProductID)
		}

		refundTx := &Transaction{
			CardID:		originalTx.CardID,
			MerchantID:	originalTx.MerchantID,
			ProductID:	originalTx.ProductID,
			Amount:		originalTx.Amount,
			Type:		"REEMBOLSO",
			Status:		"COMPLETADA",
			Description:	"Reembolso de transacción " + originalTx.ID.String(),
		}
		if err := dbTx.Create(refundTx).Error; err != nil {
			return errors.New("Error al registrar transacción de reembolso")
		}

		resultTx = refundTx

		auditDetails := fmt.Sprintf(`{"accion": "REEMBOLSO", "monto": %.2f, "original_tx_id": "%s", "admin_id": "%s"}`, originalTx.Amount, originalTx.ID.String(), adminID.String())
		audit := &AuditLog{
			UserID:		&adminID,
			Action:		"REFUND_TRANSACTION",
			EntityType:	"transaction",
			EntityID:	&refundTx.ID,
			Details:	auditDetails,
		}
		_ = dbTx.Create(audit)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &TransactionResponse{
		ID:		resultTx.ID,
		CardID:		resultTx.CardID,
		MerchantID:	resultTx.MerchantID,
		Amount:		resultTx.Amount,
		Type:		resultTx.Type,
		Status:		resultTx.Status,
		Description:	resultTx.Description,
		CreatedAt:	resultTx.CreatedAt,
	}, nil
}

func (s *PaymentService) GetMyCards(userID uuid.UUID) ([]CardResponse, error) {
	log.Println("GET /cards/me - user_id:", userID)

	var clientID uuid.UUID
	err := s.DB.Table(`"cli".clients`).Select("id").Where("user_id = ?", userID).Row().Scan(&clientID)
	if err != nil {
		return nil, errors.New("No se encontró perfil de cliente para este usuario")
	}

	cards, err := s.CardRepo.FindByClientID(clientID)
	if err != nil {
		return nil, errors.New("Error al obtener tarjetas")
	}

	var res []CardResponse
	for _, c := range cards {
		res = append(res, CardResponse{
			ID:			c.ID,
			Token:			c.CardToken,
			ClientID:		c.ClientID,
			MaskedNumber:		c.CardNumberMasked,
			ExpirationMonth:	c.ExpirationMonth,
			ExpirationYear:		c.ExpirationYear,
			Status:			c.Status,
			Balance:		c.AvailableBalance,
			CreatedAt:		c.CreatedAt,
		})
	}
	return res, nil
}

func (s *PaymentService) GetCardBalance(cardID uuid.UUID, userID uuid.UUID, role string) (float64, error) {
	card, err := s.CardRepo.FindByID(cardID)
	if err != nil {
		return 0, errors.New("Tarjeta no encontrada")
	}

	if card.ClientID != userID && role != "ADMINISTRADOR" {
		return 0, errors.New("No tienes permiso para ver el saldo de esta tarjeta")
	}

	return card.AvailableBalance, nil
}

func (s *PaymentService) UpdateCardStatus(cardID uuid.UUID, status string, adminID uuid.UUID, role string) error {
	if role != "ADMINISTRADOR" {
		return errors.New("No tienes permiso")
	}

	_, err := s.CardRepo.FindByID(cardID)
	if err != nil {
		return errors.New("Tarjeta no encontrada")
	}

	if err := s.CardRepo.UpdateStatus(cardID, status); err != nil {
		return errors.New("Error al actualizar el estado")
	}

	auditDetails := fmt.Sprintf(`{"accion": "CAMBIO_ESTADO_TARJETA", "nuevo_estado": "%s", "admin_id": "%s"}`, status, adminID.String())
	audit := &AuditLog{
		UserID:		&adminID,
		Action:		"UPDATE_CARD_STATUS",
		EntityType:	"card",
		EntityID:	&cardID,
		Details:	auditDetails,
	}
	_ = s.AuditRepo.Create(audit)
	return nil
}

func (s *PaymentService) GetTransactionByID(txID uuid.UUID, userID uuid.UUID, role string) (*TransactionResponse, error) {
	tx, err := s.TxRepo.FindByID(txID)
	if err != nil {
		return nil, errors.New("Transacción no encontrada")
	}

	return &TransactionResponse{
		ID:		tx.ID,
		CardID:		tx.CardID,
		MerchantID:	tx.MerchantID,
		Amount:		tx.Amount,
		Type:		tx.Type,
		Status:		tx.Status,
		Description:	tx.Description,
		CreatedAt:	tx.CreatedAt,
	}, nil
}
