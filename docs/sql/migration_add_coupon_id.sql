-- 为优惠券表添加 coupon_id 自增主键，并修改 coupon_code 为普通字段
-- 执行时间：2025-12-07
-- 说明：使用自增 ID 作为主键，通过唯一索引 (coupon_code, deleted_at) 解决软删除后重复创建的问题

-- 步骤1: 添加 coupon_id 字段（自增主键）
ALTER TABLE `coupon`
ADD COLUMN `coupon_id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '优惠券ID（自增主键）' FIRST;

-- 步骤2: 删除原有的主键约束
ALTER TABLE `coupon` DROP PRIMARY KEY;

-- 步骤3: 设置 coupon_id 为主键
ALTER TABLE `coupon` ADD PRIMARY KEY (`coupon_id`);

-- 步骤4: 删除原有的 coupon_code 索引（如果存在）
-- 注意：如果之前有基于 coupon_code 的索引，需要先删除
-- ALTER TABLE `coupon` DROP INDEX `idx_coupon_code`; -- 如果存在

-- 步骤5: 添加唯一索引 (coupon_code, deleted_at)
-- 这个唯一索引确保：同一个 coupon_code 只能有一个 deleted_at IS NULL 的记录
-- 可以有多个 deleted_at IS NOT NULL 的记录（已删除的历史记录）
ALTER TABLE `coupon`
ADD UNIQUE KEY `uk_coupon_code_deleted_at` (`coupon_code`, `deleted_at`);

-- 验证：查询修改后的表结构
-- SHOW CREATE TABLE `coupon`;
-- SHOW INDEX FROM `coupon`;

