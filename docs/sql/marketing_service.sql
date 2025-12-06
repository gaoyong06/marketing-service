-- ============================================
-- 营销服务数据库设计
-- 核心实体：活动(Campaign)、受众(Audience)、任务(Task)、奖励(Reward)
-- 业务数据：奖励发放(RewardGrant)、兑换码(RedeemCode)、任务完成日志(TaskCompletionLog)、库存预占(InventoryReservation)
-- ============================================

CREATE DATABASE IF NOT EXISTS `marketing_service` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `marketing_service`;

-- ============================================
-- 核心实体表
-- ============================================

-- 1. 活动表（Campaign）
CREATE TABLE `campaign` (
  `campaign_id` varchar(32) NOT NULL COMMENT '活动ID（唯一标识）',
  `tenant_id` varchar(32) NOT NULL COMMENT '租户ID',
  `app_id` varchar(32) NOT NULL COMMENT '应用ID',
  `campaign_name` varchar(128) NOT NULL COMMENT '活动名称',
  `campaign_type` varchar(32) NOT NULL COMMENT '活动类型：REDEEM_CODE/TASK_REWARD/DIRECT_SEND',
  `start_time` datetime NOT NULL COMMENT '开始时间',
  `end_time` datetime NOT NULL COMMENT '结束时间',
  `audience_config` json DEFAULT NULL COMMENT '受众配置（JSON格式，支持多受众组合）',
  `validator_config` json DEFAULT NULL COMMENT '校验规则配置（1:N关系，轻量级组合直接存JSON）',
  `status` varchar(16) NOT NULL DEFAULT 'ACTIVE' COMMENT '状态：ACTIVE/PAUSED/ENDED',
  `description` varchar(512) DEFAULT NULL COMMENT '活动描述',
  `created_by` varchar(64) DEFAULT NULL COMMENT '创建人',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间（软删除）',
  PRIMARY KEY (`campaign_id`),
  KEY `idx_tenant_app` (`tenant_id`, `app_id`),
  KEY `idx_campaign_type` (`campaign_type`),
  KEY `idx_time_range` (`start_time`, `end_time`),
  KEY `idx_status` (`status`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='活动表';

-- audience_config
-- {
--   "logic": "OR", // 逻辑关系：AND / OR
--   "items": [
--     { "type": "AUDIENCE", "id": "aud_vip_users" }, // 引用已定义的受众
--     { "type": "AUDIENCE", "id": "aud_new_users" }
--   ],
--   "exclude": [
--     { "type": "AUDIENCE", "id": "aud_black_list" } // 排除黑名单
--   ]
-- }


-- 2. 受众表（Audience）
CREATE TABLE `audience` (
  `audience_id` varchar(32) NOT NULL COMMENT '受众ID（唯一标识）',
  `tenant_id` varchar(32) NOT NULL COMMENT '租户ID',
  `app_id` varchar(32) NOT NULL COMMENT '应用ID',
  `name` varchar(128) NOT NULL COMMENT '受众名称',
  `audience_type` varchar(32) NOT NULL COMMENT '受众类型：TAG/SEGMENT/LIST/ALL',
  `rule_config` json NOT NULL COMMENT '圈选规则配置（JSON格式）',
  `status` varchar(16) NOT NULL DEFAULT 'ACTIVE' COMMENT '状态：ACTIVE/PAUSED/ENDED',
  `description` varchar(512) DEFAULT NULL COMMENT '描述',
  `created_by` varchar(64) DEFAULT NULL COMMENT '创建人',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间（软删除）',
  PRIMARY KEY (`audience_id`),
  KEY `idx_tenant_app` (`tenant_id`, `app_id`),
  KEY `idx_audience_type` (`audience_type`),
  KEY `idx_status` (`status`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='受众表（定义活动的参与人群）';



-- 3. 任务表（Task）
CREATE TABLE `task` (
  `task_id` varchar(32) NOT NULL COMMENT '任务ID（唯一标识）',
  `tenant_id` varchar(32) NOT NULL COMMENT '租户ID',
  `app_id` varchar(32) NOT NULL COMMENT '应用ID',
  `name` varchar(128) NOT NULL COMMENT '任务名称',
  `task_type` varchar(32) NOT NULL COMMENT '任务类型：INVITE/PURCHASE/SHARE/SIGN_IN',
  `trigger_config` json DEFAULT NULL COMMENT '触发配置（JSON格式：Event, Condition）',
  `condition_config` json NOT NULL COMMENT '完成条件配置（JSON格式）',
  `reward_id` varchar(32) DEFAULT NULL COMMENT '关联奖励ID（可选）',
  `status` varchar(16) NOT NULL DEFAULT 'ACTIVE' COMMENT '状态：ACTIVE/PAUSED/ENDED',
  `start_time` datetime NOT NULL COMMENT '开始时间',
  `end_time` datetime NOT NULL COMMENT '结束时间',
  `max_count` int(11) NOT NULL DEFAULT 0 COMMENT '最大完成次数（0表示无限制）',
  `description` varchar(512) DEFAULT NULL COMMENT '描述',
  `created_by` varchar(64) DEFAULT NULL COMMENT '创建人',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间（软删除）',
  PRIMARY KEY (`task_id`),
  KEY `idx_tenant_app` (`tenant_id`, `app_id`),
  KEY `idx_task_type` (`task_type`),
  KEY `idx_reward_id` (`reward_id`),
  KEY `idx_status` (`status`),
  KEY `idx_time_range` (`start_time`, `end_time`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务表（独立，包含触发器和条件）';

-- 4. 奖励表（Reward）
CREATE TABLE `reward` (
  `reward_id` varchar(32) NOT NULL COMMENT '奖励ID（唯一标识）',
  `tenant_id` varchar(32) NOT NULL COMMENT '租户ID',
  `app_id` varchar(32) NOT NULL COMMENT '应用ID',
  `reward_type` varchar(32) NOT NULL COMMENT '奖励类型：COUPON/POINTS/REDEEM_CODE/SUBSCRIPTION',
  `name` varchar(128) DEFAULT NULL COMMENT '奖励名称',
  `content_config` json NOT NULL COMMENT '奖励内容配置（JSON格式）',
  `generator_config` json DEFAULT NULL COMMENT '生成配置（JSON格式，替代Generator表）',
  `distributor_config` json DEFAULT NULL COMMENT '发放配置（JSON格式，替代Distributor表）',
  `validator_config` json DEFAULT NULL COMMENT '校验规则配置（1:N关系，轻量级组合直接存JSON）',
  `version` int(11) NOT NULL DEFAULT 1 COMMENT '版本号（每次修改时递增）',
  `valid_days` int(11) NOT NULL DEFAULT 0 COMMENT '有效期（天数）',
  `extra_config` json DEFAULT NULL COMMENT '额外配置（JSON格式）',
  `status` varchar(16) NOT NULL DEFAULT 'ACTIVE' COMMENT '状态：ACTIVE/PAUSED/ENDED',
  `description` varchar(512) DEFAULT NULL COMMENT '描述',
  `created_by` varchar(64) DEFAULT NULL COMMENT '创建人',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间（软删除）',
  PRIMARY KEY (`reward_id`),
  KEY `idx_tenant_app` (`tenant_id`, `app_id`),
  KEY `idx_reward_type` (`reward_type`),
  KEY `idx_version` (`reward_id`, `version`),
  KEY `idx_status` (`status`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='奖励表（奖励模板）';

-- ============================================
-- 组合关系表：记录实体之间的组合关系
-- 优化原则：
-- 1. 只有真正的 M:N (多对多) 且需要独立管理生命周期的关系才建立关联表
-- 2. 1:1 或 1:N (强从属) 的关系直接在主表中增加字段或 JSON 配置
-- ============================================

-- 活动组合任务关系表 (Campaign : Task = 1 : N)
-- 虽然逻辑上是 1:N，但为了 Task 的复用性（一个 Task 模板可能被多个 Campaign 引用），保留关联表
CREATE TABLE `campaign_task` (
  `campaign_task_id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `campaign_id` varchar(32) NOT NULL COMMENT '活动ID',
  `task_id` varchar(32) NOT NULL COMMENT '任务ID',
  `config` json DEFAULT NULL COMMENT '组合配置（如：覆盖任务的默认阈值）',
  `sort_order` int(11) NOT NULL DEFAULT 0 COMMENT '排序顺序',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`campaign_task_id`),
  UNIQUE KEY `uk_campaign_task` (`campaign_id`, `task_id`),
  KEY `idx_task_id` (`task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='活动组合任务关系表';

-- ============================================
-- 业务数据表：存储实际生成的业务数据
-- ============================================

-- 奖励发放表（核心表：存储每个发放的奖励发放）
CREATE TABLE `reward_grant` (
  `grant_id` varchar(32) NOT NULL COMMENT '授予ID（唯一标识）',
  `reward_id` varchar(32) NOT NULL COMMENT '奖励模板ID',
  `reward_name` varchar(128) DEFAULT NULL COMMENT '奖励名称（冗余字段，用于列表展示）',
  `reward_type` varchar(32) NOT NULL COMMENT '奖励类型（冗余字段，用于筛选）',
  `reward_version` int(11) NOT NULL COMMENT '奖励版本号',
  `content_snapshot` json NOT NULL COMMENT '奖励内容快照',
  `generator_config` json DEFAULT NULL COMMENT '生成配置快照（JSON格式，替代generator_id）',
  `campaign_id` varchar(32) DEFAULT NULL COMMENT '活动ID',
  `campaign_name` varchar(128) DEFAULT NULL COMMENT '活动名称（冗余字段，避免JOIN）',
  `task_id` varchar(32) DEFAULT NULL COMMENT '任务ID',
  `task_name` varchar(128) DEFAULT NULL COMMENT '任务名称（冗余字段，避免JOIN）',
  `tenant_id` varchar(32) NOT NULL COMMENT '租户ID',
  `app_id` varchar(32) NOT NULL COMMENT '应用ID',
  `user_id` bigint(20) DEFAULT NULL COMMENT '用户ID',
  `status` varchar(16) NOT NULL DEFAULT 'PENDING' COMMENT '状态：PENDING/GENERATED/RESERVED/DISTRIBUTED/USED/EXPIRED',
  `reserved_at` datetime DEFAULT NULL COMMENT '预占时间',
  `distributed_at` datetime DEFAULT NULL COMMENT '发放时间',
  `used_at` datetime DEFAULT NULL COMMENT '使用时间',
  `expire_time` datetime DEFAULT NULL COMMENT '过期时间',
  `error_message` varchar(512) DEFAULT NULL COMMENT '错误信息（发放失败时记录）',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`grant_id`),
  KEY `idx_tenant_app` (`tenant_id`, `app_id`),
  KEY `idx_reward_id` (`reward_id`),
  KEY `idx_reward_type` (`reward_type`),
  KEY `idx_reward_version` (`reward_id`, `reward_version`),
  KEY `idx_campaign_id` (`campaign_id`),
  KEY `idx_user_status` (`user_id`, `status`),
  KEY `idx_status` (`status`),
  KEY `idx_expire_time` (`expire_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='奖励发放表';

-- 库存预占表（防止超发，支持高并发）
CREATE TABLE `inventory_reservation` (
  `reservation_id` varchar(32) NOT NULL COMMENT '预占ID（唯一标识）',
  `resource_id` varchar(64) NOT NULL COMMENT '资源ID（如：优惠券模板ID、商品SKU、总库存Key）',
  `campaign_id` varchar(32) DEFAULT NULL COMMENT '活动ID（可选，用于活动维度的库存控制）',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `quantity` int(11) NOT NULL DEFAULT 1 COMMENT '预占数量',
  `status` varchar(16) NOT NULL DEFAULT 'PENDING' COMMENT '状态：PENDING(预占中)/CONFIRMED(已核销)/CANCELLED(已回滚)/EXPIRED(已过期)',
  `expire_at` datetime NOT NULL COMMENT '预占过期时间（超时未核销自动释放）',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`reservation_id`),
  KEY `idx_resource_status` (`resource_id`, `status`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_expire_at` (`expire_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='库存预占表';

-- 兑换码表（简化设计：只负责码的生命周期管理）
CREATE TABLE `redeem_code` (
  `code` varchar(64) NOT NULL COMMENT '兑换码（唯一标识）',
  `tenant_id` varchar(32) NOT NULL COMMENT '租户ID',
  `app_id` varchar(32) NOT NULL COMMENT '应用ID',
  `grant_id` varchar(32) DEFAULT NULL COMMENT '关联的奖励授予ID（兑换后关联）',
  `campaign_id` varchar(32) DEFAULT NULL COMMENT '所属活动ID',
  `campaign_name` varchar(128) DEFAULT NULL COMMENT '活动名称（冗余字段，用于展示）',
  `reward_id` varchar(32) NOT NULL COMMENT '关联奖励模板ID',
  `reward_name` varchar(128) DEFAULT NULL COMMENT '奖励名称（冗余字段，用于展示）',
  `batch_id` varchar(32) DEFAULT NULL COMMENT '批次ID（批量生成时使用）',
  `status` varchar(16) NOT NULL DEFAULT 'ACTIVE' COMMENT '状态：ACTIVE(可用)/REDEEMED(已兑换)/EXPIRED(已过期)/REVOKED(已作废)',
  `owner_user_id` bigint(20) DEFAULT NULL COMMENT '拥有者用户ID（预分配场景）',
  `redeemed_by` bigint(20) DEFAULT NULL COMMENT '兑换者用户ID（可能与owner不同）',
  `redeemed_at` datetime DEFAULT NULL COMMENT '兑换时间',
  `expire_at` datetime DEFAULT NULL COMMENT '过期时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`code`, `tenant_id`),
  KEY `idx_tenant_app` (`tenant_id`, `app_id`),
  KEY `idx_grant_id` (`grant_id`),
  KEY `idx_campaign_reward` (`campaign_id`, `reward_id`),
  KEY `idx_batch_id` (`batch_id`),
  KEY `idx_status` (`status`),
  KEY `idx_owner_user` (`owner_user_id`),
  KEY `idx_expire_at` (`expire_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='兑换码表';

-- 任务完成记录表（事件日志表：记录每次任务完成事件）
CREATE TABLE `task_completion_log` (
  `completion_id` varchar(32) NOT NULL COMMENT '完成记录ID（唯一标识）',
  `task_id` varchar(32) NOT NULL COMMENT '任务ID',
  `task_name` varchar(128) DEFAULT NULL COMMENT '任务名称（冗余字段，用于报表）',
  `campaign_id` varchar(32) DEFAULT NULL COMMENT '活动ID',
  `campaign_name` varchar(128) DEFAULT NULL COMMENT '活动名称（冗余字段，用于报表）',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `tenant_id` varchar(32) NOT NULL COMMENT '租户ID',
  `app_id` varchar(32) NOT NULL COMMENT '应用ID',
  `grant_id` varchar(32) DEFAULT NULL COMMENT '关联的奖励授予ID（如果触发了奖励发放）',
  `progress_data` json DEFAULT NULL COMMENT '任务进度数据（JSON格式，如：{"invited_count": 3, "target": 5}）',
  `trigger_event` varchar(64) DEFAULT NULL COMMENT '触发事件（如：USER_REGISTER, ORDER_PAID）',
  `completed_at` datetime NOT NULL COMMENT '完成时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`completion_id`),
  KEY `idx_tenant_app` (`tenant_id`, `app_id`),
  KEY `idx_task_user` (`task_id`, `user_id`),
  KEY `idx_campaign_id` (`campaign_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_grant_id` (`grant_id`),
  KEY `idx_completed_at` (`completed_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务完成记录表（事件日志）';

-- 核心实体层（4张）
-- ├── campaign          活动表
-- ├── audience          受众表
-- ├── task              任务表
-- └── reward            奖励表

-- 关系层（1张）
-- └── campaign_task     活动-任务关联表

-- 业务数据层（4张）
-- ├── reward_grant             奖励发放表
-- ├── redeem_code              兑换码表
-- ├── task_completion_log      任务完成日志表
-- └── inventory_reservation    库存预占表


-- 注意：原 reward_distribution 表已删除
-- 理由：
-- 1. 与 reward_grant 表职责重叠（都在记录奖励发放）
-- 2. reward_grant 表已经包含了发放状态、发放时间等字段
-- 3. 如需追踪发放过程，应通过 reward_grant 的状态变更日志实现
-- 4. 如需记录发放失败原因，可在 reward_grant 表增加 error_message 字段

-- ============================================
-- 版本信息
-- ============================================
-- 版本: v1.0.0
-- 创建日期: 2025-11-30
-- 最后更新: 2025-11-30
-- 
-- 设计原则:
-- 1. 极简主义: 只保留核心业务实体，避免过度设计
-- 2. 业务导向: 使用业务术语命名(grant/log)，而非技术术语(instance/record)
-- 3. 配置化: 轻量级逻辑通过JSON配置实现，重量级实体才建表
-- 4. 多租户: 所有表都支持tenant_id + app_id的资源隔离
-- 
-- 核心表结构:
-- - 4张核心实体表: campaign, audience, task, reward
-- - 1张关系表: campaign_task
-- - 4张业务数据表: reward_grant, redeem_code, task_completion_log, inventory_reservation

-- ============================================
-- 数据库查询优化索引
-- 用于提升查询性能
-- ============================================

-- ========== Campaign 表索引优化 ==========
-- 复合索引：租户+应用+状态（用于列表查询）
ALTER TABLE `campaign` ADD INDEX `idx_tenant_app_status` (`tenant_id`, `app_id`, `status`);

-- ========== Reward 表索引优化 ==========
-- 复合索引：租户+应用+状态（用于列表查询）
ALTER TABLE `reward` ADD INDEX `idx_tenant_app_status` (`tenant_id`, `app_id`, `status`);

-- ========== Task 表索引优化 ==========
-- 复合索引：租户+应用+状态（用于列表查询）
ALTER TABLE `task` ADD INDEX `idx_tenant_app_status` (`tenant_id`, `app_id`, `status`);

-- 任务类型+状态索引（用于查询活跃任务）
ALTER TABLE `task` ADD INDEX `idx_task_type_status` (`task_type`, `status`);

-- 时间范围+状态索引（用于查询活跃任务）
ALTER TABLE `task` ADD INDEX `idx_time_status` (`start_time`, `end_time`, `status`);

-- ========== RewardGrant 表索引优化 ==========
-- 复合索引：租户+应用+用户+状态（用于用户奖励查询）
ALTER TABLE `reward_grant` ADD INDEX `idx_tenant_app_user_status` (`tenant_id`, `app_id`, `user_id`, `status`);

-- 奖励ID+状态索引（用于统计）
ALTER TABLE `reward_grant` ADD INDEX `idx_reward_status` (`reward_id`, `status`);

-- 创建时间索引（用于排序）
ALTER TABLE `reward_grant` ADD INDEX `idx_created_at` (`created_at`);

-- ========== RedeemCode 表索引优化 ==========
-- 注意：idx_batch_id、idx_owner_user、idx_expire_at 已在 CREATE TABLE 中定义，无需重复添加

-- 复合索引：租户+应用+状态（用于列表查询）
ALTER TABLE `redeem_code` ADD INDEX `idx_tenant_app_status` (`tenant_id`, `app_id`, `status`);

-- ========== TaskCompletionLog 表索引优化 ==========
-- 注意：idx_task_user、idx_completed_at 已在 CREATE TABLE 中定义，无需重复添加

-- 复合索引：租户+应用+任务（用于列表查询）
ALTER TABLE `task_completion_log` ADD INDEX `idx_tenant_app_task` (`tenant_id`, `app_id`, `task_id`);

-- ========== InventoryReservation 表索引优化 ==========
-- 注意：idx_resource_status、idx_expire_at、idx_user_id 已在 CREATE TABLE 中定义，无需重复添加

-- ========== Coupon 优惠券表 ==========
-- 优惠券表（供开发者控制台使用，用于支付场景的优惠券管理）
CREATE TABLE `coupon` (
  `code` varchar(50) NOT NULL COMMENT '优惠码（唯一标识）',
  `app_id` varchar(32) NOT NULL COMMENT '应用ID',
  `discount_type` varchar(16) NOT NULL COMMENT '折扣类型: percent(百分比)/fixed(固定金额)',
  `discount_value` bigint(20) NOT NULL COMMENT '折扣值(百分比或分)',
  `valid_from` bigint(20) NOT NULL COMMENT '生效时间(timestamp)',
  `valid_until` bigint(20) NOT NULL COMMENT '过期时间(timestamp)',
  `max_uses` int(11) NOT NULL DEFAULT 1 COMMENT '最大使用次数',
  `used_count` int(11) NOT NULL DEFAULT 0 COMMENT '已使用次数',
  `min_amount` bigint(20) NOT NULL DEFAULT 0 COMMENT '最低消费金额(分)',
  `status` varchar(16) NOT NULL DEFAULT 'active' COMMENT '状态: active/inactive/expired',
  `created_at` bigint(20) NOT NULL COMMENT '创建时间(timestamp)',
  `updated_at` bigint(20) NOT NULL COMMENT '更新时间(timestamp)',
  PRIMARY KEY (`code`),
  KEY `idx_app_id` (`app_id`),
  KEY `idx_status` (`status`),
  KEY `idx_valid_time` (`valid_from`, `valid_until`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='优惠券表';

-- ========== CouponUsage 优惠券使用记录表 ==========
-- 优惠券使用记录表（记录每次优惠券的使用情况）
CREATE TABLE `coupon_usage` (
  `id` varchar(32) NOT NULL COMMENT '使用记录ID（唯一标识）',
  `coupon_code` varchar(50) NOT NULL COMMENT '优惠券码',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `order_id` varchar(64) NOT NULL COMMENT '订单ID',
  `payment_id` varchar(64) NOT NULL COMMENT '支付ID',
  `original_amount` bigint(20) NOT NULL COMMENT '原价(分)',
  `discount_amount` bigint(20) NOT NULL COMMENT '折扣金额(分)',
  `final_amount` bigint(20) NOT NULL COMMENT '实付金额(分)',
  `used_at` bigint(20) NOT NULL COMMENT '使用时间(timestamp)',
  `created_at` bigint(20) NOT NULL COMMENT '创建时间(timestamp)',
  PRIMARY KEY (`id`),
  KEY `idx_coupon_code` (`coupon_code`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_payment_id` (`payment_id`),
  KEY `idx_used_at` (`used_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='优惠券使用记录表';

-- ========== 分页查询优化建议 ==========
-- 对于大数据量的分页查询，建议使用游标分页（cursor-based pagination）
-- 而不是 OFFSET/LIMIT，可以避免深度分页的性能问题
-- 
-- 示例：
-- SELECT * FROM campaign 
-- WHERE tenant_id = ? AND app_id = ? AND campaign_id > ? 
-- ORDER BY campaign_id ASC 
-- LIMIT 20
