package validate

import (
	"net/http"
	"time"
)

// CheckReachable は base_url への到達性を軽量に確認する（HEAD）。
// 空文字の場合は false/nil を返す。
func CheckReachable(baseURL string) (bool, error) {
	if baseURL == "" {
		return false, nil
	}
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Head(baseURL)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return true, nil
}
