// services/trend.go
// 计算趋势的服务函数
package services

import "math"

// CalculatePercentageTrend 计算百分比趋势
// thresholdPercent: 变化百分比阈值（如 10 表示 10%）
func CalculatePercentageTrend(current, previous *float64, hasPrevious bool, thresholdPercent float64) string {
	if !hasPrevious || current == nil || previous == nil || *previous == 0 {
		return "stable"
	}

	diff := *current - *previous
	percentChange := (diff / math.Abs(*previous)) * 100

	if percentChange > thresholdPercent {
		return "up"
	} else if percentChange < -thresholdPercent {
		return "down"
	}
	return "stable"
}

// CalculateStepTrend 步数趋势（基于百分比）
func CalculateStepTrend(current, previous int64, hasPrevious bool, thresholdPercent float64) string {
	if !hasPrevious || previous == 0 {
		return "stable"
	}

	diff := current - previous
	percentChange := (float64(diff) / float64(previous)) * 100

	if percentChange > thresholdPercent {
		return "up"
	} else if percentChange < -thresholdPercent {
		return "down"
	}
	return "stable"
}
