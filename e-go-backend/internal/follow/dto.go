package follow

// StatusResponse is the caller's follow state for a store.
type StatusResponse struct {
	Following     bool  `json:"following"`
	FollowerCount int64 `json:"follower_count"`
}
