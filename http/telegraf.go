package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/gddo/httputil"
	"github.com/influxdata/httprouter"
	platform "github.com/influxdata/influxdb"
	pctx "github.com/influxdata/influxdb/context"
	"github.com/influxdata/influxdb/telegraf/plugins"
	"go.uber.org/zap"
)

// TelegrafBackend is all services and associated parameters required to construct
// the TelegrafHandler.
type TelegrafBackend struct {
	platform.HTTPErrorHandler
	Logger *zap.Logger

	TelegrafService            platform.TelegrafConfigStore
	UserResourceMappingService platform.UserResourceMappingService
	LabelService               platform.LabelService
	UserService                platform.UserService
	OrganizationService        platform.OrganizationService
}

// NewTelegrafBackend returns a new instance of TelegrafBackend.
func NewTelegrafBackend(b *APIBackend) *TelegrafBackend {
	return &TelegrafBackend{
		HTTPErrorHandler: b.HTTPErrorHandler,
		Logger:           b.Logger.With(zap.String("handler", "telegraf")),

		TelegrafService:            b.TelegrafService,
		UserResourceMappingService: b.UserResourceMappingService,
		LabelService:               b.LabelService,
		UserService:                b.UserService,
		OrganizationService:        b.OrganizationService,
	}
}

// TelegrafHandler is the handler for the telegraf service
type TelegrafHandler struct {
	*httprouter.Router
	platform.HTTPErrorHandler
	Logger *zap.Logger

	TelegrafService            platform.TelegrafConfigStore
	UserResourceMappingService platform.UserResourceMappingService
	LabelService               platform.LabelService
	UserService                platform.UserService
	OrganizationService        platform.OrganizationService
}

const (
	telegrafsPath            = "/api/v2/telegrafs"
	telegrafsIDPath          = "/api/v2/telegrafs/:id"
	telegrafsIDMembersPath   = "/api/v2/telegrafs/:id/members"
	telegrafsIDMembersIDPath = "/api/v2/telegrafs/:id/members/:userID"
	telegrafsIDOwnersPath    = "/api/v2/telegrafs/:id/owners"
	telegrafsIDOwnersIDPath  = "/api/v2/telegrafs/:id/owners/:userID"
	telegrafsIDLabelsPath    = "/api/v2/telegrafs/:id/labels"
	telegrafsIDLabelsIDPath  = "/api/v2/telegrafs/:id/labels/:lid"

	telegrafPluginsPath = "/api/v2/telegraf/plugins"
)

