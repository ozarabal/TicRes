package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"ticres/internal/repository" // Pastikan import path ini sesuai project Anda
)

// 1. Definisikan Tipe Job (Notification Biasa vs Refund)
type JobType int

const (
	JobNotification JobType = iota
	JobRefund
)

// 2. Update Payload agar bisa menangani Event Cancellation
type NotificationPayload struct {
	Type      JobType
	
	// Untuk Notification Biasa
	BookingID int64
	UserEmail string
	Message   string

	// Untuk Refund Massal
	EventID   int64
}

type NotificationWorker struct {
	JobQueue    chan NotificationPayload
	wg          sync.WaitGroup
	
	// Tambahkan Dependency ke Repo agar bisa akses DB
	userRepo    repository.UserRepository
	bookingRepo repository.BookingRepository
}

// 3. Update Constructor: Menerima Repository
func NewNotificationWorker(uRepo repository.UserRepository, bRepo repository.BookingRepository) *NotificationWorker {
	return &NotificationWorker{
		JobQueue:    make(chan NotificationPayload, 100),
		userRepo:    uRepo,
		bookingRepo: bRepo,
	}
}

func (w *NotificationWorker) Start() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done() // Best practice: taruh defer di awal
		log.Println("👷 Notification Worker started...")

		for job := range w.JobQueue {
			w.processJob(job)
		}

		log.Println("👷 Notification Worker stopped")
	}()
}

// 4. Update Logic ProcessJob untuk membedakan tugas
func (w *NotificationWorker) processJob(job NotificationPayload) {
	if job.Type == JobNotification {
		// Logika Lama (Kirim Email Satuan)
		w.sendEmailLog(job.UserEmail, job.BookingID, job.Message)
	} else if job.Type == JobRefund {
		// Logika Baru (Refund Massal)
		w.processEventRefund(job.EventID)
	}
}

// Fungsi helper untuk simulasi kirim email (Logika lama Anda dipindah kesini)
func (w *NotificationWorker) sendEmailLog(email string, bookingID int64, message string) {
	log.Printf("📧 Mengirim Email ke %s (Ref ID: %d): %s...", email, bookingID, message)
	time.Sleep(1 * time.Second) // Simulasi delay kirim email
	log.Printf("✅ Email terkirim ke %s!", email)
}

// 5. Logic Refund Massal (The Big Logic)
func (w *NotificationWorker) processEventRefund(eventID int64) {
	log.Printf("🔄 [START] Processing refunds for Event ID %d...", eventID)
	
	ctx := context.Background()

	// A. Ambil semua booking untuk event ini
	bookings, err := w.bookingRepo.GetBookingsByEventID(ctx, eventID)
	if err != nil {
		log.Printf("❌ Gagal mengambil data booking: %v", err)
		return
	}

	for _, b := range bookings {
		// B. Cari Email User berdasarkan UserID (Sesuai request Anda)
		user, err := w.userRepo.GetUserByID(ctx, int(b.UserID))
		if err != nil {
			log.Printf("⚠️ User ID %d tidak ditemukan, skip notifikasi.", b.UserID)
			continue
		}

		if b.Status == "PAID" {
			// C. Proses Refund (Simulasi)
			log.Printf("💰 Refunding booking %d for user %s...", b.ID, user.Email)
			time.Sleep(500 * time.Millisecond) // Simulasi delay bank

			// Update Status DB
			w.bookingRepo.UpdateBookingStatus(ctx, b.ID, "REFUNDED")

			// Kirim Email Notifikasi
			w.sendEmailLog(user.Email, b.ID, "Event dibatalkan. Uang Anda telah kami refund sepenuhnya.")

		} else if b.Status == "PENDING" {
			// Kalau belum bayar, cukup cancel saja
			w.bookingRepo.UpdateBookingStatus(ctx, b.ID, "CANCELLED")
			w.sendEmailLog(user.Email, b.ID, "Booking dibatalkan karena event ditiadakan.")
		}
	}

	log.Printf("✅ [DONE] Refund process finished for Event ID %d.", eventID)
}

// Method Public untuk Notifikasi Satuan (dipakai saat Booking Sukses)
func (w *NotificationWorker) SendNotification(bookingID int64, email, message string) {
	w.JobQueue <- NotificationPayload{
		Type:      JobNotification,
		BookingID: bookingID,
		UserEmail: email,
		Message:   message,
	}
}

// Method Public untuk Trigger Refund Massal (dipakai saat Delete Event)
func (w *NotificationWorker) EnqueueCancellation(eventID int64) {
	w.JobQueue <- NotificationPayload{
		Type:    JobRefund,
		EventID: eventID,
	}
}

func (w *NotificationWorker) Stop() {
	log.Println("🛑 Stopping worker... processing remaining jobs...")
	close(w.JobQueue)
	w.wg.Wait()
	log.Println("✅ All jobs finished. Worker safe to exit.")
}