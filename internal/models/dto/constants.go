package dto

type InteractionsResultState string

const (
	Ignore              = "ignore"
	IncludeInteractions = "include_interactions"
	IncludeAsSummary    = "include_as_summary"
)