// NewTelegrafHandler returns a new instance of TelegrafHandler.
func NewTelegrafHandler(b *TelegrafBackend) *TelegrafHandler {
	h := &TelegrafHandler{
		Router:           NewRouter(b.HTTPErrorHandler),
		HTTPErrorHandler: b.HTTPErrorHandler,
		Logger:           b.Logger,

		TelegrafService:            b.TelegrafService,
		UserResourceMappingService: b.UserResourceMappingService,
		LabelService:               b.LabelService,
		UserService:                b.UserService,
		OrganizationService:        b.OrganizationService,
	}
	h.HandlerFunc("POST", telegrafsPath, h.handlePostTelegraf)
	h.HandlerFunc("GET", telegrafsPath, h.handleGetTelegrafs)
	h.HandlerFunc("GET", telegrafsIDPath, h.handleGetTelegraf)
	h.HandlerFunc("DELETE", telegrafsIDPath, h.handleDeleteTelegraf)
	h.HandlerFunc("PUT", telegrafsIDPath, h.handlePutTelegraf)

	h.HandlerFunc("GET", telegrafPluginsPath, h.handleGetTelegrafPlugins)

	memberBackend := MemberBackend{
		HTTPErrorHandler:           b.HTTPErrorHandler,
		Logger:                     b.Logger.With(zap.String("handler", "member")),
		ResourceType:               platform.TelegrafsResourceType,
		UserType:                   platform.Member,
		UserResourceMappingService: b.UserResourceMappingService,
		UserService:                b.UserService,
	}
	h.HandlerFunc("POST", telegrafsIDMembersPath, newPostMemberHandler(memberBackend))
	h.HandlerFunc("GET", telegrafsIDMembersPath, newGetMembersHandler(memberBackend))
	h.HandlerFunc("DELETE", telegrafsIDMembersIDPath, newDeleteMemberHandler(memberBackend))

	ownerBackend := MemberBackend{
		HTTPErrorHandler:           b.HTTPErrorHandler,
		Logger:                     b.Logger.With(zap.String("handler", "member")),
		ResourceType:               platform.TelegrafsResourceType,
		UserType:                   platform.Owner,
		UserResourceMappingService: b.UserResourceMappingService,
		UserService:                b.UserService,
	}
	h.HandlerFunc("POST", telegrafsIDOwnersPath, newPostMemberHandler(ownerBackend))
	h.HandlerFunc("GET", telegrafsIDOwnersPath, newGetMembersHandler(ownerBackend))
	h.HandlerFunc("DELETE", telegrafsIDOwnersIDPath, newDeleteMemberHandler(ownerBackend))

	labelBackend := &LabelBackend{
		HTTPErrorHandler: b.HTTPErrorHandler,
		Logger:           b.Logger.With(zap.String("handler", "label")),
		LabelService:     b.LabelService,
		ResourceType:     platform.TelegrafsResourceType,
	}
	h.HandlerFunc("GET", telegrafsIDLabelsPath, newGetLabelsHandler(labelBackend))
	h.HandlerFunc("POST", telegrafsIDLabelsPath, newPostLabelHandler(labelBackend))
	h.HandlerFunc("DELETE", telegrafsIDLabelsIDPath, newDeleteLabelHandler(labelBackend))

	return h
}

type telegrafLinks struct {
	Self    string `json:"self"`
	Labels  string `json:"labels"`
	Members string `json:"members"`
	Owners  string `json:"owners"`
}

type telegrafResponse struct {
	*platform.TelegrafConfig
	Labels []platform.Label `json:"labels"`
	Links  telegrafLinks    `json:"links"`
}

type telegrafResponses struct {
	TelegrafConfigs []*telegrafResponse `json:"configurations"`
}

func newTelegrafResponse(tc *platform.TelegrafConfig, labels []*platform.Label) *telegrafResponse {
	res := &telegrafResponse{
		TelegrafConfig: tc,
		Links: telegrafLinks{
			Self:    fmt.Sprintf("/api/v2/telegrafs/%s", tc.ID),
			Labels:  fmt.Sprintf("/api/v2/telegrafs/%s/labels", tc.ID),
			Members: fmt.Sprintf("/api/v2/telegrafs/%s/members", tc.ID),
			Owners:  fmt.Sprintf("/api/v2/telegrafs/%s/owners", tc.ID),
		},
		Labels: []platform.Label{},
	}

	for _, l := range labels {
		res.Labels = append(res.Labels, *l)
	}

	return res
}

func newTelegrafResponses(ctx context.Context, tcs []*platform.TelegrafConfig, labelService platform.LabelService) *telegrafResponses {
	resp := &telegrafResponses{
		TelegrafConfigs: make([]*telegrafResponse, len(tcs)),
	}
	for i, c := range tcs {
		labels, _ := labelService.FindResourceLabels(ctx, platform.LabelMappingFilter{ResourceID: c.ID})
		resp.TelegrafConfigs[i] = newTelegrafResponse(c, labels)
	}
	return resp
}

func decodeGetTelegrafRequest(ctx context.Context, r *http.Request) (i platform.ID, err error) {
	params := httprouter.ParamsFromContext(ctx)
	id := params.ByName("id")
	if id == "" {
		return i, &platform.Error{
			Code: platform.EInvalid,
			Msg:  "url missing id",
		}
	}

	if err := i.DecodeFromString(id); err != nil {
		return i, err
	}
	return i, nil
}

