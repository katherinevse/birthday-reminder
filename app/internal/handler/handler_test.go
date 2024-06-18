package handler

import (
	"birthdayReminder/app/internal/handler/auth"
	mock_auth "birthdayReminder/app/internal/handler/auth/mocks"
	"birthdayReminder/app/internal/handler/login"
	mock_handler "birthdayReminder/app/internal/handler/mocks"
	"birthdayReminder/app/internal/handler/subscribe"
	"birthdayReminder/app/internal/repository/user"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRegister(t *testing.T) {
	testCases := []struct {
		name           string
		payload        interface{}
		setupMock      func(mockUserRepo *mock_handler.MockUserRepository)
		expectedError  bool
		expectedStatus int
		expectedOutput string
	}{
		{
			name:           "Invalid payload",
			payload:        "invalid json",
			setupMock:      func(mockUserRepo *mock_handler.MockUserRepository) {},
			expectedError:  true,
			expectedStatus: http.StatusBadRequest,
			expectedOutput: "Invalid request payload",
		},
		{
			name:           "Missing user data",
			payload:        user.User{Name: "", Email: "", Password: "", DateOfBirth: time.Time{}},
			setupMock:      func(mockUserRepo *mock_handler.MockUserRepository) {},
			expectedError:  true,
			expectedStatus: http.StatusBadRequest,
			expectedOutput: "Invalid user data",
		},
		{
			name:    "Error saving user",
			payload: user.User{Name: "John Doe", Email: "john@example.com", Password: "password", DateOfBirth: time.Now()},
			setupMock: func(mockUserRepo *mock_handler.MockUserRepository) {
				mockUserRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(errors.New("db error"))
			},
			expectedError:  true,
			expectedStatus: http.StatusInternalServerError,
			expectedOutput: "Error saving user to database: db error",
		},
		{
			name:    "Successful registration",
			payload: user.User{Name: "John Doe", Email: "john@example.com", Password: "password", DateOfBirth: time.Now()},
			setupMock: func(mockUserRepo *mock_handler.MockUserRepository) {
				mockUserRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedError:  false,
			expectedStatus: http.StatusCreated,
			expectedOutput: "User registered successfully",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserRepo := mock_handler.NewMockUserRepository(ctrl)
			handler := &Handler{userRepo: mockUserRepo}
			tt.setupMock(mockUserRepo)

			reqBody, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(reqBody))
			w := httptest.NewRecorder()

			handler.Register(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.expectedStatus, res.StatusCode)

			body, _ := io.ReadAll(res.Body)
			assert.Contains(t, string(body), tt.expectedOutput)
		})
	}

}

