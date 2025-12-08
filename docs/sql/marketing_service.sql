-- ============================================
-- 优惠券服务数据库设计
-- 核心功能：优惠券管理、使用记录、统计
-- 设计理念：极简、刚需、克制
-- ============================================

CREATE DATABASE IF NOT EXISTS `marketing_service` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `marketing_service`;

-- ============================================
-- 优惠券表
-- ============================================

-- 优惠券表（供开发者控制台使用，用于支付场景的优惠券管理）
CREATE TABLE `coupon` (
  `coupon_id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '优惠券ID（自增主键）',
  `coupon_code` varchar(50) NOT NULL COMMENT '优惠码（业务唯一标识）',
  `app_id` varchar(64) NOT NULL COMMENT '应用ID',
  `discount_type` varchar(16) NOT NULL COMMENT '折扣类型: percent(百分比)/fixed(固定金额)',
  `discount_value` bigint(20) NOT NULL COMMENT '折扣值(百分比或分)',
  `currency` enum('CNY','USD','EUR') NOT NULL DEFAULT 'CNY' COMMENT '货币单位: CNY(人民币)/USD(美元)/EUR(欧元)，仅固定金额类型需要',
  `valid_from` datetime NOT NULL COMMENT '生效时间',
  `valid_until` datetime NOT NULL COMMENT '过期时间',
  `max_uses` int(11) NOT NULL DEFAULT 1 COMMENT '最大使用次数',
  `used_count` int(11) NOT NULL DEFAULT 0 COMMENT '已使用次数',
  `min_amount` bigint(20) NOT NULL DEFAULT 0 COMMENT '最低消费金额(分)',
  `status` enum('active','inactive','expired') NOT NULL DEFAULT 'active' COMMENT '优惠券状态: active(激活-可使用)/inactive(停用-不可使用)/expired(已过期-系统自动标记)',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间（软删除）',
  PRIMARY KEY (`coupon_id`),
  UNIQUE KEY `uk_coupon_code_deleted_at` (`coupon_code`, `deleted_at`),
  KEY `idx_app_id` (`app_id`),
  KEY `idx_status` (`status`),
  KEY `idx_valid_time` (`valid_from`, `valid_until`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='优惠券表';

-- ============================================
-- 优惠券使用记录表
-- ============================================

-- 优惠券使用记录表（记录每次优惠券的使用情况）
CREATE TABLE `coupon_usage` (
  `coupon_usage_id` varchar(32) NOT NULL COMMENT '使用记录ID（唯一标识）',
  `coupon_code` varchar(50) NOT NULL COMMENT '优惠券码',
  `uid` bigint(20) NOT NULL COMMENT '用户ID',
  `payment_order_id` varchar(64) NOT NULL COMMENT '支付订单ID（payment-service的业务订单号orderId）',
  `payment_id` varchar(64) NOT NULL COMMENT '支付流水号（payment-service返回的payment_id）',
  `original_amount` bigint(20) NOT NULL COMMENT '原价(分)',
  `discount_amount` bigint(20) NOT NULL COMMENT '折扣金额(分)',
  `final_amount` bigint(20) NOT NULL COMMENT '实付金额(分)',
  `used_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '使用时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`coupon_usage_id`),
  KEY `idx_coupon_code` (`coupon_code`),
  KEY `idx_uid` (`uid`),
  KEY `idx_payment_order_id` (`payment_order_id`),
  KEY `idx_payment_id` (`payment_id`),
  KEY `idx_used_at` (`used_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='优惠券使用记录表';

-- ============================================
-- 版本信息
-- ============================================
-- 版本: v2.0.0 (极简重构版)
-- 创建日期: 2025-11-30
-- 最后更新: 2025-12-07
-- 
-- 设计原则:
-- 1. 极简主义: 只保留核心优惠券功能，移除复杂营销活动系统
-- 2. 业务导向: 专注于支付场景的优惠券管理和使用
-- 3. 克制设计: 避免过度设计，保持简单易用
-- 
-- 核心表结构:
-- - 1张优惠券表: coupon
-- - 1张使用记录表: coupon_usage
--
-- 移除的表（v1.0 版本）:
-- - campaign (活动表)
-- - audience (受众表)
-- - task (任务表)
-- - reward (奖励表)
-- - campaign_task (活动-任务关联表)
-- - reward_grant (奖励发放表)
-- - redeem_code (兑换码表)
-- - task_completion_log (任务完成日志表)
-- - inventory_reservation (库存预占表)
