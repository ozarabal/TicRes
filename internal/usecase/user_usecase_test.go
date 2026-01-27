package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"
	
	"ticres/internal/entity"
	"ticres/internal/usecase"
	"ticres/internal/usecase/mocks"

	"golang.org/x/crypto/bcrypt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserUsecase_Register(t *testing.T) {
	// 1. Setup Mock
	mockRepo := new(mocks.MockUserRepo)
	
	// 2. Setup Usecase dengan Mock Repo
	// jwtSecret & expiry asal saja karena Register tidak pakai JWT
	u := usecase.NewUserUsecase(mockRepo, time.Second*2, "secret", 1)

	// 3. Definisi Tabel Test Case
	tests := []struct {
		name        string
		input       *entity.User
		mockBehavior func() // Disini kita atur ekspektasi mock
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Success Register",
			input: &entity.User{
				Name:     "Test User",
				Email:    "test@mail.com",
				Password: "password123",
			},
			mockBehavior: func() {
				// Kita berekspektasi CreateUser DIPANGGIL 1x dengan argumen apapun
				// Dan kita suruh dia Return nil (tidak error)
				mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "Failed Register - Email Duplicate",
			input: &entity.User{
				Name:     "Test User",
				Email:    "duplicate@mail.com",
				Password: "password123",
			},
			mockBehavior: func() {
				// Kita suruh Mock return error Duplicate
				mockRepo.On("CreateUser", mock.Anything, mock.Anything).Return(entity.ErrUserAlreadyExsist).Once()
			},
			wantErr:     true,
			expectedErr: entity.ErrUserAlreadyExsist,
		},
		{
			name: "Failed Register - Database Error",
			input: &entity.User{
				Name:     "Test User",
				Email:    "error@mail.com",
				Password: "password123",
			},
			mockBehavior: func() {
				// Kita suruh Mock return error generic DB
				mockRepo.On("CreateUser", mock.Anything, mock.Anything).Return(errors.New("db down")).Once()
			},
			wantErr: true,
		},
	}

	// 4. Eksekusi Loop Test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock setiap iterasi agar bersih
			// (Tergantung versi testify, kadang perlu reset manual atau buat object baru)
			// Disini kita timpa behaviornya lewat fungsi mockBehavior
			tt.mockBehavior()

			err := u.Register(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
			}
			
			// Verifikasi bahwa Mock benar-benar dipanggil sesuai rencana
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserUsercase_Login(t *testing.T) {
	
	password := "password123"
	// Kita harus men-generate hash asli agar bcrypt di usecase tidak error
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	mockUser := &entity.User{
		ID:       1,
		Email:    "test@example.com",
		Password: string(hashedPassword), // Password di DB ceritanya sudah di-hash
	}

	tests := []struct {
		name        string
		email       string
		password	string
		mockBehavior func(m *mocks.MockUserRepo) 
		wantErr     bool
		expectedErr error
	}{
		// berhasil
		{
			name: "Success Login",
			email: "test@example.com",
			password: "password123",
			mockBehavior: func(m *mocks.MockUserRepo) {
				m.On("GetUserByEmail", mock.Anything, "test@example.com").Return(mockUser, nil).Once()
			},
			wantErr: false,
		},

		// email tidak ditemukan 
		{
			name : "Failed Login - ",
			email: "unknown@example.com",
			password: "password123",
			mockBehavior: func(m *mocks.MockUserRepo) {
				m.On("GetUserByEmail", mock.Anything, "unknown@example.com").
					Return(nil, entity.ErrInternalServer).Once() 
			},
			wantErr: true,
		},

		// salah passord

		{
			name: "Failed Login - Password Unmatch",
			email: "test@example.com",
			password: "WRONG_Pass",
			mockBehavior: func(m *mocks.MockUserRepo) {
				m.On("GetUserByEmail", mock.Anything, "test@example.com").
					Return(mockUser, nil).Once()
			},
			wantErr: true,
		},
	}

	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepo)

			tt.mockBehavior(mockRepo)

			u :=usecase.NewUserUsecase(mockRepo, time.Second*2, "secret", 1)

			// Execute
			token, err := u.Login(context.Background(), tt.email, tt.password)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}