package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/arasvet/microtube/internal/domain"
	"github.com/arasvet/microtube/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSearchUC - мок для тестирования
type MockSearchUC struct {
	mock.Mock
}

func (m *MockSearchUC) SearchVideos(ctx context.Context, params domain.SearchParams) ([]domain.SearchResult, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]domain.SearchResult), args.Error(1)
}

func TestSearchHandler_SearchVideos(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		limit          string
		offset         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "успешный поиск",
			query:          "test",
			limit:          "10",
			offset:         "0",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "отсутствует обязательный параметр q",
			query:          "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "parameter 'q' is required\n",
		},
		{
			name:           "неверный limit",
			query:          "test",
			limit:          "invalid",
			expectedStatus: http.StatusOK, // используем значение по умолчанию
		},
		{
			name:           "неверный offset",
			query:          "test",
			offset:         "invalid",
			expectedStatus: http.StatusOK, // используем значение по умолчанию
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок
			mockUC := new(MockSearchUC)

			// Настраиваем ожидания для успешных случаев
			if tt.expectedStatus == http.StatusOK {
				expectedParams := domain.SearchParams{
					Query:  tt.query,
					Limit:  20, // значение по умолчанию
					Offset: 0,
				}

				// Парсим limit если он передан
				if tt.limit != "" {
					if limit, err := parseLimit(tt.limit); err == nil {
						expectedParams.Limit = limit
					}
				}

				// Парсим offset если он передан
				if tt.offset != "" {
					if offset, err := parseOffset(tt.offset); err == nil {
						expectedParams.Offset = offset
					}
				}

				// Создаем тестовые данные
				testResults := []domain.SearchResult{
					{
						Video: domain.Video{
							ID:          uuid.New(),
							Title:       "Test Video",
							Description: "Test Description",
							Lang:        "en",
							Tags:        []string{"test", "video"},
							DurationS:   120,
							UploadedAt:  time.Now(),
						},
						Score: 0.85,
					},
				}

				mockUC.On("SearchVideos", mock.Anything, expectedParams).Return(testResults, nil)
			}

			// Создаем handler с интерфейсом
			var searchUC usecase.SearchUCInterface = mockUC
			handler := &SearchHandler{UC: searchUC}

			// Создаем тестовый запрос
			req := httptest.NewRequest("GET", "/search", nil)
			q := req.URL.Query()
			if tt.query != "" {
				q.Add("q", tt.query)
			}
			if tt.limit != "" {
				q.Add("limit", tt.limit)
			}
			if tt.offset != "" {
				q.Add("offset", tt.offset)
			}
			req.URL.RawQuery = q.Encode()

			// Создаем ResponseRecorder
			w := httptest.NewRecorder()

			// Выполняем запрос
			handler.searchVideos(w, req)

			// Проверяем статус
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Проверяем тело ответа для ошибок
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, w.Body.String())
			}

			// Проверяем, что мок был вызван для успешных случаев
			if tt.expectedStatus == http.StatusOK {
				mockUC.AssertExpectations(t)
			}
		})
	}
}

// Вспомогательные функции для парсинга (копируем логику из handler)
func parseLimit(limitStr string) (int, error) {
	if limitStr == "" {
		return 20, nil
	}
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		return l, nil
	}
	return 20, nil
}

func parseOffset(offsetStr string) (int, error) {
	if offsetStr == "" {
		return 0, nil
	}
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		return o, nil
	}
	return 0, nil
}
