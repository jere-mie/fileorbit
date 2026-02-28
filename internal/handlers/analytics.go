package handlers

import (
	"net/http"
	"time"

	"github.com/jere-mie/fileorbit/internal/models"
)

// DailyCount represents download count for a single day.
type DailyCount struct {
	Date  string `db:"date" json:"date"`
	Count int    `db:"count" json:"count"`
}

// ReferrerCount represents download count from a single referrer.
type ReferrerCount struct {
	Referrer string `db:"referrer" json:"referrer"`
	Count    int    `db:"count" json:"count"`
}

// AnalyticsData holds template data for the analytics page.
type AnalyticsData struct {
	File      models.File
	Events    []models.AnalyticsEvent
	Daily     []DailyCount
	Referrers []ReferrerCount
	BaseURL   string
}

// FileAnalytics renders the analytics page for a specific file.
func (a *App) FileAnalytics(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var file models.File
	if err := a.DB.Get(&file, fileFindQuery, id); err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Recent events
	var events []models.AnalyticsEvent
	a.DB.Select(&events,
		`SELECT id, file_id, accessed_at, referrer, user_agent, ip_address
		FROM analytics WHERE file_id = ? ORDER BY accessed_at DESC LIMIT 50`, file.ID)

	// Daily downloads for last 30 days
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	var daily []DailyCount
	a.DB.Select(&daily,
		`SELECT DATE(accessed_at) as date, COUNT(*) as count
		FROM analytics
		WHERE file_id = ? AND DATE(accessed_at) >= ?
		GROUP BY DATE(accessed_at)
		ORDER BY date ASC`, file.ID, thirtyDaysAgo)

	// Top referrers
	var referrers []ReferrerCount
	a.DB.Select(&referrers,
		`SELECT CASE WHEN referrer = '' THEN 'Direct' ELSE referrer END as referrer,
			COUNT(*) as count
		FROM analytics
		WHERE file_id = ?
		GROUP BY referrer
		ORDER BY count DESC
		LIMIT 10`, file.ID)

	a.render(w, "file_analytics", AnalyticsData{
		File:      file,
		Events:    events,
		Daily:     daily,
		Referrers: referrers,
		BaseURL:   a.Config.BaseURL,
	})
}