func decodeGetTelegrafPluginRequest(ctx context.Context, r *http.Request) (*plugins.TelegrafPlugins, error) {
	t, ok := r.URL.Query()["type"]
	if !ok || len(t) < 1 {
		return plugins.AvailablePlugins()
	}

	var telPlugins *plugins.TelegrafPlugins
	var err error

	switch t[0] {
	case "inputs":
		telPlugins, err = plugins.AvailableInputs()
		if err != nil {
			return nil, err
		}
	case "outputs":
		telPlugins, err = plugins.AvailableOutputs()
		if err != nil {
			return nil, err
		}
	case "processors":
		telPlugins, err = plugins.AvailableProcessors()
		if err != nil {
			return nil, err
		}
	case "aggregators":
		telPlugins, err = plugins.AvailableAggregators()
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown plugin type '%s'", t)
	}

	return telPlugins, nil
}

func (h *TelegrafHandler) handleGetTelegrafPlugins(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	telPlugins, err := decodeGetTelegrafPluginRequest(ctx, r)
	if err != nil {
		h.HandleHTTPError(ctx, err, w)
		return
	}

	h.Logger.Debug("telegraf plugins retrieved", zap.String("plugins", fmt.Sprint(telPlugins)))

	if err := encodeResponse(ctx, w, http.StatusOK, telPlugins); err != nil {
		logEncodingError(h.Logger, r, err)
		return
	}
}

func (h *TelegrafHandler) handleGetTelegrafs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter, err := decodeTelegrafConfigFilter(ctx, r)
	if err != nil {
		h.Logger.Debug("failed to decode request", zap.Error(err))
		h.HandleHTTPError(ctx, err, w)
		return
	}
	tcs, _, err := h.TelegrafService.FindTelegrafConfigs(ctx, *filter)
	if err != nil {
		h.HandleHTTPError(ctx, err, w)
		return
	}
	h.Logger.Debug("telegrafs retrieved", zap.String("telegrafs", fmt.Sprint(tcs)))

	if err := encodeResponse(ctx, w, http.StatusOK, newTelegrafResponses(ctx, tcs, h.LabelService)); err != nil {
		logEncodingError(h.Logger, r, err)
		return
	}
}

func (h *TelegrafHandler) handleGetTelegraf(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := decodeGetTelegrafRequest(ctx, r)
	if err != nil {
		h.HandleHTTPError(ctx, err, w)
		return
	}
	tc, err := h.TelegrafService.FindTelegrafConfigByID(ctx, id)
	if err != nil {
		h.HandleHTTPError(ctx, err, w)
		return
	}
	h.Logger.Debug("telegraf retrieved", zap.String("telegraf", fmt.Sprint(tc)))

	offers := []string{"application/toml", "application/json", "application/octet-stream"}
	defaultOffer := "application/toml"
	mimeType := httputil.NegotiateContentType(r, offers, defaultOffer)
	switch mimeType {
	case "application/octet-stream":
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.toml\"", strings.Replace(strings.TrimSpace(tc.Name), " ", "_", -1)))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(tc.Config))
	case "application/json":
		labels, err := h.LabelService.FindResourceLabels(ctx, platform.LabelMappingFilter{ResourceID: tc.ID})
		if err != nil {
			h.HandleHTTPError(ctx, err, w)
			return
		}

		if err := encodeResponse(ctx, w, http.StatusOK, newTelegrafResponse(tc, labels)); err != nil {
			logEncodingError(h.Logger, r, err)
			return
		}
	case "application/toml":
		w.Header().Set("Content-Type", "application/toml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(tc.Config))
	}
}

func decodeTelegrafConfigFilter(ctx context.Context, r *http.Request) (*platform.TelegrafConfigFilter, error) {
	f := &platform.TelegrafConfigFilter{}
	urm, err := decodeUserResourceMappingFilter(ctx, r, platform.TelegrafsResourceType)
	if err == nil {
		f.UserResourceMappingFilter = *urm
	}

	q := r.URL.Query()
	if orgIDStr := q.Get("orgID"); orgIDStr != "" {
		orgID, err := platform.IDFromString(orgIDStr)
		if err != nil {
			return f, &platform.Error{
				Code: platform.EInvalid,
				Msg:  "orgID is invalid",
				Err:  err,
			}
		}
		f.OrgID = orgID
	} else if orgNameStr := q.Get("org"); orgNameStr != "" {
		*f.Organization = orgNameStr
	}
	return f, err
}

