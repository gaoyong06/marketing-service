-- 为优惠券表添加货币单位字段
-- 执行时间：2025-12-07
-- 说明：为固定金额类型的优惠券添加货币单位配置，支持多币种场景
-- 使用 enum 类型，在数据库层面约束货币代码，避免无效值

ALTER TABLE `coupon`
ADD COLUMN `currency` enum('CNY','USD','EUR') NOT NULL DEFAULT 'CNY' COMMENT '货币单位: CNY(人民币)/USD(美元)/EUR(欧元)，仅固定金额类型需要' AFTER `discount_value`;

-- 为现有数据设置默认货币单位（如果 discount_type 为 fixed，默认使用 CNY）
-- 注意：如果系统之前只支持 CNY，可以跳过此步骤
-- UPDATE `coupon` SET `currency` = 'CNY' WHERE `discount_type` = 'fixed' AND `currency` IS NULL;

