// utils/email.go
package utils

import (
	"fmt"
	"log"
	"net/smtp"
	"os"

	"backend/internal/config"
)

var (
	smtpHost string
	smtpPort string
	smtpUser string
	smtpPass string
	from     = "Hotel Mutiara <no-reply@hotelmutiara.com>"
	loaded   bool
)

// Load config sekali saja, saat pertama kali dipakai
func loadSMTPConfig() {
	if loaded {
		return
	}
	cfg := config.GetConfig()
	smtpHost = cfg.SMTPHost
	smtpPort = cfg.SMTPPort
	smtpUser = cfg.SMTPUser
	smtpPass = cfg.SMTPPass
	loaded = true
}

func SendApprovalPendingEmail(to, name string) {
	loadSMTPConfig() // baru load saat benar-benar kirim email

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	subject := "Menunggu Persetujuan Akun Admin"
	body := fmt.Sprintf(`
		<h2>Halo %s,</h2>
		<p>Akun admin Anda telah berhasil dibuat!</p>
		<p>Sedang <strong>menunggu persetujuan Superadmin</strong>.</p>
		<br>
		<p>Terima kasih,<br>Tim Hotel Mutiara</p>
	`, name)

	msg := []byte("To: " + to + "\r\n" +
		"From: " + from + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\r\n\r\n" + body)

	go func() {
		addr := smtpHost + ":" + smtpPort
		if err := smtp.SendMail(addr, auth, smtpUser, []string{to}, msg); err != nil {
			log.Printf("[EMAIL GAGAL] %s → %v", to, err)
		} else {
			log.Printf("[EMAIL BERHASIL] → %s (Pending Approval)", to)
		}
	}()
}

func SendApprovalSuccessEmail(to, name string) {
	loadSMTPConfig()

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	subject := "Akun Anda Telah Disetujui!"
	body := fmt.Sprintf(`
		<h2>Selamat %s!</h2>
		<p>Akun admin Anda telah <strong>disetujui</strong> oleh Superadmin.</p>
		<p style="margin:30px 0;">
			<a href="https://hotelmutiara.vercel.app/auth/signin"
			   style="background:#f59e0b;color:white;padding:14px 32px;text-decoration:none;border-radius:10px;font-weight:bold;">
			   Login Sekarang
			</a>
		</p>
		<br>
		<p>Terima kasih,<br>Tim Hotel Mutiara</p>
	`, name)

	msg := []byte("To: " + to + "\r\n" +
		"From: " + from + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\r\n\r\n" + body)

	go func() {
		addr := smtpHost + ":" + smtpPort
		if err := smtp.SendMail(addr, auth, smtpUser, []string{to}, msg); err != nil {
			log.Printf("[EMAIL GAGAL] %s → %v", to, err)
		} else {
			log.Printf("[EMAIL BERHASIL] → %s (Approved)", to)
		}
	}()
}

// BARU: kirim OTP ke guest
func SendGuestOTPEmail(toEmail, fullName, otp string) {
	from := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	if from == "" || pass == "" {
		log.Println("SMTP tidak dikonfigurasi, OTP tidak dikirim")
		return
	}

	msg := fmt.Sprintf(`From: %s
To: %s
Subject: Kode Verifikasi Akun Guest

Halo %s,

Terima kasih telah mendaftar sebagai Guest.

Kode OTP Anda adalah: <b>%s</b>

Kode ini berlaku selama 5 menit.

Jika Anda tidak mendaftar, abaikan email ini.

Terima kasih,
Tim Wisata
`, from, toEmail, fullName, otp)

	auth := smtp.PlainAuth("", from, pass, smtpHost)
	addr := smtpHost + ":" + smtpPort

	go func() {
		if err := smtp.SendMail(addr, auth, from, []string{toEmail}, []byte(msg)); err != nil {
			log.Printf("Gagal kirim OTP ke %s: %v", toEmail, err)
		} else {
			log.Printf("OTP berhasil dikirim ke %s", toEmail)
		}
	}()
}