// Package v1 provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.12.2 DO NOT EDIT.
package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	openapi_types "github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/go-chi/chi/v5"
	google_uuid "github.com/google/uuid"
)

// Defines values for SortOrder.
const (
	ASC  SortOrder = "ASC"
	DESC SortOrder = "DESC"
)

// Error defines model for Error.
type Error struct {
	// Code error code, unique per error
	Code string `json:"code"`

	// Message error message
	Message string `json:"message"`
}

// InternalError defines model for InternalError.
type InternalError struct {
	Message string `json:"message"`
}

// Match defines model for Match.
type Match struct {
	MapName       string           `json:"map_name"`
	MatchDuration time.Duration    `json:"match_duration"`
	MatchID       google_uuid.UUID `json:"match_id"`
	Team1         MatchTeam        `json:"team1"`
	Team2         MatchTeam        `json:"team2"`

	// UploadTime RFC3339 datetime string
	UploadTime time.Time `json:"upload_time"`
}

// MatchList defines model for MatchList.
type MatchList struct {
	Matches       []Match `json:"matches"`
	NextPageToken string  `json:"next_page_token"`
}

// MatchTeam defines model for MatchTeam.
type MatchTeam struct {
	ClanName string `json:"clan_name"`

	// FlagCode ISO 3166 flag code
	FlagCode       string   `json:"flag_code"`
	PlayerSteamIds []uint64 `json:"player_steam_ids"`
	Score          uint8    `json:"score"`
}

// PlayerMatchListRequest defines model for PlayerMatchListRequest.
type PlayerMatchListRequest struct {
	// PageSize uint16 integer
	PageSize uint64 `json:"page_size"`

	// PageToken base64 string
	PageToken string              `json:"page_token"`
	Sort      PlayerMatchListSort `json:"sort"`
}

// PlayerMatchListSort defines model for PlayerMatchListSort.
type PlayerMatchListSort struct {
	UploadTime SortOrder `json:"upload_time"`
}

// PlayerProfile defines model for PlayerProfile.
type PlayerProfile struct {
	// CreateTime RFC3339 datetime string
	CreateTime   time.Time   `json:"create_time"`
	MainTeamName string      `json:"main_team_name"`
	Stats        PlayerStats `json:"stats"`
	SteamID      uint64      `json:"steam_id"`

	// UpdateTime RFC3339 datetime string
	UpdateTime       time.Time        `json:"update_time"`
	WeaponClassStats WeaponClassStats `json:"weapon_class_stats"`
	WeaponStats      WeaponStats      `json:"weapon_stats"`
}

// PlayerStats defines model for PlayerStats.
type PlayerStats struct {
	AssistsPerRound       float32 `json:"assists_per_round"`
	BlindPerRound         float32 `json:"blind_per_round"`
	BlindedPerRound       float32 `json:"blinded_per_round"`
	DamagePerRound        float32 `json:"damage_per_round"`
	DeathsPerRound        float32 `json:"deaths_per_round"`
	GrenadeDamagePerRound float32 `json:"grenade_damage_per_round"`
	HeadshotPercentage    float32 `json:"headshot_percentage"`
	KillDeathRatio        float32 `json:"kill_death_ratio"`
	KillsPerRound         float32 `json:"kills_per_round"`
	MatchesPlayed         uint16  `json:"matches_played"`
	RoundsPlayed          uint32  `json:"rounds_played"`
	TotalDeaths           uint32  `json:"total_deaths"`
	TotalKills            uint32  `json:"total_kills"`
}

// SortOrder defines model for SortOrder.
type SortOrder string

// WeaponClassStat defines model for WeaponClassStat.
type WeaponClassStat struct {
	TotalKills  uint32 `json:"total_kills"`
	WeaponClass string `json:"weapon_class"`
}

// WeaponClassStats defines model for WeaponClassStats.
type WeaponClassStats = []WeaponClassStat

// WeaponStat defines model for WeaponStat.
type WeaponStat struct {
	TotalKills uint32 `json:"total_kills"`
	WeaponName string `json:"weapon_name"`
}

// WeaponStats defines model for WeaponStats.
type WeaponStats = []WeaponStat

// PlayerMatchListRequestBody defines model for PlayerMatchListRequestBody.
type PlayerMatchListRequestBody = PlayerMatchListRequest

