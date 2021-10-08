package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

var (
	projectID   string
	serviceID   string
	serviceName string
	urlMapName  string
	token       string
)

func main() {
	flag.StringVar(&projectID, "project-id", "", "[Required] Set a project ID. You can find it by executing 'gcloud projects list'.")
	flag.StringVar(&serviceID, "service-id", uuidString(), "[Optinal] Set a service ID which should be unique.")
	flag.StringVar(&serviceName, "service-name", defaultServiceName(), "[Optinal] Set a service name.")
	flag.StringVar(&urlMapName, "url-map-name", "", "[Required] Set a url map name(load balancing name). You can find it by executing 'gcloud compute url-maps list'.")
	flag.StringVar(&token, "token", "", "[Required] Set an access token. You can get access token by executing 'gcloud auth print-access-token'")
	flag.Parse()

	if projectID == "" {
		log.Fatal(errors.New(`project ID is required (set -project-id)`))
	}
	if urlMapName == "" {
		log.Fatal(errors.New(`url map name is required (set -url-map-name)`))
	}
	if token == "" {
		log.Fatal(errors.New(`token is required (set -token)`))
	}

	if err := createService(); err != nil {
		log.Fatal(err)
	}
	if err := createSLO(); err != nil {
		log.Fatal(err)
	}
	fmt.Println(`Succeeded!`)
	fmt.Println(`Please visit and check the created service and SLO at https://console.cloud.google.com/monitoring/services`)
	fmt.Println(`They will appear in a few minutes`)
}

func uuidString() string {
	u, err := uuid.NewRandom()
	if err != nil {
		log.Fatal(err)
	}
	return u.String()
}

func defaultServiceName() string {
	return fmt.Sprintf("Service created by gcp-slo-setter (since %s)", time.Now().Format(`2006-01-02 15:04`))
}

func createService() error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	body, err := json.Marshal(Service{
		DisplayName: serviceName,
		Custom:      map[string]string{},
	})
	if err != nil {
		return fmt.Errorf("failed to marshal service request : %w", err)
	}

	// 内部的に HTTP/1.1 利用している
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf(
			"https://monitoring.googleapis.com/v3/projects/%s/services?service_id=%s",
			projectID,
			serviceID,
		),
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("failed to create service request : %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("got error during http request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		log.Printf("GCP API status code: %d\n", res.StatusCode)
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to read create service response from GCP : %w", err)
		}
		log.Printf("service creation error details: %s", b)
		return errors.New(`failed to create service`)
	}
	return nil
}

func createSLO() error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	body, err := json.Marshal(SLO{
		ServiceLevelIndicator: SLI{
			RequestBased: RequestBased{
				GoodTotalRatio: GoodTotalRatio{
					// Filter は OR をサポートしない点に注意
					BadServiceFilter: fmt.Sprintf(
						`metric.type="loadbalancing.googleapis.com/https/request_count"
resource.type="https_lb_rule"
resource.label."url_map_name"="%s"
metric.label."response_code_class">="500"`, urlMapName),
					GoodServiceFilter: fmt.Sprintf(
						`metric.type="loadbalancing.googleapis.com/https/request_count"
resource.type="https_lb_rule"
resource.label."url_map_name"="%s"
metric.label."response_code_class"<"300"`, urlMapName),
				},
			},
		},
		Goal:          0.99,      // 99% 固定
		RollingPeriod: "604800s", // 7週間固定
		DisplayName:   "99% Availability in Rolling 7 Days",
	})
	if err != nil {
		return fmt.Errorf("failed to marshal slo request : %w", err)
	}

	// 内部的に HTTP/1.1 利用している
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf(
			"https://monitoring.googleapis.com/v3/projects/%s/services/%s/serviceLevelObjectives",
			projectID,
			serviceID,
		),
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("failed to create slo request : %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("got error during http request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		log.Printf("GCP API status code: %d\n", res.StatusCode)
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to read create slo response from GCP : %w", err)
		}
		log.Printf("slo creation error details: %s", b)
		return errors.New(`failed to create slo`)
	}
	return nil
}
