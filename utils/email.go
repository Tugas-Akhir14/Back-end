// utils/email.go
package utils

import (
	"fmt"
	"log"
	"net/smtp"
)

func SendApprovalPendingEmail(to, name string) {
	from := "no-reply@hotelmutiara.com"
	subject := "Menunggu Persetujuan Akun Admin"
	body := fmt.Sprintf(`
		<h2>Halo %s,</h2>
		<p>Akun admin Anda telah berhasil dibuat!</p>
		<p>Silakan tunggu persetujuan dari <strong>Superadmin</strong>.</p>
		<br>
		<p>Terima kasih,<br>Tim Hotel Mutiara</p>
	`, name)
	sendEmail(from, to, subject, body)
}

func SendApprovalSuccessEmail(to, name string) {
	from := "no-reply@hotelmutiara.com"
	subject := "Akun Anda Telah Disetujui!"
	body := fmt.Sprintf(`
		<h2>Selamat %s!</h2>
		<p>Akun admin Anda telah <strong>disetujui</strong> oleh Superadmin.</p>
		<p><a href="http://localhost:3000/auth/signin" style="background:#000;color:#fff;padding:10px 20px;text-decoration:none;border-radius:8px;">Login Sekarang</a></p>
		<br>
		<p>Terima kasih,<br>Tim Hotel Mutiara</p>
	`, name)
	sendEmail(from, to, subject, body)
}

func sendEmail(from, to, subject, body string) {
	auth := smtp.PlainAuth("", "YOUR_GMAIL@gmail.com", "YOUR_APP_PASSWORD", "smtp.gmail.com")
	msg := []byte("To: " + to + "\r\n" +
		"From: " + from + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\r\n\r\n" +
		body)

	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, msg)
	if err != nil {
		log.Println("Email gagal:", err)
	}
}