// UploadReplayMultipartBody defines parameters for UploadReplay.
type UploadReplayMultipartBody struct {
	// ReplayArchive архив с демкой, только 1 демка из архива будет загружена, макс. размер архива - 500 мб
	ReplayArchive *openapi_types.File `json:"replay_archive,omitempty"`
}

// GetPlayerMatchesJSONRequestBody defines body for GetPlayerMatches for application/json ContentType.
type GetPlayerMatchesJSONRequestBody = PlayerMatchListRequest

// UploadReplayMultipartRequestBody defines body for UploadReplay for multipart/form-data ContentType.
type UploadReplayMultipartRequestBody UploadReplayMultipartBody

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Get player profile
	// (GET /players/{steam_id})
	GetPlayerProfile(w http.ResponseWriter, r *http.Request, steamId uint64)
	// Get player matches
	// (POST /players/{steam_id}/matches)
	GetPlayerMatches(w http.ResponseWriter, r *http.Request, steamId uint64)
	// Upload replay
	// (POST /replays)
	UploadReplay(w http.ResponseWriter, r *http.Request)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.Handler) http.Handler

// GetPlayerProfile operation middleware
func (siw *ServerInterfaceWrapper) GetPlayerProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "steam_id" -------------
	var steamId uint64

	err = runtime.BindStyledParameterWithLocation("simple", false, "steam_id", runtime.ParamLocationPath, chi.URLParam(r, "steam_id"), &steamId)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "steam_id", Err: err})
		return
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetPlayerProfile(w, r, steamId)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// GetPlayerMatches operation middleware
func (siw *ServerInterfaceWrapper) GetPlayerMatches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "steam_id" -------------
	var steamId uint64

	err = runtime.BindStyledParameterWithLocation("simple", false, "steam_id", runtime.ParamLocationPath, chi.URLParam(r, "steam_id"), &steamId)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "steam_id", Err: err})
		return
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetPlayerMatches(w, r, steamId)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// UploadReplay operation middleware
func (siw *ServerInterfaceWrapper) UploadReplay(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.UploadReplay(w, r)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

type UnescapedCookieParamError struct {
	ParamName string
	Err       error
}

func (e *UnescapedCookieParamError) Error() string {
	return fmt.Sprintf("error unescaping cookie parameter '%s'", e.ParamName)
}

func (e *UnescapedCookieParamError) Unwrap() error {
	return e.Err
}

type UnmarshallingParamError struct {
	ParamName string
	Err       error
}

func (e *UnmarshallingParamError) Error() string {
	return fmt.Sprintf("Error unmarshalling parameter %s as JSON: %s", e.ParamName, e.Err.Error())
}

func (e *UnmarshallingParamError) Unwrap() error {
	return e.Err
}

type RequiredParamError struct {
	ParamName string
}

func (e *RequiredParamError) Error() string {
	return fmt.Sprintf("Query argument %s is required, but not found", e.ParamName)
}

type RequiredHeaderError struct {
	ParamName string
	Err       error
}

func (e *RequiredHeaderError) Error() string {
	return fmt.Sprintf("Header parameter %s is required, but not found", e.ParamName)
}

func (e *RequiredHeaderError) Unwrap() error {
	return e.Err
}

type InvalidParamFormatError struct {
	ParamName string
	Err       error
}

func (e *InvalidParamFormatError) Error() string {
	return fmt.Sprintf("Invalid format for parameter %s: %s", e.ParamName, e.Err.Error())
}

func (e *InvalidParamFormatError) Unwrap() error {
	return e.Err
}

type TooManyValuesForParamError struct {
	ParamName string
	Count     int
}

func (e *TooManyValuesForParamError) Error() string {
	return fmt.Sprintf("Expected one value for %s, got %d", e.ParamName, e.Count)
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{})
}

type ChiServerOptions struct {
	BaseURL          string
	BaseRouter       chi.Router
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r chi.Router) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseRouter: r,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, r chi.Router, baseURL string) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseURL:    baseURL,
		BaseRouter: r,
	})
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options ChiServerOptions) http.Handler {
	r := options.BaseRouter

	if r == nil {
		r = chi.NewRouter()
	}
	if options.ErrorHandlerFunc == nil {
		options.ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandlerFunc:   options.ErrorHandlerFunc,
	}

	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/players/{steam_id}", wrapper.GetPlayerProfile)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/players/{steam_id}/matches", wrapper.GetPlayerMatches)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/replays", wrapper.UploadReplay)
	})

	return r
}
