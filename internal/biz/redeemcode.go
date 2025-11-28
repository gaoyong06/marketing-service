package biz

import (
	"context"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// RedeemCode 兑换码领域模型
// 兑换码状态常量
const (
	CodeStatusUnassigned = 0 // 未分配
	CodeStatusActive     = 1 // 已分配
	CodeStatusRedeemed   = 2 // 已核销
	CodeStatusExpired    = 3 // 已失效
)

// 兑换码类型常量
const (
	CodeTypeDiscount = "DISCOUNT" // 折扣码
	CodeTypeCoupon   = "COUPON"   // 优惠码
	CodeTypeGift     = "GIFT"     // 礼品码
)

// 兑换操作类型
const (
	RedemptionTypeNormal = "NORMAL" // 普通兑换
	RedemptionTypeForce  = "FORCE"  // 强制兑换
)

// RedemptionType 兑换类型
type RedemptionType string

// 兑换类型常量
const (
	RedemptionTypeDefault RedemptionType = "REDEMPTION" // 默认兑换状态
)

// Redemption 兑换记录模型
type Redemption struct {
	RedemptionID   string                 // 兑换记录ID
	RedeemCodeID   int64                  // 兑换码ID
	Code           string                 // 兑换码
	CampaignID     string                 // 关联活动ID
	TenantID       string                 // 所属租户
	UserID         string                 // 用户ID
	RedemptionType RedemptionType         // 兑换类型
	RedemptionAt   time.Time              // 兑换时间
	RedemptionIP   string                 // 兑换IP
	RedemptionEnv  string                 // 兑换环境
	Channel        string                 // 兑换渠道
	Operator       string                 // 操作人
	Metadata       map[string]interface{} // 元数据
	CreatedAt      time.Time              // 创建时间
	UpdatedAt      time.Time              // 更新时间
}

type RedeemCode struct {
	RedeemCodeID  int64                  // 主键ID
	Code          string                 // 兑换码
	CampaignID    string                 // 关联活动ID
	TenantID      string                 // 所属租户
	BatchID       string                 // 批次ID
	ProductCode   string                 // 适用产品线
	CodeType      string                 // 码类型：DISCOUNT-折扣码 COUPON-优惠码 GIFT-礼品码
	UserID        int64                  // 分配用户ID
	Status        int32                  // 状态：0-未分配 1-已分配 2-已核销 3-已失效
	ValueAmount   float64                // 面额（金额类）
	ValuePercent  float64                // 折扣比例（百分比类）
	RewardItems   map[string]string      // 奖励物品（实物类）
	Metadata      map[string]interface{} // 元数据，存储额外信息
	ValidFrom     time.Time              // 生效时间
	ValidUntil    time.Time              // 过期时间
	RedemptionAt  *time.Time             // 核销时间
	RedeemChannel string                 // 核销渠道
	CreatedAt     time.Time              // 创建时间
	UpdatedAt     time.Time              // 更新时间
}

// CodeBatch u5151u6362u7801u6279u6b21u9886u57dfu6a21u578b
type CodeBatch struct {
	BatchID      string            // u6279u6b21ID
	BatchName    string            // u6279u6b21u540du79f0
	CampaignID   string            // u5173u8054u6d3bu52a8ID
	TenantID     string            // u6240u5c5eu79dfu6237
	GenerateType string            // u751fu6210u65b9u5f0fuff1aMANUAL-u624bu52a8 AUTO-u81eau52a8 IMPORT-u5bfcu5165
	TotalCount   int32             // u603bu6570u91cf
	UsedCount    int32             // u5df2u4f7fu7528u6570u91cf
	CodeCount    int32             // u7801u6570u91cf
	CodePrefix   string            // u7801u524du7f00
	CodeType     string            // u7801u7c7bu578b
	CodeLength   int32             // u7801u957fu5ea6
	ValidFrom    time.Time         // u751fu6548u65f6u95f4
	ValidUntil   time.Time         // u8fc7u671fu65f6u95f4
	Description  string            // u63cfu8ff0
	Operator     string            // u64cdu4f5cu4eba
	GenerateRule map[string]string // u751fu6210u89c4u5219
	CreatedAt    time.Time         // u521bu5efau65f6u95f4
	UpdatedAt    time.Time         // u66f4u65b0u65f6u95f4
}

// RedemptionRecord u5151u6362u8bb0u5f55u9886u57dfu6a21u578b
type RedemptionRecord struct {
	RecordID      int64             // 记录ID
	RedeemCodeID  int64             // 兑换码ID
	Code          string            // 兑换码值
	UserID        int64             // 用户ID
	TenantID      string            // 所属租户
	CampaignID    string            // 关联活动ID
	ProductCode   string            // 产品线
	RedeemChannel string            // 核销渠道
	RedemptionAt  time.Time         // 核销时间
	DeviceInfo    string            // 设备信息
	IPAddress     string            // IP地址
	Location      string            // 地理位置
	OrderID       string            // 关联订单ID
	RewardDetail  map[string]string // 奖励详情
}

// CodeGenerateRule u7801u751fu6210u89c4u5219
type CodeGenerateRule struct {
	CodeType     string   // u7801u7c7bu578buff1aRANDOM_NUM/ALPHANUM/UUID
	Length       int      // u957fu5ea6
	Prefix       string   // u524du7f00
	ExcludeChars []string // u9700u8981u6392u9664u7684u5b57u7b26
	CampaignID   string   // u5173u8054u6d3bu52a8ID
}

// CodeGenerator u7801u751fu6210u5668u63a5u53e3
type CodeGenerator interface {
	Generate(rule CodeGenerateRule) (string, error)
}

// RedeemCodeRepo u5151u6362u7801u4ed3u50a8u63a5u53e3
type RedeemCodeRepo interface {
	CreateCode(ctx context.Context, code *RedeemCode) (*RedeemCode, error)
	GetCode(ctx context.Context, codeStr string) (*RedeemCode, error)
	UpdateCode(ctx context.Context, code *RedeemCode) (*RedeemCode, error)
	ListCodes(ctx context.Context, campaignID, tenantID, productCode, codeType string, status int32, userID int64, pageNum, pageSize int32) ([]*RedeemCode, int32, error)
	CreateBatch(ctx context.Context, batch *CodeBatch, codes []*RedeemCode) (*CodeBatch, error)
	GetBatch(ctx context.Context, batchID string) (*CodeBatch, error)
	CreateRedemptionRecord(ctx context.Context, record *RedemptionRecord) (*RedemptionRecord, error)
	GetRedemptionRecords(ctx context.Context, code string) ([]*RedemptionRecord, error)
	AssignCode(ctx context.Context, code string, userID int64) (*RedeemCode, error)
	RedeemCode(ctx context.Context, code string, userID int64, productCode, redeemChannel, deviceInfo, ipAddress, location, orderID string) (*RedeemCode, *RedemptionRecord, error)
}

// RedeemCodeUsecase u5151u6362u7801u7528u4f8b
type RedeemCodeUsecase struct {
	repo         RedeemCodeRepo
	campaignRepo CampaignRepo
	tenantRepo   TenantRepo
	generator    CodeGenerator
	log          *log.Helper
}

// NewRedeemCodeUsecase u521bu5efau5151u6362u7801u7528u4f8b
func NewRedeemCodeUsecase(repo RedeemCodeRepo, campaignRepo CampaignRepo, tenantRepo TenantRepo, generator CodeGenerator, logger log.Logger) *RedeemCodeUsecase {
	return &RedeemCodeUsecase{
		repo:         repo,
		campaignRepo: campaignRepo,
		tenantRepo:   tenantRepo,
		generator:    generator,
		log:          log.NewHelper(logger),
	}
}

// GenerateRedeemCodes u751fu6210u5151u6362u7801
func (uc *RedeemCodeUsecase) GenerateRedeemCodes(ctx context.Context, campaignID, codeType, generateType string, count int32, generateRule map[string]string, valueAmount, valuePercent float64, rewardItems map[string]string, expireTime time.Time, operator string) (*CodeBatch, []string, error) {
	uc.log.WithContext(ctx).Infof("GenerateRedeemCodes: campaignID=%v, count=%v", campaignID, count)

	// u83b7u53d6u6d3bu52a8u4fe1u606f
	campaign, err := uc.campaignRepo.Get(ctx, campaignID)
	if err != nil {
		return nil, nil, err
	}

	// u68c0u67e5u79dfu6237u914du989d
	success, _, err := uc.tenantRepo.CheckAndConsumeQuota(ctx, campaign.TenantID, "REDEEM_CODE", "TOTAL", count, campaign.ProductCode)
	if err != nil || !success {
		return nil, nil, err
	}

	// u521bu5efau6279u6b21
	batch := &CodeBatch{
		BatchID:      generateBatchID(),
		CampaignID:   campaignID,
		TenantID:     campaign.TenantID,
		GenerateType: generateType,
		TotalCount:   count,
		UsedCount:    0,
		Operator:     operator,
		GenerateRule: generateRule,
		CreatedAt:    time.Now(),
	}

	// u751fu6210u5151u6362u7801
	codes := make([]*RedeemCode, 0, count)
	sampleCodes := make([]string, 0, 10) // u6700u591au8fd4u56de10u4e2au6837u4f8bu7801

	// u6784u5efau751fu6210u89c4u5219
	rule := CodeGenerateRule{
		CodeType:   generateRule["codeType"],
		Length:     getIntFromMap(generateRule, "length", 8),
		Prefix:     generateRule["prefix"],
		CampaignID: campaignID,
	}

	for i := int32(0); i < count; i++ {
		// u751fu6210u552fu4e00u7801
		codeStr, err := uc.generator.Generate(rule)
		if err != nil {
			return nil, nil, err
		}

		// u521bu5efau5151u6362u7801u5bf9u8c61
		code := &RedeemCode{
			Code:         codeStr,
			CampaignID:   campaignID,
			TenantID:     campaign.TenantID,
			ProductCode:  campaign.ProductCode,
			CodeType:     codeType,
			Status:       0, // u672au5206u914d
			ValueAmount:  valueAmount,
			ValuePercent: valuePercent,
			RewardItems:  rewardItems,
			ValidFrom:    time.Now(),
			ValidUntil:   expireTime,
			CreatedAt:    time.Now(),
		}

		codes = append(codes, code)

		// u6536u96c6u6837u4f8bu7801
		if len(sampleCodes) < 10 {
			sampleCodes = append(sampleCodes, codeStr)
		}
	}

	// u6279u91cfu521bu5efau5151u6362u7801
	createdBatch, err := uc.repo.CreateBatch(ctx, batch, codes)
	if err != nil {
		return nil, nil, err
	}

	return createdBatch, sampleCodes, nil
}

// AssignRedeemCode u5206u914du5151u6362u7801
func (uc *RedeemCodeUsecase) AssignRedeemCode(ctx context.Context, code string, userID int64) (*RedeemCode, error) {
	uc.log.WithContext(ctx).Infof("AssignRedeemCode: code=%v, userID=%v", code, userID)
	return uc.repo.AssignCode(ctx, code, userID)
}

// RedeemCode u5151u6362u7801u6838u9500
func (uc *RedeemCodeUsecase) RedeemCode(ctx context.Context, code string, userID int64, productCode, redeemChannel, deviceInfo, ipAddress, location, orderID string) (*RedeemCode, *RedemptionRecord, error) {
	uc.log.WithContext(ctx).Infof("RedeemCode: code=%v, userID=%v", code, userID)
	return uc.repo.RedeemCode(ctx, code, userID, productCode, redeemChannel, deviceInfo, ipAddress, location, orderID)
}

// ListRedeemCodes u5217u51fau5151u6362u7801
func (uc *RedeemCodeUsecase) ListRedeemCodes(ctx context.Context, campaignID, tenantID, productCode, codeType string, status int32, userID int64, pageNum, pageSize int32) ([]*RedeemCode, int32, error) {
	uc.log.WithContext(ctx).Infof("ListRedeemCodes: campaignID=%v, tenantID=%v", campaignID, tenantID)
	return uc.repo.ListCodes(ctx, campaignID, tenantID, productCode, codeType, status, userID, pageNum, pageSize)
}

// GetRedeemCode u83b7u53d6u5151u6362u7801
func (uc *RedeemCodeUsecase) GetRedeemCode(ctx context.Context, code string) (*RedeemCode, error) {
	uc.log.WithContext(ctx).Infof("GetRedeemCode: code=%v", code)
	return uc.repo.GetCode(ctx, code)
}

// Helper functions

// generateBatchID u751fu6210u6279u6b21ID
func generateBatchID() string {
	// u5b9eu9645u5b9eu73b0u4e2du5e94u8be5u4f7fu7528UUIDu6216u5176u4ed6u552fu4e00IDu751fu6210u7b97u6cd5
	return "BATCH_" + time.Now().Format("20060102150405")
}

// getIntFromMap u4eceu6620u5c04u4e2du83b7u53d6u6574u6570u503c
func getIntFromMap(m map[string]string, key string, defaultValue int) int {
	if val, ok := m[key]; ok {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}
