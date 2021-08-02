package contract

// Level 日志等级类型
type Level int

//日志等级
//@see https://tools.ietf.org/html/rfc5424
const (
	// LevelEmergency 紧急情况：系统无法使用
	LevelEmergency Level = iota

	// LevelAlert 警报：必须立即采取措施
	LevelAlert

	// LevelCritical 严重：危急情况
	LevelCritical

	// LevelError 错误
	LevelError

	// LevelWarning 警告
	LevelWarning

	// LevelNotice 注意：正常但重要条件
	LevelNotice

	// LevelInfo 信息
	LevelInfo

	// LevelDebug 调试
	LevelDebug
)

//日志等级名称
var levelToName map[Level]string
var nameToLevel map[string]Level

func init() {
	levelToName = map[Level]string{
		LevelEmergency: "emergency",
		LevelAlert:     "alert",
		LevelCritical:  "critical",
		LevelError:     "error",
		LevelWarning:   "warning",
		LevelNotice:    "notice",
		LevelInfo:      "info",
		LevelDebug:     "debug",
	}
	nameToLevel = map[string]Level{
		"emergency": LevelEmergency,
		"alert": LevelAlert,
		"critical": LevelCritical,
		"error": LevelError,
		"warning": LevelWarning,
		"notice": LevelNotice,
		"info":LevelInfo,
		"debug":LevelDebug,
	}
}

func GetLevelByName(name string) (level Level) {
	var ok bool
	if level, ok = nameToLevel[name]; !ok {
		level = LevelDebug
	}
	return
}

func GetNameByLevel(level Level) (name string) {
	var ok bool
	if name, ok = levelToName[level]; !ok {
		name = levelToName[LevelDebug]
	}
	return
}
