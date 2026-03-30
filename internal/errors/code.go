package errors

import (
	pkgErrors "github.com/gaoyong06/go-pkg/errors"
	i18nPkg "github.com/gaoyong06/go-pkg/middleware/i18n"
)

func init() {
	// 初始化全局错误管理器（使用项目特定的配置）
	pkgErrors.InitGlobalErrorManager("i18n", i18nPkg.Language)
}
