package mediakit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay/channel"
	channelmediakit "github.com/QuantumNous/new-api/relay/channel/mediakit"
	"github.com/QuantumNous/new-api/relay/channel/task/taskcommon"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

const (
	actionEnhanceVideo        = "enhance-video"
	actionEraseSubtitle       = "erase-video-subtitle"
	actionEraseSubtitlePro    = "erase-video-subtitle-pro"
	defaultEstimatedMinutes   = 1.0
	defaultEstimatedFPS       = 30.0
	defaultEstimatedShortSide = 1080
)

type TaskAdaptor struct {
	taskcommon.BaseBilling
	baseURL string
	apiKey  string
}

type mediaKitSubmitRequest struct {
	VideoURL        string  `json:"video_url"`
	ToolVersion     string  `json:"tool_version,omitempty"`
	Resolution      string  `json:"resolution,omitempty"`
	ResolutionLimit int     `json:"resolution_limit,omitempty"`
	FPS             float64 `json:"fps,omitempty"`
	Mode            string  `json:"mode,omitempty"`
	QueueID         string  `json:"queue_id,omitempty"`
	ClientToken     string  `json:"client_token,omitempty"`
	CallbackArgs    string  `json:"callback_args,omitempty"`
	EraseLocations  []any   `json:"erase_ratio_location,omitempty"`
}

type mediaKitError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Type    string `json:"type,omitempty"`
	Param   string `json:"param,omitempty"`
}

type mediaKitSubmitResponse struct {
	Success   bool           `json:"success"`
	TaskID    string         `json:"task_id,omitempty"`
	RequestID string         `json:"request_id,omitempty"`
	Error     *mediaKitError `json:"error,omitempty"`
	Message   string         `json:"message,omitempty"`
}

type mediaKitTaskOutput struct {
	VideoURL    string  `json:"video_url,omitempty"`
	Duration    float64 `json:"duration,omitempty"`
	Resolution  string  `json:"resolution,omitempty"`
	FPS         float64 `json:"fps,omitempty"`
	ToolVersion string  `json:"tool_version,omitempty"`
}

type mediaKitTaskResponse struct {
	Success    *bool                 `json:"success,omitempty"`
	TaskID     string                `json:"task_id,omitempty"`
	TaskType   string                `json:"task_type,omitempty"`
	Status     string                `json:"status,omitempty"`
	RequestID  string                `json:"request_id,omitempty"`
	Result     *mediaKitTaskOutput   `json:"result,omitempty"`
	Output     *mediaKitTaskOutput   `json:"output,omitempty"`
	Error      *mediaKitError        `json:"error,omitempty"`
	Data       *mediaKitTaskResponse `json:"data,omitempty"`
	CreatedAt  int64                 `json:"created_at,omitempty"`
	FinishedAt int64                 `json:"finished_at,omitempty"`
	ExpiresAt  int64                 `json:"expires_at,omitempty"`
}

func (a *TaskAdaptor) Init(info *relaycommon.RelayInfo) {
	a.baseURL = info.ChannelBaseUrl
	a.apiKey = info.ApiKey
}

func (a *TaskAdaptor) ValidateRequestAndSetAction(c *gin.Context, info *relaycommon.RelayInfo) *dto.TaskError {
	var req mediaKitSubmitRequest
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		return service.TaskErrorWrapperLocal(fmt.Errorf("invalid request body: %w", err), "invalid_request", http.StatusBadRequest)
	}
	if strings.TrimSpace(req.VideoURL) == "" {
		return service.TaskErrorWrapperLocal(fmt.Errorf("video_url is required"), "invalid_request", http.StatusBadRequest)
	}

	modelName := channelmediakit.InferModelFromPathAndBody(info.RequestURLPath, mustBodyBytes(c))
	if modelName == "" {
		return service.TaskErrorWrapperLocal(fmt.Errorf("unsupported MediaKit path: %s", info.RequestURLPath), "invalid_request", http.StatusBadRequest)
	}
	info.OriginModelName = modelName
	info.UpstreamModelName = modelName
	info.Action = actionForModel(modelName)
	return nil
}

