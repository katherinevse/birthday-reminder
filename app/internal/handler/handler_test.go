package handler

import (
	mock_handler "birthdayReminder/app/internal/handler/mocks"
	"birthdayReminder/app/internal/repository/user"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandler_Register(t *testing.T) {
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
