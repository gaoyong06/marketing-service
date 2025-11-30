package biz

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// Validator 校验器接口
type Validator interface {
	Validate(ctx context.Context, req *ValidationRequest) error
}

// ValidationRequest 校验请求
type ValidationRequest struct {
	RewardID    string
	CampaignID  string
	UserID      int64
	TenantID    string
	AppID       string
	Config      map[string]interface{} // 校验器配置
	Reward      *Reward                // 奖励信息
	Campaign    *Campaign              // 活动信息
}

// ValidatorService 校验器服务
type ValidatorService struct {
	validators map[string]Validator
	log        *log.Helper
}

// NewValidatorService 创建校验器服务
func NewValidatorService(logger log.Logger) *ValidatorService {
	vs := &ValidatorService{
		validators: make(map[string]Validator),
		log:        log.NewHelper(logger),
	}

	// 注册内置校验器
	vs.Register("TIME", NewTimeValidator())
	vs.Register("USER", NewUserValidator())
	vs.Register("LIMIT", NewLimitValidator())
	vs.Register("INVENTORY", NewInventoryValidator())

	return vs
}

// Register 注册校验器
func (vs *ValidatorService) Register(validatorType string, validator Validator) {
	vs.validators[validatorType] = validator
}

// Validate 执行校验链
func (vs *ValidatorService) Validate(ctx context.Context, req *ValidationRequest) error {
	if req.Config == nil {
		return nil // 无配置，跳过校验
	}

	// 解析校验器配置
	validators, ok := req.Config["validators"].([]interface{})
	if !ok {
		// 单个校验器配置
		return vs.validateSingle(ctx, req)
	}

	// 多个校验器链式校验
	for _, v := range validators {
		validatorConfig, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		validatorType, ok := validatorConfig["type"].(string)
		if !ok {
			continue
		}

		validator, exists := vs.validators[validatorType]
		if !exists {
			vs.log.Warnf("unknown validator type: %s", validatorType)
			continue
		}

		// 创建子请求
		subReq := &ValidationRequest{
			RewardID:   req.RewardID,
			CampaignID: req.CampaignID,
			UserID:     req.UserID,
			TenantID:   req.TenantID,
			AppID:      req.AppID,
			Config:     validatorConfig,
			Reward:     req.Reward,
			Campaign:   req.Campaign,
		}

		if err := validator.Validate(ctx, subReq); err != nil {
			return err
		}
	}

	return nil
}

// validateSingle 校验单个校验器
func (vs *ValidatorService) validateSingle(ctx context.Context, req *ValidationRequest) error {
	validatorType, ok := req.Config["type"].(string)
	if !ok {
		return nil
	}

	validator, exists := vs.validators[validatorType]
	if !exists {
		vs.log.Warnf("unknown validator type: %s", validatorType)
		return nil
	}

	return validator.Validate(ctx, req)
}

// ========== 内置校验器实现 ==========

// TimeValidator 时间校验器
type TimeValidator struct{}

// NewTimeValidator 创建时间校验器
func NewTimeValidator() Validator {
	return &TimeValidator{}
}

// Validate 校验时间范围
func (v *TimeValidator) Validate(ctx context.Context, req *ValidationRequest) error {
	startTimeStr, ok := req.Config["start_time"].(string)
	if ok && startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err == nil && time.Now().Before(startTime) {
			return ErrValidationFailed("reward not yet available")
		}
	}

	endTimeStr, ok := req.Config["end_time"].(string)
	if ok && endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err == nil && time.Now().After(endTime) {
			return ErrValidationFailed("reward expired")
		}
	}

	return nil
}

// UserValidator 用户资格校验器
type UserValidator struct{}

// NewUserValidator 创建用户资格校验器
func NewUserValidator() Validator {
	return &UserValidator{}
}

// Validate 校验用户资格
func (v *UserValidator) Validate(ctx context.Context, req *ValidationRequest) error {
	// TODO: 实现用户资格校验逻辑
	// 需要配合 Audience 进行用户圈选验证
	// 这里简化处理，实际应该调用用户服务或查询 Audience 配置
	return nil
}

// LimitValidator 频次限制校验器
type LimitValidator struct {
	repo RewardGrantRepo
}

// NewLimitValidator 创建频次限制校验器
func NewLimitValidator() Validator {
	// 注意：这里需要注入 RewardGrantRepo，但为了简化，先返回基础实现
	// 实际使用时应该通过依赖注入传入
	return &LimitValidator{}
}

// SetRepo 设置 Repository（用于依赖注入）
func (v *LimitValidator) SetRepo(repo RewardGrantRepo) {
	v.repo = repo
}

// Validate 校验频次限制
func (v *LimitValidator) Validate(ctx context.Context, req *ValidationRequest) error {
	if v.repo == nil {
		return nil // 无 repo，跳过校验
	}

	// 检查用户限制
	userLimit, ok := req.Config["user_limit"].(float64)
	if ok && userLimit > 0 {
		count, err := v.repo.CountByStatus(ctx, req.RewardID, "DISTRIBUTED")
		if err != nil {
			return err
		}
		if count >= int64(userLimit) {
			return ErrValidationFailed("user limit exceeded")
		}
	}

	// 检查活动总限制
	totalLimit, ok := req.Config["total_limit"].(float64)
	if ok && totalLimit > 0 {
		// TODO: 需要按活动统计，这里简化处理
	}

	return nil
}

// InventoryValidator 库存校验器
type InventoryValidator struct {
	repo InventoryReservationRepo
}

// NewInventoryValidator 创建库存校验器
func NewInventoryValidator() Validator {
	return &InventoryValidator{}
}

// SetRepo 设置 Repository
func (v *InventoryValidator) SetRepo(repo InventoryReservationRepo) {
	v.repo = repo
}

// Validate 校验库存
func (v *InventoryValidator) Validate(ctx context.Context, req *ValidationRequest) error {
	if v.repo == nil {
		return nil
	}

	// 检查库存配置
	maxInventory, ok := req.Config["max_inventory"].(float64)
	if !ok || maxInventory <= 0 {
		return nil // 无库存限制
	}

	// 统计已发放数量
	resourceID := req.RewardID
	pendingCount, err := v.repo.CountPendingByResource(ctx, resourceID)
	if err != nil {
		return err
	}

	if int64(pendingCount) >= int64(maxInventory) {
		return ErrValidationFailed("inventory exhausted")
	}

	return nil
}

// ErrValidationFailed 校验失败错误
type ErrValidationFailed string

func (e ErrValidationFailed) Error() string {
	return string(e)
}

// ParseValidatorConfig 解析校验器配置
func ParseValidatorConfig(configJSON string) (map[string]interface{}, error) {
	if configJSON == "" {
		return nil, nil
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, err
	}

	return config, nil
}

