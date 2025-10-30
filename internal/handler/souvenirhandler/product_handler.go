// backend/internal/handler/souvenirhandler/product_handler.go
package souvenirhandler

import (
    "backend/internal/models/souvenir"
    "backend/internal/service/souvenirservice"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
)

type ProductHandler struct {
    service souvenirservice.ProductService
}

func NewProductHandler(service souvenirservice.ProductService) *ProductHandler {
    return &ProductHandler{service}
}

// getBaseURL helper
func getBaseURL() string {
    baseURL := os.Getenv("BASE_URL")
    if baseURL == "" {
        baseURL = "http://localhost:8080"
    }
    return baseURL
}

// CreateProduct
func (h *ProductHandler) CreateProduct(c *gin.Context) {
    var req souvenir.ProductCreate
    if err := c.ShouldBind(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data: " + err.Error()})
        return
    }

    form, err := c.MultipartForm()
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "No files uploaded"})
        return
    }

    fileHeaders := form.File["gambar"]
    if len(fileHeaders) == 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "At least one image required"})
        return
    }

    uploadDir := "./uploads"
    if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload dir"})
        return
    }

    baseURL := getBaseURL()
    var imagePaths []string

    for _, fileHeader := range fileHeaders {
        if fileHeader.Size > 5*1024*1024 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "File too large (max 5MB)"})
            return
        }

        ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
        if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Only JPG/PNG/WEBP allowed"})
            return
        }

        file, err := fileHeader.Open()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
            return
        }
        defer file.Close()

        timestamp := time.Now().UnixNano()
        filename := fmt.Sprintf("%d_%s", timestamp, fileHeader.Filename)
        savePath := filepath.Join(uploadDir, filename)

        dst, err := os.Create(savePath)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
            return
        }
        defer dst.Close()

        if _, err := io.Copy(dst, file); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to copy file"})
            return
        }

        imagePaths = append(imagePaths, baseURL+"/uploads/"+filename)
    }

    req.Gambar = strings.Join(imagePaths, ",")

    product, err := h.service.CreateProduct(&req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"data": product})
}

// UpdateProduct
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
        return
    }

    var req souvenir.ProductUpdate
    if err := c.ShouldBind(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    form, err := c.MultipartForm()
    if err == nil && form != nil {
        fileHeaders := form.File["gambar"]
        if len(fileHeaders) > 0 {
            uploadDir := "./uploads"
            os.MkdirAll(uploadDir, os.ModePerm)

            baseURL := getBaseURL()
            var imagePaths []string

            for _, fileHeader := range fileHeaders {
                if fileHeader.Size > 5*1024*1024 {
                    continue
                }

                ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
                if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
                    continue
                }

                file, err := fileHeader.Open()
                if err != nil {
                    continue
                }
                defer file.Close()

                timestamp := time.Now().UnixNano()
                filename := fmt.Sprintf("%d_%s", timestamp, fileHeader.Filename)
                savePath := filepath.Join(uploadDir, filename)

                dst, err := os.Create(savePath)
                if err != nil {
                    continue
                }
                defer dst.Close()

                io.Copy(dst, file)
                imagePaths = append(imagePaths, baseURL+"/uploads/"+filename)
            }

            if len(imagePaths) > 0 {
                gambarStr := strings.Join(imagePaths, ",")
                req.Gambar = &gambarStr
            }
        }
    }

    product, err := h.service.UpdateProduct(uint(id), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": product})
}

// === Existing Methods (tidak diubah) ===
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