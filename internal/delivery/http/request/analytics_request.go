package request

// AnalyticsActivitiesQuery represents the optional query params for the activities endpoint.
type AnalyticsActivitiesQuery struct {
	Limit int `json:"limit"`
}

// AnalyticsLeaderboardQuery represents the optional query params for the leaderboard endpoint.
type AnalyticsLeaderboardQuery struct {
	Period string `json:"period"`
}
