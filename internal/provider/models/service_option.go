package models

import (
	"context"
	"fmt"

	api "github.com/cachefly/cachefly-go-sdk/pkg/cachefly/api/v2_5"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Terraform model for service options
type ServiceOptionsModel struct {
	ServiceID              types.String             `tfsdk:"service_id"`
	FTP                    types.Bool               `tfsdk:"ftp"`
	CORS                   types.Bool               `tfsdk:"cors"`
	AutoRedirect           types.Bool               `tfsdk:"auto_redirect"`
	BrotliCompression      types.Bool               `tfsdk:"brotli_compression"`
	BrotliSupport          types.Bool               `tfsdk:"brotli_support"`
	Livestreaming          types.Bool               `tfsdk:"livestreaming"`
	NoCache                types.Bool               `tfsdk:"nocache"`
	CacheByGeoCountry      types.Bool               `tfsdk:"cache_by_geo_country"`
	CacheByRegion          types.Bool               `tfsdk:"cache_by_region"`
	CacheByReferer         types.Bool               `tfsdk:"cache_by_referer"`
	NormalizeQueryString   types.Bool               `tfsdk:"normalize_query_string"`
	AllowRetry             types.Bool               `tfsdk:"allow_retry"`
	LinkPreheat            types.Bool               `tfsdk:"link_preheat"`
	EdgeToOrigin           types.Bool               `tfsdk:"edge_to_origin"`
	FollowRedirect         types.Bool               `tfsdk:"follow_redirect"`
	PurgeNoQuery           types.Bool               `tfsdk:"purge_no_query"`
	ForceOrigQString       types.Bool               `tfsdk:"force_orig_qstring"`
	ServeStale             types.Bool               `tfsdk:"serve_stale"`
	CachePostRequests      types.Bool               `tfsdk:"cache_post_requests"`
	SendXFF                types.Bool               `tfsdk:"send_xff"`
	UseCFDooTEncoding      types.Bool               `tfsdk:"use_cf_doot_encoding"`
	SkipURLEncoding        types.Bool               `tfsdk:"skip_url_encoding"`
	Trace                  types.Bool               `tfsdk:"trace"`
	UseSlicer              types.Bool               `tfsdk:"use_slicer"`
	ProtectServeKeyEnabled types.Bool               `tfsdk:"protect_serve_key_enabled"`
	APIKeyEnabled          types.Bool               `tfsdk:"api_key_enabled"`
	ReverseProxy           *ReverseProxyConfigModel `tfsdk:"reverse_proxy"`
	ErrorTTL               basetypes.ObjectValue    `tfsdk:"error_ttl"`
	ConTimeout             basetypes.ObjectValue    `tfsdk:"con_timeout"`
	MaxCons                basetypes.ObjectValue    `tfsdk:"max_cons"`
	TTFBTimeout            basetypes.ObjectValue    `tfsdk:"ttfb_timeout"`
	OriginHostHeader       basetypes.ObjectValue    `tfsdk:"origin_hostheader"`
	SharedShield           basetypes.ObjectValue    `tfsdk:"shared_shield"`
}

// represents reverse proxy configuration
type ReverseProxyConfigModel struct {
	Enabled           types.Bool   `tfsdk:"enabled"`
	Hostname          types.String `tfsdk:"hostname"`
	Prepend           types.String `tfsdk:"prepend"`
	TTL               types.Int64  `tfsdk:"ttl"`
	CacheByQueryParam types.Bool   `tfsdk:"cache_by_query_param"`
	OriginScheme      types.String `tfsdk:"origin_scheme"`
	UseRobotsTXT      types.Bool   `tfsdk:"use_robots_txt"`
	Mode              types.String `tfsdk:"mode"`
}

// represents an option with enabled/value structure for Terraform
type OptionModel struct {
	Enabled types.Bool  `tfsdk:"enabled"`
	Value   types.Int64 `tfsdk:"value"`
}

// represents an option with enabled/array value structure
type OptionArrayModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
	Value   types.List `tfsdk:"value"`
}

// represents an option with enabled/string value structure
type OptionStringModel struct {
	Enabled types.Bool   `tfsdk:"enabled"`
	Value   types.String `tfsdk:"value"`
}

// implements planmodifier.Object
type ReverseProxyPlanModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m ReverseProxyPlanModifier) Description(_ context.Context) string {
	return "Automatically sets use_robots_txt to true when reverse proxy is enabled"
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m ReverseProxyPlanModifier) MarkdownDescription(_ context.Context) string {
	return "Automatically sets `use_robots_txt` to `true` when reverse proxy is enabled"
}

