// backend/internal/handler/bookhandler/category_handler.go
package bookhandler

import (
	"net/http"
	"strconv"

	// HANYA DIPAKAI DI KOMENTAR → HAPUS IMPORT
	// "backend/internal/models/book"
	"backend/internal/service/bookservice"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categoryService bookservice.CategoryService
}

func NewCategoryHandler(categoryService bookservice.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

// Create
func (h *CategoryHandler) Create(c *gin.Context) {
	var input struct {
		Nama string `json:"nama" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.categoryService.Create(input.Nama)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, category)
}

// GetAll
func (h *CategoryHandler) GetAll(c *gin.Context) {
	categories, err := h.categoryService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, categories)
}

// GetByID
func (h *CategoryHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	category, err := h.categoryService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}
	c.JSON(http.StatusOK, category)
}

// Update
func (h *CategoryHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var input struct {
		Nama *string `json:"nama"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Nama == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nama is required"})
		return
	}

	category, err := h.categoryService.Update(uint(id), *input.Nama)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

// Delete
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.categoryService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Category deleted"})
}