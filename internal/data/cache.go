package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"marketing-service/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

// CacheService 缓存服务
type CacheService struct {
	rdb  *redis.Client
	log  *log.Helper
}

// NewCacheService 创建缓存服务
func NewCacheService(rdb *redis.Client, logger log.Logger) *CacheService {
	return &CacheService{
		rdb: rdb,
		log: log.NewHelper(logger),
	}
}

// ========== Campaign 缓存 ==========

// GetCampaign 从缓存获取活动
func (c *CacheService) GetCampaign(ctx context.Context, campaignID string) (*biz.Campaign, error) {
	key := fmt.Sprintf("campaign:%s", campaignID)
	data, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		c.log.Warnf("failed to get campaign from cache: %v", err)
		return nil, err
	}

	var campaign biz.Campaign
	if err := json.Unmarshal([]byte(data), &campaign); err != nil {
		c.log.Warnf("failed to unmarshal campaign from cache: %v", err)
		return nil, err
	}

	return &campaign, nil
}

// SetCampaign 设置活动缓存
func (c *CacheService) SetCampaign(ctx context.Context, campaign *biz.Campaign, ttl time.Duration) error {
	key := fmt.Sprintf("campaign:%s", campaign.ID)
	data, err := json.Marshal(campaign)
	if err != nil {
		return err
	}

	if err := c.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		c.log.Warnf("failed to set campaign cache: %v", err)
		return err
	}

	return nil
}

// DeleteCampaign 删除活动缓存
func (c *CacheService) DeleteCampaign(ctx context.Context, campaignID string) error {
	key := fmt.Sprintf("campaign:%s", campaignID)
	if err := c.rdb.Del(ctx, key).Err(); err != nil {
		c.log.Warnf("failed to delete campaign cache: %v", err)
		return err
	}
	return nil
}

// ========== Reward 缓存 ==========

// GetReward 从缓存获取奖励
func (c *CacheService) GetReward(ctx context.Context, rewardID string) (*biz.Reward, error) {
	key := fmt.Sprintf("reward:%s", rewardID)
	data, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		c.log.Warnf("failed to get reward from cache: %v", err)
		return nil, err
	}

	var reward biz.Reward
	if err := json.Unmarshal([]byte(data), &reward); err != nil {
		c.log.Warnf("failed to unmarshal reward from cache: %v", err)
		return nil, err
	}

	return &reward, nil
}

// SetReward 设置奖励缓存
func (c *CacheService) SetReward(ctx context.Context, reward *biz.Reward, ttl time.Duration) error {
	key := fmt.Sprintf("reward:%s", reward.ID)
	data, err := json.Marshal(reward)
	if err != nil {
		return err
	}

	if err := c.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		c.log.Warnf("failed to set reward cache: %v", err)
		return err
	}

	return nil
}

// DeleteReward 删除奖励缓存
func (c *CacheService) DeleteReward(ctx context.Context, rewardID string) error {
	key := fmt.Sprintf("reward:%s", rewardID)
	if err := c.rdb.Del(ctx, key).Err(); err != nil {
		c.log.Warnf("failed to delete reward cache: %v", err)
		return err
	}
	return nil
}

// ========== Task 缓存 ==========

// GetTask 从缓存获取任务
func (c *CacheService) GetTask(ctx context.Context, taskID string) (*biz.Task, error) {
	key := fmt.Sprintf("task:%s", taskID)
	data, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		c.log.Warnf("failed to get task from cache: %v", err)
		return nil, err
	}

	var task biz.Task
	if err := json.Unmarshal([]byte(data), &task); err != nil {
		c.log.Warnf("failed to unmarshal task from cache: %v", err)
		return nil, err
	}

	return &task, nil
}

// SetTask 设置任务缓存
func (c *CacheService) SetTask(ctx context.Context, task *biz.Task, ttl time.Duration) error {
	key := fmt.Sprintf("task:%s", task.ID)
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}

	if err := c.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		c.log.Warnf("failed to set task cache: %v", err)
		return err
	}

	return nil
}

// DeleteTask 删除任务缓存
func (c *CacheService) DeleteTask(ctx context.Context, taskID string) error {
	key := fmt.Sprintf("task:%s", taskID)
	if err := c.rdb.Del(ctx, key).Err(); err != nil {
		c.log.Warnf("failed to delete task cache: %v", err)
		return err
	}
	return nil
}

// ========== 缓存失效策略 ==========

// InvalidateCampaignTasks 失效活动的任务列表缓存
func (c *CacheService) InvalidateCampaignTasks(ctx context.Context, campaignID string) error {
	pattern := fmt.Sprintf("campaign:%s:tasks:*", campaignID)
	keys, err := c.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		c.log.Warnf("failed to get keys for pattern %s: %v", pattern, err)
		return err
	}

	if len(keys) > 0 {
		if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
			c.log.Warnf("failed to delete keys: %v", err)
			return err
		}
	}

	return nil
}

// InvalidateRewardGrants 失效奖励发放记录缓存
func (c *CacheService) InvalidateRewardGrants(ctx context.Context, rewardID string) error {
	pattern := fmt.Sprintf("reward:%s:grants:*", rewardID)
	keys, err := c.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		c.log.Warnf("failed to get keys for pattern %s: %v", pattern, err)
		return err
	}

	if len(keys) > 0 {
		if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
			c.log.Warnf("failed to delete keys: %v", err)
			return err
		}
	}

	return nil
}

// 默认缓存过期时间
const (
	DefaultCacheTTL = 30 * time.Minute  // 默认30分钟
	CampaignCacheTTL = 1 * time.Hour   // 活动缓存1小时
	RewardCacheTTL = 1 * time.Hour      // 奖励缓存1小时
	TaskCacheTTL = 1 * time.Hour        // 任务缓存1小时
)