// PlanModifyObject implements the plan modification logic.
func (m ReverseProxyPlanModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	// Do nothing if there is no configuration value
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Extract the reverse proxy configuration
	var config ReverseProxyConfigModel
	diags := req.ConfigValue.As(ctx, &config, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If reverse proxy is enabled, ensure use_robots_txt is true and set defaults
	if !config.Enabled.IsNull() && config.Enabled.ValueBool() {
		modified := false

		// Auto-set use_robots_txt to true (API requirement)
		if config.UseRobotsTXT.IsNull() || !config.UseRobotsTXT.ValueBool() {
			config.UseRobotsTXT = types.BoolValue(true)
			modified = true
		}

		// Set sensible defaults for other fields if not provided
		if config.OriginScheme.IsNull() || config.OriginScheme.ValueString() == "" || config.OriginScheme.ValueString() == "http" {
			config.OriginScheme = types.StringValue("FOLLOW")
			modified = true
		}

		if config.Mode.IsNull() || config.Mode.ValueString() == "" {
			config.Mode = types.StringValue("WEB")
			modified = true
		}

		if config.TTL.IsNull() || config.TTL.ValueInt64() == 0 {
			config.TTL = types.Int64Value(2678400) // 31 days
			modified = true
		}

		// If we made any modifications, update the plan value
		if modified {
			newValue, diags := types.ObjectValueFrom(ctx, req.ConfigValue.Type(ctx).(basetypes.ObjectType).AttrTypes, config)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			resp.PlanValue = newValue
		}
	}
}

// getOptionFromObjectValue extracts OptionModel from ObjectValue
func getOptionFromObjectValue(ctx context.Context, objVal basetypes.ObjectValue) (*OptionModel, error) {
	if objVal.IsNull() || objVal.IsUnknown() {
		return nil, nil
	}

	var option OptionModel
	diags := objVal.As(ctx, &option, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, fmt.Errorf("failed to convert ObjectValue to OptionModel")
	}

	return &option, nil
}

// getOptionArrayFromObjectValue extracts OptionArrayModel from ObjectValue
func getOptionArrayFromObjectValue(ctx context.Context, objVal basetypes.ObjectValue) (*OptionArrayModel, error) {
	if objVal.IsNull() || objVal.IsUnknown() {
		return nil, nil
	}

	var option OptionArrayModel
	diags := objVal.As(ctx, &option, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, fmt.Errorf("failed to convert ObjectValue to OptionArrayModel")
	}

	return &option, nil
}

// getOptionStringFromObjectValue extracts OptionStringModel from ObjectValue
func getOptionStringFromObjectValue(ctx context.Context, objVal basetypes.ObjectValue) (*OptionStringModel, error) {
	if objVal.IsNull() || objVal.IsUnknown() {
		return nil, nil
	}

	var option OptionStringModel
	diags := objVal.As(ctx, &option, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, fmt.Errorf("failed to convert ObjectValue to OptionStringModel")
	}

	return &option, nil
}

// Helper function to create the plan modifier
func ReverseProxyAutoConfigurePlanModifier() planmodifier.Object {
	return ReverseProxyPlanModifier{}
}

// ValidateReverseProxyConfig validates the reverse proxy configuration
func (m *ServiceOptionsModel) ValidateReverseProxyConfig() error {
	if m.ReverseProxy == nil {
		return nil // No reverse proxy config is fine
	}

	// If reverse proxy is enabled, validate required fields
	if m.ReverseProxy.Enabled.ValueBool() {
		// Hostname is required when enabled
		if m.ReverseProxy.Hostname.IsNull() || m.ReverseProxy.Hostname.ValueString() == "" {
			return fmt.Errorf("reverse_proxy.hostname is required when reverse_proxy.enabled is true")
		}

		// Validate origin_scheme values
		if !m.ReverseProxy.OriginScheme.IsNull() {
			scheme := m.ReverseProxy.OriginScheme.ValueString()
			if scheme != "" && scheme != "FOLLOW" && scheme != "HTTP" && scheme != "HTTPS" {
				return fmt.Errorf("reverse_proxy.origin_scheme must be one of: FOLLOW, HTTP, HTTPS")
			}
		}

		// Validate mode values
		if !m.ReverseProxy.Mode.IsNull() {
			mode := m.ReverseProxy.Mode.ValueString()
			if mode != "" && mode != "WEB" && mode != "API" {
				return fmt.Errorf("reverse_proxy.mode must be one of: WEB, API")
			}
		}
	}

	return nil
}

// NormalizeReverseProxyConfig automatically sets required values during plan stage
func (m *ServiceOptionsModel) NormalizeReverseProxyConfig() {
	if m.ReverseProxy == nil {
		return
	}

	// If reverse proxy is enabled, automatically set use_robots_txt to true
	if m.ReverseProxy.Enabled.ValueBool() {
		// Auto-set use_robots_txt to true (API requirement)
		if m.ReverseProxy.UseRobotsTXT.IsNull() || !m.ReverseProxy.UseRobotsTXT.ValueBool() {
			m.ReverseProxy.UseRobotsTXT = types.BoolValue(true)
		}

		// Set default values if not provided
		if m.ReverseProxy.OriginScheme.IsNull() || m.ReverseProxy.OriginScheme.ValueString() == "" {
			m.ReverseProxy.OriginScheme = types.StringValue("FOLLOW")
		}

		if m.ReverseProxy.Mode.IsNull() || m.ReverseProxy.Mode.ValueString() == "" {
			m.ReverseProxy.Mode = types.StringValue("WEB")
		}

		if m.ReverseProxy.TTL.IsNull() || m.ReverseProxy.TTL.ValueInt64() == 0 {
			m.ReverseProxy.TTL = types.Int64Value(2678400) // 31 days
		}
	}
}

// ToSDKServiceOptions converts Terraform model to SDK model
func (m *ServiceOptionsModel) ToSDKServiceOptions(ctx context.Context) *api.ServiceOptions {
	opts := &api.ServiceOptions{
		FTP:                    m.FTP.ValueBool(),
		CORS:                   m.CORS.ValueBool(),
		AutoRedirect:           m.AutoRedirect.ValueBool(),
		BrotliCompression:      m.BrotliCompression.ValueBool(),
		BrotliSupport:          m.BrotliSupport.ValueBool(),
		Livestreaming:          m.Livestreaming.ValueBool(),
		NoCache:                m.NoCache.ValueBool(),
		CacheByGeoCountry:      m.CacheByGeoCountry.ValueBool(),
		CacheByRegion:          m.CacheByRegion.ValueBool(),
		CacheByReferer:         m.CacheByReferer.ValueBool(),
		NormalizeQueryString:   m.NormalizeQueryString.ValueBool(),
		AllowRetry:             m.AllowRetry.ValueBool(),
		LinkPreheat:            m.LinkPreheat.ValueBool(),
		EdgeToOrigin:           m.EdgeToOrigin.ValueBool(),
		FollowRedirect:         m.FollowRedirect.ValueBool(),
		PurgeNoQuery:           m.PurgeNoQuery.ValueBool(),
		ForceOrigQString:       m.ForceOrigQString.ValueBool(),
		ServeStale:             m.ServeStale.ValueBool(),
		CachePostRequests:      m.CachePostRequests.ValueBool(),
		SendXFF:                m.SendXFF.ValueBool(),
		UseCFDooTEncoding:      m.UseCFDooTEncoding.ValueBool(),
		SkipURLEncoding:        m.SkipURLEncoding.ValueBool(),
		Trace:                  m.Trace.ValueBool(),
		UseSlicer:              m.UseSlicer.ValueBool(),
		ProtectServeKeyEnabled: m.ProtectServeKeyEnabled.ValueBool(),
		APIKeyEnabled:          m.APIKeyEnabled.ValueBool(),
	}

	// Handle ErrorTTL option
	errorTTL, _ := getOptionFromObjectValue(ctx, m.ErrorTTL)
	if errorTTL != nil && errorTTL.Enabled.ValueBool() {
		opts.ErrorTTL = api.Option{
			Enabled: true,
			Value:   int(errorTTL.Value.ValueInt64()),
		}
	} else {
		opts.ErrorTTL = api.Option{
			Enabled: false,
			Value:   120, // Default value when disabled
		}
	}

	// Handle ConTimeout option
	conTimeout, _ := getOptionFromObjectValue(ctx, m.ConTimeout)
	if conTimeout != nil && conTimeout.Enabled.ValueBool() {
		opts.ConTimeout = api.Option{
			Enabled: true,
			Value:   int(conTimeout.Value.ValueInt64()),
		}
	} else {
		opts.ConTimeout = api.Option{
			Enabled: false,
			Value:   3, // Default value when disabled
		}
	}

	// Handle MaxCons option
	maxCons, _ := getOptionFromObjectValue(ctx, m.MaxCons)
	if maxCons != nil && maxCons.Enabled.ValueBool() {
		opts.MaxCons = api.Option{
			Enabled: true,
			Value:   int(maxCons.Value.ValueInt64()),
		}
	} else {
		opts.MaxCons = api.Option{
			Enabled: false,
			Value:   10, // Default value when disabled
		}
	}

	// Handle TTFBTimeout option
	ttfbTimeout, _ := getOptionFromObjectValue(ctx, m.TTFBTimeout)
	if ttfbTimeout != nil && ttfbTimeout.Enabled.ValueBool() {
		opts.TTFBTimeout = api.Option{
			Enabled: true,
			Value:   int(ttfbTimeout.Value.ValueInt64()),
		}
	} else {
		opts.TTFBTimeout = api.Option{
			Enabled: false,
			Value:   3, // Default value when disabled
		}
	}

	// Handle OriginHostHeader option
	originHostHeader, _ := getOptionArrayFromObjectValue(ctx, m.OriginHostHeader)
	if originHostHeader != nil && originHostHeader.Enabled.ValueBool() {
		// Convert types.List to slice for API
		var valueSlice []interface{}
		if !originHostHeader.Value.IsNull() && !originHostHeader.Value.IsUnknown() {
			elements := originHostHeader.Value.Elements()
			for _, element := range elements {
				if strVal, ok := element.(types.String); ok {
					valueSlice = append(valueSlice, strVal.ValueString())
				}
			}
		}

		opts.OriginHostHeader = api.Option{
			Enabled: true,
			Value:   valueSlice,
		}
	} else {
		opts.OriginHostHeader = api.Option{
			Enabled: false,
			Value:   []interface{}{}, // Empty array as default
		}
	}

	// Handle SharedShield option
	sharedShield, _ := getOptionStringFromObjectValue(ctx, m.SharedShield)
	if sharedShield != nil && sharedShield.Enabled.ValueBool() {
		opts.SharedShield = api.Option{
			Enabled: true,
			Value:   sharedShield.Value.ValueString(),
		}
	} else {
		opts.SharedShield = api.Option{
			Enabled: false,
			Value:   "", // Empty string as default
		}
	}

	// Convert reverse proxy config
	if m.ReverseProxy != nil && m.ReverseProxy.Enabled.ValueBool() {
		// When reverse proxy is enabled, include all required fields with proper defaults
		originScheme := m.ReverseProxy.OriginScheme.ValueString()
		if originScheme == "" {
			originScheme = "FOLLOW" // Default from working example
		}

		mode := m.ReverseProxy.Mode.ValueString()
		if mode == "" {
			mode = "WEB" // Default
		}

		ttl := int(m.ReverseProxy.TTL.ValueInt64())
		if ttl == 0 {
			ttl = 2678400 // Default TTL (31 days)
		}

		// Ensure use_robots_txt is true when enabled (API requirement)
		useRobotsTXT := m.ReverseProxy.UseRobotsTXT.ValueBool()
		if !useRobotsTXT {
			// This should have been caught by validation, but ensure it's true for API
			useRobotsTXT = true
		}

		opts.ReverseProxy = api.ReverseProxyConfig{
			Enabled:           true,
			Mode:              mode,
			CacheByQueryParam: m.ReverseProxy.CacheByQueryParam.ValueBool(),
			Hostname:          m.ReverseProxy.Hostname.ValueString(),
			OriginScheme:      originScheme,
			Prepend:           m.ReverseProxy.Prepend.ValueString(),
			TTL:               ttl,
			UseRobotsTXT:      useRobotsTXT,
		}
	} else {
		// When reverse proxy is disabled, only set enabled to false
		opts.ReverseProxy = api.ReverseProxyConfig{
			Enabled: false,
		}
	}

	// Initialize required arrays
	opts.MimeTypesOverrides = make([]api.MimeTypeOverride, 0)
	opts.ExpiryHeaders = make([]api.ExpiryHeader, 0)

	// Initialize all Option fields with defaults
	opts.RawLogs = api.Option{Enabled: false, Value: ""}
	opts.BWThrottle = api.Option{Enabled: false, Value: ""}
	opts.PurgeMode = api.Option{Enabled: false, Value: ""}
	opts.DirPurgeSkip = api.Option{Enabled: false, Value: ""}
	opts.SkipPserveExt = api.Option{Enabled: false, Value: ""}
	opts.HTTPMethods = api.Option{Enabled: false, Value: ""}
	opts.BWThrottleQuery = api.Option{Enabled: false, Value: ""}
	opts.Slice = api.Option{Enabled: false, Value: ""}
	opts.Redirect = api.Option{Enabled: false, Value: ""}
	opts.SkipEncodingExt = api.Option{Enabled: false, Value: ""}

	return opts
}

// FromSDKServiceOptions converts SDK model to Terraform model
func (m *ServiceOptionsModel) FromSDKServiceOptions(ctx context.Context, opts *api.ServiceOptions) {
	m.FTP = types.BoolValue(opts.FTP)
	m.CORS = types.BoolValue(opts.CORS)
	m.AutoRedirect = types.BoolValue(opts.AutoRedirect)
	m.BrotliCompression = types.BoolValue(opts.BrotliCompression)
	m.BrotliSupport = types.BoolValue(opts.BrotliSupport)
	m.Livestreaming = types.BoolValue(opts.Livestreaming)
	m.NoCache = types.BoolValue(opts.NoCache)
	m.CacheByGeoCountry = types.BoolValue(opts.CacheByGeoCountry)
	m.CacheByRegion = types.BoolValue(opts.CacheByRegion)
	m.CacheByReferer = types.BoolValue(opts.CacheByReferer)
	m.NormalizeQueryString = types.BoolValue(opts.NormalizeQueryString)
	m.AllowRetry = types.BoolValue(opts.AllowRetry)
	m.LinkPreheat = types.BoolValue(opts.LinkPreheat)
	m.EdgeToOrigin = types.BoolValue(opts.EdgeToOrigin)
	m.FollowRedirect = types.BoolValue(opts.FollowRedirect)
	m.PurgeNoQuery = types.BoolValue(opts.PurgeNoQuery)
	m.ForceOrigQString = types.BoolValue(opts.ForceOrigQString)
	m.ServeStale = types.BoolValue(opts.ServeStale)
	m.CachePostRequests = types.BoolValue(opts.CachePostRequests)
	m.SendXFF = types.BoolValue(opts.SendXFF)
	m.UseCFDooTEncoding = types.BoolValue(opts.UseCFDooTEncoding)
	m.SkipURLEncoding = types.BoolValue(opts.SkipURLEncoding)
	m.Trace = types.BoolValue(opts.Trace)
	m.UseSlicer = types.BoolValue(opts.UseSlicer)
	m.ProtectServeKeyEnabled = types.BoolValue(opts.ProtectServeKeyEnabled)
	m.APIKeyEnabled = types.BoolValue(opts.APIKeyEnabled)

	// Convert ErrorTTL option
	errorTTLOption := &OptionModel{
		Enabled: types.BoolValue(opts.ErrorTTL.Enabled),
		Value:   types.Int64Value(60), // Default value
	}

	// Handle the ErrorTTL value
	if opts.ErrorTTL.Enabled {
		switch v := opts.ErrorTTL.Value.(type) {
		case int:
			errorTTLOption.Value = types.Int64Value(int64(v))
		case int64:
			errorTTLOption.Value = types.Int64Value(v)
		case float64:
			errorTTLOption.Value = types.Int64Value(int64(v))
		default:
			errorTTLOption.Value = types.Int64Value(60) // Fallback default
		}
	}

	m.ErrorTTL, _ = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"enabled": types.BoolType,
		"value":   types.Int64Type,
	}, errorTTLOption)

	// Convert ConTimeout option
	conTimeoutOption := &OptionModel{
		Enabled: types.BoolValue(opts.ConTimeout.Enabled),
		Value:   types.Int64Value(3), // Default value
	}

	// Handle the ConTimeout value
	switch v := opts.ConTimeout.Value.(type) {
	case int:
		conTimeoutOption.Value = types.Int64Value(int64(v))
	case int64:
		conTimeoutOption.Value = types.Int64Value(v)
	case float64:
		conTimeoutOption.Value = types.Int64Value(int64(v))
	default:
		if opts.ConTimeout.Enabled {
			conTimeoutOption.Value = types.Int64Value(3) // Conservative fallback
		} else {
			conTimeoutOption.Value = types.Int64Value(3) // Default when disabled
		}
	}

	m.ConTimeout, _ = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"enabled": types.BoolType,
		"value":   types.Int64Type,
	}, conTimeoutOption)

	// Convert MaxCons option
	maxConsOption := &OptionModel{
		Enabled: types.BoolValue(opts.MaxCons.Enabled),
		Value:   types.Int64Value(10), // Default value
	}

	// Handle the MaxCons value
	switch v := opts.MaxCons.Value.(type) {
	case int:
		maxConsOption.Value = types.Int64Value(int64(v))
	case int64:
		maxConsOption.Value = types.Int64Value(v)
	case float64:
		maxConsOption.Value = types.Int64Value(int64(v))
	default:
		if opts.MaxCons.Enabled {
			maxConsOption.Value = types.Int64Value(10) // Conservative fallback
		} else {
			maxConsOption.Value = types.Int64Value(10) // Default when disabled
		}
	}

	m.MaxCons, _ = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"enabled": types.BoolType,
		"value":   types.Int64Type,
	}, maxConsOption)

	// Convert TTFBTimeout option
	ttfbTimeoutOption := &OptionModel{
		Enabled: types.BoolValue(opts.TTFBTimeout.Enabled),
		Value:   types.Int64Value(3), // Default value
	}

	// Handle the TTFBTimeout value
	switch v := opts.TTFBTimeout.Value.(type) {
	case int:
		ttfbTimeoutOption.Value = types.Int64Value(int64(v))
	case int64:
		ttfbTimeoutOption.Value = types.Int64Value(v)
	case float64:
		ttfbTimeoutOption.Value = types.Int64Value(int64(v))
	default:
		if opts.TTFBTimeout.Enabled {
			ttfbTimeoutOption.Value = types.Int64Value(3) // Conservative fallback
		} else {
			ttfbTimeoutOption.Value = types.Int64Value(3) // Default when disabled
		}
	}

	m.TTFBTimeout, _ = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"enabled": types.BoolType,
		"value":   types.Int64Type,
	}, ttfbTimeoutOption)

	// Convert OriginHostHeader option
	originHostHeaderOption := &OptionArrayModel{
		Enabled: types.BoolValue(opts.OriginHostHeader.Enabled),
		Value:   types.ListNull(types.StringType), // Default empty list
	}

	// Handle the OriginHostHeader value (array)
	if opts.OriginHostHeader.Value != nil {
		switch v := opts.OriginHostHeader.Value.(type) {
		case []interface{}:
			// Convert slice to types.List
			var stringValues []attr.Value
			for _, item := range v {
				if str, ok := item.(string); ok {
					stringValues = append(stringValues, types.StringValue(str))
				}
			}
			if len(stringValues) > 0 {
				originHostHeaderOption.Value, _ = types.ListValue(types.StringType, stringValues)
			}
		case []string:
			// Handle direct string slice
			var stringValues []attr.Value
			for _, str := range v {
				stringValues = append(stringValues, types.StringValue(str))
			}
			if len(stringValues) > 0 {
				originHostHeaderOption.Value, _ = types.ListValue(types.StringType, stringValues)
			}
		default:
			// Keep as empty list for unknown types
			originHostHeaderOption.Value = types.ListNull(types.StringType)
		}
	}

	m.OriginHostHeader, _ = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"enabled": types.BoolType,
		"value":   types.ListType{ElemType: types.StringType},
	}, originHostHeaderOption)

	// Convert SharedShield option
	sharedShieldOption := &OptionStringModel{
		Enabled: types.BoolValue(opts.SharedShield.Enabled),
		Value:   types.StringValue(""), // Default empty string
	}

	// Handle the SharedShield value (string)
	if opts.SharedShield.Value != nil {
		switch v := opts.SharedShield.Value.(type) {
		case string:
			sharedShieldOption.Value = types.StringValue(v)
		default:
			// Keep as empty string for unknown types
			sharedShieldOption.Value = types.StringValue("")
		}
	}

	m.SharedShield, _ = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"enabled": types.BoolType,
		"value":   types.StringType,
	}, sharedShieldOption)

	// Convert reverse proxy config - always populate
	m.ReverseProxy = &ReverseProxyConfigModel{
		Enabled:           types.BoolValue(opts.ReverseProxy.Enabled),
		Hostname:          types.StringValue(opts.ReverseProxy.Hostname),
		Prepend:           types.StringValue(opts.ReverseProxy.Prepend),
		TTL:               types.Int64Value(int64(opts.ReverseProxy.TTL)),
		CacheByQueryParam: types.BoolValue(opts.ReverseProxy.CacheByQueryParam),
		OriginScheme:      types.StringValue(opts.ReverseProxy.OriginScheme),
		UseRobotsTXT:      types.BoolValue(opts.ReverseProxy.UseRobotsTXT),
		Mode:              types.StringValue(opts.ReverseProxy.Mode),
	}
}
