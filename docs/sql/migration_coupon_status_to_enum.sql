-- 将优惠券表的 status 字段从 varchar 改为 enum 类型
-- 执行时间：2025-12-07
-- 说明：使用 enum 类型在数据库层面约束状态值，提高数据完整性和查询性能
-- 状态值：active(激活-可使用)/inactive(禁用-不可使用)/expired(已过期-系统自动标记)

-- 步骤1: 检查并清理无效的状态值（如果有）
-- 注意：如果数据库中存在其他状态值，需要先处理这些数据
-- UPDATE `coupon` SET `status` = 'inactive' WHERE `status` NOT IN ('active', 'inactive', 'expired');

-- 步骤2: 修改字段类型为 enum
ALTER TABLE `coupon`
MODIFY COLUMN `status` enum('active','inactive','expired') NOT NULL DEFAULT 'active' COMMENT '优惠券状态: active(激活-可使用)/inactive(禁用-不可使用)/expired(已过期-系统自动标记)';

-- 验证：查询修改后的字段信息
-- SHOW COLUMNS FROM `coupon` WHERE Field = 'status';

