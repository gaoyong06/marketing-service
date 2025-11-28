CREATE DATABASE `marketing_service` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 该库包含：
-- campaigns (活动表)
-- redeem_codes (兑换码表)
-- code_batches (批次表)
-- redemption_records (核销记录表)
-- campaign_rules (活动规则表)


-- 活动表（campaigns）

CREATE TABLE `campaigns` (
  `campaign_id` varchar(32) NOT NULL COMMENT '活动ID',
  `campaign_name` varchar(64) NOT NULL COMMENT '活动名称',
  `tenant_id` varchar(24) NOT NULL COMMENT '所属租户',
  `product_code` varchar(16) NOT NULL COMMENT '适用产品线',
  `campaign_type` varchar(32) NOT NULL COMMENT '活动类型',
  `start_time` datetime NOT NULL COMMENT '开始时间',
  `end_time` datetime NOT NULL COMMENT '结束时间',
  `total_budget` decimal(12,2) DEFAULT NULL COMMENT '总预算',
  `rule_config` json NOT NULL COMMENT '规则配置',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态：0-未开始 1-进行中 2-已结束 3-手动终止',
  `created_by` varchar(64) DEFAULT NULL COMMENT '创建人',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`campaign_id`),
  KEY `idx_tenant_product` (`tenant_id`,`product_code`),
  KEY `idx_time_range` (`start_time`,`end_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='营销活动表';


-- 兑换码表（redeem_codes）
CREATE TABLE `redeem_codes` (
  `redeem_code_id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `code` varchar(32) NOT NULL COMMENT '兑换码',
  `campaign_id` varchar(32) NOT NULL COMMENT '关联活动ID',
  `tenant_id` varchar(24) NOT NULL COMMENT '所属租户',
  `product_code` varchar(16) NOT NULL COMMENT '适用产品线',
  `code_type` varchar(16) NOT NULL COMMENT '码类型：DISCOUNT-折扣码 COUPON-优惠码 GIFT-礼品码',
  `user_id` bigint(20) DEFAULT NULL COMMENT '分配用户ID',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态：0-未分配 1-已分配 2-已核销 3-已失效',
  `value_amount` decimal(10,2) DEFAULT NULL COMMENT '面额（金额类）',
  `value_percent` decimal(5,2) DEFAULT NULL COMMENT '折扣比例（百分比类）',
  `reward_items` json DEFAULT NULL COMMENT '奖励物品（实物类）',
  `expire_time` datetime DEFAULT NULL COMMENT '过期时间',
  `redeem_time` datetime DEFAULT NULL COMMENT '核销时间',
  `redeem_channel` varchar(16) DEFAULT NULL COMMENT '核销渠道',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`redeem_code_id`),
  UNIQUE KEY `uk_code_tenant` (`code`,`tenant_id`),
  KEY `idx_campaign` (`campaign_id`),
  KEY `idx_tenant_product` (`tenant_id`,`product_code`),
  KEY `idx_user_status` (`user_id`,`status`),
  KEY `idx_expire_time` (`expire_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='兑换码表';


-- 兑换码批次表（code_batches）
CREATE TABLE `code_batches` (
  `batch_id` varchar(32) NOT NULL COMMENT '批次ID',
  `campaign_id` varchar(32) NOT NULL COMMENT '关联活动ID',
  `tenant_id` varchar(24) NOT NULL COMMENT '所属租户',
  `generate_type` varchar(16) NOT NULL COMMENT '生成方式：MANUAL-手动 AUTO-自动 IMPORT-导入',
  `total_count` int(11) NOT NULL COMMENT '总数量',
  `used_count` int(11) NOT NULL DEFAULT '0' COMMENT '已使用数量',
  `operator` varchar(64) DEFAULT NULL COMMENT '操作人',
  `generate_rule` json DEFAULT NULL COMMENT '生成规则',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`batch_id`),
  KEY `idx_campaign` (`campaign_id`),
  KEY `idx_tenant` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='兑换码批次表';


-- 兑换记录表（redemption_records）
CREATE TABLE `redemption_records` (
  `record_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `redeem_code_id` bigint(20) NOT NULL COMMENT '兑换码ID',
  `code` varchar(32) NOT NULL COMMENT '兑换码值',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `tenant_id` varchar(24) NOT NULL COMMENT '所属租户',
  `campaign_id` varchar(32) NOT NULL COMMENT '关联活动ID',
  `product_code` varchar(16) NOT NULL COMMENT '产品线',
  `redeem_channel` varchar(16) NOT NULL COMMENT '核销渠道',
  `redeem_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '核销时间',
  `device_info` varchar(128) DEFAULT NULL COMMENT '设备信息',
  `ip_address` varchar(64) DEFAULT NULL COMMENT 'IP地址',
  `location` varchar(64) DEFAULT NULL COMMENT '地理位置',
  `order_id` varchar(32) DEFAULT NULL COMMENT '关联订单ID',
  `reward_detail` json DEFAULT NULL COMMENT '奖励详情',
  PRIMARY KEY (`record_id`),
  KEY `idx_code` (`code`),
  KEY `idx_user` (`user_id`),
  KEY `idx_tenant_campaign` (`tenant_id`,`campaign_id`),
  KEY `idx_redeem_time` (`redeem_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='兑换记录表';


-- 索引优化
-- 兑换码表
ALTER TABLE `redeem_codes` ADD INDEX `idx_tenant_status` (`tenant_id`, `status`);
ALTER TABLE `redeem_codes` ADD INDEX `idx_campaign_status` (`campaign_id`, `status`);

-- 活动表
ALTER TABLE `campaigns` ADD INDEX `idx_status_time` (`status`, `start_time`, `end_time`);