package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MarketingMetrics 营销服务指标
type MarketingMetrics struct {
	// 活动相关指标
	CampaignCreatedTotal    prometheus.Counter
	CampaignActiveTotal     prometheus.Gauge
	CampaignCompletedTotal  prometheus.Counter

	// 任务相关指标
	TaskCreatedTotal        prometheus.Counter
	TaskCompletedTotal      prometheus.Counter
	TaskTriggeredTotal      prometheus.Counter

	// 奖励相关指标
	RewardCreatedTotal      prometheus.Counter
	RewardGrantedTotal      prometheus.Counter
	RewardGrantedByType     *prometheus.CounterVec

	// 兑换码相关指标
	RedeemCodeGeneratedTotal prometheus.Counter
	RedeemCodeRedeemedTotal   prometheus.Counter
	RedeemCodeExpiredTotal    prometheus.Counter

	// 库存相关指标
	InventoryReservedTotal   prometheus.Counter
	InventoryConfirmedTotal  prometheus.Counter
	InventoryCancelledTotal  prometheus.Counter

	// 业务操作耗时
	TaskTriggerDuration     *prometheus.HistogramVec
	RewardGenerationDuration *prometheus.HistogramVec
	RewardDistributionDuration *prometheus.HistogramVec
}

// NewMarketingMetrics 创建营销服务指标
func NewMarketingMetrics() *MarketingMetrics {
	return &MarketingMetrics{
		CampaignCreatedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "marketing_campaign_created_total",
			Help: "Total number of campaigns created",
		}),
		CampaignActiveTotal: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "marketing_campaign_active_total",
			Help: "Total number of active campaigns",
		}),
		CampaignCompletedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "marketing_campaign_completed_total",
			Help: "Total number of completed campaigns",
		}),
		TaskCreatedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "marketing_task_created_total",
			Help: "Total number of tasks created",
		}),
		TaskCompletedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "marketing_task_completed_total",
			Help: "Total number of tasks completed",
		}),
		TaskTriggeredTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "marketing_task_triggered_total",
			Help: "Total number of tasks triggered",
		}),
		RewardCreatedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "marketing_reward_created_total",
			Help: "Total number of rewards created",
		}),
		RewardGrantedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "marketing_reward_granted_total",
			Help: "Total number of rewards granted",
		}),
		RewardGrantedByType: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "marketing_reward_granted_by_type_total",
			Help: "Total number of rewards granted by type",
		}, []string{"reward_type"}),
		RedeemCodeGeneratedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "marketing_redeem_code_generated_total",
			Help: "Total number of redeem codes generated",
		}),
		RedeemCodeRedeemedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "marketing_redeem_code_redeemed_total",
			Help: "Total number of redeem codes redeemed",
		}),
		RedeemCodeExpiredTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "marketing_redeem_code_expired_total",
			Help: "Total number of redeem codes expired",
		}),
		InventoryReservedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "marketing_inventory_reserved_total",
			Help: "Total number of inventory reservations",
		}),
		InventoryConfirmedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "marketing_inventory_confirmed_total",
			Help: "Total number of inventory confirmations",
		}),
		InventoryCancelledTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "marketing_inventory_cancelled_total",
			Help: "Total number of inventory cancellations",
		}),
		TaskTriggerDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "marketing_task_trigger_duration_seconds",
			Help:    "Duration of task trigger operations",
			Buckets: prometheus.DefBuckets,
		}, []string{"task_type", "status"}),
		RewardGenerationDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "marketing_reward_generation_duration_seconds",
			Help:    "Duration of reward generation operations",
			Buckets: prometheus.DefBuckets,
		}, []string{"reward_type"}),
		RewardDistributionDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "marketing_reward_distribution_duration_seconds",
			Help:    "Duration of reward distribution operations",
			Buckets: prometheus.DefBuckets,
		}, []string{"distribution_type", "status"}),
	}
}

// 全局指标实例
var defaultMetrics *MarketingMetrics

// InitMetrics 初始化全局指标
func InitMetrics() {
	defaultMetrics = NewMarketingMetrics()
}

// GetMetrics 获取全局指标实例
func GetMetrics() *MarketingMetrics {
	if defaultMetrics == nil {
		InitMetrics()
	}
	return defaultMetrics
}

