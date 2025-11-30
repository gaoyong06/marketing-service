package data

import (
	"context"

	"marketing-service/internal/biz"
	"marketing-service/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/datatypes"
)

// campaignTaskRepo 活动-任务关联仓储实现
type campaignTaskRepo struct {
	data *Data
	log  *log.Helper
}

// NewCampaignTaskRepo 创建活动-任务关联仓储
func NewCampaignTaskRepo(data *Data, logger log.Logger) biz.CampaignTaskRepo {
	return &campaignTaskRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// Save 保存活动-任务关联
func (r *campaignTaskRepo) Save(ctx context.Context, ct *biz.CampaignTask) (*biz.CampaignTask, error) {
	m := &model.CampaignTask{
		CampaignID: ct.CampaignID,
		TaskID:     ct.TaskID,
		SortOrder:  ct.SortOrder,
	}
	if ct.Config != "" {
		m.Config = datatypes.JSON(ct.Config)
	}

	if err := r.data.db.WithContext(ctx).Create(m).Error; err != nil {
		r.log.Errorf("failed to save campaign task: %v", err)
		return nil, err
	}

	return r.toBizModel(m), nil
}

// Delete 删除活动-任务关联
func (r *campaignTaskRepo) Delete(ctx context.Context, campaignID, taskID string) error {
	if err := r.data.db.WithContext(ctx).
		Where("campaign_id = ? AND task_id = ?", campaignID, taskID).
		Delete(&model.CampaignTask{}).Error; err != nil {
		r.log.Errorf("failed to delete campaign task: %v", err)
		return err
	}
	return nil
}

// ListByCampaign 列出活动的所有任务
func (r *campaignTaskRepo) ListByCampaign(ctx context.Context, campaignID string) ([]*biz.CampaignTask, error) {
	var models []model.CampaignTask
	if err := r.data.db.WithContext(ctx).
		Where("campaign_id = ?", campaignID).
		Order("sort_order ASC, created_at ASC").
		Find(&models).Error; err != nil {
		r.log.Errorf("failed to list campaign tasks: %v", err)
		return nil, err
	}

	result := make([]*biz.CampaignTask, 0, len(models))
	for _, m := range models {
		result = append(result, r.toBizModel(&m))
	}
	return result, nil
}

// toBizModel 转换为业务模型
func (r *campaignTaskRepo) toBizModel(m *model.CampaignTask) *biz.CampaignTask {
	config := ""
	if len(m.Config) > 0 {
		config = string(m.Config)
	}
	return &biz.CampaignTask{
		CampaignTaskID: m.CampaignTaskID,
		CampaignID:     m.CampaignID,
		TaskID:         m.TaskID,
		Config:         config,
		SortOrder:      m.SortOrder,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

