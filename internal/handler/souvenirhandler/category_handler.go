// backend/internal/handler/souvenirhandler/category_handler.go
package souvenirhandler

import (
    "backend/internal/models/souvenir"
    "backend/internal/service/souvenirservice"
    "github.com/gin-gonic/gin"
    "net/http"
    "strconv"
)

type CategoryHandler struct {
    service souvenirservice.CategoryService
}

func NewCategoryHandler(service souvenirservice.CategoryService) *CategoryHandler {
    return &CategoryHandler{service}
}

func (h *CategoryHandler) Create(c *gin.Context) {
    var req souvenir.CategoryCreate
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    cat, err := h.service.Create(&req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, gin.H{"data": cat})
}

func (h *CategoryHandler) GetAll(c *gin.Context) {
    cats, err := h.service.GetAll()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": cats})
}

func (h *CategoryHandler) GetByID(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    cat, err := h.service.GetByID(uint(id))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": cat})
}

func (h *CategoryHandler) Update(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    var req souvenir.CategoryUpdate
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    cat, err := h.service.Update(uint(id), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": cat})
}

func (h *CategoryHandler) Delete(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    if err := h.service.Delete(uint(id)); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Category deleted"})
}