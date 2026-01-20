// backend/internal/handler/cafehandler/product_handler.go
package cafehandler

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"backend/internal/models/cafe"
	"backend/internal/service/cafeservice"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

    type ProductHandler struct {
        productService  cafeservice.ProductService
        categoryService cafeservice.CategoryService
    }

    func NewProductHandler(productService cafeservice.ProductService, categoryService cafeservice.CategoryService) *ProductHandler {
        return &ProductHandler{productService, categoryService}
    }

    // CREATE + IMAGE
    func (h *ProductHandler) CreateProduct(c *gin.Context) {
        var input struct {
            Nama       string  `form:"nama" binding:"required"`
            Deskripsi  string  `form:"deskripsi"`
            Harga      float64 `form:"harga" binding:"required"`
            Stok       int     `form:"stok" binding:"required,min=0"`
            CategoryID uint    `form:"category_id" binding:"required"`
        }

        if err := c.ShouldBind(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        gambarPath, err := h.uploadImage(c)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        createInput := cafe.ProductCafeCreate{
            Nama:       input.Nama,
            Deskripsi:  input.Deskripsi,
            Harga:      input.Harga,
            Stok:       input.Stok,
            CategoryID: input.CategoryID,
            Gambar:     gambarPath,
        }

        product, err := h.productService.Create(createInput)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusCreated, product)
    }

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	// VALIDASI ID
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var input cafe.ProductCafeUpdate

	// === AMBIL FIELD MANUAL ===
	if v := c.PostForm("nama"); v != "" {
		input.Nama = &v
	}

	if v := c.PostForm("deskripsi"); v != "" {
		input.Deskripsi = &v
	}

	if v := c.PostForm("harga"); v != "" {
		harga, err := strconv.ParseFloat(v, 64)
		if err != nil || harga <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "harga tidak valid"})
			return
		}
		input.Harga = &harga
	}

	if v := c.PostForm("stok"); v != "" {
		stok, err := strconv.Atoi(v)
		if err != nil || stok < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "stok tidak valid"})
			return
		}
		input.Stok = &stok
	}

	if v := c.PostForm("category_id"); v != "" {
		cid, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "category_id tidak valid"})
			return
		}
		categoryID := uint(cid)
		input.CategoryID = &categoryID
	}

	// === GAMBAR (OPTIONAL) ===
	if file, err := c.FormFile("gambar"); err == nil {
		path, err := h.uploadImageFromFile(c, file)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		input.Gambar = &path
	}

	// === UPDATE ===
	product, err := h.productService.Update(uint(id), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}



    func (h *ProductHandler) ListProducts(c *gin.Context) {
        var categoryID *uint
        if idStr := c.Query("category_id"); idStr != "" {
            if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
                uid := uint(id)
                categoryID = &uid
            }
        }

        products, err := h.productService.GetAll(categoryID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, products)
    }

    func (h *ProductHandler) GetProduct(c *gin.Context) {
        id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
        product, err := h.productService.GetByID(uint(id))
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
            return
        }
        c.JSON(http.StatusOK, product)
    }

    func (h *ProductHandler) DeleteProduct(c *gin.Context) {
        id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
        if err := h.productService.Delete(uint(id)); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"message": "Product deleted"})
    }

    func (h *ProductHandler) GetProductsByCategory(c *gin.Context) {
        idStr := c.Param("category_id")
        if idStr == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "category_id is required"})
            return
        }
        id, err := strconv.ParseUint(idStr, 10, 32)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
            return
        }
        categoryID := uint(id)
        products, err := h.productService.GetAll(&categoryID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, products)
    }

    // UPLOAD HELPERS
    func (h *ProductHandler) uploadImage(c *gin.Context) (string, error) {
        file, err := c.FormFile("gambar")
        if err != nil {
            return "", fmt.Errorf("gambar wajib diunggah")
        }
        return h.uploadImageFromFile(c, file)
    }

    func (h *ProductHandler) uploadImageFromFile(c *gin.Context, file *multipart.FileHeader) (string, error) {
        ext := strings.ToLower(filepath.Ext(file.Filename))
        if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
            return "", fmt.Errorf("format gambar tidak didukung")
        }
        if file.Size > 5*1024*1024 {
            return "", fmt.Errorf("gambar maksimal 5MB")
        }
        filename := uuid.New().String() + ext
        path := filepath.Join("uploads", filename)
        if err := c.SaveUploadedFile(file, path); err != nil {
            return "", err
        }
        return "/uploads/" + filename, nil
    }

    func (h *ProductHandler) ListPublic(c *gin.Context) {
        var categoryID *uint
        if idStr := c.Query("category_id"); idStr != "" {
            if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
                uid := uint(id)
                categoryID = &uid
            }
        }

        products, err := h.productService.GetAll(categoryID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, products)
    }

    func (h *ProductHandler) GetPublicByID(c *gin.Context) {
        id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
        product, err := h.productService.GetByID(uint(id))
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
            return
        }
        c.JSON(http.StatusOK, product)
    }

    func (h *ProductHandler) GetPublicByCategory(c *gin.Context) {
        idStr := c.Param("category_id")
        if idStr == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "category_id is required"})
            return
        }
        id, err := strconv.ParseUint(idStr, 10, 32)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
            return
        }
        categoryID := uint(id)
        products, err := h.productService.GetAll(&categoryID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, products)
    }