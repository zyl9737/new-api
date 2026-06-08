package mediakit

import "github.com/QuantumNous/new-api/relay/channel/openai"

// ChannelName is the display identifier used in model marketplace listings.
var ChannelName = "volc-mediakit"

// ModelList contains synthetic task models used for channel selection, billing,
// and logging for Volc MediaKit async video tools.
var ModelList = []string{
	ModelEnhanceVideoStandard,
	ModelEnhanceVideoProfessional,
	ModelEraseVideoSubtitleStandard,
	ModelEraseVideoSubtitlePro,
}

// Adaptor embeds the OpenAI adaptor only to satisfy generic channel metadata
// flows (model marketplace / owner lookup). Actual request handling for this
// channel type is task-only and is implemented in relay/channel/task/mediakit.
type Adaptor struct {
	openai.Adaptor
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}
