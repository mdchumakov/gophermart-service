package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// HTTPClient представляет HTTP клиент для работы с системой ACCRUAL
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHTTPClient создает новый HTTP клиент для системы ACCRUAL
func NewHTTPClient(baseURL string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// GetOrderInfo получает информацию о расчёте начислений баллов лояльности для заказа
func (c *HTTPClient) GetOrderInfo(ctx context.Context, orderNumber string) (*OrderInfo, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return c.parseOrderInfo(resp)
	case http.StatusNoContent:
		return nil, nil // заказ не зарегистрирован в системе расчёта
	case http.StatusTooManyRequests:
		return nil, c.parseRateLimitError(resp)
	case http.StatusInternalServerError:
		return nil, fmt.Errorf("internal server error from accrual system")
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

// parseOrderInfo парсит ответ с информацией о заказе
func (c *HTTPClient) parseOrderInfo(resp *http.Response) (*OrderInfo, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var orderInfo OrderInfo
	if err := json.Unmarshal(body, &orderInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &orderInfo, nil
}

// parseRateLimitError парсит ошибку превышения лимита запросов
func (c *HTTPClient) parseRateLimitError(resp *http.Response) *RateLimitError {
	retryAfter := 60 * time.Second // значение по умолчанию

	if retryAfterStr := resp.Header.Get("Retry-After"); retryAfterStr != "" {
		if seconds, err := strconv.Atoi(retryAfterStr); err == nil {
			retryAfter = time.Duration(seconds) * time.Second
		}
	}

	body, _ := io.ReadAll(resp.Body)
	message := string(body)
	if message == "" {
		message = "rate limit exceeded"
	}

	return &RateLimitError{
		RetryAfter: retryAfter,
		Message:    message,
	}
}
