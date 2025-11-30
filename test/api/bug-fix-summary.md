# Campaign 和 Reward 更新 API Bug 修复总结

## 问题描述

Campaign 和 Reward 的更新 API 在更新名称时没有生效，更新后返回的名称仍然是旧值。

## 问题分析

### 根本原因

GORM 的 `Updates` 方法在使用结构体时，会忽略零值字段。当使用 `Updates(m)` 更新时，如果结构体中的某些字段是零值（如 `time.Time{}`、空字符串等），这些字段不会被更新。

虽然 `CampaignName` 和 `Name` 字段不是零值，但由于 `Updates` 方法在处理结构体时的行为，可能导致某些字段更新失败。

### 代码位置

1. **Campaign 更新**：`internal/data/campaign.go:86-103`
2. **Reward 更新**：`internal/data/reward.go:103-145`

## 修复方案

### 修复方法

将 `Updates(m)` 改为使用 `map[string]interface{}` 明确指定要更新的字段，避免零值字段被忽略的问题。

### 修复代码

#### Campaign Update 修复

**修复前**：
```go
func (r *campaignRepo) Update(ctx context.Context, c *biz.Campaign) (*biz.Campaign, error) {
	m := r.toDataModel(c)
	if err := r.data.db.WithContext(ctx).Model(&model.Campaign{}).
		Where("campaign_id = ?", m.CampaignID).Updates(m).Error; err != nil {
		// ...
	}
	// ...
}
```

**修复后**：
```go
func (r *campaignRepo) Update(ctx context.Context, c *biz.Campaign) (*biz.Campaign, error) {
	m := r.toDataModel(c)
	// 使用 map 明确指定要更新的字段，避免零值字段被忽略
	updateFields := map[string]interface{}{
		"campaign_name":    m.CampaignName,
		"campaign_type":    m.CampaignType,
		"start_time":       m.StartTime,
		"end_time":         m.EndTime,
		"audience_config":  m.AudienceConfig,
		"validator_config": m.ValidatorConfig,
		"status":           m.Status,
		"description":      m.Description,
		"updated_at":       m.UpdatedAt,
	}
	if err := r.data.db.WithContext(ctx).Model(&model.Campaign{}).
		Where("campaign_id = ?", m.CampaignID).Updates(updateFields).Error; err != nil {
		// ...
	}
	// ...
}
```

#### Reward Update 修复

**修复前**：
```go
func (r *rewardRepo) Update(ctx context.Context, reward *biz.Reward) (*biz.Reward, error) {
	// ...
	m := r.toDataModel(reward)
	m.Version = current.Version + 1

	if err := r.data.db.WithContext(ctx).Model(&model.Reward{}).
		Where("reward_id = ?", m.RewardID).Updates(m).Error; err != nil {
		// ...
	}
	// ...
}
```

**修复后**：
```go
func (r *rewardRepo) Update(ctx context.Context, reward *biz.Reward) (*biz.Reward, error) {
	// ...
	m := r.toDataModel(reward)
	m.Version = current.Version + 1

	// 使用 map 明确指定要更新的字段，避免零值字段被忽略
	updateFields := map[string]interface{}{
		"reward_type":        m.RewardType,
		"name":               m.Name,
		"content_config":     m.ContentConfig,
		"generator_config":   m.GeneratorConfig,
		"distributor_config": m.DistributorConfig,
		"validator_config":   m.ValidatorConfig,
		"version":            m.Version,
		"valid_days":         m.ValidDays,
		"extra_config":       m.ExtraConfig,
		"status":             m.Status,
		"description":        m.Description,
		"updated_at":         m.UpdatedAt,
	}
	if err := r.data.db.WithContext(ctx).Model(&model.Reward{}).
		Where("reward_id = ?", m.RewardID).Updates(updateFields).Error; err != nil {
		// ...
	}
	// ...
}
```

## 修复效果

### 预期效果

- ✅ Campaign 更新 API 可以正确更新 `campaign_name` 字段
- ✅ Reward 更新 API 可以正确更新 `name` 字段
- ✅ 所有其他字段也能正确更新

### 验证方法

1. 创建 Campaign/Reward
2. 调用更新 API 修改名称
3. 验证返回的名称是否已更新
4. 查询数据库确认名称已更新

## 注意事项

1. **服务重启**：修复后需要重启服务才能生效
2. **测试覆盖**：测试用例已经覆盖更新功能，修复后所有测试应该通过
3. **向后兼容**：此修复不影响现有功能，只是修复了更新逻辑

## 相关文件

- `internal/data/campaign.go` - Campaign Repository Update 方法
- `internal/data/reward.go` - Reward Repository Update 方法
- `internal/service/marketing.go` - Campaign 和 Reward 更新服务方法（添加了调试日志）

## 状态

- ✅ **已修复**：代码已修改，使用 `map[string]interface{}` 明确指定更新字段
- ⚠️ **待验证**：需要重启服务并验证修复是否生效

---

**修复时间**：2025-11-30
**修复人员**：AI Assistant
**修复原因**：GORM `Updates` 方法忽略零值字段导致更新失败