func (a *TaskAdaptor) EstimateBilling(c *gin.Context, info *relaycommon.RelayInfo) map[string]float64 {
	params := ExtractBillingParams(info.RequestURLPath, info.OriginModelName, mustBodyBytes(c))
	if params == nil {
		return nil
	}
	ratios := map[string]float64{
		"duration_minutes": defaultEstimatedMinutes,
	}
	switch params.Tool {
	case actionEnhanceVideo:
		coefficient, _ := enhanceCoefficient(params.ToolVersion, params.Resolution, params.ResolutionLimit, params.FPS)
		ratios["billing_coefficient"] = coefficient
	case actionEraseSubtitle, actionEraseSubtitlePro:
		ratios["billing_coefficient"] = subtitleCoefficient(params.Tool)
	}
	return ratios
}

func (a *TaskAdaptor) BuildRequestURL(info *relaycommon.RelayInfo) (string, error) {
	upstreamPath := channelmediakit.UpstreamPathForRequestPath(info.RequestURLPath)
	if upstreamPath == "" {
		return "", fmt.Errorf("unsupported MediaKit path: %s", info.RequestURLPath)
	}
	return channelmediakit.BuildBaseURL(a.baseURL) + upstreamPath, nil
}

func (a *TaskAdaptor) BuildRequestHeader(_ *gin.Context, req *http.Request, _ *relaycommon.RelayInfo) error {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	return nil
}

func (a *TaskAdaptor) BuildRequestBody(c *gin.Context, info *relaycommon.RelayInfo) (io.Reader, error) {
	storage, err := common.GetBodyStorage(c)
	if err != nil {
		return nil, err
	}
	if _, err = storage.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	rawBody, err := storage.Bytes()
	if err != nil {
		return nil, err
	}
	if len(info.ParamOverride) > 0 {
		rawBody, err = relaycommon.ApplyParamOverrideWithRelayInfo(rawBody, info)
		if err != nil {
			return nil, err
		}
	}
	return bytes.NewReader(rawBody), nil
}

func (a *TaskAdaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (*http.Response, error) {
	return channel.DoTaskApiRequest(a, c, info, requestBody)
}

func (a *TaskAdaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (taskID string, taskData []byte, taskErr *dto.TaskError) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, service.TaskErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
	}
	_ = resp.Body.Close()

	var submitResp mediaKitSubmitResponse
	if err := common.Unmarshal(responseBody, &submitResp); err != nil {
		return "", nil, service.TaskErrorWrapper(fmt.Errorf("unmarshal MediaKit submit response failed: %w", err), "invalid_response", http.StatusInternalServerError)
	}
	if !submitResp.Success {
		return "", nil, service.TaskErrorWrapperLocal(fmt.Errorf("%s", extractMediaKitErrorMessage(submitResp.Error, submitResp.Message)), "mediakit_submit_failed", http.StatusBadRequest)
	}
	if strings.TrimSpace(submitResp.TaskID) == "" {
		return "", nil, service.TaskErrorWrapperLocal(fmt.Errorf("task_id is empty"), "invalid_response", http.StatusInternalServerError)
	}

	patched := patchMediaKitTaskID(responseBody, info.PublicTaskID)
	c.Data(http.StatusOK, "application/json", patched)
	return submitResp.TaskID, responseBody, nil
}

func (a *TaskAdaptor) FetchTask(baseURL, key string, body map[string]any, proxy string) (*http.Response, error) {
	taskID, ok := body["task_id"].(string)
	if !ok || strings.TrimSpace(taskID) == "" {
		return nil, fmt.Errorf("invalid task_id")
	}
	url := channelmediakit.BuildBaseURL(baseURL) + "/tasks/" + taskID
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

	client, err := service.GetHttpClientWithProxy(proxy)
	if err != nil {
		return nil, fmt.Errorf("new proxy http client failed: %w", err)
	}
	return client.Do(req)
}

