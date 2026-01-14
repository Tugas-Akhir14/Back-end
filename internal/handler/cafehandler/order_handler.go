package cafehandler

import (
	"backend/internal/models/cafe"
	"backend/internal/service/cafeservice"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService cafeservice.OrderService
}

func NewOrderHandler(orderService cafeservice.OrderService) *OrderHandler {
	return &OrderHandler{orderService}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var input cafe.OrderCafeCreate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orderService.Create(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate link wa.me untuk notifikasi ke admin
	waLink := h.generateWaLink(order)

	// Response dengan tambahan wa_link
	c.JSON(http.StatusCreated, gin.H{
		"order":   order,
		"wa_link": waLink,
		"message": "Pesanan berhasil dibuat. Gunakan link wa_link untuk mengirim notifikasi ke admin WhatsApp",
	})
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
	orders, err := h.orderService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	order, err := h.orderService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) UpdateOrder(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var input cafe.OrderCafeUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orderService.Update(uint(id), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

// generateWaLink membuat link wa.me dengan pesan pesanan yang sudah terformat
func (h *OrderHandler) generateWaLink(order *cafe.OrderCafe) string {
	var itemsStr strings.Builder
	for _, item := range order.Items {
		subtotal := item.Price * float64(item.Quantity)
		itemsStr.WriteString(fmt.Sprintf(
			"- %s x %d (@ Rp %.0f) = Rp %.0f\n",
			item.Product.Nama,
			item.Quantity,
			item.Price,
			subtotal,
		))
	}

	// Format pesan yang rapi dan mudah dibaca di WhatsApp
	message := fmt.Sprintf(
		"*PESANAN BARU #%d*\n\n"+
			"👤 Nama: %s\n"+
			"🪑 Meja: %s\n\n"+
			"*Daftar Pesanan:*\n"+
			"%s"+
			"\n💰 *Total:* Rp %.0f\n"+
			"📊 Status: %s\n\n"+
			"Silakan proses pesanan ini secepatnya ya! ☕\n"+
			"Terima kasih! 🙏",
		order.ID,
		order.CustomerName,
		itemsStr.String(),
		order.TotalPrice,
		order.Status,
	)

	// Encode agar aman di URL
	encodedMsg := url.QueryEscape(message)

	// Nomor tujuan dari env atau fallback
	waNumber := os.Getenv("CAFE_WHATSAPP_NUMBER")
	if waNumber == "" {
		waNumber = "6281396554949" // fallback sesuai permintaan kamu
	}

	// Link wa.me resmi
	return fmt.Sprintf("https://wa.me/%s?text=%s", waNumber, encodedMsg)
}

// Helper untuk handle table_number yang mungkin nil
func getTableNumber(ptr *string) string {
	if ptr == nil || *ptr == "" {
		return "- (Tanpa Meja)"
	}
	return *ptr
}


// ConfirmPayment - Khusus konfirmasi pembayaran oleh admin
func (h *OrderHandler) ConfirmPayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	order, err := h.orderService.ConfirmPayment(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Optional: Generate wa_link lagi untuk konfirmasi ke customer (jika perlu)
	waLink := h.generateWaLink(order) // reuse fungsi yang sudah ada

	c.JSON(http.StatusOK, gin.H{
		"order":   order,
		"wa_link": waLink,
		"message": "Pembayaran berhasil dikonfirmasi. Status order sekarang: paid",
	})
}

// CancelOrder - Membatalkan pesanan oleh admin
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	order, err := h.orderService.CancelOrder(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order":   order,
		"message": "Pesanan berhasil dibatalkan",
	})
}