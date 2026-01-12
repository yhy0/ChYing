package config

// ScanProfile defines the high-level scanning strategy.
type ScanProfile int

const (
	// Intelligent is the default mode. It balances speed and thoroughness
	// by making smart decisions based on the target's characteristics.
	Intelligent ScanProfile = iota
	// Deep enables all payload generation techniques for the most comprehensive scan.
	// This mode is slow but has the highest chance of finding vulnerabilities.
	Deep
	// Stealthy uses a limited set of common, low-noise payloads to avoid detection.
	// It's faster but less thorough.
	Stealthy
	// Custom allows advanced users to manually tweak all configuration options.
	Custom
)

// PayloadConfig holds the configuration for payload generation.
type PayloadConfig struct {
	EnableURLEncoding           bool `json:"enableUrlEncoding"`
	EnableHTMLEncoding          bool `json:"enableHtmlEncoding"`
	EnableUnicodeEncoding       bool `json:"enableUnicodeEncoding"`
	EnableHexEncoding           bool `json:"enableHexEncoding"`
	EnableMixedEncoding         bool `json:"enableMixedEncoding"`
	EnableWAFBypass             bool `json:"enableWafBypass"`
	EnableEventHandlerInjection bool `json:"enableEventHandlerInjection"`
	MaxPayloadsPerContext       int  `json:"maxPayloadsPerContext"`
}

// VerifierConfig holds the configuration for the verification process.
type VerifierConfig struct {
	UseASTVerification bool `json:"useAstVerification"`
	CheckForAlerts     bool `json:"checkForAlerts"`
}

// EngineConfig holds the general configuration for the scan engine.
type EngineConfig struct {
	MaxConcurrentRequests     int `json:"maxConcurrentRequests"`
	RequestTimeoutSeconds     int `json:"requestTimeoutSeconds"`
	DynamicAnalysisTimeoutSec int `json:"dynamicAnalysisTimeoutSec"`
	MaxRetries                int `json:"maxRetries"`
	RetryDelaySeconds         int `json:"retryDelaySeconds"`
}

// Config is the main configuration structure for the XSS scanner.
type Config struct {
	// Profile is the selected scan profile, which determines the default settings.
	Profile ScanProfile
	// Payload contains detailed settings for payload generation.
	Payload PayloadConfig
	// Engine contains general engine settings
	Engine EngineConfig
}

// NewConfig creates a new configuration based on the selected scan profile.
// This is the primary way to get a configuration object.
func NewConfig(profile ScanProfile) *Config {
	cfg := &Config{
		Profile: profile,
	}

	switch profile {
	case Deep:
		// Enable everything for the deepest scan
		cfg.Payload = PayloadConfig{
			EnableHTMLEncoding:          true,
			EnableURLEncoding:           true,
			EnableUnicodeEncoding:       true,
			EnableHexEncoding:           true,
			EnableMixedEncoding:         true,
			EnableWAFBypass:             true,
			EnableEventHandlerInjection: true,
			MaxPayloadsPerContext:       100, // 深度模式允许更多 payload
		}
		cfg.Engine = EngineConfig{
			MaxConcurrentRequests:     5,
			RequestTimeoutSeconds:     30,
			DynamicAnalysisTimeoutSec: 60, // 深度模式允许更长的动态分析时间
			MaxRetries:                3,
			RetryDelaySeconds:         2,
		}
	case Stealthy:
		// Enable only the most common and least noisy techniques
		cfg.Payload = PayloadConfig{
			EnableHTMLEncoding:          false,
			EnableURLEncoding:           true, // URL encoding is very common
			EnableUnicodeEncoding:       false,
			EnableHexEncoding:           false,
			EnableMixedEncoding:         false,
			EnableWAFBypass:             false,
			EnableEventHandlerInjection: false, // Event handlers can be noisy
			MaxPayloadsPerContext:       20,    // 隐蔽模式限制 payload 数量
		}
		cfg.Engine = EngineConfig{
			MaxConcurrentRequests:     2, // 隐蔽模式限制并发
			RequestTimeoutSeconds:     15,
			DynamicAnalysisTimeoutSec: 20, // 更短的动态分析时间
			MaxRetries:                2,
			RetryDelaySeconds:         3, // 更长的重试延迟以降低噪音
		}
	case Intelligent:
		fallthrough
	default:
		// A balanced default set for the intelligent mode.
		// In the future, this mode could dynamically adjust these settings.
		cfg.Payload = PayloadConfig{
			EnableHTMLEncoding:          true,
			EnableURLEncoding:           true,
			EnableUnicodeEncoding:       true,
			EnableHexEncoding:           false,
			EnableMixedEncoding:         true,
			EnableWAFBypass:             true,
			EnableEventHandlerInjection: true,
			MaxPayloadsPerContext:       50, // 智能模式的平衡设置
		}
		cfg.Engine = EngineConfig{
			MaxConcurrentRequests:     3,
			RequestTimeoutSeconds:     20,
			DynamicAnalysisTimeoutSec: 30, // 智能模式的平衡超时
			MaxRetries:                3,
			RetryDelaySeconds:         1,
		}
	}

	return cfg
}