func (a *TaskAdaptor) ParseTaskResult(respBody []byte) (*relaycommon.TaskInfo, error) {
	taskResp, effective, err := parseTaskResponse(respBody)
	if err != nil {
		return nil, err
	}

	ti := &relaycommon.TaskInfo{}
	status := normalizeMediaKitStatus(effective.Status)
	switch status {
	case "queued":
		ti.Status = model.TaskStatusQueued
		ti.Progress = "20%"
	case "running":
		ti.Status = model.TaskStatusInProgress
		ti.Progress = "50%"
	case "success":
		ti.Status = model.TaskStatusSuccess
		ti.Progress = "100%"
		if out := effectiveOutput(effective); out != nil {
			ti.Url = strings.TrimSpace(out.VideoURL)
		}
	case "failure":
		ti.Status = model.TaskStatusFailure
		ti.Progress = "100%"
		ti.Reason = extractMediaKitErrorMessage(taskResp.Error, "")
		if ti.Reason == "" {
			ti.Reason = extractMediaKitErrorMessage(effective.Error, "")
		}
		if ti.Reason == "" && strings.TrimSpace(effective.Status) != "" {
			ti.Reason = strings.TrimSpace(effective.Status)
		}
	default:
		ti.Status = model.TaskStatusInProgress
		ti.Progress = "30%"
	}
	return ti, nil
}

func (a *TaskAdaptor) AdjustBillingOnComplete(task *model.Task, _ *relaycommon.TaskInfo) int {
	bc := task.PrivateData.BillingContext
	if bc == nil || bc.ModelPrice <= 0 || bc.GroupRatio == 0 {
		return 0
	}

	_, effective, err := parseTaskResponse(task.Data)
	if err != nil {
		return 0
	}
	out := effectiveOutput(effective)
	if out == nil || out.Duration <= 0 {
		return 0
	}

	params := bc.MediaKit
	if params == nil {
		params = &model.TaskMediaKitParams{}
	}
	if params.Tool == "" {
		params.Tool = actionForModel(resolveTaskModelName(task))
	}

	switch params.Tool {
	case actionEnhanceVideo:
		toolVersion := channelmediakit.NormalizeToolVersion(out.ToolVersion)
		if toolVersion == "standard" && params.ToolVersion != "" {
			toolVersion = channelmediakit.NormalizeToolVersion(params.ToolVersion)
		}
		resolution := normalizeResolution(out.Resolution)
		if resolution == "" {
			resolution = normalizeResolution(params.Resolution)
		}
		fps := out.FPS
		if fps <= 0 {
			fps = params.FPS
		}
		if fps <= 0 {
			fps = defaultEstimatedFPS
		}
		coefficient, bucket := enhanceCoefficient(toolVersion, resolution, params.ResolutionLimit, fps)
		actualQuota := quotaFromBilling(bc.ModelPrice, bc.GroupRatio, out.Duration, coefficient)
		if actualQuota <= 0 {
			return 0
		}
		bc.OtherRatios = map[string]float64{
			"duration_seconds":    out.Duration,
			"duration_minutes":    out.Duration / 60,
			"billing_coefficient": coefficient,
			"resolution_bucket":   float64(bucket),
			"output_fps":          fps,
		}
		bc.SettlementReason = fmt.Sprintf(
			"MediaKit结算：tool=enhance-video, toolVersion=%s, duration=%.2fs, resolution=%s, fps=%.2f, coefficient=%.2f",
			toolVersion, out.Duration, resolutionLabel(bucket, resolution), fps, coefficient,
		)
		return actualQuota
	case actionEraseSubtitle, actionEraseSubtitlePro:
		coefficient := subtitleCoefficient(params.Tool)
		actualQuota := quotaFromBilling(bc.ModelPrice, bc.GroupRatio, out.Duration, coefficient)
		if actualQuota <= 0 {
			return 0
		}
		bc.OtherRatios = map[string]float64{
			"duration_seconds":    out.Duration,
			"duration_minutes":    out.Duration / 60,
			"billing_coefficient": coefficient,
		}
		bc.SettlementReason = fmt.Sprintf(
			"MediaKit结算：tool=%s, duration=%.2fs, coefficient=%.2f",
			params.Tool, out.Duration, coefficient,
		)
		return actualQuota
	default:
		return 0
	}
}

