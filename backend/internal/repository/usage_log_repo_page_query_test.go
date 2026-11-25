package repository

import (
	"strings"
	"testing"

	"github.com/pengbin9472/ggbond/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

func TestUsageLogPageQuerySortsNarrowPageBeforeLoadingFullRows(t *testing.T) {
	query := usageLogPageQuery("WHERE created_at >= $1", pagination.PaginationParams{
		SortBy:    "created_at",
		SortOrder: "desc",
	}, 2, 3)

	pageCTE := strings.SplitN(query, ")\n\t\tSELECT", 2)[0]
	require.Contains(t, pageCTE, "SELECT id, created_at AS sort_key")
	require.Contains(t, pageCTE, "ORDER BY sort_key DESC, id DESC")
	require.Contains(t, pageCTE, "LIMIT $2 OFFSET $3")
	require.NotContains(t, pageCTE, "user_agent")
	require.Contains(t, query, "FROM usage_logs ul\n\t\tJOIN page_ids p ON p.id = ul.id")
	require.Contains(t, query, "ORDER BY p.sort_key DESC, p.id DESC")
}

func TestUsageLogPageQueryKeepsModelSortExpressionAndDirection(t *testing.T) {
	query := usageLogPageQuery("", pagination.PaginationParams{
		SortBy:    "model",
		SortOrder: "asc",
	}, 1, 2)

	require.Contains(t, query, "COALESCE(NULLIF(TRIM(requested_model), ''), model) AS sort_key")
	require.Contains(t, query, "ORDER BY sort_key ASC, id ASC")
	require.Contains(t, query, "ORDER BY p.sort_key ASC, p.id ASC")
}
