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

// MockFeedUC - мок для тестирования
type MockFeedUC struct {
	mock.Mock
}

func (m *MockFeedUC) GetFeed(ctx context.Context, params domain.FeedParams) ([]domain.Video, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]domain.Video), args.Error(1)
}

func TestFeedHandler_GetFeed(t *testing.T) {
	tests := []struct {
		name           string
		feedType       string
		limit          string
		expectedStatus int
		expectedType   string
		expectedLimit  int
	}{
		{
			name:           "популярные видео (по умолчанию)",
			feedType:       "",
			limit:          "",
			expectedStatus: http.StatusOK,
			expectedType:   "popular",
			expectedLimit:  20,
		},
		{
			name:           "популярные видео",
			feedType:       "popular",
			limit:          "10",
			expectedStatus: http.StatusOK,
			expectedType:   "popular",
			expectedLimit:  10,
		},
		{
			name:           "комментируемые видео",
			feedType:       "commented",
			limit:          "15",
			expectedStatus: http.StatusOK,
			expectedType:   "commented",
			expectedLimit:  15,
		},
		{
			name:           "случайные видео",
			feedType:       "random",
			limit:          "5",
			expectedStatus: http.StatusOK,
			expectedType:   "random",
			expectedLimit:  5,
		},
		{
			name:           "неверный limit (используется значение по умолчанию)",
			feedType:       "popular",
			limit:          "invalid",
			expectedStatus: http.StatusOK,
			expectedType:   "popular",
			expectedLimit:  20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок
			mockUC := new(MockFeedUC)

			// Создаем тестовые данные
			testVideos := []domain.Video{
				{
					ID:          uuid.New(),
					Title:       "Test Video 1",
					Description: "Test Description 1",
					Lang:        "en",
					Tags:        []string{"test", "video"},
					DurationS:   120,
					UploadedAt:  time.Now(),
				},
				{
					ID:          uuid.New(),
					Title:       "Test Video 2",
					Description: "Test Description 2",
					Lang:        "en",
					Tags:        []string{"test", "video"},
					DurationS:   180,
					UploadedAt:  time.Now(),
				},
			}

			// Настраиваем ожидания
			expectedParams := domain.FeedParams{
				Type:  domain.FeedType(tt.expectedType),
				Limit: tt.expectedLimit,
			}
			mockUC.On("GetFeed", mock.Anything, expectedParams).Return(testVideos, nil)

			// Создаем handler
			var feedUC usecase.FeedUCInterface = mockUC
			handler := &FeedHandler{UC: feedUC}

			// Создаем тестовый запрос
			req := httptest.NewRequest("GET", "/videos/feed", nil)
			q := req.URL.Query()
			if tt.feedType != "" {
				q.Add("type", tt.feedType)
			}
			if tt.limit != "" {
				q.Add("limit", tt.limit)
			}
			req.URL.RawQuery = q.Encode()

			// Создаем ResponseRecorder
			w := httptest.NewRecorder()

			// Выполняем запрос
			handler.getFeed(w, req)

			// Проверяем статус
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Проверяем, что мок был вызван
			mockUC.AssertExpectations(t)
		})
	}
}

func TestFeedHandler_GetFeed_Error(t *testing.T) {
	// Создаем мок
	mockUC := new(MockFeedUC)

	// Настраиваем ожидания для ошибки
	expectedParams := domain.FeedParams{
		Type:  domain.FeedTypePopular,
		Limit: 20,
	}
	mockUC.On("GetFeed", mock.Anything, expectedParams).Return([]domain.Video{}, assert.AnError)

	// Создаем handler
	var feedUC usecase.FeedUCInterface = mockUC
	handler := &FeedHandler{UC: feedUC}

	// Создаем тестовый запрос
	req := httptest.NewRequest("GET", "/videos/feed", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос
	handler.getFeed(w, req)

	// Проверяем, что вернулась ошибка
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "внутренняя ошибка сервера")

	// Проверяем, что мок был вызван
	mockUC.AssertExpectations(t)
}