func (a *TaskAdaptor) GetModelList() []string {
	return channelmediakit.ModelList
}

func (a *TaskAdaptor) GetChannelName() string {
	return channelmediakit.ChannelName
}

func ExtractBillingParams(path, modelName string, body []byte) *model.TaskMediaKitParams {
	tool := actionForModel(modelName)
	if tool == "" {
		modelName = channelmediakit.InferModelFromPathAndBody(path, body)
		tool = actionForModel(modelName)
	}
	if tool == "" {
		return nil
	}

	params := &model.TaskMediaKitParams{
		Tool: tool,
	}
	if tool != actionEnhanceVideo {
		return params
	}

	var req mediaKitSubmitRequest
	if len(body) > 0 {
		_ = common.Unmarshal(body, &req)
	}
	params.ToolVersion = channelmediakit.NormalizeToolVersion(req.ToolVersion)
	params.Resolution = normalizeResolution(req.Resolution)
	params.ResolutionLimit = req.ResolutionLimit
	if req.FPS > 0 {
		params.FPS = req.FPS
	}
	return params
}

func mustBodyBytes(c *gin.Context) []byte {
	storage, err := common.GetBodyStorage(c)
	if err != nil {
		return nil
	}
	raw, err := storage.Bytes()
	if err != nil {
		return nil
	}
	_, _ = storage.Seek(0, io.SeekStart)
	return raw
}

func parseTaskResponse(body []byte) (*mediaKitTaskResponse, *mediaKitTaskResponse, error) {
	var taskResp mediaKitTaskResponse
	if err := common.Unmarshal(body, &taskResp); err != nil {
		return nil, nil, fmt.Errorf("unmarshal MediaKit task response failed: %w", err)
	}
	return &taskResp, effectiveTaskBody(&taskResp), nil
}

func effectiveTaskBody(resp *mediaKitTaskResponse) *mediaKitTaskResponse {
	if resp == nil {
		return nil
	}
	if resp.Data != nil {
		if resp.Data.Status != "" || resp.Data.TaskID != "" || resp.Data.Result != nil || resp.Data.Output != nil {
			return resp.Data
		}
	}
	return resp
}

func effectiveOutput(resp *mediaKitTaskResponse) *mediaKitTaskOutput {
	if resp == nil {
		return nil
	}
	if resp.Output != nil {
		return resp.Output
	}
	return resp.Result
}

func patchMediaKitTaskID(body []byte, taskID string) []byte {
	if strings.TrimSpace(taskID) == "" {
		return body
	}
	var payload map[string]json.RawMessage
	if err := common.Unmarshal(body, &payload); err != nil {
		return body
	}
	taskIDJSON, err := common.Marshal(taskID)
	if err != nil {
		return body
	}
	payload["task_id"] = json.RawMessage(taskIDJSON)
	patched, err := common.Marshal(payload)
	if err != nil {
		return body
	}
	return patched
}

func extractMediaKitErrorMessage(apiErr *mediaKitError, fallback string) string {
	if apiErr != nil {
		if msg := strings.TrimSpace(apiErr.Message); msg != "" {
			return msg
		}
		if code := strings.TrimSpace(apiErr.Code); code != "" {
			return code
		}
	}
	return strings.TrimSpace(fallback)
}

func actionForModel(modelName string) string {
	switch modelName {
	case channelmediakit.ModelEnhanceVideoStandard, channelmediakit.ModelEnhanceVideoProfessional:
		return actionEnhanceVideo
	case channelmediakit.ModelEraseVideoSubtitleStandard:
		return actionEraseSubtitle
	case channelmediakit.ModelEraseVideoSubtitlePro:
		return actionEraseSubtitlePro
	default:
		return ""
	}
}

