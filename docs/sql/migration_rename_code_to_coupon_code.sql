-- 将优惠券表的 code 字段重命名为 coupon_code
-- 执行时间：2025-12-07
-- 说明：统一字段命名规范，使字段名更具描述性

-- 步骤1: 修改主键约束（需要先删除主键，再添加新的主键）
ALTER TABLE `coupon` DROP PRIMARY KEY;

-- 步骤2: 重命名字段
ALTER TABLE `coupon` 
CHANGE COLUMN `code` `coupon_code` varchar(50) NOT NULL COMMENT '优惠码（唯一标识）';

-- 步骤3: 重新添加主键约束
ALTER TABLE `coupon` ADD PRIMARY KEY (`coupon_code`);

-- 步骤4: 更新索引（如果有基于 code 的索引，需要重新创建）
-- 注意：如果之前有基于 code 的复合索引，需要手动调整

-- 验证：查询修改后的字段信息
-- SHOW COLUMNS FROM `coupon` WHERE Field = 'coupon_code';
-- SHOW INDEX FROM `coupon`;

