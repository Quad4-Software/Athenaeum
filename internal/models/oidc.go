package models

// OIDCMatchBy selects how an OIDC identity maps to a local account.
type OIDCMatchBy string

const (
	OIDCMatchUsername OIDCMatchBy = "username"
	OIDCMatchEmail    OIDCMatchBy = "email"
	OIDCMatchSub      OIDCMatchBy = "sub"
)

// OIDCConfig holds OpenID Connect provider settings.
type OIDCConfig struct {
	Enabled          bool        `json:"enabled"`
	LoginLocal       bool        `json:"loginLocal"`
	IssuerURL        string      `json:"issuerUrl"`
	AuthorizeURL     string      `json:"authorizeUrl"`
	TokenURL         string      `json:"tokenUrl"`
	UserinfoURL      string      `json:"userinfoUrl"`
	JWKSURL          string      `json:"jwksUrl"`
	LogoutURL        string      `json:"logoutUrl,omitempty"`
	ClientID         string      `json:"clientId"`
	ClientSecret     string      `json:"clientSecret,omitempty"`
	ClientSecretSet  bool        `json:"clientSecretSet"`
	SigningAlgorithm string      `json:"signingAlgorithm"`
	ButtonText       string      `json:"buttonText"`
	MatchBy          OIDCMatchBy `json:"matchBy"`
	AutoRegister     bool        `json:"autoRegister"`
	AutoLaunch       bool        `json:"autoLaunch"`
	GroupClaim       string      `json:"groupClaim"`
	AdminGroups      string      `json:"adminGroups"`
}

// OIDCDiscovery is populated from an issuer's well-known configuration.
type OIDCDiscovery struct {
	IssuerURL    string `json:"issuerUrl"`
	AuthorizeURL string `json:"authorizeUrl"`
	TokenURL     string `json:"tokenUrl"`
	UserinfoURL  string `json:"userinfoUrl"`
	JWKSURL      string `json:"jwksUrl"`
	LogoutURL    string `json:"logoutUrl,omitempty"`
}

// AuthMethods describes which sign-in options are available.
type AuthMethods struct {
	AuthEnabled       bool            `json:"authEnabled"`
	LoginLocal        bool            `json:"loginLocal"`
	LoginOIDC         bool            `json:"loginOidc"`
	OIDCButtonText    string          `json:"oidcButtonText,omitempty"`
	OIDCAutoLaunch    bool            `json:"oidcAutoLaunch"`
	AllowRegistration bool            `json:"allowRegistration"`
	PasswordPolicy    *PasswordPolicy `json:"passwordPolicy,omitempty"`
	Altcha            *AltchaPublic   `json:"altcha,omitempty"`
}

// PasswordPolicy is the browser-visible password strength policy.
type PasswordPolicy struct {
	MinLength     int  `json:"minLength"`
	LongLength    int  `json:"longLength"`
	MinKinds      int  `json:"minKinds"`
	RequireLower  bool `json:"requireLower"`
	RequireUpper  bool `json:"requireUpper"`
	RequireDigit  bool `json:"requireDigit"`
	RequireSymbol bool `json:"requireSymbol"`
}

// AltchaPublic is browser-safe ALTCHA widget configuration.
type AltchaPublic struct {
	Enabled         bool               `json:"enabled"`
	ChallengeURL    string             `json:"challengeUrl,omitempty"`
	ProtectLogin    bool               `json:"protectLogin"`
	ProtectSetup    bool               `json:"protectSetup"`
	ProtectRegister bool               `json:"protectRegister"`
	Widget          AltchaWidgetPublic `json:"widget"`
}

// AltchaWidgetPublic customizes the ALTCHA web component.
type AltchaWidgetPublic struct {
	Auto       string `json:"auto,omitempty"`
	Display    string `json:"display,omitempty"`
	HideFooter bool   `json:"hideFooter,omitempty"`
	HideLogo   bool   `json:"hideLogo,omitempty"`
	Language   string `json:"language,omitempty"`
	Name       string `json:"name,omitempty"`
	Theme      string `json:"theme,omitempty"`
	Type       string `json:"type,omitempty"`
	Workers    int    `json:"workers,omitempty"`
}
