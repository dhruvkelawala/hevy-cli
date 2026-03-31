package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type whoopRecoverySnapshot struct {
	RecoveryScore float64 `json:"recovery_score"`
	HRVRMSSD      float64 `json:"hrv_rmssd_milli"`
	RestingHR     float64 `json:"resting_heart_rate"`
}

type whoopHistoryPoint struct {
	Day           string  `json:"day"`
	RecoveryScore float64 `json:"recovery_score"`
}

type whoopRecoveryResponse struct {
	Records []struct {
		CycleStart string `json:"cycle_start"`
		ScoreState string `json:"score_state"`
		Score      struct {
			RecoveryScore    float64 `json:"recovery_score"`
			RestingHeartRate float64 `json:"resting_heart_rate"`
			HRVRMSSD         float64 `json:"hrv_rmssd_milli"`
		} `json:"score"`
	} `json:"records"`
}

func whoopScriptPath() (string, error) {
	base := strings.TrimSpace(defaultWhoopPath())
	if app.config != nil && strings.TrimSpace(app.config.WhoopPath) != "" {
		base = strings.TrimSpace(app.config.WhoopPath)
	}
	expanded, err := expandHome(base)
	if err != nil {
		return "", err
	}
	if strings.HasSuffix(expanded, ".py") {
		return expanded, nil
	}
	return filepath.Join(expanded, "scripts", "get_recovery.py"), nil
}

func expandHome(path string) (string, error) {
	if path == "" || path[0] != '~' {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if path == "~" {
		return home, nil
	}
	return filepath.Join(home, strings.TrimPrefix(path, "~/")), nil
}

func fetchWhoopRecovery(args ...string) (*whoopRecoveryResponse, error) {
	scriptPath, err := whoopScriptPath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(scriptPath); err != nil {
		return nil, err
	}
	cmdArgs := append([]string{scriptPath}, args...)
	cmd := exec.Command("python3", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			return nil, err
		}
		return nil, fmt.Errorf("%w: %s", err, message)
	}
	var resp whoopRecoveryResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return nil, fmt.Errorf("parse WHOOP response: %w", err)
	}
	return &resp, nil
}

func parseWhoopSnapshot(resp *whoopRecoveryResponse) *whoopRecoverySnapshot {
	if resp == nil || len(resp.Records) == 0 {
		return nil
	}
	record := resp.Records[0]
	if record.ScoreState != "SCORED" && record.Score.RecoveryScore == 0 {
		return nil
	}
	return &whoopRecoverySnapshot{
		RecoveryScore: record.Score.RecoveryScore,
		HRVRMSSD:      record.Score.HRVRMSSD,
		RestingHR:     record.Score.RestingHeartRate,
	}
}

func parseWhoopHistory(resp *whoopRecoveryResponse) []whoopHistoryPoint {
	if resp == nil {
		return nil
	}
	history := make([]whoopHistoryPoint, 0, len(resp.Records))
	for _, record := range resp.Records {
		label := "-"
		if startedAt, err := time.Parse(time.RFC3339, record.CycleStart); err == nil {
			label = startedAt.Local().Format("Mon")
		}
		history = append(history, whoopHistoryPoint{Day: label, RecoveryScore: record.Score.RecoveryScore})
	}
	return history
}

func whoopStatus(score float64) (string, string) {
	switch {
	case score >= 67:
		return "GREEN", "Full send. Heavy compounds OK."
	case score >= 34:
		return "YELLOW", "Moderate session. Reduce volume 20%, skip heavy singles."
	default:
		return "RED", "Active recovery only. Light cardio or mobility."
	}
}
