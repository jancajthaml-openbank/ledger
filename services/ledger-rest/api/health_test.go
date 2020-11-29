package api

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jancajthaml-openbank/ledger-rest/system"
)

type mockMonitor struct {
	system.CapacityCheck
	healthy bool
}

func (monitor mockMonitor) IsHealthy() bool {
	return monitor.healthy
}

func (monitor mockMonitor) GetFree() uint64 {
	return uint64(0)
}

func (monitor mockMonitor) GetUsed() uint64 {
	return uint64(0)
}

func TestHealthCheckHandler(t *testing.T) {
	t.Log("HEAD - healthy")
	{
		monitor := new(mockMonitor)
		monitor.healthy = true

		router := echo.New()
		router.HEAD("/health", HealtCheckPing(monitor, monitor))

		req := httptest.NewRequest(http.MethodHead, "/health", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Body.String())
	}

	t.Log("HEAD - unhealthy")
	{
		monitor := new(mockMonitor)
		monitor.healthy = false

		router := echo.New()
		router.HEAD("/health", HealtCheckPing(monitor, monitor))

		req := httptest.NewRequest(http.MethodHead, "/health", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		assert.Empty(t, rec.Body.String())
	}

	t.Log("GET - healthy")
	{
		monitor := new(mockMonitor)
		monitor.healthy = true

		router := echo.New()
		router.GET("/health", HealtCheck(monitor, monitor))

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `
            {
                "storage": {
                    "free": 0,
                    "used": 0,
                    "healthy": true
                },
                "memory": {
                    "free": 0,
                    "used": 0,
                    "healthy": true
                }
            }
        `, rec.Body.String())
	}

	t.Log("GET - unhealthy")
	{
		monitor := new(mockMonitor)
		monitor.healthy = false

		router := echo.New()
		router.GET("/health", HealtCheck(monitor, monitor))

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		assert.JSONEq(t, `
            {
                "storage": {
                    "free": 0,
                    "used": 0,
                    "healthy": false
                },
                "memory": {
                    "free": 0,
                    "used": 0,
                    "healthy": false
                }
            }
        `, rec.Body.String())
	}
}
