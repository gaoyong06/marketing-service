package data

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gaoyong06/middleground/marketing-service/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// CampaignModel 营销活动模型
type CampaignModel struct {
	ID              int64      `gorm:"primaryKey;autoIncrement"`
	CampaignID      string     `gorm:"type:varchar(32);uniqueIndex;not null"`
	TenantID        string     `gorm:"type:varchar(32);index;not null"`
	CampaignName    string     `gorm:"type:varchar(128);not null"`
	CampaignType    string     `gorm:"type:varchar(32);not null"`
	Status          int32      `gorm:"type:tinyint;not null;default:1"`
	StartTime       time.Time  `gorm:"type:datetime;not null"`
	EndTime         time.Time  `gorm:"type:datetime;not null"`
	Rules           string     `gorm:"type:text"`
	Description     string     `gorm:"type:varchar(512)"`
	CreatedAt       time.Time  `gorm:"type:datetime;not null"`
	UpdatedAt       time.Time  `gorm:"type:datetime;not null"`
	DeletedAt       *time.Time `gorm:"index"`
}

// TableName 设置表名
func (CampaignModel) TableName() string {
	return "campaign"
}

// CampaignRuleModel 活动规则模型
type CampaignRuleModel struct {
	ID         int64      `gorm:"primaryKey;autoIncrement"`
	CampaignID string     `gorm:"type:varchar(32);index;not null"`
	RuleType   string     `gorm:"type:varchar(32);not null"`
	RuleConfig string     `gorm:"type:text;not null"`
	CreatedAt  time.Time  `gorm:"type:datetime;not null"`
	UpdatedAt  time.Time  `gorm:"type:datetime;not null"`
	DeletedAt  *time.Time `gorm:"index"`
}

// TableName 设置表名
func (CampaignRuleModel) TableName() string {
	return "campaign_rule"
}

type campaignRepo struct {
	data *Data
	log  *log.Helper
}

