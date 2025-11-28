package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/gaoyong06/middleground/marketing-service/internal/biz"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RedeemCodeModel 兑换码模型
type RedeemCodeModel struct {
	ID           int64      `gorm:"primaryKey;autoIncrement"`
	Code         string     `gorm:"type:varchar(32);uniqueIndex;not null"`
	CampaignID   string     `gorm:"type:varchar(32);index;not null"`
	TenantID     string     `gorm:"type:varchar(32);index;not null"`
	BatchID      string     `gorm:"type:varchar(32);index"`
	Status       int32      `gorm:"type:tinyint;not null;default:1"`
	ValidFrom    time.Time  `gorm:"type:datetime;not null"`
	ValidUntil   time.Time  `gorm:"type:datetime;not null"`
	Metadata     string     `gorm:"type:text"`
	RedemptionAt *time.Time `gorm:"type:datetime"`
	CreatedAt    time.Time  `gorm:"type:datetime;not null"`
	UpdatedAt    time.Time  `gorm:"type:datetime;not null"`
	DeletedAt    *time.Time `gorm:"index"`
}

// TableName 设置表名
func (RedeemCodeModel) TableName() string {
	return "redeem_code"
}

// CodeBatchModel 兑换码批次模型
type CodeBatchModel struct {
	ID          string     `gorm:"primaryKey;type:varchar(32)"`
	BatchID     string     `gorm:"type:varchar(32);uniqueIndex;not null"`
	TenantID    string     `gorm:"type:varchar(32);index;not null"`
	CampaignID  string     `gorm:"type:varchar(32);index;not null"`
	BatchName   string     `gorm:"type:varchar(64);not null"`
	CodeCount   int32      `gorm:"type:int;not null"`
	CodePrefix  string     `gorm:"type:varchar(16)"`
	CodeType    string     `gorm:"type:varchar(16);not null"`
	CodeLength  int32      `gorm:"type:int;not null"`
	ValidFrom   time.Time  `gorm:"type:datetime;not null"`
	ValidUntil  time.Time  `gorm:"type:datetime;not null"`
	Description string     `gorm:"type:varchar(255)"`
	CreatedAt   time.Time  `gorm:"type:datetime;not null"`
	UpdatedAt   time.Time  `gorm:"type:datetime;not null"`
	DeletedAt   *time.Time `gorm:"index"`
}

// TableName 设置表名
func (CodeBatchModel) TableName() string {
	return "code_batch"
}

// RedemptionRecordModel 兑换记录模型
type RedemptionRecordModel struct {
	ID            string     `gorm:"primaryKey;type:varchar(32)"`
	RedemptionID   string     `gorm:"type:varchar(32);uniqueIndex;not null"`
	RedeemCodeID   int64      `gorm:"type:bigint;index;not null"`
	Code          string     `gorm:"type:varchar(32);index;not null"`
	CampaignID    string     `gorm:"type:varchar(32);index;not null"`
	TenantID      string     `gorm:"type:varchar(32);index;not null"`
	UserID        string     `gorm:"type:varchar(32);index;not null"`
	RedemptionType string     `gorm:"type:varchar(16);not null"`
	RedemptionAt   time.Time  `gorm:"type:datetime;not null"`
	RedemptionIP   string     `gorm:"type:varchar(32)"`
	RedemptionEnv  string     `gorm:"type:varchar(16)"`
	RedeemChannel string     `gorm:"type:varchar(32)"`
	DeviceInfo    string     `gorm:"type:text"`
	IPAddress     string     `gorm:"type:varchar(64)"`
	Location      string     `gorm:"type:varchar(255)"`
	OrderID       string     `gorm:"type:varchar(64)"`
	RewardDetail  string     `gorm:"type:text"`
	Metadata      string     `gorm:"type:text"`
	CreatedAt     time.Time  `gorm:"type:datetime;not null"`
	UpdatedAt     time.Time  `gorm:"type:datetime;not null"`
	DeletedAt     *time.Time `gorm:"index"`
}

// TableName 设置表名
func (RedemptionRecordModel) TableName() string {
	return "redemption_record"
}

type redeemCodeRepo struct {
	data *Data
	log  *log.Helper
}

