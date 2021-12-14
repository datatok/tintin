package utils

type Filter struct {
	Schedule, Team, Pipelines string
	Status                    []string
}

type ExecutionTimeline struct {
	Start, End string
	Duration   int
}
