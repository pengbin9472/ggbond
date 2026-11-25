package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pengbin9472/ggbond/internal/pkg/usagestats"
	"github.com/stretchr/testify/require"
)

func TestGetUsageAnalyticsWithFiltersUsesOneGroupedAggregate(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &usageLogRepository{sql: db}
	start := time.Date(2026, 11, 25, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	filters := usagestats.UsageLogFilters{StartTime: &start, EndTime: &end}

	columns := []string{
		"inbound_grouped", "upstream_grouped", "model_grouped", "group_grouped",
		"inbound_endpoint", "upstream_endpoint", "model_key", "group_key", "group_name",
		"requests", "input_tokens", "output_tokens", "cache_creation_tokens", "cache_read_tokens",
		"total_tokens", "cost", "actual_cost", "account_cost", "avg_duration_ms",
	}
	rows := sqlmock.NewRows(columns).
		AddRow(1, 1, 1, 1, nil, nil, nil, nil, "", 3, 10, 20, 4, 5, 39, 7.0, 8.0, 6.0, 100.0).
		AddRow(0, 1, 1, 1, "openai", nil, nil, nil, "", 3, 10, 20, 4, 5, 39, 7.0, 8.0, 6.0, 100.0).
		AddRow(1, 0, 1, 1, nil, "responses", nil, nil, "", 3, 10, 20, 4, 5, 39, 7.0, 8.0, 6.0, 100.0).
		AddRow(0, 0, 1, 1, "openai", "responses", nil, nil, "", 3, 10, 20, 4, 5, 39, 7.0, 8.0, 6.0, 100.0).
		AddRow(1, 1, 0, 1, nil, nil, "gpt-5", nil, "", 3, 10, 20, 4, 5, 39, 7.0, 8.0, 6.0, 100.0).
		AddRow(1, 1, 1, 0, nil, nil, nil, 4, "test-group", 3, 10, 20, 4, 5, 39, 7.0, 8.0, 6.0, 100.0)
	mock.ExpectQuery("GROUP BY GROUPING SETS").WithArgs(start, end).WillReturnRows(rows)

	stats, err := repo.GetUsageAnalyticsWithFilters(context.Background(), filters)
	require.NoError(t, err)
	require.Equal(t, int64(3), stats.TotalRequests)
	require.Equal(t, int64(39), stats.TotalTokens)
	require.Equal(t, 7.0, stats.TotalCost)
	require.Len(t, stats.Endpoints, 1)
	require.Equal(t, "openai", stats.Endpoints[0].Endpoint)
	require.Len(t, stats.UpstreamEndpoints, 1)
	require.Equal(t, "responses", stats.UpstreamEndpoints[0].Endpoint)
	require.Len(t, stats.EndpointPaths, 1)
	require.Equal(t, "openai -> responses", stats.EndpointPaths[0].Endpoint)
	require.Len(t, stats.Models, 1)
	require.Equal(t, "gpt-5", stats.Models[0].Model)
	require.Equal(t, int64(39), stats.Models[0].TotalTokens)
	require.Len(t, stats.Groups, 1)
	require.Equal(t, int64(4), stats.Groups[0].GroupID)
	require.Equal(t, "test-group", stats.Groups[0].GroupName)
	require.NoError(t, mock.ExpectationsWereMet())
}
