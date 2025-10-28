// backend/internal/handler/souvenirhandler/product_handler.go
package souvenirhandler

import (
    "backend/internal/models/souvenir"
    "backend/internal/service/souvenirservice"
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
)

type ProductHandler struct {
    service souvenirservice.ProductService
}

func NewProductHandler(service souvenirservice.ProductService) *ProductHandler {
    return &ProductHandler{service}
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
    var req souvenir.ProductCreate
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    product, err := h.service.CreateProduct(&req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, gin.H{"data": product})
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    product, err := h.service.GetProductByID(uint(id))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": product})
}

func (h *ProductHandler) GetAllProducts(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

    products, total, err := h.service.GetAllProducts(page, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "data":  products,
        "total": total,
        "page":  page,
        "limit": limit,
    })
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    var req souvenir.ProductUpdate
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    product, err := h.service.UpdateProduct(uint(id), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": product})
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    if err := h.service.DeleteProduct(uint(id)); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

func (h *ProductHandler) GetProductsByCategory(c *gin.Context) {
    categoryID, _ := strconv.ParseUint(c.Param("category_id"), 10, 32)
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

    products, total, err := h.service.GetProductsByCategoryID(uint(categoryID), page, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "data":  products,
        "total": total,
        "page":  page,
        "limit": limit,
    })
}