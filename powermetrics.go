package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// PowermetricsData is the JSON output from `powermetrics --format json`.
type PowermetricsData struct {
	ThermalPressure string           `json:"thermal_pressure"`
	SMC             *PowerSMC        `json:"smc"`
	Processor       *PowerProcessor  `json:"processor"`
}

type PowerSMC struct {
	Fan []PowerFan `json:"fan"`
}

type PowerFan struct {
	Name string  `json:"name"`
	RPM  float64 `json:"rpm"`
}

type PowerProcessor struct {
	CpuEnergy    float64          `json:"cpu_energy"`
	GpuEnergy    float64          `json:"gpu_energy"`
	Clusters     []PowerCluster   `json:"clusters"`
}

type PowerCluster struct {
	Name         string  `json:"name"`
	Freq         float64 `json:"freq_hz"`
	Active       float64 `json:"active_ratio"`
	Dvfm         []PowerDvfm `json:"dvfm_states"`
}

type PowerDvfm struct {
	Freq       uint64  `json:"freq"`
	ActiveTime float64 `json:"active_ns"`
}

// PowermetricsResult holds parsed powermetrics output.
type PowermetricsResult struct {
	ThermalPressure string
	Fans            []PowerFan
	CPUClusters     []PowerCluster
	CPUEnergyW      float64
	GPUEnergyW      float64
}

// RunPowermetrics executes powermetrics and returns parsed data.
// Returns nil error with nil result if powermetrics is unavailable (e.g. no root).
func RunPowermetrics() (*PowermetricsResult, error) {
	cmd := exec.Command("powermetrics",
		"-s", "thermal,smc",
		"-n", "1",
		"-i", "500",
		"--format", "json",
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// powermetrics may output multiple JSON objects; take the first complete one
	raw := string(out)
	idx := strings.Index(raw, "\n{")
	if idx > 0 {
		raw = raw[idx+1:]
	}
	// Find end of first JSON object
	depth := 0
	end := -1
	for i, ch := range raw {
		if ch == '{' {
			depth++
		} else if ch == '}' {
			depth--
			if depth == 0 {
				end = i + 1
				break
			}
		}
	}
	if end > 0 {
		raw = raw[:end]
	}

	var pm PowermetricsData
	if err := json.Unmarshal([]byte(raw), &pm); err != nil {
		return nil, fmt.Errorf("failed to parse powermetrics JSON: %w", err)
	}

	result := &PowermetricsResult{
		ThermalPressure: pm.ThermalPressure,
	}
	if pm.SMC != nil {
		result.Fans = pm.SMC.Fan
	}
	if pm.Processor != nil {
		result.CPUClusters = pm.Processor.Clusters
		result.CPUEnergyW = pm.Processor.CpuEnergy
		result.GPUEnergyW = pm.Processor.GpuEnergy
	}
	return result, nil
}
