package api

// authentication types
const (
	authMethodOAuth = "oauth"
	headless        = "headless"
	// TODO add new types of authentication method here...
)

// oath related data
const (
	oauth2ClientId       = "default"
	oauth2AuthUriSuffix  = "oauth2/authorize" // API for obtaining authorization codes
	oauth2TokenUriSuffix = "oauth2/token"     // API for exchanging the auth code for a token
	oauthRedirectUri     = "http://127.0.0.1:3101/callback"
	SHA256Hash           = "S256" // the SHA-256 hashing alorithm used to generate the code challenge for PKCE
)

const (
	typeAPIPath   = "/knowledge-store/v1/types/"
	objectAPIPath = "/knowledge-store/v1/objects/"
)
