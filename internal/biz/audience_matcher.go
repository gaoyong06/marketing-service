package biz

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
)

// AudienceMatcher 受众圈选匹配器
type AudienceMatcher interface {
	// MatchUser 检查用户是否匹配受众规则
	MatchUser(ctx context.Context, userID int64, audience *Audience) (bool, error)
}

// AudienceMatcherService 受众圈选服务
type AudienceMatcherService struct {
	repo   AudienceRepo
	log    *log.Helper
	matchers map[string]AudienceMatcher
}

// NewAudienceMatcherService 创建受众圈选服务
// 注意：需要传入 AudienceRepo，但为了简化 Wire 依赖，这里使用接口
func NewAudienceMatcherService(repo AudienceRepo, logger log.Logger) *AudienceMatcherService {
	ams := &AudienceMatcherService{
		repo:     repo,
		log:      log.NewHelper(logger),
		matchers: make(map[string]AudienceMatcher),
	}

	// 注册内置匹配器
	ams.Register("TAG", NewTagMatcher())
	ams.Register("SEGMENT", NewSegmentMatcher())
	ams.Register("LIST", NewListMatcher())
	ams.Register("ALL", NewAllMatcher())

	return ams
}

// Register 注册匹配器
func (ams *AudienceMatcherService) Register(audienceType string, matcher AudienceMatcher) {
	ams.matchers[audienceType] = matcher
}

// MatchUser 检查用户是否匹配受众
func (ams *AudienceMatcherService) MatchUser(ctx context.Context, userID int64, audience *Audience) (bool, error) {
	if audience == nil {
		return false, fmt.Errorf("audience is nil")
	}

	matcher, exists := ams.matchers[audience.AudienceType]
	if !exists {
		ams.log.Warnf("unknown audience type: %s", audience.AudienceType)
		return false, fmt.Errorf("unknown audience type: %s", audience.AudienceType)
	}

	return matcher.MatchUser(ctx, userID, audience)
}

// MatchAudienceConfig 检查用户是否匹配受众配置（支持多受众组合）
func (ams *AudienceMatcherService) MatchAudienceConfig(ctx context.Context, userID int64, audienceConfig string) (bool, error) {
	if audienceConfig == "" {
		return true, nil // 无受众配置，默认通过
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(audienceConfig), &config); err != nil {
		return false, fmt.Errorf("failed to parse audience config: %w", err)
	}

	// 解析逻辑关系（AND/OR）
	logic, _ := config["logic"].(string)
	if logic == "" {
		logic = "OR" // 默认 OR
	}

	// 解析受众列表
	items, _ := config["items"].([]interface{})
	exclude, _ := config["exclude"].([]interface{})

	// 检查排除列表
	if len(exclude) > 0 {
		for _, item := range exclude {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			itemType, _ := itemMap["type"].(string)
			itemID, _ := itemMap["id"].(string)

			if itemType == "AUDIENCE" && itemID != "" {
				audience, err := ams.repo.FindByID(ctx, itemID)
				if err != nil {
					ams.log.Warnf("failed to find audience %s: %v", itemID, err)
					continue
				}

				matched, err := ams.MatchUser(ctx, userID, audience)
				if err != nil {
					return false, err
				}

				if matched {
					return false, nil // 在排除列表中，不匹配
				}
			}
		}
	}

	// 检查包含列表
	if len(items) == 0 {
		return true, nil // 无包含列表，默认通过
	}

	results := make([]bool, 0, len(items))
	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		itemType, _ := itemMap["type"].(string)
		itemID, _ := itemMap["id"].(string)

		if itemType == "AUDIENCE" && itemID != "" {
			audience, err := ams.repo.FindByID(ctx, itemID)
			if err != nil {
				ams.log.Warnf("failed to find audience %s: %v", itemID, err)
				continue
			}

			matched, err := ams.MatchUser(ctx, userID, audience)
			if err != nil {
				return false, err
			}

			results = append(results, matched)
		}
	}

	if len(results) == 0 {
		return true, nil
	}

	// 根据逻辑关系判断
	if logic == "AND" {
		// AND: 所有结果都为 true
		for _, result := range results {
			if !result {
				return false, nil
			}
		}
		return true, nil
	} else {
		// OR: 至少一个结果为 true
		for _, result := range results {
			if result {
				return true, nil
			}
		}
		return false, nil
	}
}

// ========== 内置匹配器实现 ==========

// TagMatcher 标签匹配器
type TagMatcher struct{}

// NewTagMatcher 创建标签匹配器
func NewTagMatcher() AudienceMatcher {
	return &TagMatcher{}
}

// MatchUser 检查用户是否匹配标签
func (m *TagMatcher) MatchUser(ctx context.Context, userID int64, audience *Audience) (bool, error) {
	// TODO: 实现标签匹配逻辑
	// 需要调用用户服务或查询用户标签表
	// 这里简化处理，实际应该根据 RuleConfig 中的标签规则进行匹配
	return true, nil
}

// SegmentMatcher 画像分群匹配器
type SegmentMatcher struct{}

// NewSegmentMatcher 创建画像分群匹配器
func NewSegmentMatcher() AudienceMatcher {
	return &SegmentMatcher{}
}

// MatchUser 检查用户是否匹配画像分群
func (m *SegmentMatcher) MatchUser(ctx context.Context, userID int64, audience *Audience) (bool, error) {
	// TODO: 实现画像分群匹配逻辑
	// 需要调用用户服务或查询用户画像数据
	// 这里简化处理，实际应该根据 RuleConfig 中的分群规则进行匹配
	return true, nil
}

// ListMatcher 名单匹配器
type ListMatcher struct{}

// NewListMatcher 创建名单匹配器
func NewListMatcher() AudienceMatcher {
	return &ListMatcher{}
}

// MatchUser 检查用户是否在名单中
func (m *ListMatcher) MatchUser(ctx context.Context, userID int64, audience *Audience) (bool, error) {
	// TODO: 实现名单匹配逻辑
	// 需要查询名单表或从 RuleConfig 中解析用户ID列表
	// 这里简化处理，实际应该根据 RuleConfig 中的名单进行匹配
	return true, nil
}

// AllMatcher 全量用户匹配器
type AllMatcher struct{}

// NewAllMatcher 创建全量用户匹配器
func NewAllMatcher() AudienceMatcher {
	return &AllMatcher{}
}

// MatchUser 全量用户匹配（总是返回 true）
func (m *AllMatcher) MatchUser(ctx context.Context, userID int64, audience *Audience) (bool, error) {
	return true, nil
}

