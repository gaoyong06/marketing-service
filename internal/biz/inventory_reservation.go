package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// InventoryReservation 库存预占领域对象
type InventoryReservation struct {
	ReservationID string
	ResourceID    string
	CampaignID    string
	UserID        int64
	Quantity      int
	Status        string
	ExpireAt      time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// InventoryReservationRepo 库存预占仓储接口
type InventoryReservationRepo interface {
	Save(context.Context, *InventoryReservation) (*InventoryReservation, error)
	FindByID(context.Context, string) (*InventoryReservation, error)
	UpdateStatus(context.Context, string, string) error
	CountPendingByResource(context.Context, string) (int, error)
	ListExpired(context.Context) ([]*InventoryReservation, error)
	CancelExpired(context.Context) (int64, error)
	List(context.Context, string, string, int64, string, int, int) ([]*InventoryReservation, int64, error)
}

// InventoryReservationUseCase 库存预占用例
type InventoryReservationUseCase struct {
	repo InventoryReservationRepo
	log  *log.Helper
}

// NewInventoryReservationUseCase 创建库存预占用例
func NewInventoryReservationUseCase(repo InventoryReservationRepo, logger log.Logger) *InventoryReservationUseCase {
	return &InventoryReservationUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// Reserve 预占库存
func (uc *InventoryReservationUseCase) Reserve(ctx context.Context, ir *InventoryReservation) (*InventoryReservation, error) {
	if ir.ReservationID == "" {
		ir.ReservationID = GenerateShortID()
	}
	if ir.Status == "" {
		ir.Status = "PENDING"
	}
	ir.CreatedAt = time.Now()
	ir.UpdatedAt = time.Now()
	return uc.repo.Save(ctx, ir)
}

// Confirm 确认预占（核销）
func (uc *InventoryReservationUseCase) Confirm(ctx context.Context, reservationID string) error {
	return uc.repo.UpdateStatus(ctx, reservationID, "CONFIRMED")
}

// Cancel 取消预占
func (uc *InventoryReservationUseCase) Cancel(ctx context.Context, reservationID string) error {
	return uc.repo.UpdateStatus(ctx, reservationID, "CANCELLED")
}

// GetPendingCount 获取资源的待确认预占数量
func (uc *InventoryReservationUseCase) GetPendingCount(ctx context.Context, resourceID string) (int, error) {
	return uc.repo.CountPendingByResource(ctx, resourceID)
}

// CleanupExpired 清理过期的预占
func (uc *InventoryReservationUseCase) CleanupExpired(ctx context.Context) (int64, error) {
	return uc.repo.CancelExpired(ctx)
}

// List 列出库存预占记录
func (uc *InventoryReservationUseCase) List(ctx context.Context, resourceID, campaignID string, userID int64, status string, page, pageSize int) ([]*InventoryReservation, int64, error) {
	return uc.repo.List(ctx, resourceID, campaignID, userID, status, page, pageSize)
}
