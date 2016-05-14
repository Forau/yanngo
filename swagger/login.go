package swagger

type Login struct {
	Environment string `json:"environment,omitempty"`
	SessionKey  string `json:"session_key,omitempty"`
	ExpiresIn   int64  `json:"expires_in,omitempty"`
	PrivateFeed Feed   `json:"private_feed,omitempty"`
	PublicFeed  Feed   `json:"public_feed,omitempty"`
}
