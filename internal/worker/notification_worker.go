package worker

import (
	"context"
	"sync"
	"time"

	"ticres/internal/entity"
	"ticres/internal/repository"
	"ticres/pkg/logger"
)

type JobType int

const (
	JobNotification JobType = iota
	JobRefund
)

type NotificationPayload struct {
	Type      JobType
	BookingID int64
	UserEmail string
	Message   string
	EventID   int64
}

type NotificationWorker struct {
	JobQueue        chan NotificationPayload
	wg              sync.WaitGroup
	userRepo        repository.UserRepository
	bookingRepo     repository.BookingRepository
	transactionRepo repository.TransactionRepository
	refundRepo      repository.RefundRepository
}

func NewNotificationWorker(
	uRepo repository.UserRepository,
	bRepo repository.BookingRepository,
	txnRepo repository.TransactionRepository,
	refundRepo repository.RefundRepository,
) *NotificationWorker {
	return &NotificationWorker{
		JobQueue:        make(chan NotificationPayload, 100),
		userRepo:        uRepo,
		bookingRepo:     bRepo,
		transactionRepo: txnRepo,
		refundRepo:      refundRepo,
	}
}

func (w *NotificationWorker) Start() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		logger.Info("worker: notification worker started")

		for job := range w.JobQueue {
			w.processJob(job)
		}

		logger.Info("worker: notification worker stopped")
	}()
}

func (w *NotificationWorker) processJob(job NotificationPayload) {
	if job.Type == JobNotification {
		w.sendEmailLog(job.UserEmail, job.BookingID, job.Message)
	} else if job.Type == JobRefund {
		w.processEventRefund(job.EventID)
	}
}

func (w *NotificationWorker) sendEmailLog(email string, bookingID int64, message string) {
	logger.Debug("worker: sending email",
		logger.String("email", email),
		logger.Int64("booking_id", bookingID),
		logger.String("message", message),
	)
	time.Sleep(1 * time.Second) // Simulate email delay
	logger.Info("worker: email sent",
		logger.String("email", email),
		logger.Int64("booking_id", bookingID),
	)
}

func (w *NotificationWorker) processEventRefund(eventID int64) {
	logger.Info("worker: starting refund process", logger.Int64("event_id", eventID))

	ctx := context.Background()

	bookings, err := w.bookingRepo.GetBookingsByEventID(ctx, eventID)
	if err != nil {
		logger.Error("worker: failed to get bookings for refund",
			logger.Int64("event_id", eventID),
			logger.Err(err),
		)
		return
	}

	logger.Debug("worker: processing refunds",
		logger.Int64("event_id", eventID),
		logger.Int("booking_count", len(bookings)),
	)

	for _, b := range bookings {
		user, err := w.userRepo.GetUserByID(ctx, int(b.UserID))
		if err != nil {
			logger.Warn("worker: user not found, skipping notification",
				logger.Int64("user_id", b.UserID),
				logger.Int64("booking_id", b.ID),
			)
			continue
		}

		if b.Status == "PAID" {
			logger.Debug("worker: processing refund",
				logger.Int64("booking_id", b.ID),
				logger.String("email", user.Email),
			)
			time.Sleep(500 * time.Millisecond) // Simulate bank delay

			// Get the transaction and update its status to REFUNDED
			txn, err := w.transactionRepo.GetTransactionByBookingID(ctx, b.ID)
			if err != nil {
				logger.Error("worker: failed to get transaction for refund",
					logger.Int64("booking_id", b.ID),
					logger.Err(err),
				)
			}

			if txn != nil {
				if err := w.transactionRepo.UpdateTransactionStatus(ctx, txn.ID, "REFUNDED", ""); err != nil {
					logger.Error("worker: failed to update transaction to REFUNDED",
						logger.Int64("payment_id", txn.ID),
						logger.Err(err),
					)
				}

				// Create refund record
				refund := &entity.Refund{
					BookingID: b.ID,
					Amount:    txn.Amount,
					Reason:    "Event cancelled by administrator",
					Status:    "COMPLETED",
				}
				if err := w.refundRepo.CreateRefund(ctx, refund); err != nil {
					logger.Error("worker: failed to create refund record",
						logger.Int64("booking_id", b.ID),
						logger.Err(err),
					)
				}
			}

			// Update booking status to REFUNDED
			if err := w.bookingRepo.UpdateBookingStatus(ctx, b.ID, "REFUNDED"); err != nil {
				logger.Error("worker: failed to update booking status to REFUNDED",
					logger.Int64("booking_id", b.ID),
					logger.Err(err),
				)
				continue
			}

			// Release seats back
			if err := w.bookingRepo.ReleaseSeatsByBookingID(ctx, b.ID); err != nil {
				logger.Error("worker: failed to release seats",
					logger.Int64("booking_id", b.ID),
					logger.Err(err),
				)
			}

			w.sendEmailLog(user.Email, b.ID, "Event dibatalkan. Uang Anda telah kami refund sepenuhnya.")
			logger.Info("worker: booking refunded",
				logger.Int64("booking_id", b.ID),
				logger.String("email", user.Email),
			)

		} else if b.Status == "PENDING" {
			// Cancel pending transaction if exists
			txn, _ := w.transactionRepo.GetTransactionByBookingID(ctx, b.ID)
			if txn != nil {
				w.transactionRepo.UpdateTransactionStatus(ctx, txn.ID, "CANCELLED", "")
			}

			if err := w.bookingRepo.UpdateBookingStatus(ctx, b.ID, "CANCELLED"); err != nil {
				logger.Error("worker: failed to update booking status to CANCELLED",
					logger.Int64("booking_id", b.ID),
					logger.Err(err),
				)
				continue
			}

			// Release seats back
			if err := w.bookingRepo.ReleaseSeatsByBookingID(ctx, b.ID); err != nil {
				logger.Error("worker: failed to release seats",
					logger.Int64("booking_id", b.ID),
					logger.Err(err),
				)
			}

			w.sendEmailLog(user.Email, b.ID, "Booking dibatalkan karena event ditiadakan.")
			logger.Info("worker: booking cancelled",
				logger.Int64("booking_id", b.ID),
				logger.String("email", user.Email),
			)
		}
	}

	logger.Info("worker: refund process completed", logger.Int64("event_id", eventID))
}

func (w *NotificationWorker) SendNotification(bookingID int64, email, message string) {
	logger.Debug("worker: enqueuing notification",
		logger.Int64("booking_id", bookingID),
		logger.String("email", email),
	)
	w.JobQueue <- NotificationPayload{
		Type:      JobNotification,
		BookingID: bookingID,
		UserEmail: email,
		Message:   message,
	}
}

func (w *NotificationWorker) EnqueueCancellation(eventID int64) {
	logger.Info("worker: enqueuing cancellation refund", logger.Int64("event_id", eventID))
	w.JobQueue <- NotificationPayload{
		Type:    JobRefund,
		EventID: eventID,
	}
}

func (w *NotificationWorker) Stop() {
	logger.Info("worker: stopping, processing remaining jobs...")
	close(w.JobQueue)
	w.wg.Wait()
	logger.Info("worker: all jobs finished, safe to exit")
}
