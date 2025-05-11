package health

import (
	"time"

	"gorm.io/gorm"
)

type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]Status `json:"services"`
}

type Status struct {
	Status    string        `json:"status"`
	Latency   time.Duration `json:"latency"`
	Error     string        `json:"error,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

type HealthChecker struct {
	db *gorm.DB
}

func NewHealthChecker(db *gorm.DB) *HealthChecker {
	return &HealthChecker{
		db: db,
	}
}

func (h *HealthChecker) Check() HealthStatus {
	health := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Services:  make(map[string]Status),
	}

	// Check database
	dbStart := time.Now()
	sqlDB, err := h.db.DB()
	if err != nil {
		health.Services["database"] = Status{
			Status:    "unhealthy",
			Latency:   time.Since(dbStart),
			Error:     err.Error(),
			Timestamp: time.Now(),
		}
		health.Status = "unhealthy"
	} else {
		err = sqlDB.Ping()
		if err != nil {
			health.Services["database"] = Status{
				Status:    "unhealthy",
				Latency:   time.Since(dbStart),
				Error:     err.Error(),
				Timestamp: time.Now(),
			}
			health.Status = "unhealthy"
		} else {
			health.Services["database"] = Status{
				Status:    "healthy",
				Latency:   time.Since(dbStart),
				Timestamp: time.Now(),
			}
		}
	}

	return health
}
