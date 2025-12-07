-- 为优惠券表添加 deleted_at 字段，实现软删除功能
-- 执行时间：2025-12-07
-- 说明：使用软删除可以保留历史数据，便于数据恢复和审计

-- 步骤1: 添加 deleted_at 字段
ALTER TABLE `coupon`
ADD COLUMN `deleted_at` datetime DEFAULT NULL COMMENT '删除时间（软删除）' AFTER `updated_at`;

-- 步骤2: 添加索引以优化查询性能
ALTER TABLE `coupon`
ADD KEY `idx_deleted_at` (`deleted_at`);

-- 验证：查询修改后的字段信息
-- SHOW COLUMNS FROM `coupon` WHERE Field = 'deleted_at';
-- SHOW INDEX FROM `coupon` WHERE Key_name = 'idx_deleted_at';

-- 注意：
-- 1. GORM 会自动在查询时过滤 deleted_at IS NULL 的记录
-- 2. 使用 Delete 方法时会自动设置 deleted_at 为当前时间
-- 3. 如需查询已删除的记录，可以使用 Unscoped() 方法
-- 4. 如需永久删除，可以使用 Unscoped().Delete() 方法