// NewCampaignRepo 创建活动仓储实现
func NewCampaignRepo(data *Data, logger log.Logger) biz.CampaignRepo {
	return &campaignRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// Create 创建营销活动
func (r *campaignRepo) Create(ctx context.Context, campaign *biz.Campaign) (*biz.Campaign, error) {
	// 序列化规则
	rulesJSON, err := json.Marshal(campaign.Rules)
	if err != nil {
		return nil, err
	}

	// 创建活动模型
	model := &CampaignModel{
		CampaignID:   campaign.CampaignID,
		TenantID:     campaign.TenantID,
		CampaignName: campaign.CampaignName,
		CampaignType: campaign.CampaignType,
		Status:       campaign.Status,
		StartTime:    campaign.StartTime,
		EndTime:      campaign.EndTime,
		Rules:        string(rulesJSON),
		Description:  campaign.Description,
	}

	// 开启事务
	err = r.data.db.Transaction(func(tx *gorm.DB) error {
		// 创建活动记录
		if err := tx.Create(model).Error; err != nil {
			return err
		}

		// 创建活动规则
		for _, rule := range campaign.Rules {
			ruleConfig, err := json.Marshal(rule.Config)
			if err != nil {
				return err
			}

			ruleModel := &CampaignRuleModel{
				CampaignID: campaign.CampaignID,
				RuleType:   rule.RuleType,
				RuleConfig: string(ruleConfig),
			}

			if err := tx.Create(ruleModel).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 返回创建的活动
	return r.Get(ctx, campaign.CampaignID)
}

// Update 更新营销活动
func (r *campaignRepo) Update(ctx context.Context, campaign *biz.Campaign) (*biz.Campaign, error) {
	// 序列化规则
	rulesJSON, err := json.Marshal(campaign.Rules)
	if err != nil {
		return nil, err
	}

	// 更新活动
	updates := map[string]interface{}{
		"campaign_name": campaign.CampaignName,
		"status":        campaign.Status,
		"start_time":    campaign.StartTime,
		"end_time":      campaign.EndTime,
		"rules":         string(rulesJSON),
		"description":   campaign.Description,
		"updated_at":    time.Now(),
	}

	// 开启事务
	err = r.data.db.Transaction(func(tx *gorm.DB) error {
		// 更新活动记录
		if err := tx.Model(&CampaignModel{}).Where("campaign_id = ?", campaign.CampaignID).Updates(updates).Error; err != nil {
			return err
		}

		// 删除旧规则
		if err := tx.Where("campaign_id = ?", campaign.CampaignID).Delete(&CampaignRuleModel{}).Error; err != nil {
			return err
		}

		// 创建新规则
		for _, rule := range campaign.Rules {
			ruleConfig, err := json.Marshal(rule.Config)
			if err != nil {
				return err
			}

			ruleModel := &CampaignRuleModel{
				CampaignID: campaign.CampaignID,
				RuleType:   rule.RuleType,
				RuleConfig: string(ruleConfig),
			}

			if err := tx.Create(ruleModel).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 返回更新后的活动
	return r.Get(ctx, campaign.CampaignID)
}

// Get 获取营销活动
func (r *campaignRepo) Get(ctx context.Context, campaignID string) (*biz.Campaign, error) {
	var model CampaignModel
	if err := r.data.db.Where("campaign_id = ?", campaignID).First(&model).Error; err != nil {
		return nil, err
	}

	// 查询活动规则
	var ruleModels []CampaignRuleModel
	if err := r.data.db.Where("campaign_id = ?", campaignID).Find(&ruleModels).Error; err != nil {
		return nil, err
	}

	// 解析规则
	rules := make([]*biz.CampaignRule, 0, len(ruleModels))
	for _, ruleModel := range ruleModels {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(ruleModel.RuleConfig), &config); err != nil {
			return nil, err
		}

		rules = append(rules, &biz.CampaignRule{
			RuleType: ruleModel.RuleType,
			Config:   config,
		})
	}

	// 构建返回对象
	return &biz.Campaign{
		CampaignID:   model.CampaignID,
		TenantID:     model.TenantID,
		CampaignName: model.CampaignName,
		CampaignType: model.CampaignType,
		Status:       model.Status,
		StartTime:    model.StartTime,
		EndTime:      model.EndTime,
		Rules:        rules,
		Description:  model.Description,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}, nil
}

// List 列出营销活动
func (r *campaignRepo) List(ctx context.Context, tenantID, productCode, campaignType string, status int32, pageNum, pageSize int32) ([]*biz.Campaign, int32, error) {
	var count int64
	var models []CampaignModel

	db := r.data.db.Model(&CampaignModel{})

	// 应用租户过滤
	if tenantID != "" {
		db = db.Where("tenant_id = ?", tenantID)
	}
	
	// 应用产品代码过滤
	if productCode != "" {
		db = db.Where("product_code = ?", productCode)
	}
	
	// 应用活动类型过滤
	if campaignType != "" {
		db = db.Where("campaign_type = ?", campaignType)
	}
	
	// 应用状态过滤
	if status >= 0 {
		db = db.Where("status = ?", status)
	}

	// 获取总数
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// 计算分页参数
	offset := int((pageNum - 1) * pageSize)
	limit := int(pageSize)
	
	// 获取分页数据
	if err := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	// 转换为业务对象
	campaigns := make([]*biz.Campaign, 0, len(models))
	for _, model := range models {
		// 解析规则JSON
		var rules []*biz.CampaignRule
		if model.Rules != "" {
			if err := json.Unmarshal([]byte(model.Rules), &rules); err != nil {
				r.log.Errorf("failed to unmarshal rules for campaign %s: %v", model.CampaignID, err)
				// 继续处理，不中断
			}
		}

		campaigns = append(campaigns, &biz.Campaign{
			CampaignID:   model.CampaignID,
			TenantID:     model.TenantID,
			CampaignName: model.CampaignName,
			CampaignType: model.CampaignType,
			Status:       model.Status,
			StartTime:    model.StartTime,
			EndTime:      model.EndTime,
			Rules:        rules,
			Description:  model.Description,
			CreatedAt:    model.CreatedAt,
			UpdatedAt:    model.UpdatedAt,
		})
	}

	return campaigns, int32(count), nil
}

// Delete 删除营销活动
func (r *campaignRepo) Delete(ctx context.Context, campaignID string) error {
	return r.data.db.Transaction(func(tx *gorm.DB) error {
		// 删除活动规则
		if err := tx.Where("campaign_id = ?", campaignID).Delete(&CampaignRuleModel{}).Error; err != nil {
			return err
		}

		// 删除活动
		if err := tx.Where("campaign_id = ?", campaignID).Delete(&CampaignModel{}).Error; err != nil {
			return err
		}

		return nil
	})
}