// NewRedeemCodeRepo 创建兑换码仓库实现
func NewRedeemCodeRepo(data *Data, logger log.Logger) biz.RedeemCodeRepo {
	return &redeemCodeRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// RedeemCode 兑换码
func (r *redeemCodeRepo) RedeemCode(ctx context.Context, code string, userID int64, productCode, redeemChannel, deviceInfo, ipAddress, location, orderID string) (*biz.RedeemCode, *biz.RedemptionRecord, error) {
	r.log.WithContext(ctx).Infof("RedeemCode: code=%v, userID=%v", code, userID)

	// 开始事务
	tx := r.data.db.Begin()
	if tx.Error != nil {
		return nil, nil, tx.Error
	}

	// 查询兑换码
	var model RedeemCodeModel
	result := tx.Where("code = ?", code).First(&model)
	if result.Error != nil {
		tx.Rollback()
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("code not found: %s", code)
		}
		return nil, nil, result.Error
	}

	// 检查兑换码状态
	if model.Status != biz.CodeStatusActive {
		tx.Rollback()
		return nil, nil, fmt.Errorf("code is not active: %s, status: %d", code, model.Status)
	}

	// 检查兑换码有效期
	now := time.Now()
	if now.Before(model.ValidFrom) || now.After(model.ValidUntil) {
		tx.Rollback()
		return nil, nil, fmt.Errorf("code is not valid at this time: %s", code)
	}

	// 更新兑换码状态为已使用
	model.Status = biz.CodeStatusRedeemed
	model.RedemptionAt = &now
	result = tx.Save(&model)
	if result.Error != nil {
		tx.Rollback()
		return nil, nil, result.Error
	}

	// 解析元数据
	var metadata map[string]interface{}
	if model.Metadata != "" {
		if err := json.Unmarshal([]byte(model.Metadata), &metadata); err != nil {
			tx.Rollback()
			return nil, nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	} else {
		metadata = make(map[string]interface{})
	}

	// 创建兑换记录
	redemptionID := uuid.New().String()
	redemptionRecord := RedemptionRecordModel{
		ID:            uuid.New().String(),
		RedemptionID:   redemptionID,
		RedeemCodeID:   model.ID,
		Code:          code,
		CampaignID:    model.CampaignID,
		TenantID:      model.TenantID,
		UserID:        fmt.Sprintf("%d", userID),
		RedemptionType: "NORMAL",
		RedeemChannel: redeemChannel,
		DeviceInfo:    deviceInfo,
		IPAddress:     ipAddress,
		Location:      location,
		OrderID:       orderID,
		RedemptionAt:  now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	result = tx.Create(&redemptionRecord)
	if result.Error != nil {
		tx.Rollback()
		return nil, nil, result.Error
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, nil, err
	}

	// 构建返回的领域模型
	redeemCode := &biz.RedeemCode{
		RedeemCodeID: model.ID,
		Code:        model.Code,
		CampaignID:  model.CampaignID,
		BatchID:     model.BatchID,
		TenantID:    model.TenantID,
		Status:      model.Status,
		ValidFrom:   model.ValidFrom,
		ValidUntil:  model.ValidUntil,
		Metadata:    metadata,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		RedemptionAt: model.RedemptionAt,
	}

	// 将 ID 转换为 int64
	recordID, err := strconv.ParseInt(redemptionRecord.ID, 10, 64)
	if err != nil {
		r.log.WithContext(ctx).Errorf("Failed to parse record ID: %v", err)
		recordID = 0 // 使用默认值
	}
	
	// 将 RedeemCodeID 转换为 int64
	redeemCodeID := redemptionRecord.RedeemCodeID
	
	record := &biz.RedemptionRecord{
		RecordID:      recordID,
		RedeemCodeID:  redeemCodeID,
		Code:         redemptionRecord.Code,
		UserID:       userID,
		TenantID:     redemptionRecord.TenantID,
		CampaignID:   redemptionRecord.CampaignID,
		RedeemChannel: redemptionRecord.RedeemChannel,
		RedemptionAt: redemptionRecord.RedemptionAt,
		DeviceInfo:   redemptionRecord.DeviceInfo,
		IPAddress:    redemptionRecord.IPAddress,
		Location:     redemptionRecord.Location,
		OrderID:      redemptionRecord.OrderID,
	}

	return redeemCode, record, nil
}

// AssignCode 分配兑换码给用户
func (r *redeemCodeRepo) AssignCode(ctx context.Context, code string, userID int64) (*biz.RedeemCode, error) {
	r.log.WithContext(ctx).Infof("AssignCode: code=%v, userID=%v", code, userID)

	// 查找兑换码
	redeemCode, err := r.GetCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// 检查兑换码状态
	if redeemCode.Status != biz.CodeStatusActive {
		return nil, fmt.Errorf("redeem code is not active")
	}

	// 更新兑换码状态和用户ID
	redeemCode.UserID = userID

	// 更新到数据库
	updatedCode, err := r.UpdateCode(ctx, redeemCode)
	if err != nil {
		return nil, err
	}

	return updatedCode, nil
}

// GetCode 获取兑换码
func (r *redeemCodeRepo) GetCode(ctx context.Context, codeStr string) (*biz.RedeemCode, error) {
	r.log.WithContext(ctx).Infof("GetCode: code=%v", codeStr)

	// 从数据库中查询兑换码
	var model RedeemCodeModel
	result := r.data.db.Where("code = ?", codeStr).First(&model)
	if result.Error != nil {
		return nil, result.Error
	}

	// 解析元数据
	var metadata map[string]interface{}
	if model.Metadata != "" {
		if err := json.Unmarshal([]byte(model.Metadata), &metadata); err != nil {
			r.log.WithContext(ctx).Errorf("Failed to unmarshal metadata: %v", err)
			metadata = make(map[string]interface{})
		}
	}

	// 转换为领域模型
	return &biz.RedeemCode{
		RedeemCodeID: model.ID,
		Code:        model.Code,
		CampaignID:  model.CampaignID,
		TenantID:    model.TenantID,
		BatchID:     model.BatchID,
		Status:      model.Status,
		ValidFrom:   model.ValidFrom,
		ValidUntil:  model.ValidUntil,
		Metadata:    metadata,
		RedemptionAt: model.RedemptionAt,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}, nil
}

// UpdateCode 更新兑换码
func (r *redeemCodeRepo) UpdateCode(ctx context.Context, code *biz.RedeemCode) (*biz.RedeemCode, error) {
	r.log.WithContext(ctx).Infof("UpdateCode: code=%v", code.Code)

	// 序列化元数据
	var metadataStr string
	if len(code.Metadata) > 0 {
		metadataBytes, err := json.Marshal(code.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataStr = string(metadataBytes)
	}

	// 更新兑换码
	model := &RedeemCodeModel{
		ID:          code.RedeemCodeID,
		Code:        code.Code,
		CampaignID:  code.CampaignID,
		TenantID:    code.TenantID,
		BatchID:     code.BatchID,
		Status:      code.Status,
		ValidFrom:   code.ValidFrom,
		ValidUntil:  code.ValidUntil,
		Metadata:    metadataStr,
		RedemptionAt: code.RedemptionAt,
		UpdatedAt:   time.Now(),
	}

	result := r.data.db.Save(model)
	if result.Error != nil {
		return nil, result.Error
	}

	return code, nil
}

// CreateBatch 创建兑换码批次
func (r *redeemCodeRepo) CreateBatch(ctx context.Context, batch *biz.CodeBatch, codes []*biz.RedeemCode) (*biz.CodeBatch, error) {
	// 开始事务
	tx := r.data.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	// 确保事务最终提交或回滚
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建批次模型
	model := &CodeBatchModel{
		ID:          batch.BatchID,
		TenantID:    batch.TenantID,
		CampaignID:  batch.CampaignID,
		BatchName:   batch.BatchName,
		CodeCount:   batch.CodeCount,
		CodePrefix:  batch.CodePrefix,
		CodeType:    batch.CodeType,
		CodeLength:  batch.CodeLength,
		ValidFrom:   batch.ValidFrom,
		ValidUntil:  batch.ValidUntil,
		Description: batch.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存批次到数据库
	result := tx.Create(model)
	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}

	// 更新批次ID
	batch.BatchID = model.ID

	// 批量创建兑换码
	if len(codes) > 0 {
		codeModels := make([]RedeemCodeModel, 0, len(codes))
		for _, code := range codes {
			// 序列化元数据
			var metadataStr string
			if len(code.Metadata) > 0 {
				metadataBytes, err := json.Marshal(code.Metadata)
				if err != nil {
					tx.Rollback()
					return nil, fmt.Errorf("failed to marshal metadata: %w", err)
				}
				metadataStr = string(metadataBytes)
			}

			codeModels = append(codeModels, RedeemCodeModel{
				Code:       code.Code,
				CampaignID: code.CampaignID,
				BatchID:    batch.BatchID,
				TenantID:   code.TenantID,
				Status:     code.Status,
				ValidFrom:  code.ValidFrom,
				ValidUntil: code.ValidUntil,
				Metadata:   metadataStr,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			})
		}

		// 批量插入兑换码
		result = tx.CreateInBatches(codeModels, 100)
		if result.Error != nil {
			tx.Rollback()
			return nil, result.Error
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return batch, nil
}

// CreateCode 创建单个兑换码
func (r *redeemCodeRepo) CreateCode(ctx context.Context, code *biz.RedeemCode) (*biz.RedeemCode, error) {
	r.log.WithContext(ctx).Infof("CreateCode: code=%v", code.Code)

	// 序列化元数据
	var metadataStr string
	if len(code.Metadata) > 0 {
		metadataBytes, err := json.Marshal(code.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataStr = string(metadataBytes)
	}

	// 创建兑换码模型
	model := &RedeemCodeModel{
		Code:       code.Code,
		CampaignID: code.CampaignID,
		BatchID:    code.BatchID,
		TenantID:   code.TenantID,
		Status:     code.Status,
		ValidFrom:  code.ValidFrom,
		ValidUntil: code.ValidUntil,
		Metadata:   metadataStr,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 保存到数据库
	result := r.data.db.Create(model)
	if result.Error != nil {
		return nil, result.Error
	}

	// 更新ID
	code.RedeemCodeID = model.ID

	return code, nil
}

// CreateRedemptionRecord 创建兑换记录
func (r *redeemCodeRepo) CreateRedemptionRecord(ctx context.Context, record *biz.RedemptionRecord) (*biz.RedemptionRecord, error) {
	r.log.WithContext(ctx).Infof("CreateRedemptionRecord: code=%v, userID=%v", record.Code, record.UserID)

	// 序列化奖励详情
	var rewardDetailStr string
	if len(record.RewardDetail) > 0 {
		rewardDetailBytes, err := json.Marshal(record.RewardDetail)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal reward detail: %w", err)
		}
		rewardDetailStr = string(rewardDetailBytes)
	}

	// 创建兑换记录模型
	model := &RedemptionRecordModel{
		ID:           fmt.Sprintf("%d", record.RecordID),
		RedeemCodeID: record.RedeemCodeID,
		Code:         record.Code,
		CampaignID:   record.CampaignID,
		TenantID:     record.TenantID,
		UserID:       fmt.Sprintf("%d", record.UserID),
		RedemptionAt: record.RedemptionAt,
		Metadata:     rewardDetailStr,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 保存到数据库
	result := r.data.db.Create(model)
	if result.Error != nil {
		return nil, result.Error
	}

	// 更新ID
	recordID, _ := strconv.ParseInt(model.ID, 10, 64)
	record.RecordID = recordID

	return record, nil
}

// GetRedemptionRecords 获取兑换记录
func (r *redeemCodeRepo) GetRedemptionRecords(ctx context.Context, codeID string) ([]*biz.RedemptionRecord, error) {
	r.log.WithContext(ctx).Infof("GetRedemptionRecords: codeID=%v", codeID)

	// 从数据库中查询兑换记录
	var models []RedemptionRecordModel
	// 尝试将 codeID 转换为 int64
	codeIDInt, err := strconv.ParseInt(codeID, 10, 64)
	if err != nil {
		r.log.WithContext(ctx).Errorf("Failed to parse codeID: %v", err)
		return nil, fmt.Errorf("invalid code ID: %w", err)
	}
	result := r.data.db.Where("redeem_code_id = ?", codeIDInt).Find(&models)
	if result.Error != nil {
		return nil, result.Error
	}

	// 转换为领域模型
	records := make([]*biz.RedemptionRecord, 0, len(models))
	for _, model := range models {
		// 解析奖励详情
		var rewardDetail map[string]string
		if model.Metadata != "" {
			if err := json.Unmarshal([]byte(model.Metadata), &rewardDetail); err != nil {
				r.log.WithContext(ctx).Errorf("Failed to unmarshal reward detail: %v", err)
				rewardDetail = make(map[string]string)
			}
		}

		// 解析用户ID
		userID, _ := strconv.ParseInt(model.UserID, 10, 64)

		// 解析记录ID
		recordID, _ := strconv.ParseInt(model.ID, 10, 64)

		records = append(records, &biz.RedemptionRecord{
			RecordID:      recordID,
			RedeemCodeID:  model.RedeemCodeID,
			Code:         model.Code,
			UserID:       userID,
			TenantID:     model.TenantID,
			CampaignID:   model.CampaignID,
			RedeemChannel: model.RedeemChannel,
			RedemptionAt: model.RedemptionAt,
			RewardDetail: rewardDetail,
		})
	}

	return records, nil
}

// BatchCreateCodes 批量创建兑换码
func (r *redeemCodeRepo) BatchCreateCodes(ctx context.Context, codes []*biz.RedeemCode) error {
	// 使用事务批量创建
	return r.data.db.Transaction(func(tx *gorm.DB) error {
		// 批量插入，每批500个
		batchSize := 500
		for i := 0; i < len(codes); i += batchSize {
			end := i + batchSize
			if end > len(codes) {
				end = len(codes)
			}

			batch := codes[i:end]
			models := make([]RedeemCodeModel, len(batch))

			for j, code := range batch {
				// 序列化元数据
				var metadataJSON string
				if code.Metadata != nil {
					data, err := json.Marshal(code.Metadata)
					if err != nil {
						return err
					}
					metadataJSON = string(data)
				}

				models[j] = RedeemCodeModel{
					Code:       code.Code,
					CampaignID: code.CampaignID,
					TenantID:   code.TenantID,
					BatchID:    code.BatchID,
					Status:     code.Status,
					ValidFrom:  code.ValidFrom,
					ValidUntil: code.ValidUntil,
					Metadata:   metadataJSON,
				}
			}

			if err := tx.Create(&models).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetByCode 通过兑换码获取
func (r *redeemCodeRepo) GetByCode(ctx context.Context, code string) (*biz.RedeemCode, error) {
	// 首先尝试从Redis缓存获取
	cacheKey := fmt.Sprintf("redeem_code:%s", code)
	cachedData, err := r.data.redis.Get(ctx, cacheKey).Result()

	if err == nil {
		// 缓存命中，解析数据
		var redeemCode biz.RedeemCode
		if err := json.Unmarshal([]byte(cachedData), &redeemCode); err == nil {
			return &redeemCode, nil
		}
		// 解析失败，继续从数据库获取
	}

	// 从数据库获取
	var model RedeemCodeModel
	if err := r.data.db.Where("code = ?", code).First(&model).Error; err != nil {
		return nil, err
	}

	// 解析元数据
	var metadata map[string]interface{}
	if model.Metadata != "" {
		if err := json.Unmarshal([]byte(model.Metadata), &metadata); err != nil {
			r.log.Warnf("failed to unmarshal metadata for code %s: %v", code, err)
		}
	}

	// 构建返回对象
	redeemCode := &biz.RedeemCode{
		Code:       model.Code,
		CampaignID: model.CampaignID,
		TenantID:   model.TenantID,
		BatchID:    model.BatchID,
		Status:     model.Status,
		ValidFrom:  model.ValidFrom,
		ValidUntil: model.ValidUntil,
		Metadata:   metadata,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
	}

	if model.RedemptionAt != nil {
		redeemCode.RedemptionAt = model.RedemptionAt
	}

	// 缓存到Redis，有效期1小时
	cacheData, err := json.Marshal(redeemCode)
	if err == nil {
		r.data.redis.Set(ctx, cacheKey, string(cacheData), time.Hour)
	}

	return redeemCode, nil
}

// Redeem 兑换码兑换
func (r *redeemCodeRepo) Redeem(ctx context.Context, redemption *biz.Redemption) error {
	// 使用分布式锁确保并发安全
	lockKey := fmt.Sprintf("redeem_lock:%s", redemption.Code)
	lockValue := redemption.RedemptionID

	// 尝试获取锁，超时时间5秒
	locked, err := r.data.redis.SetNX(ctx, lockKey, lockValue, 5*time.Second).Result()
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !locked {
		return fmt.Errorf("code is being redeemed by another request")
	}

	defer r.data.redis.Del(ctx, lockKey)

	// 开启事务
	return r.data.db.Transaction(func(tx *gorm.DB) error {
		// 查询兑换码并锁定行
		var codeModel RedeemCodeModel
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("code = ?", redemption.Code).First(&codeModel).Error; err != nil {
			return err
		}

		// 检查兑换码状态
		if codeModel.Status != biz.CodeStatusActive {
			return fmt.Errorf("code is not active")
		}

		// 检查兑换码是否已被兑换
		if codeModel.RedemptionAt != nil {
			return fmt.Errorf("code has already been redeemed")
		}

		// 检查兑换码有效期
		now := time.Now()
		if now.Before(codeModel.ValidFrom) || now.After(codeModel.ValidUntil) {
			return fmt.Errorf("code is not valid at this time")
		}

		// 序列化元数据
		var metadataJSON string
		if redemption.Metadata != nil {
			data, err := json.Marshal(redemption.Metadata)
			if err != nil {
				return err
			}
			metadataJSON = string(data)
		}

		// 创建兑换记录
		redemptionModel := &RedemptionRecordModel{
			RedemptionID:  redemption.RedemptionID,
			Code:          redemption.Code,
			CampaignID:    codeModel.CampaignID,
			TenantID:      codeModel.TenantID,
			UserID:        redemption.UserID,
			RedemptionAt:  now,
			RedemptionIP:  redemption.RedemptionIP,
			RedemptionEnv: redemption.RedemptionEnv,
			Metadata:      metadataJSON,
		}

		if err := tx.Create(redemptionModel).Error; err != nil {
			return err
		}

		// 更新兑换码状态
		if err := tx.Model(&codeModel).Updates(map[string]interface{}{
			"status":        biz.CodeStatusRedeemed,
			"redemption_at": now,
			"updated_at":    now,
		}).Error; err != nil {
			return err
		}

		// 清除缓存
		cacheKey := fmt.Sprintf("redeem_code:%s", redemption.Code)
		r.data.redis.Del(ctx, cacheKey)

		return nil
	})
}

// ListCodes 列出兑换码
func (r *redeemCodeRepo) ListCodes(ctx context.Context, campaignID, tenantID, productCode, codeType string, status int32, userID int64, pageNum, pageSize int32) ([]*biz.RedeemCode, int32, error) {
	var count int64
	var models []RedeemCodeModel

	db := r.data.db.Model(&RedeemCodeModel{})

	// 应用过滤条件
	if campaignID != "" {
		db = db.Where("campaign_id = ?", campaignID)
	}

	if tenantID != "" {
		db = db.Where("tenant_id = ?", tenantID)
	}

	if status >= 0 {
		db = db.Where("status = ?", status)
	}
	
	// 处理产品代码和码类型，这些可能存储在元数据中
	if productCode != "" {
		// 假设 productCode 存储在 metadata 字段中
		db = db.Where("metadata LIKE ?", "%\"productCode\":\"%"+productCode+"%\"")
	}
	
	if codeType != "" {
		// 假设 codeType 存储在 metadata 字段中
		db = db.Where("metadata LIKE ?", "%\"codeType\":\"%"+codeType+"%\"")
	}
	
	if userID > 0 {
		// 假设 userID 存储在 metadata 字段中
		userIDStr := fmt.Sprintf("%d", userID)
		db = db.Where("metadata LIKE ?", "%\"userID\":\"%"+userIDStr+"%\"")
	}

	// 获取总数
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := int((pageNum - 1) * pageSize)
	limit := int(pageSize)
	if err := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	// 转换为业务对象
	codes := make([]*biz.RedeemCode, 0, len(models))
	for _, model := range models {
		// 解析元数据
		var metadata map[string]interface{}
		if model.Metadata != "" {
			if err := json.Unmarshal([]byte(model.Metadata), &metadata); err != nil {
				r.log.Warnf("failed to unmarshal metadata for code %s: %v", model.Code, err)
			}
		}

		code := &biz.RedeemCode{
			Code:       model.Code,
			CampaignID: model.CampaignID,
			TenantID:   model.TenantID,
			BatchID:    model.BatchID,
			Status:     model.Status,
			ValidFrom:  model.ValidFrom,
			ValidUntil: model.ValidUntil,
			Metadata:   metadata,
			CreatedAt:  model.CreatedAt,
			UpdatedAt:  model.UpdatedAt,
		}

		if model.RedemptionAt != nil {
			code.RedemptionAt = model.RedemptionAt
		}

		codes = append(codes, code)
	}

	return codes, int32(count), nil
}

// GetBatch 获取批次信息
func (r *redeemCodeRepo) GetBatch(ctx context.Context, batchID string) (*biz.CodeBatch, error) {
	var model CodeBatchModel
	if err := r.data.db.Where("batch_id = ?", batchID).First(&model).Error; err != nil {
		return nil, err
	}

	return &biz.CodeBatch{
		BatchID:     model.BatchID,
		CampaignID:  model.CampaignID,
		TenantID:    model.TenantID,
		BatchName:   model.BatchName,
		CodeCount:   model.CodeCount,
		CodePrefix:  model.CodePrefix,
		CodeType:    model.CodeType,
		CodeLength:  model.CodeLength,
		ValidFrom:   model.ValidFrom,
		ValidUntil:  model.ValidUntil,
		Description: model.Description,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}, nil
}

// ListBatches 列出批次
func (r *redeemCodeRepo) ListBatches(ctx context.Context, campaignID string, offset, limit int) ([]*biz.CodeBatch, int64, error) {
	var count int64
	var models []CodeBatchModel

	db := r.data.db.Model(&CodeBatchModel{})

	// 应用过滤条件
	if campaignID != "" {
		db = db.Where("campaign_id = ?", campaignID)
	}

	// 获取总数
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	if err := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	// 转换为业务对象
	batches := make([]*biz.CodeBatch, 0, len(models))
	for _, model := range models {
		batches = append(batches, &biz.CodeBatch{
			BatchID:     model.BatchID,
			CampaignID:  model.CampaignID,
			TenantID:    model.TenantID,
			BatchName:   model.BatchName,
			CodeCount:   model.CodeCount,
			CodePrefix:  model.CodePrefix,
			CodeType:    model.CodeType,
			CodeLength:  model.CodeLength,
			ValidFrom:   model.ValidFrom,
			ValidUntil:  model.ValidUntil,
			Description: model.Description,
			CreatedAt:   model.CreatedAt,
			UpdatedAt:   model.UpdatedAt,
		})
	}

	return batches, count, nil
}

// GetRedemptionRecord 获取兑换记录
func (r *redeemCodeRepo) GetRedemptionRecord(ctx context.Context, redemptionID string) (*biz.Redemption, error) {
	var model RedemptionRecordModel
	if err := r.data.db.Where("redemption_id = ?", redemptionID).First(&model).Error; err != nil {
		return nil, err
	}

	// 解析元数据
	var metadata map[string]interface{}
	if model.Metadata != "" {
		if err := json.Unmarshal([]byte(model.Metadata), &metadata); err != nil {
			r.log.Warnf("failed to unmarshal metadata for redemption %s: %v", redemptionID, err)
		}
	}

	return &biz.Redemption{
		RedemptionID:  model.RedemptionID,
		Code:          model.Code,
		CampaignID:    model.CampaignID,
		TenantID:      model.TenantID,
		UserID:        model.UserID,
		RedemptionAt:  model.RedemptionAt,
		RedemptionIP:  model.RedemptionIP,
		RedemptionEnv: model.RedemptionEnv,
		Metadata:      metadata,
	}, nil
}

// ListRedemptionRecords 列出兑换记录
func (r *redeemCodeRepo) ListRedemptionRecords(ctx context.Context, campaignID, userID string, offset, limit int) ([]*biz.Redemption, int64, error) {
	var count int64
	var models []RedemptionRecordModel

	db := r.data.db.Model(&RedemptionRecordModel{})

	// 应用过滤条件
	if campaignID != "" {
		db = db.Where("campaign_id = ?", campaignID)
	}

	if userID != "" {
		db = db.Where("user_id = ?", userID)
	}

	// 获取总数
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	if err := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	// 转换为业务对象
	redemptions := make([]*biz.Redemption, 0, len(models))
	for _, model := range models {
		// 解析元数据
		var metadata map[string]interface{}
		if model.Metadata != "" {
			if err := json.Unmarshal([]byte(model.Metadata), &metadata); err != nil {
				r.log.Warnf("failed to unmarshal metadata for redemption %s: %v", model.RedemptionID, err)
			}
		}

		redemptions = append(redemptions, &biz.Redemption{
			RedemptionID:  model.RedemptionID,
			Code:          model.Code,
			CampaignID:    model.CampaignID,
			TenantID:      model.TenantID,
			UserID:        model.UserID,
			RedemptionAt:  model.RedemptionAt,
			RedemptionIP:  model.RedemptionIP,
			RedemptionEnv: model.RedemptionEnv,
			Metadata:      metadata,
		})
	}

	return redemptions, count, nil
}
