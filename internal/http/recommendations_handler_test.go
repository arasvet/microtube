package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"time"

	"github.com/arasvet/microtube/internal/domain"
	"github.com/arasvet/microtube/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRecommendationsUC - мок для тестирования
type MockRecommendationsUC struct {
	mock.Mock
}

func (m *MockRecommendationsUC) GetRecommendations(ctx context.Context, params domain.RecommendationParams) ([]domain.RecommendationResult, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]domain.RecommendationResult), args.Error(1)
}

func TestRecommendationsHandler_GetRecommendations(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		sessionID      string
		limit          string
		expectedStatus int
		expectedType   string
		expectedLimit  int
	}{
		{
			name:           "персональные рекомендации для пользователя",
			userID:         "550e8400-e29b-41d4-a716-446655440000",
			sessionID:      "",
			limit:          "10",
			expectedStatus: http.StatusOK,
			expectedType:   "personal",
			expectedLimit:  10,
		},
		{
			name:           "холодные рекомендации для сессии",
			userID:         "",
			sessionID:      "session-123",
			limit:          "15",
			expectedStatus: http.StatusOK,
			expectedType:   "cold",
			expectedLimit:  15,
		},
		{
			name:           "рекомендации без параметров (должна быть ошибка)",
			userID:         "",
			sessionID:      "",
			limit:          "20",
			expectedStatus: http.StatusInternalServerError,
			expectedType:   "",
			expectedLimit:  20,
		},
		{
			name:           "неверный limit (используется значение по умолчанию)",
			userID:         "550e8400-e29b-41d4-a716-446655440000",
			sessionID:      "",
			limit:          "invalid",
			expectedStatus: http.StatusOK,
			expectedType:   "personal",
			expectedLimit:  20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок
			mockUC := new(MockRecommendationsUC)

			// Настраиваем ожидания для успешных случаев
			if tt.expectedStatus == http.StatusOK {
				expectedParams := domain.RecommendationParams{
					Limit: tt.expectedLimit,
				}

				// Устанавливаем user_id или session_id
				if tt.userID != "" {
					expectedParams.UserID = &tt.userID
				}
				if tt.sessionID != "" {
					expectedParams.SessionID = &tt.sessionID
				}

				// Создаем тестовые данные
				testResults := []domain.RecommendationResult{
					{
						Video: domain.Video{
							ID:          uuid.New(),
							Title:       "Test Video 1",
							Description: "Test Description 1",
							Lang:        "en",
							Tags:        []string{"test", "video"},
							DurationS:   120,
							UploadedAt:  time.Now(),
						},
						Reason: domain.ReasonPopular,
						Score:  0.8,
					},
					{
						Video: domain.Video{
							ID:          uuid.New(),
							Title:       "Test Video 2",
							Description: "Test Description 2",
							Lang:        "en",
							Tags:        []string{"test", "video"},
							DurationS:   180,
							UploadedAt:  time.Now(),
						},
						Reason: domain.ReasonUserTags,
						Score:  0.9,
					},
				}

				mockUC.On("GetRecommendations", mock.Anything, expectedParams).Return(testResults, nil)
			} else {
				// Для случая с ошибкой
				expectedParams := domain.RecommendationParams{
					Limit: tt.expectedLimit,
				}
				mockUC.On("GetRecommendations", mock.Anything, expectedParams).Return([]domain.RecommendationResult{}, assert.AnError)
			}

			// Создаем handler
			var recommendationsUC usecase.RecommendationsUCInterface = mockUC
			handler := &RecommendationsHandler{UC: recommendationsUC}

			// Создаем тестовый запрос
			req := httptest.NewRequest("GET", "/recommendations", nil)
			q := req.URL.Query()
			if tt.userID != "" {
				q.Add("user_id", tt.userID)
			}
			if tt.sessionID != "" {
				q.Add("session_id", tt.sessionID)
			}
			if tt.limit != "" {
				q.Add("limit", tt.limit)
			}
			req.URL.RawQuery = q.Encode()

			// Создаем ResponseRecorder
			w := httptest.NewRecorder()

			// Выполняем запрос
			handler.getRecommendations(w, req)

			// Проверяем статус
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Проверяем, что мок был вызван
			mockUC.AssertExpectations(t)
		})
	}
}

func TestRecommendationsHandler_GetRecommendations_Error(t *testing.T) {
	// Создаем мок
	mockUC := new(MockRecommendationsUC)

	// Настраиваем ожидания для ошибки
	expectedParams := domain.RecommendationParams{
		UserID: &[]string{"550e8400-e29b-41d4-a716-446655440000"}[0],
		Limit:  20,
	}
	mockUC.On("GetRecommendations", mock.Anything, expectedParams).Return([]domain.RecommendationResult{}, assert.AnError)

	// Создаем handler
	var recommendationsUC usecase.RecommendationsUCInterface = mockUC
	handler := &RecommendationsHandler{UC: recommendationsUC}

	// Создаем тестовый запрос
	req := httptest.NewRequest("GET", "/recommendations?user_id=550e8400-e29b-41d4-a716-446655440000", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос
	handler.getRecommendations(w, req)

	// Проверяем, что вернулась ошибка
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "внутренняя ошибка сервера")

	// Проверяем, что мок был вызван
	mockUC.AssertExpectations(t)
}
