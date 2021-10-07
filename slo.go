package main

type SLO struct {
	ServiceLevelIndicator SLI     `json:"serviceLevelIndicator"`
	Goal                  float64 `json:"goal"`
	RollingPeriod         string  `json:"rollingPeriod"`
	DisplayName           string  `json:"displayName"`
}

type SLI struct {
	RequestBased RequestBased `json:"requestBased"`
}

type RequestBased struct {
	GoodTotalRatio GoodTotalRatio `json:"goodTotalRatio"`
}

type GoodTotalRatio struct {
	GoodServiceFilter string `json:"goodServiceFilter"`
	BadServiceFilter  string `json:"badServiceFilter"`
}
