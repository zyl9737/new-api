package mediakit

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
)

const (
	ModelEnhanceVideoStandard         = "volc-mediakit-enhance-video-standard"
	ModelEnhanceVideoProfessional     = "volc-mediakit-enhance-video-professional"
	ModelEraseVideoSubtitleStandard   = "volc-mediakit-erase-video-subtitle-standard"
	ModelEraseVideoSubtitlePro        = "volc-mediakit-erase-video-subtitle-pro"
	RouteEnhanceVideo                 = "/api/v1/tools/enhance-video"
	RouteEraseVideoSubtitle           = "/api/v1/tools/erase-video-subtitle"
	RouteEraseVideoSubtitlePro        = "/api/v1/tools/erase-video-subtitle-pro"
	RouteTaskFetchPrefix              = "/api/v1/tasks/"
	UpstreamEnhanceVideoPath          = "/tools/enhance-video"
	UpstreamEraseVideoSubtitlePath    = "/tools/erase-video-subtitle"
	UpstreamEraseVideoSubtitleProPath = "/tools/erase-video-subtitle-pro"
)

type enhanceVideoRequestProbe struct {
	ToolVersion string `json:"tool_version,omitempty"`
}

func NormalizeToolVersion(toolVersion string) string {
	switch strings.ToLower(strings.TrimSpace(toolVersion)) {
	case "professional":
		return "professional"
	default:
		return "standard"
	}
}

func InferModelFromPathAndBody(path string, body []byte) string {
	path = stripQuery(path)
	switch path {
	case RouteEnhanceVideo:
		probe := enhanceVideoRequestProbe{}
		if len(body) > 0 {
			_ = common.Unmarshal(body, &probe)
		}
		if NormalizeToolVersion(probe.ToolVersion) == "professional" {
			return ModelEnhanceVideoProfessional
		}
		return ModelEnhanceVideoStandard
	case RouteEraseVideoSubtitle:
		return ModelEraseVideoSubtitleStandard
	case RouteEraseVideoSubtitlePro:
		return ModelEraseVideoSubtitlePro
	default:
		return ""
	}
}

func IsSubmitPath(path string) bool {
	switch stripQuery(path) {
	case RouteEnhanceVideo, RouteEraseVideoSubtitle, RouteEraseVideoSubtitlePro:
		return true
	default:
		return false
	}
}

func IsFetchPath(path string) bool {
	return strings.HasPrefix(stripQuery(path), RouteTaskFetchPrefix)
}

func UpstreamPathForRequestPath(path string) string {
	switch stripQuery(path) {
	case RouteEnhanceVideo:
		return UpstreamEnhanceVideoPath
	case RouteEraseVideoSubtitle:
		return UpstreamEraseVideoSubtitlePath
	case RouteEraseVideoSubtitlePro:
		return UpstreamEraseVideoSubtitleProPath
	default:
		return ""
	}
}

func BuildBaseURL(baseURL string) string {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = "https://mediakit.cn-beijing.volces.com/api/v1"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if strings.HasSuffix(baseURL, "/api/v1") {
		return baseURL
	}
	return baseURL + "/api/v1"
}

func stripQuery(path string) string {
	if idx := strings.Index(path, "?"); idx >= 0 {
		return path[:idx]
	}
	return path
}
