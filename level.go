package flog

import "strings"

//日志等级类型
type Level int

/**
 * 日志等级
 * @see https://tools.ietf.org/html/rfc5424
 */
const (
	/**
	 * 紧急情况：系统无法使用
	 */
	LEVEL_EMERGENCY Level = iota

	/**
	 * 警报：必须立即采取措施
	 */
	LEVEL_ALERT

	/**
	 * 严重：危急情况
	 */
	LEVEL_CRITICAL

	/**
	 * 错误
	 */
	LEVEL_ERROR

	/**
	 * 警告
	 */
	LEVEL_WARNING

	/**
	 * 注意：正常但重要条件
	 */
	LEVEL_NOTICE

	/**
	 * 信息
	 */
	LEVEL_INFO

	/**
	 * 调试
	 */
	LEVEL_DEBUG
)

//日志等级名称
var levelMap map[Level]string

func init() {
	levelMap = map[Level]string{
		LEVEL_EMERGENCY: "EMERGENCY",
		LEVEL_ALERT:     "ALERT",
		LEVEL_CRITICAL:  "CRITICAL",
		LEVEL_ERROR:     "ERROR",
		LEVEL_WARNING:   "WARNING",
		LEVEL_NOTICE:    "NOTICE",
		LEVEL_INFO:      "INFO",
		LEVEL_DEBUG:     "DEBUG",
	}
}

func GetLevelByName(level string) Level {
	for k, v := range levelMap {
		if strings.EqualFold(level, v) {
			return k
		}
	}
	return LEVEL_DEBUG
}

func GetNameByLevel(level Level) (name string) {
	var ok bool
	if name, ok = levelMap[level]; !ok {
		name = levelMap[LEVEL_DEBUG]
	}
	return
}
