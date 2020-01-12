package flog

/**
 * 日志等级接口
 * @see https://tools.ietf.org/html/rfc5424
 */
type LoggerInterface interface {
	/**
	 * 紧急情况：系统无法使用
	 */
	Emergency(message string, context ...interface{})
	EmergencyF(format string, v ...interface{})

	/**
	 * 警报：必须立即采取措施
	 */
	Alert(message string, context ...interface{})
	AlertF(format string, v ...interface{})

	/**
	 * 严重：危急情况
	 */
	Critical(message string, context ...interface{})
	CriticalF(format string, v ...interface{})

	/**
	 * 错误
	 */
	Error(message string, context ...interface{})
	ErrorF(format string, v ...interface{})

	/**
	 * 警告
	 */
	Warning(message string, context ...interface{})
	WarningF(format string, v ...interface{})

	/**
	 * 注意：正常但重要条件
	 */
	Notice(message string, context ...interface{})
	NoticeF(format string, v ...interface{})

	/**
	 * 信息
	 */
	Info(message string, context ...interface{})
	InfoF(format string, v ...interface{})

	/**
	 * 调试
	 */
	Debug(message string, context ...interface{})
	DebugF(format string, v ...interface{})
}
