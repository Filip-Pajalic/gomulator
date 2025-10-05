package common

type Config struct {
	EnableBootAnimation bool
	BootAnimationSpeed  float64 // Speed multiplier for boot animation

	EnableDebugWindow bool
	LogLevel          string
}

var GlobalConfig = Config{
	EnableBootAnimation: true,
	BootAnimationSpeed:  1.0,
	EnableDebugWindow:   true,
	LogLevel:            "INFO",
}

func SetBootAnimationEnabled(enabled bool) {
	GlobalConfig.EnableBootAnimation = enabled
}

func IsBootAnimationEnabled() bool {
	return GlobalConfig.EnableBootAnimation
}
