package biz

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

// Generator 生成器接口
type Generator interface {
	Generate(ctx context.Context, req *GenerationRequest) (string, error)
}

// GenerationRequest 生成请求
type GenerationRequest struct {
	RewardID   string
	RewardType string
	UserID     int64
	Config     map[string]interface{} // 生成器配置
	Reward     *Reward                // 奖励信息
}

// GeneratorService 生成器服务
type GeneratorService struct {
	generators map[string]Generator
	log        *log.Helper
}

// NewGeneratorService 创建生成器服务
func NewGeneratorService(logger log.Logger) *GeneratorService {
	gs := &GeneratorService{
		generators: make(map[string]Generator),
		log:        log.NewHelper(logger),
	}

	// 注册内置生成器
	gs.Register("CODE", NewCodeGenerator())
	gs.Register("COUPON", NewCouponGenerator())
	gs.Register("POINTS", NewPointsGenerator())

	return gs
}

// Register 注册生成器
func (gs *GeneratorService) Register(generatorType string, generator Generator) {
	gs.generators[generatorType] = generator
}

// Generate 生成奖励内容
func (gs *GeneratorService) Generate(ctx context.Context, req *GenerationRequest) (string, error) {
	if req.Config == nil {
		// 无配置，使用默认生成逻辑
		return gs.generateDefault(ctx, req)
	}

	generatorType, ok := req.Config["type"].(string)
	if !ok {
		generatorType = req.RewardType // 使用奖励类型作为生成器类型
	}

	generator, exists := gs.generators[generatorType]
	if !exists {
		gs.log.Warnf("unknown generator type: %s, using default", generatorType)
		return gs.generateDefault(ctx, req)
	}

	return generator.Generate(ctx, req)
}

// generateDefault 默认生成逻辑
func (gs *GeneratorService) generateDefault(ctx context.Context, req *GenerationRequest) (string, error) {
	// 根据奖励类型选择生成器
	switch req.RewardType {
	case "REDEEM_CODE":
		return NewCodeGenerator().Generate(ctx, req)
	case "COUPON":
		return NewCouponGenerator().Generate(ctx, req)
	case "POINTS":
		return NewPointsGenerator().Generate(ctx, req)
	default:
		// 返回基础内容配置
		return req.Reward.ContentConfig, nil
	}
}

// ========== 内置生成器实现 ==========

// CodeGenerator 兑换码生成器
type CodeGenerator struct{}

// NewCodeGenerator 创建兑换码生成器
func NewCodeGenerator() Generator {
	return &CodeGenerator{}
}

// Generate 生成兑换码
func (g *CodeGenerator) Generate(ctx context.Context, req *GenerationRequest) (string, error) {
	// 生成兑换码
	code := generateRedeemCode()

	// 构建内容配置
	content := map[string]interface{}{
		"code":      code,
		"code_type": getStringValue(req.Config, "code_type", "COUPON"),
		"value":     getFloatValue(req.Config, "value", 0),
	}

	contentJSON, err := json.Marshal(content)
	if err != nil {
		return "", err
	}

	return string(contentJSON), nil
}

// CouponGenerator 优惠券生成器
type CouponGenerator struct{}

// NewCouponGenerator 创建优惠券生成器
func NewCouponGenerator() Generator {
	return &CouponGenerator{}
}

// Generate 生成优惠券
func (g *CouponGenerator) Generate(ctx context.Context, req *GenerationRequest) (string, error) {
	// 生成优惠券ID
	couponID := uuid.New().String()

	// 构建内容配置
	content := map[string]interface{}{
		"coupon_id":      couponID,
		"discount_type":  getStringValue(req.Config, "discount_type", "AMOUNT"),
		"discount_value": getFloatValue(req.Config, "discount_value", 0),
		"min_amount":     getFloatValue(req.Config, "min_amount", 0),
	}

	contentJSON, err := json.Marshal(content)
	if err != nil {
		return "", err
	}

	return string(contentJSON), nil
}

// PointsGenerator 积分生成器
type PointsGenerator struct{}

// NewPointsGenerator 创建积分生成器
func NewPointsGenerator() Generator {
	return &PointsGenerator{}
}

// Generate 生成积分
func (g *PointsGenerator) Generate(ctx context.Context, req *GenerationRequest) (string, error) {
	// 获取积分数量
	points := getFloatValue(req.Config, "points", 0)
	if points <= 0 {
		// 从奖励内容配置中获取
		var contentConfig map[string]interface{}
		if err := json.Unmarshal([]byte(req.Reward.ContentConfig), &contentConfig); err == nil {
			if p, ok := contentConfig["points"].(float64); ok {
				points = p
			}
		}
	}

	// 构建内容配置
	content := map[string]interface{}{
		"points":      points,
		"points_type": getStringValue(req.Config, "points_type", "CASH"),
	}

	contentJSON, err := json.Marshal(content)
	if err != nil {
		return "", err
	}

	return string(contentJSON), nil
}

// generateRedeemCode 生成兑换码
func generateRedeemCode() string {
	chars := "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	code := make([]byte, 8)
	rand.Seed(time.Now().UnixNano())
	for i := range code {
		code[i] = chars[rand.Intn(len(chars))]
	}
	return string(code)
}

// getStringValue 从配置中获取字符串值
func getStringValue(config map[string]interface{}, key, defaultValue string) string {
	if val, ok := config[key].(string); ok {
		return val
	}
	return defaultValue
}

// getFloatValue 从配置中获取浮点数值
func getFloatValue(config map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := config[key].(float64); ok {
		return val
	}
	return defaultValue
}

// ParseGeneratorConfig 解析生成器配置
func ParseGeneratorConfig(configJSON string) (map[string]interface{}, error) {
	if configJSON == "" {
		return nil, nil
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, err
	}

	return config, nil
}
