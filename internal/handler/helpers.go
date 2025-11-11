package handler

import (
	"net"
	"net/http"
	"strings"
)

// getRealIP извлекает реальный IP адрес клиента
// WHY: RemoteAddr может содержать IP прокси, а не клиента
// HOW: Проверяем заголовки от reverse proxy (nginx, cloudflare)
//
// Порядок проверки:
// 1. X-Forwarded-For (стандартный заголовок от nginx/proxy)
// 2. X-Real-IP (альтернативный заголовок)
// 3. RemoteAddr (fallback, если нет прокси)
//
// Returns: IP адрес без порта
func getRealIP(r *http.Request) string {
	// X-Forwarded-For может содержать несколько IP через запятую
	// Формат: "client, proxy1, proxy2"
	// Берем первый IP (это реальный клиент)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" {
				return ip
			}
		}
	}

	// X-Real-IP - альтернативный заголовок от некоторых proxy
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}

	// Fallback: берем RemoteAddr и убираем порт
	// RemoteAddr формат: "192.168.1.1:12345" или "[::1]:12345"
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// Если не удалось распарсить (нет порта), возвращаем как есть
		return r.RemoteAddr
	}

	return ip
}