func decodePostTelegrafRequest(ctx context.Context, r *http.Request) (*platform.TelegrafConfig, error) {
	tc := new(platform.TelegrafConfig)
	err := json.NewDecoder(r.Body).Decode(tc)
	// fmt.Printf("GOT THING: %+v\n", tc)
	return tc, err
}

func decodePutTelegrafRequest(ctx context.Context, r *http.Request) (*platform.TelegrafConfig, error) {
	tc := new(platform.TelegrafConfig)
	if err := json.NewDecoder(r.Body).Decode(tc); err != nil {
		return nil, err
	}
	params := httprouter.ParamsFromContext(ctx)
	id := params.ByName("id")
	if id == "" {
		return nil, &platform.Error{
			Code: platform.EInvalid,
			Msg:  "url missing id",
		}
	}
	i := new(platform.ID)
	if err := i.DecodeFromString(id); err != nil {
		return nil, err
	}
	tc.ID = *i
	return tc, nil
}

// handlePostTelegraf is the HTTP handler for the POST /api/v2/telegrafs route.
func (h *TelegrafHandler) handlePostTelegraf(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tc, err := decodePostTelegrafRequest(ctx, r)
	if err != nil {
		h.Logger.Debug("failed to decode request", zap.Error(err))
		h.HandleHTTPError(ctx, err, w)
		return
	}
	auth, err := pctx.GetAuthorizer(ctx)
	if err != nil {
		h.HandleHTTPError(ctx, err, w)
		return
	}

	if err := h.TelegrafService.CreateTelegrafConfig(ctx, tc, auth.GetUserID()); err != nil {
		h.HandleHTTPError(ctx, err, w)
		return
	}
	h.Logger.Debug("telegraf created", zap.String("telegraf", fmt.Sprint(tc)))

	if err := encodeResponse(ctx, w, http.StatusCreated, newTelegrafResponse(tc, []*platform.Label{})); err != nil {
		logEncodingError(h.Logger, r, err)
		return
	}
}

// handlePutTelegraf is the HTTP handler for the POST /api/v2/telegrafs route.
func (h *TelegrafHandler) handlePutTelegraf(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tc, err := decodePutTelegrafRequest(ctx, r)
	if err != nil {
		h.Logger.Debug("failed to decode request", zap.Error(err))
		h.HandleHTTPError(ctx, err, w)
		return
	}
	auth, err := pctx.GetAuthorizer(ctx)
	if err != nil {
		h.HandleHTTPError(ctx, err, w)
		return
	}

	tc, err = h.TelegrafService.UpdateTelegrafConfig(ctx, tc.ID, tc, auth.GetUserID())
	if err != nil {
		h.HandleHTTPError(ctx, err, w)
		return
	}

	labels, err := h.LabelService.FindResourceLabels(ctx, platform.LabelMappingFilter{ResourceID: tc.ID})
	if err != nil {
		h.HandleHTTPError(ctx, err, w)
		return
	}
	h.Logger.Debug("telegraf updated", zap.String("telegraf", fmt.Sprint(tc)))

	if err := encodeResponse(ctx, w, http.StatusOK, newTelegrafResponse(tc, labels)); err != nil {
		logEncodingError(h.Logger, r, err)
		return
	}
}

func (h *TelegrafHandler) handleDeleteTelegraf(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	i, err := decodeGetTelegrafRequest(ctx, r)
	if err != nil {
		h.HandleHTTPError(ctx, err, w)
		return
	}

	if err = h.TelegrafService.DeleteTelegrafConfig(ctx, i); err != nil {
		h.HandleHTTPError(ctx, err, w)
		return
	}
	h.Logger.Debug("telegraf deleted", zap.String("telegrafID", fmt.Sprint(i)))

	w.WriteHeader(http.StatusNoContent)
}
