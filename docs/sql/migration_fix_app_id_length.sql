-- 修复 coupon 表中 app_id 字段长度不足的问题
-- 问题：app_id 字段定义为 varchar(32)，但实际 UUID 格式的 app_id 长度为 36 个字符
-- 解决：将 app_id 字段长度从 varchar(32) 改为 varchar(64)

ALTER TABLE `coupon` 
MODIFY COLUMN `app_id` varchar(64) NOT NULL COMMENT '应用ID';

