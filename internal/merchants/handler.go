package merchants

import (
	"math"
	"net/http"
	"strconv"

	"genesis-pay-backend/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MerchantHandler struct {
	MerchantService	*MerchantService
	ProductService	*ProductService
}

func NewMerchantHandler(ms *MerchantService, ps *ProductService) *MerchantHandler {
	return &MerchantHandler{MerchantService: ms, ProductService: ps}
}

func (h *MerchantHandler) CreateMerchant(c *gin.Context) {
	role := c.GetString("role")

	var req CreateMerchantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Datos obligatorios faltantes"})
		return
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))
	merchant, err := h.MerchantService.Create(req, role, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Comercio creado", "data": merchant})
}

func (h *MerchantHandler) GetMerchants(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	category := c.Query("category")
	search := c.Query("search")

	merchants, total, err := h.MerchantService.FindAll(page, pageSize, category, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error al obtener comercios"})
		return
	}

	c.JSON(http.StatusOK, PaginatedResponse{
		Success:	true,
		Message:	"Comercios obtenidos de forma exitosa",
		Data:		merchants,
		Page:		page,
		PageSize:	pageSize,
		TotalRows:	total,
		TotalPages:	int(math.Ceil(float64(total) / float64(pageSize))),
	})
}

func (h *MerchantHandler) GetMe(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, _ := uuid.Parse(userIDStr)

	merchant, err := h.MerchantService.FindByUserID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "No tiene perfil de comercio"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Perfil de comercio obtenido", "data": merchant})
}

func (h *MerchantHandler) GetMerchant(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ID inválido"})
		return
	}

	merchant, err := h.MerchantService.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": merchant})
}

func (h *MerchantHandler) UpdateMerchant(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ID inválido"})
		return
	}

	var req UpdateMerchantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Datos inválidos"})
		return
	}

	userIDStr := c.GetString("user_id")
	userID, _ := uuid.Parse(userIDStr)
	role := c.GetString("role")

	merchant, err := h.MerchantService.Update(id, req, userID, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Comercio actualizado", "data": merchant})
}

func (h *MerchantHandler) DeleteMerchant(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ID inválido"})
		return
	}
	role := c.GetString("role")

	if err := h.MerchantService.Delete(id, role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Comercio eliminado/desactivado"})
}

func (h *MerchantHandler) CreateProduct(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ID inválido"})
		return
	}

	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Datos inválidos"})
		return
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))
	role := c.GetString("role")

	product, err := h.ProductService.Create(merchantID, req, userID, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Producto creado exitosamente", "data": product})
}

func (h *MerchantHandler) GetProducts(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ID inválido"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	products, total, err := h.ProductService.FindByMerchantID(merchantID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error al obtener productos"})
		return
	}

	c.JSON(http.StatusOK, PaginatedResponse{
		Success:	true,
		Message:	"Productos obtenidos",
		Data:		products,
		Page:		page,
		PageSize:	pageSize,
		TotalRows:	total,
		TotalPages:	int(math.Ceil(float64(total) / float64(pageSize))),
	})
}

func (h *MerchantHandler) UpdateProduct(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("id"))
	productID, err2 := uuid.Parse(c.Param("pid"))
	if err != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ID inválido"})
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Datos inválidos"})
		return
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))
	role := c.GetString("role")

	product, err := h.ProductService.Update(productID, merchantID, req, userID, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Producto actualizado", "data": product})
}

func (h *MerchantHandler) DeleteProduct(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("id"))
	productID, err2 := uuid.Parse(c.Param("pid"))
	if err != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ID inválido"})
		return
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))
	role := c.GetString("role")

	if err := h.ProductService.Delete(productID, merchantID, userID, role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Producto eliminado"})
}

func RegisterMerchantRoutes(router *gin.RouterGroup, handler *MerchantHandler, jwtSecret string) {
	authMw := middleware.AuthRequired(jwtSecret)
	adminMw := middleware.RequireRole("ADMINISTRADOR")

	router.GET("", handler.GetMerchants)
	router.GET("/me", authMw, handler.GetMe)
	router.GET("/:id", handler.GetMerchant)
	router.GET("/:id/products", handler.GetProducts)

	router.POST("", authMw, handler.CreateMerchant)
	router.PUT("/:id", authMw, handler.UpdateMerchant)
	router.DELETE("/:id", authMw, adminMw, handler.DeleteMerchant)

	router.POST("/:id/products", authMw, handler.CreateProduct)
	router.PUT("/:id/products/:pid", authMw, handler.UpdateProduct)
	router.DELETE("/:id/products/:pid", authMw, handler.DeleteProduct)
}
