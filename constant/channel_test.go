package constant

import (
	"testing"
)

func TestChannelTypeVolcAdapterRegistration(t *testing.T) {
	// Verify the constant value is 58 (next available after ChannelTypeCodex=57).
	if ChannelTypeVolcAdapter != 58 {
		t.Errorf("expected ChannelTypeVolcAdapter=58, got %d", ChannelTypeVolcAdapter)
	}
	if ChannelTypeVolcMediaKit != 59 {
		t.Errorf("expected ChannelTypeVolcMediaKit=59, got %d", ChannelTypeVolcMediaKit)
	}

	// ChannelTypeDummy is explicitly assigned to equal ChannelTypeVolcMediaKit
	// (the highest defined channel type). When adding a new channel type, append
	// it after the last existing one with an explicit numeric value, then update
	// ChannelTypeDummy to equal the new highest. The explicit assignment avoids
	// confusion with iota's implicit-repeat semantics.
	if ChannelTypeDummy != ChannelTypeVolcMediaKit {
		t.Errorf("expected ChannelTypeDummy == ChannelTypeVolcMediaKit (%d), got %d",
			ChannelTypeVolcMediaKit, ChannelTypeDummy)
	}
}

func TestChannelTypeVolcAdapterDisplayName(t *testing.T) {
	name := GetChannelTypeName(ChannelTypeVolcAdapter)
	if name != "VolcAdapter" {
		t.Errorf("expected display name %q, got %q", "VolcAdapter", name)
	}
}

func TestChannelTypeVolcAdapterBaseURL(t *testing.T) {
	const wantURL = "https://ark.cn-beijing.volces.com"
	if ChannelTypeVolcAdapter >= len(ChannelBaseURLs) {
		t.Fatalf("ChannelBaseURLs too short: len=%d, ChannelTypeVolcAdapter=%d", len(ChannelBaseURLs), ChannelTypeVolcAdapter)
	}
	got := ChannelBaseURLs[ChannelTypeVolcAdapter]
	if got != wantURL {
		t.Errorf("expected base URL %q, got %q", wantURL, got)
	}
}

func TestChannelTypeVolcMediaKitBaseURL(t *testing.T) {
	const wantURL = "https://mediakit.cn-beijing.volces.com"
	if ChannelTypeVolcMediaKit >= len(ChannelBaseURLs) {
		t.Fatalf("ChannelBaseURLs too short: len=%d, ChannelTypeVolcMediaKit=%d", len(ChannelBaseURLs), ChannelTypeVolcMediaKit)
	}
	got := ChannelBaseURLs[ChannelTypeVolcMediaKit]
	if got != wantURL {
		t.Errorf("expected base URL %q, got %q", wantURL, got)
	}
}

// TestChannelBaseURLsLength verifies the ChannelBaseURLs slice covers all channel
// types including ChannelTypeVolcMediaKit, so no index-out-of-bounds can occur.
func TestChannelBaseURLsLength(t *testing.T) {
	// ChannelBaseURLs must have at least ChannelTypeVolcMediaKit+1 entries (indices 0..ChannelTypeVolcMediaKit).
	if len(ChannelBaseURLs) < ChannelTypeVolcMediaKit+1 {
		t.Errorf("ChannelBaseURLs has %d entries but needs at least %d to cover ChannelTypeVolcMediaKit=%d",
			len(ChannelBaseURLs), ChannelTypeVolcMediaKit+1, ChannelTypeVolcMediaKit)
	}
}

func TestChannelTypeVolcAdapterInNames(t *testing.T) {
	if _, ok := ChannelTypeNames[ChannelTypeVolcAdapter]; !ok {
		t.Errorf("ChannelTypeVolcAdapter (%d) not found in ChannelTypeNames map", ChannelTypeVolcAdapter)
	}
	if _, ok := ChannelTypeNames[ChannelTypeVolcMediaKit]; !ok {
		t.Errorf("ChannelTypeVolcMediaKit (%d) not found in ChannelTypeNames map", ChannelTypeVolcMediaKit)
	}
}