func TestLogin(t *testing.T) {
	testCases := []struct {
		name           string
		payload        interface{}
		setupMock      func(mockUserRepo *mock_handler.MockUserRepository, mockTokenManager *mock_auth.MockTokenManager)
		expectedStatus int
		expectedOutput string
	}{
		{
			name:           "Invalid payload",
			payload:        "invalid json",
			setupMock:      func(mockUserRepo *mock_handler.MockUserRepository, mockTokenManager *mock_auth.MockTokenManager) {},
			expectedStatus: http.StatusBadRequest,
			expectedOutput: "Invalid request payload",
		},
		{
			name:    "User not found",
			payload: login.Dto{Email: "nonexistent@example.com", Password: "password"},
			setupMock: func(mockUserRepo *mock_handler.MockUserRepository, mockTokenManager *mock_auth.MockTokenManager) {
				mockUserRepo.EXPECT().GetUserByEmail("nonexistent@example.com").Return(nil, errors.New("user not found"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedOutput: "Invalid email or password",
		},
		{
			name:    "Invalid password",
			payload: login.Dto{Email: "john@example.com", Password: "wrongpassword"},
			setupMock: func(mockUserRepo *mock_handler.MockUserRepository, mockTokenManager *mock_auth.MockTokenManager) {
				dbUser := &user.User{Email: "john@example.com", Password: "$2a$10$7wZT5nNPo8/W9cC5ZjJrBu57oG6e8CxPf49E7NmBztG7vZKbA35Oe"}
				mockUserRepo.EXPECT().GetUserByEmail("john@example.com").Return(dbUser, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedOutput: "Invalid email or password",
		},
		// TODO add testcases
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserRepo := mock_handler.NewMockUserRepository(ctrl)
			mockTokenManager := mock_auth.NewMockTokenManager(ctrl)
			handler := &Handler{userRepo: mockUserRepo, JWTSecretKey: "secret"}

			tt.setupMock(mockUserRepo, mockTokenManager)

			reqBody, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
			w := httptest.NewRecorder()

			handler.Login(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.expectedStatus, res.StatusCode)

			body, _ := io.ReadAll(res.Body)
			assert.Contains(t, string(body), tt.expectedOutput)
		})
	}
}

func TestSubscribe(t *testing.T) {
	testCases := []struct {
		name           string
		token          string
		payload        interface{}
		setupMock      func(mockSubscriptionRepo *mock_handler.MockSubscriptionRepository, mockTokenManager *mock_auth.MockTokenManager)
		expectedStatus int
		expectedOutput string
	}{
		{
			name:    "Missing Authorization header",
			token:   "",
			payload: subscribe.RequestDto{RelatedUserID: 2},
			setupMock: func(mockSubscriptionRepo *mock_handler.MockSubscriptionRepository, mockTokenManager *mock_auth.MockTokenManager) {
			},
			expectedStatus: http.StatusUnauthorized,
			expectedOutput: "Authorization header missing",
		},
		{
			name:    "Invalid JWT token",
			token:   "invalid_token",
			payload: subscribe.RequestDto{RelatedUserID: 2},
			setupMock: func(mockSubscriptionRepo *mock_handler.MockSubscriptionRepository, mockTokenManager *mock_auth.MockTokenManager) {
				mockTokenManager.EXPECT().ParseJWT("invalid_token", "secret").Return(nil, errors.New("invalid token"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedOutput: "Unauthorized",
		},
		{
			name:    "Error decoding request body",
			token:   "valid_token",
			payload: "invalid_json",
			setupMock: func(mockSubscriptionRepo *mock_handler.MockSubscriptionRepository, mockTokenManager *mock_auth.MockTokenManager) {
				mockTokenManager.EXPECT().ParseJWT("valid_token", "secret").Return(&auth.Claims{UserID: 1}, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedOutput: "Invalid request payload",
		},
		{
			name:    "Error creating subscription",
			token:   "valid_token",
			payload: subscribe.RequestDto{RelatedUserID: 2},
			setupMock: func(mockSubscriptionRepo *mock_handler.MockSubscriptionRepository, mockTokenManager *mock_auth.MockTokenManager) {
				mockTokenManager.EXPECT().ParseJWT("valid_token", "secret").Return(&auth.Claims{UserID: 1}, nil)
				mockSubscriptionRepo.EXPECT().CreateSubscription(1, 2).Return(errors.New("error creating subscription"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedOutput: "Error creating subscription",
		},
		{
			name:    "Successful subscription creation",
			token:   "valid_token",
			payload: subscribe.RequestDto{RelatedUserID: 2},
			setupMock: func(mockSubscriptionRepo *mock_handler.MockSubscriptionRepository, mockTokenManager *mock_auth.MockTokenManager) {
				mockTokenManager.EXPECT().ParseJWT("valid_token", "secret").Return(&auth.Claims{UserID: 1}, nil)
				mockSubscriptionRepo.EXPECT().CreateSubscription(1, 2).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedOutput: "Subscription created successfully",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSubscriptionRepo := mock_handler.NewMockSubscriptionRepository(ctrl)
			mockTokenManager := mock_auth.NewMockTokenManager(ctrl)

			handler := &Handler{
				subscriptionRepo: mockSubscriptionRepo,
				tokenManager:     mockTokenManager,
				JWTSecretKey:     "secret",
			}

			tt.setupMock(mockSubscriptionRepo, mockTokenManager)

			reqBody, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/subscribe", bytes.NewBuffer(reqBody))
			req.Header.Set("Authorization", tt.token)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			handler.Subscribe(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d; got %d", tt.expectedStatus, res.StatusCode)
			}

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			if !strings.Contains(string(body), tt.expectedOutput) {
				t.Errorf("Expected response body to contain '%s'; got '%s'", tt.expectedOutput, string(body))
			}
		})
	}
}

func TestGetAvailableUsers(t *testing.T) {
	testCases := []struct {
		name           string
		token          string
		setupMock      func(mockUserRepo *mock_handler.MockUserRepository, mockTokenManager *mock_auth.MockTokenManager)
		expectedStatus int
		expectedOutput string
	}{
		{
			name:  "Missing Authorization header",
			token: "",
			setupMock: func(mockUserRepo *mock_handler.MockUserRepository, mockTokenManager *mock_auth.MockTokenManager) {
			},
			expectedStatus: http.StatusUnauthorized,
			expectedOutput: "Authorization header missing",
		},
		{
			name:  "Invalid JWT token",
			token: "invalid.token",
			setupMock: func(mockUserRepo *mock_handler.MockUserRepository, mockTokenManager *mock_auth.MockTokenManager) {
				mockTokenManager.EXPECT().ParseJWT("invalid.token", "secret").Return(nil, errors.New("invalid token"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedOutput: "Unauthorized",
		},
		{
			name:  "Invalid user ID in token",
			token: "valid.token.with.invalid.userID",
			setupMock: func(mockUserRepo *mock_handler.MockUserRepository, mockTokenManager *mock_auth.MockTokenManager) {
				mockTokenManager.EXPECT().ParseJWT("valid.token.with.invalid.userID", "secret").Return(&auth.Claims{UserID: 0}, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedOutput: "Unauthorized",
		},
		{
			name:  "Error fetching available users",
			token: "valid.token",
			setupMock: func(mockUserRepo *mock_handler.MockUserRepository, mockTokenManager *mock_auth.MockTokenManager) {
				mockTokenManager.EXPECT().ParseJWT("valid.token", "secret").Return(&auth.Claims{UserID: 1}, nil)
				mockUserRepo.EXPECT().GetAvailableUsersForSubscription(1).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedOutput: "Error fetching available users",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserRepo := mock_handler.NewMockUserRepository(ctrl)
			mockTokenManager := mock_auth.NewMockTokenManager(ctrl)
			handler := &Handler{
				JWTSecretKey:     "secret",
				userRepo:         mockUserRepo,
				subscriptionRepo: nil,
				tokenManager:     mockTokenManager,
			}

			tt.setupMock(mockUserRepo, mockTokenManager)

			req := httptest.NewRequest(http.MethodGet, "/users/available", nil)
			req.Header.Set("Authorization", tt.token)
			w := httptest.NewRecorder()

			handler.GetAvailableUsers(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.expectedStatus, res.StatusCode)

			body, _ := io.ReadAll(res.Body)
			assert.Contains(t, string(body), tt.expectedOutput)
		})
	}
}

func TestUnsubscribe(t *testing.T) {
	testCases := []struct {
		name           string
		token          string
		requestBody    subscribe.RequestDto
		setupMock      func(mockSubscriptionRepo *mock_handler.MockSubscriptionRepository, mockTokenManager *mock_auth.MockTokenManager)
		expectedStatus int
		expectedOutput string
	}{
		{
			name:        "Missing Authorization header",
			token:       "",
			requestBody: subscribe.RequestDto{},
			setupMock: func(mockSubscriptionRepo *mock_handler.MockSubscriptionRepository, mockTokenManager *mock_auth.MockTokenManager) {
			},
			expectedStatus: http.StatusUnauthorized,
			expectedOutput: "Authorization header missing\n",
		},
		{
			name:        "Invalid JWT token",
			token:       "invalid.token",
			requestBody: subscribe.RequestDto{},
			setupMock: func(mockSubscriptionRepo *mock_handler.MockSubscriptionRepository, mockTokenManager *mock_auth.MockTokenManager) {
				mockTokenManager.EXPECT().ParseJWT("invalid.token", "secret").Return(nil, errors.New("invalid token"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedOutput: "Unauthorized\n",
		},
		{
			name:        "Invalid user ID in token",
			token:       "valid.token.with.invalid.userID",
			requestBody: subscribe.RequestDto{},
			setupMock: func(mockSubscriptionRepo *mock_handler.MockSubscriptionRepository, mockTokenManager *mock_auth.MockTokenManager) {
				mockTokenManager.EXPECT().ParseJWT("valid.token.with.invalid.userID", "secret").Return(&auth.Claims{UserID: 0}, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedOutput: "Unauthorized\n",
		},
		//{
		//	name:        "Error decoding request body",
		//	token:       "valid.token",
		//	requestBody: subscribe.RequestDto{}, // Пустой экземпляр структуры
		//	setupMock: func(mockSubscriptionRepo *mock_handler.MockSubscriptionRepository, mockTokenManager *mock_auth.MockTokenManager) {
		//		mockTokenManager.EXPECT().ParseJWT("valid.token", "secret").Return(&auth.Claims{UserID: 1}, nil)
		//	},
		//	expectedStatus: http.StatusBadRequest,
		//	expectedOutput: "Invalid request payload\n",
		//},

		{
			name:  "Subscription does not exist",
			token: "valid.token",
			requestBody: subscribe.RequestDto{
				RelatedUserID: 2,
			},
			setupMock: func(mockSubscriptionRepo *mock_handler.MockSubscriptionRepository, mockTokenManager *mock_auth.MockTokenManager) {
				mockTokenManager.EXPECT().ParseJWT("valid.token", "secret").Return(&auth.Claims{UserID: 1}, nil)
				mockSubscriptionRepo.EXPECT().UnsubscribeUser(1, 2).Return(errors.New("subscription does not exist"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedOutput: "You are not subscribed to this user\n",
		},
		{
			name:  "Error unsubscribing user",
			token: "valid.token",
			requestBody: subscribe.RequestDto{
				RelatedUserID: 2,
			},
			setupMock: func(mockSubscriptionRepo *mock_handler.MockSubscriptionRepository, mockTokenManager *mock_auth.MockTokenManager) {
				mockTokenManager.EXPECT().ParseJWT("valid.token", "secret").Return(&auth.Claims{UserID: 1}, nil)
				mockSubscriptionRepo.EXPECT().UnsubscribeUser(1, 2).Return(errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedOutput: "Error unsubscribing user\n",
		},
		{
			name:  "Successful unsubscription",
			token: "valid.token",
			requestBody: subscribe.RequestDto{
				RelatedUserID: 2,
			},
			setupMock: func(mockSubscriptionRepo *mock_handler.MockSubscriptionRepository, mockTokenManager *mock_auth.MockTokenManager) {
				mockTokenManager.EXPECT().ParseJWT("valid.token", "secret").Return(&auth.Claims{UserID: 1}, nil)
				mockSubscriptionRepo.EXPECT().UnsubscribeUser(1, 2).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedOutput: "Unsubscribed successfully",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSubscriptionRepo := mock_handler.NewMockSubscriptionRepository(ctrl)
			mockTokenManager := mock_auth.NewMockTokenManager(ctrl)
			handler := &Handler{
				JWTSecretKey:     "secret",
				subscriptionRepo: mockSubscriptionRepo,
				tokenManager:     mockTokenManager,
			}

			tt.setupMock(mockSubscriptionRepo, mockTokenManager)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/unsubscribe", bytes.NewBuffer(body))
			req.Header.Set("Authorization", tt.token)
			w := httptest.NewRecorder()

			handler.Unsubscribe(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.expectedStatus, res.StatusCode)

			responseBody, _ := io.ReadAll(res.Body)
			assert.Equal(t, tt.expectedOutput, string(responseBody))
		})
	}
}
