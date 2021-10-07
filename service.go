package main

type Service struct {
	DisplayName string            `json:"displayName"`
	Custom      map[string]string `json:"custom"`
}