func resolveTaskModelName(task *model.Task) string {
	if task == nil {
		return ""
	}
	if bc := task.PrivateData.BillingContext; bc != nil && strings.TrimSpace(bc.OriginModelName) != "" {
		return bc.OriginModelName
	}
	if strings.TrimSpace(task.Properties.OriginModelName) != "" {
		return task.Properties.OriginModelName
	}
	return strings.TrimSpace(task.Properties.UpstreamModelName)
}

func normalizeMediaKitStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "queued", "pending", "waiting", "created", "submitted":
		return "queued"
	case "running", "processing", "in_progress":
		return "running"
	case "completed", "succeeded", "success", "done", "finished", "finish":
		return "success"
	case "failed", "error", "fail", "cancelled", "canceled":
		return "failure"
	default:
		return ""
	}
}

func subtitleCoefficient(tool string) float64 {
	if tool == actionEraseSubtitlePro {
		return 1
	}
	return 0.4
}

func enhanceCoefficient(toolVersion, resolution string, resolutionLimit int, fps float64) (float64, int) {
	toolVersion = channelmediakit.NormalizeToolVersion(toolVersion)
	bucket := resolutionBucket(resolution, resolutionLimit)
	highFPS := fps > 30

	switch toolVersion {
	case "professional":
		switch bucket {
		case 720:
			if highFPS {
				return 20, 720
			}
			return 10, 720
		case 1080:
			if highFPS {
				return 40, 1080
			}
			return 20, 1080
		case 1440:
			if highFPS {
				return 80, 1440
			}
			return 40, 1440
		default:
			if highFPS {
				return 160, 2160
			}
			return 80, 2160
		}
	default:
		switch bucket {
		case 720:
			if highFPS {
				return 2, 720
			}
			return 1, 720
		case 1080:
			if highFPS {
				return 4, 1080
			}
			return 2, 1080
		case 1440:
			if highFPS {
				return 8, 1440
			}
			return 4, 1440
		default:
			if highFPS {
				return 16, 2160
			}
			return 8, 2160
		}
	}
}

func resolutionBucket(resolution string, resolutionLimit int) int {
	shortSide := resolutionShortSide(resolution)
	if shortSide <= 0 {
		shortSide = resolutionLimit
	}
	if shortSide <= 0 {
		shortSide = defaultEstimatedShortSide
	}
	switch {
	case shortSide <= 720:
		return 720
	case shortSide <= 1080:
		return 1080
	case shortSide <= 1440:
		return 1440
	default:
		return 2160
	}
}

func resolutionShortSide(resolution string) int {
	switch normalizeResolution(resolution) {
	case "240p":
		return 240
	case "360p":
		return 360
	case "480p":
		return 480
	case "540p":
		return 540
	case "720p":
		return 720
	case "1080p":
		return 1080
	case "2k":
		return 1440
	case "4k":
		return 2160
	default:
		return 0
	}
}

func normalizeResolution(resolution string) string {
	switch strings.ToLower(strings.TrimSpace(resolution)) {
	case "240p", "360p", "480p", "540p", "720p", "1080p", "2k", "4k":
		return strings.ToLower(strings.TrimSpace(resolution))
	default:
		return strings.ToLower(strings.TrimSpace(resolution))
	}
}

func resolutionLabel(bucket int, resolution string) string {
	if normalized := normalizeResolution(resolution); normalized != "" {
		return normalized
	}
	switch bucket {
	case 720:
		return "720p-or-below"
	case 1080:
		return "1080p"
	case 1440:
		return "2k"
	default:
		return "4k"
	}
}

func quotaFromBilling(modelPrice, groupRatio, durationSeconds, coefficient float64) int {
	if modelPrice <= 0 || groupRatio < 0 || durationSeconds <= 0 || coefficient <= 0 {
		return 0
	}
	durationMinutes := durationSeconds / 60
	if durationMinutes <= 0 {
		return 0
	}
	return int(math.Round(modelPrice * coefficient * durationMinutes * groupRatio * common.QuotaPerUnit))
}
