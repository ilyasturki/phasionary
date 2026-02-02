package app

import (
	"fmt"
	"time"
)

func FormatDate(timestamp string) string {
	if timestamp == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return timestamp
	}
	return t.Local().Format("Jan 2, 2006 at 3:04 PM")
}

func FormatRelativeTime(timestamp string) string {
	if timestamp == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return ""
	}
	now := time.Now()
	diff := now.Sub(t)

	if diff < 0 {
		diff = -diff
		return formatFutureDuration(diff)
	}

	return formatPastDuration(diff)
}

func formatPastDuration(diff time.Duration) string {
	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case diff < 365*24*time.Hour:
		months := int(diff.Hours() / (24 * 30))
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(diff.Hours() / (24 * 365))
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

func formatFutureDuration(diff time.Duration) string {
	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "in 1 minute"
		}
		return fmt.Sprintf("in %d minutes", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "in 1 hour"
		}
		return fmt.Sprintf("in %d hours", hours)
	default:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "in 1 day"
		}
		return fmt.Sprintf("in %d days", days)
	}
}

func FormatDateWithRelative(timestamp string) string {
	if timestamp == "" {
		return ""
	}
	date := FormatDate(timestamp)
	relative := FormatRelativeTime(timestamp)
	if relative == "" {
		return date
	}
	return fmt.Sprintf("%s (%s)", date, relative)
}

func FormatEstimate(minutes int) string {
	if minutes == 0 {
		return ""
	}
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	hours := minutes / 60
	if hours < 8 {
		return fmt.Sprintf("%dh", hours)
	}
	days := hours / 8
	return fmt.Sprintf("%dd", days)
}

func FormatEstimateLabel(minutes int) string {
	if minutes == 0 {
		return "None"
	}
	if minutes < 60 {
		if minutes == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", minutes)
	}
	hours := minutes / 60
	if hours < 8 {
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}
	days := hours / 8
	if days == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", days)
}
