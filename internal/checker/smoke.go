package checker

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/HXong/predeploy-guard/internal/config"
)

type SmokeResult struct {
	Name           string
	Method         string
	URL            string
	ExpectedStatus int
	ActualStatus   int
	Passed         bool
	Error          string
	Duration       time.Duration
}

func WaitUntilReady(baseURL string, healthPath string, timeoutSeconds int) error {
	deadline := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)
	url := joinURL(baseURL, healthPath)

	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			_ = resp.Body.Close()

			if resp.StatusCode >= 200 && resp.StatusCode < 500 {
				return nil
			}
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("service did not become ready within %d seconds: %s", timeoutSeconds, url)
}

func RunSmokeChecks(baseURL string, checks []config.SmokeCheck) []SmokeResult {
	results := make([]SmokeResult, 0, len(checks))

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for _, check := range checks {
		result := runSingleSmokeCheck(client, baseURL, check)
		results = append(results, result)
	}

	return results
}

func runSingleSmokeCheck(client *http.Client, baseURL string, check config.SmokeCheck) SmokeResult {
	method := strings.ToUpper(check.Method)
	url := joinURL(baseURL, check.Path)

	result := SmokeResult{
		Name:           check.Name,
		Method:         method,
		URL:            url,
		ExpectedStatus: check.ExpectedStatus,
	}

	start := time.Now()

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		result.Error = err.Error()
		result.Duration = time.Since(start)
		return result
	}

	resp, err := client.Do(req)
	if err != nil {
		result.Error = err.Error()
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	result.ActualStatus = resp.StatusCode
	result.Duration = time.Since(start)
	result.Passed = resp.StatusCode == check.ExpectedStatus

	return result
}

func AllPassed(results []SmokeResult) bool {
	for _, result := range results {
		if !result.Passed {
			return false
		}
	}

	return true
}
