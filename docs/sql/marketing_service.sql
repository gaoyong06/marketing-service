/*
 Navicat Premium Dump SQL

 Source Server         : localhost
 Source Server Type    : MySQL
 Source Server Version : 90500 (9.5.0)
 Source Host           : localhost:3306
 Source Schema         : marketing_service

 Target Server Type    : MySQL
 Target Server Version : 90500 (9.5.0)
 File Encoding         : 65001

 Date: 22/01/2026 11:35:12
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for coupon
-- ----------------------------
DROP TABLE IF EXISTS `coupon`;
CREATE TABLE `coupon` (
  `coupon_id` bigint NOT NULL AUTO_INCREMENT COMMENT '优惠券ID（自增主键）',
  `coupon_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '优惠码（业务唯一标识）',
  `app_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '应用ID',
  `discount_type` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '折扣类型: percent(百分比)/fixed(固定金额)',
  `discount_value` bigint NOT NULL COMMENT '折扣值(百分比或分)',
  `currency` enum('CNY','USD','EUR') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'CNY' COMMENT '货币单位: CNY(人民币)/USD(美元)/EUR(欧元)，仅固定金额类型需要',
  `valid_from` datetime(3) NOT NULL COMMENT '生效时间(UTC时间)',
  `valid_until` datetime(3) NOT NULL COMMENT '过期时间(UTC时间)',
  `max_uses` int NOT NULL DEFAULT '1' COMMENT '最大使用次数',
  `used_count` int NOT NULL DEFAULT '0' COMMENT '已使用次数',
  `min_amount` bigint NOT NULL DEFAULT '0' COMMENT '最低消费金额(分)',
  `status` enum('active','inactive','expired') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active' COMMENT '优惠券状态: active(激活-可使用)/inactive(停用-不可使用)/expired(已过期-系统自动标记)',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间(UTC时间)',
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间(UTC时间)',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT '删除时间(UTC时间)',
  PRIMARY KEY (`coupon_id`),
  UNIQUE KEY `uk_coupon_code` (`coupon_code`),
  KEY `idx_app_id` (`app_id`),
  KEY `idx_status` (`status`),
  KEY `idx_valid_time` (`valid_from`,`valid_until`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='优惠券表';

-- ----------------------------
-- Table structure for coupon_usage
-- ----------------------------
DROP TABLE IF EXISTS `coupon_usage`;
CREATE TABLE `coupon_usage` (
  `coupon_usage_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '使用记录ID（唯一标识）',
  `coupon_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '优惠券码',
  `app_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '应用ID',
  `user_id` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '用户ID',
  `payment_order_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '支付订单ID（payment-service的业务订单号orderId）',
  `payment_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '支付流水号（payment-service返回的payment_id）',
  `original_amount` bigint NOT NULL COMMENT '原价(分)',
  `discount_amount` bigint NOT NULL COMMENT '折扣金额(分)',
  `final_amount` bigint NOT NULL COMMENT '实付金额(分)',
  `used_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '使用时间(UTC时间)',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间(UTC时间)',
  PRIMARY KEY (`coupon_usage_id`),
  KEY `idx_coupon_code` (`coupon_code`),
  KEY `idx_app_id` (`app_id`),
  KEY `idx_app_id_used_at` (`app_id`,`used_at`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_payment_order_id` (`payment_order_id`),
  KEY `idx_payment_id` (`payment_id`),
  KEY `idx_used_at` (`used_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='优惠券使用记录表';

SET FOREIGN_KEY_CHECKS = 1;
