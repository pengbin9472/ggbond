package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	dbent "github.com/pengbin9472/ggbond/ent"
	"github.com/pengbin9472/ggbond/ent/apikey"
	"github.com/pengbin9472/ggbond/ent/group"
	"github.com/pengbin9472/ggbond/internal/pkg/logger"
	"github.com/pengbin9472/ggbond/internal/pkg/pagination"
	"github.com/pengbin9472/ggbond/internal/service"
	"github.com/lib/pq"
)

type sqlExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type groupRepository struct {
	client *dbent.Client
	sql    sqlExecutor
}

func NewGroupRepository(client *dbent.Client, sqlDB *sql.DB) service.GroupRepository {
	return newGroupRepositoryWithSQL(client, sqlDB)
}

func newGroupRepositoryWithSQL(client *dbent.Client, sqlq sqlExecutor) *groupRepository {
	return &groupRepository{client: client, sql: sqlq}
}

func (r *groupRepository) Create(ctx context.Context, groupIn *service.Group) error {
	builder := r.client.Group.Create().
		SetName(groupIn.Name).
		SetDescription(groupIn.Description).
		SetPlatform(groupIn.Platform).
		SetRateMultiplier(groupIn.RateMultiplier).
		SetIsExclusive(groupIn.IsExclusive).
		SetStatus(groupIn.Status).
		SetSubscriptionType(groupIn.SubscriptionType).
		SetNillableDailyLimitUsd(groupIn.DailyLimitUSD).
		SetNillableWeeklyLimitUsd(groupIn.WeeklyLimitUSD).
		SetNillableMonthlyLimitUsd(groupIn.MonthlyLimitUSD).
		SetNillableImagePrice1k(groupIn.ImagePrice1K).
		SetNillableImagePrice2k(groupIn.ImagePrice2K).
		SetNillableImagePrice4k(groupIn.ImagePrice4K).
		SetNillableSoraImagePrice360(groupIn.SoraImagePrice360).
		SetNillableSoraImagePrice540(groupIn.SoraImagePrice540).
		SetNillableSoraVideoPricePerRequest(groupIn.SoraVideoPricePerRequest).
		SetNillableSoraVideoPricePerRequestHd(groupIn.SoraVideoPricePerRequestHD).
		SetDefaultValidityDays(groupIn.DefaultValidityDays).
		SetClaudeCodeOnly(groupIn.ClaudeCodeOnly).
		SetNillableFallbackGroupID(groupIn.FallbackGroupID).
		SetNillableFallbackGroupIDOnInvalidRequest(groupIn.FallbackGroupIDOnInvalidRequest).
		SetModelRoutingEnabled(groupIn.ModelRoutingEnabled).
		SetMcpXMLInject(groupIn.MCPXMLInject).
		SetSoraStorageQuotaBytes(groupIn.SoraStorageQuotaBytes)

	// 设置模型路由配置
	if groupIn.ModelRouting != nil {
		builder = builder.SetModelRouting(groupIn.ModelRouting)
	}

	// 设置支持的模型系列（始终设置，空数组表示不限制）
	builder = builder.SetSupportedModelScopes(groupIn.SupportedModelScopes)

	created, err := builder.Save(ctx)
	if err == nil {
		groupIn.ID = created.ID
		groupIn.CreatedAt = created.CreatedAt
		groupIn.UpdatedAt = created.UpdatedAt
		if err := enqueueSchedulerOutbox(ctx, r.sql, service.SchedulerOutboxEventGroupChanged, nil, &groupIn.ID, nil); err != nil {
			logger.LegacyPrintf("repository.group", "[SchedulerOutbox] enqueue group create failed: group=%d err=%v", groupIn.ID, err)
		}
	}
	return translatePersistenceError(err, nil, service.ErrGroupExists)
}

func (r *groupRepository) GetByID(ctx context.Context, id int64) (*service.Group, error) {
	out, err := r.GetByIDLite(ctx, id)
	if err != nil {
		return nil, err
	}
	count, _ := r.GetAccountCount(ctx, out.ID)
	out.AccountCount = count
	return out, nil
}

func (r *groupRepository) GetByIDLite(ctx context.Context, id int64) (*service.Group, error) {
	// AccountCount is intentionally not loaded here; use GetByID when needed.
	m, err := r.client.Group.Query().
		Where(group.IDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrGroupNotFound, nil)
	}
	return groupEntityToService(m), nil
}

func (r *groupRepository) Update(ctx context.Context, groupIn *service.Group) error {
	builder := r.client.Group.UpdateOneID(groupIn.ID).
		SetName(groupIn.Name).
		SetDescription(groupIn.Description).
		SetPlatform(groupIn.Platform).
		SetRateMultiplier(groupIn.RateMultiplier).
		SetIsExclusive(groupIn.IsExclusive).
		SetStatus(groupIn.Status).
		SetSubscriptionType(groupIn.SubscriptionType).
		SetNillableDailyLimitUsd(groupIn.DailyLimitUSD).
		SetNillableWeeklyLimitUsd(groupIn.WeeklyLimitUSD).
		SetNillableMonthlyLimitUsd(groupIn.MonthlyLimitUSD).
		SetNillableImagePrice1k(groupIn.ImagePrice1K).
		SetNillableImagePrice2k(groupIn.ImagePrice2K).
		SetNillableImagePrice4k(groupIn.ImagePrice4K).
		SetNillableSoraImagePrice360(groupIn.SoraImagePrice360).
		SetNillableSoraImagePrice540(groupIn.SoraImagePrice540).
		SetNillableSoraVideoPricePerRequest(groupIn.SoraVideoPricePerRequest).
		SetNillableSoraVideoPricePerRequestHd(groupIn.SoraVideoPricePerRequestHD).
		SetDefaultValidityDays(groupIn.DefaultValidityDays).
		SetClaudeCodeOnly(groupIn.ClaudeCodeOnly).
		SetModelRoutingEnabled(groupIn.ModelRoutingEnabled).
		SetMcpXMLInject(groupIn.MCPXMLInject).
		SetSoraStorageQuotaBytes(groupIn.SoraStorageQuotaBytes)

	// 显式处理可空字段：nil 需要 clear，非 nil 需要 set。
	if groupIn.DailyLimitUSD != nil {
		builder = builder.SetDailyLimitUsd(*groupIn.DailyLimitUSD)
	} else {
		builder = builder.ClearDailyLimitUsd()
	}
	if groupIn.WeeklyLimitUSD != nil {
		builder = builder.SetWeeklyLimitUsd(*groupIn.WeeklyLimitUSD)
	} else {
		builder = builder.ClearWeeklyLimitUsd()
	}
	if groupIn.MonthlyLimitUSD != nil {
		builder = builder.SetMonthlyLimitUsd(*groupIn.MonthlyLimitUSD)
	} else {
		builder = builder.ClearMonthlyLimitUsd()
	}
	if groupIn.ImagePrice1K != nil {
		builder = builder.SetImagePrice1k(*groupIn.ImagePrice1K)
	} else {
		builder = builder.ClearImagePrice1k()
	}
	if groupIn.ImagePrice2K != nil {
		builder = builder.SetImagePrice2k(*groupIn.ImagePrice2K)
	} else {
		builder = builder.ClearImagePrice2k()
	}
	if groupIn.ImagePrice4K != nil {
		builder = builder.SetImagePrice4k(*groupIn.ImagePrice4K)
	} else {
		builder = builder.ClearImagePrice4k()
	}

	// 处理 FallbackGroupID：nil 时清除，否则设置
	if groupIn.FallbackGroupID != nil {
		builder = builder.SetFallbackGroupID(*groupIn.FallbackGroupID)
	} else {
		builder = builder.ClearFallbackGroupID()
	}
	// 处理 FallbackGroupIDOnInvalidRequest：nil 时清除，否则设置
	if groupIn.FallbackGroupIDOnInvalidRequest != nil {
		builder = builder.SetFallbackGroupIDOnInvalidRequest(*groupIn.FallbackGroupIDOnInvalidRequest)
	} else {
		builder = builder.ClearFallbackGroupIDOnInvalidRequest()
	}

	// 处理 ModelRouting：nil 时清除，否则设置
	if groupIn.ModelRouting != nil {
		builder = builder.SetModelRouting(groupIn.ModelRouting)
	} else {
		builder = builder.ClearModelRouting()
	}

	// 处理 SupportedModelScopes（始终设置，空数组表示不限制）
	builder = builder.SetSupportedModelScopes(groupIn.SupportedModelScopes)

	updated, err := builder.Save(ctx)
	if err != nil {
		return translatePersistenceError(err, service.ErrGroupNotFound, service.ErrGroupExists)
	}
	groupIn.UpdatedAt = updated.UpdatedAt
	if err := enqueueSchedulerOutbox(ctx, r.sql, service.SchedulerOutboxEventGroupChanged, nil, &groupIn.ID, nil); err != nil {
		logger.LegacyPrintf("repository.group", "[SchedulerOutbox] enqueue group update failed: group=%d err=%v", groupIn.ID, err)
	}
	return nil
}

func (r *groupRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.client.Group.Delete().Where(group.IDEQ(id)).Exec(ctx)
	if err != nil {
		return translatePersistenceError(err, service.ErrGroupNotFound, nil)
	}
	if err := enqueueSchedulerOutbox(ctx, r.sql, service.SchedulerOutboxEventGroupChanged, nil, &id, nil); err != nil {
		logger.LegacyPrintf("repository.group", "[SchedulerOutbox] enqueue group delete failed: group=%d err=%v", id, err)
	}
	return nil
}

func (r *groupRepository) List(ctx context.Context, params pagination.PaginationParams) ([]service.Group, *pagination.PaginationResult, error) {
	return r.ListWithFilters(ctx, params, "", "", "", nil)
}

func (r *groupRepository) ListWithFilters(ctx context.Context, params pagination.PaginationParams, platform, status, search string, isExclusive *bool) ([]service.Group, *pagination.PaginationResult, error) {
	q := r.client.Group.Query()

	if platform != "" {
		q = q.Where(group.PlatformEQ(platform))
	}
	if status != "" {
		q = q.Where(group.StatusEQ(status))
	}
	if search != "" {
		q = q.Where(group.Or(
			group.NameContainsFold(search),
			group.DescriptionContainsFold(search),
		))
	}
	if isExclusive != nil {
		q = q.Where(group.IsExclusiveEQ(*isExclusive))
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	groups, err := q.
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(dbent.Asc(group.FieldSortOrder), dbent.Asc(group.FieldID)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	groupIDs := make([]int64, 0, len(groups))
	outGroups := make([]service.Group, 0, len(groups))
	for i := range groups {
		g := groupEntityToService(groups[i])
		outGroups = append(outGroups, *g)
		groupIDs = append(groupIDs, g.ID)
	}

	counts, err := r.loadAccountCounts(ctx, groupIDs)
	if err == nil {
		for i := range outGroups {
			outGroups[i].AccountCount = counts[outGroups[i].ID]
		}
	}

	return outGroups, paginationResultFromTotal(int64(total), params), nil
}

func (r *groupRepository) ListActive(ctx context.Context) ([]service.Group, error) {
	groups, err := r.client.Group.Query().
		Where(group.StatusEQ(service.StatusActive)).
		Order(dbent.Asc(group.FieldSortOrder), dbent.Asc(group.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	groupIDs := make([]int64, 0, len(groups))
	outGroups := make([]service.Group, 0, len(groups))
	for i := range groups {
		g := groupEntityToService(groups[i])
		outGroups = append(outGroups, *g)
		groupIDs = append(groupIDs, g.ID)
	}

	counts, err := r.loadAccountCounts(ctx, groupIDs)
	if err == nil {
		for i := range outGroups {
			outGroups[i].AccountCount = counts[outGroups[i].ID]
		}
	}

	return outGroups, nil
}

func (r *groupRepository) ListActiveByPlatform(ctx context.Context, platform string) ([]service.Group, error) {
	groups, err := r.client.Group.Query().
		Where(group.StatusEQ(service.StatusActive), group.PlatformEQ(platform)).
		Order(dbent.Asc(group.FieldSortOrder), dbent.Asc(group.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	groupIDs := make([]int64, 0, len(groups))
	outGroups := make([]service.Group, 0, len(groups))
	for i := range groups {
		g := groupEntityToService(groups[i])
		outGroups = append(outGroups, *g)
		groupIDs = append(groupIDs, g.ID)
	}

	counts, err := r.loadAccountCounts(ctx, groupIDs)
	if err == nil {
		for i := range outGroups {
			outGroups[i].AccountCount = counts[outGroups[i].ID]
		}
	}

	return outGroups, nil
}

func (r *groupRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	return r.client.Group.Query().Where(group.NameEQ(name)).Exist(ctx)
}

// ExistsByIDs 批量检查分组是否存在（仅检查未软删除记录）。
// 返回结构：map[groupID]exists。
func (r *groupRepository) ExistsByIDs(ctx context.Context, ids []int64) (map[int64]bool, error) {
	result := make(map[int64]bool, len(ids))
	if len(ids) == 0 {
		return result, nil
	}

	uniqueIDs := make([]int64, 0, len(ids))
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
		result[id] = false
	}
	if len(uniqueIDs) == 0 {
		return result, nil
	}

	rows, err := r.sql.QueryContext(ctx, `
		SELECT id
		FROM groups
		WHERE id = ANY($1) AND deleted_at IS NULL
	`, pq.Array(uniqueIDs))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result[id] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *groupRepository) GetAccountCount(ctx context.Context, groupID int64) (int64, error) {
	var count int64
	if err := scanSingleRow(ctx, r.sql, "SELECT COUNT(*) FROM account_groups WHERE group_id = $1", []any{groupID}, &count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *groupRepository) DeleteAccountGroupsByGroupID(ctx context.Context, groupID int64) (int64, error) {
	res, err := r.sql.ExecContext(ctx, "DELETE FROM account_groups WHERE group_id = $1", groupID)
	if err != nil {
		return 0, err
	}
	affected, _ := res.RowsAffected()
	if err := enqueueSchedulerOutbox(ctx, r.sql, service.SchedulerOutboxEventGroupChanged, nil, &groupID, nil); err != nil {
		logger.LegacyPrintf("repository.group", "[SchedulerOutbox] enqueue group account clear failed: group=%d err=%v", groupID, err)
	}
	return affected, nil
}

func (r *groupRepository) DeleteCascade(ctx context.Context, id int64) ([]int64, error) {
	g, err := r.client.Group.Query().Where(group.IDEQ(id)).Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrGroupNotFound, nil)
	}
	groupSvc := groupEntityToService(g)

	// 使用 ent 事务统一包裹：避免手工基于 *sql.Tx 构造 ent client 带来的驱动断言问题，
	// 同时保证级联删除的原子性。
	tx, err := r.client.Tx(ctx)
	if err != nil && !errors.Is(err, dbent.ErrTxStarted) {
		return nil, err
	}
	exec := r.client
	txClient := r.client
	if err == nil {
		defer func() { _ = tx.Rollback() }()
		exec = tx.Client()
		txClient = exec
	}
	// err 为 dbent.ErrTxStarted 时，复用当前 client 参与同一事务。

	// Lock the group row to avoid concurrent writes while we cascade.
	// 这里使用 exec.QueryContext 手动扫描，确保同一事务内加锁并能区分"未找到"与其他错误。
	rows, err := exec.QueryContext(ctx, "SELECT id FROM groups WHERE id = $1 AND deleted_at IS NULL FOR UPDATE", id)
	if err != nil {
		return nil, err
	}
	var lockedID int64
	if rows.Next() {
		if err := rows.Scan(&lockedID); err != nil {
			_ = rows.Close()
			return nil, err
		}
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if lockedID == 0 {
		return nil, service.ErrGroupNotFound
	}

	var affectedUserIDs []int64
	if groupSvc.IsSubscriptionType() {
		// 只查询未软删除的订阅，避免通知已取消订阅的用户
		rows, err := exec.QueryContext(ctx, "SELECT user_id FROM user_subscriptions WHERE group_id = $1 AND deleted_at IS NULL", id)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var userID int64
			if scanErr := rows.Scan(&userID); scanErr != nil {
				_ = rows.Close()
				return nil, scanErr
			}
			affectedUserIDs = append(affectedUserIDs, userID)
		}
		if err := rows.Close(); err != nil {
			return nil, err
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}

		// 软删除订阅：设置 deleted_at 而非硬删除
		if _, err := exec.ExecContext(ctx, "UPDATE user_subscriptions SET deleted_at = NOW() WHERE group_id = $1 AND deleted_at IS NULL", id); err != nil {
			return nil, err
		}
	}

	// 2. Clear group_id for api keys bound to this group.
	// 仅更新未软删除的记录，避免修改已删除数据，保证审计与历史回溯一致性。
	// 与 APIKeyRepository 的软删除语义保持一致，减少跨模块行为差异。
	if _, err := txClient.APIKey.Update().
		Where(apikey.GroupIDEQ(id), apikey.DeletedAtIsNil()).
		ClearGroupID().
		Save(ctx); err != nil {
		return nil, err
	}

	// 3. Remove the group id from user_allowed_groups join table.
	// Legacy users.allowed_groups 列已弃用，不再同步。
	if _, err := exec.ExecContext(ctx, "DELETE FROM user_allowed_groups WHERE group_id = $1", id); err != nil {
		return nil, err
	}

	// 4. Delete account_groups join rows.
	if _, err := exec.ExecContext(ctx, "DELETE FROM account_groups WHERE group_id = $1", id); err != nil {
		return nil, err
	}

	// 5. Soft-delete group itself.
	if _, err := txClient.Group.Delete().Where(group.IDEQ(id)).Exec(ctx); err != nil {
		return nil, err
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	}
	if err := enqueueSchedulerOutbox(ctx, r.sql, service.SchedulerOutboxEventGroupChanged, nil, &id, nil); err != nil {
		logger.LegacyPrintf("repository.group", "[SchedulerOutbox] enqueue group cascade delete failed: group=%d err=%v", id, err)
	}

	return affectedUserIDs, nil
}

func (r *groupRepository) loadAccountCounts(ctx context.Context, groupIDs []int64) (counts map[int64]int64, err error) {
	counts = make(map[int64]int64, len(groupIDs))
	if len(groupIDs) == 0 {
		return counts, nil
	}

	rows, err := r.sql.QueryContext(
		ctx,
		"SELECT group_id, COUNT(*) FROM account_groups WHERE group_id = ANY($1) GROUP BY group_id",
		pq.Array(groupIDs),
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
			counts = nil
		}
	}()

	for rows.Next() {
		var groupID int64
		var count int64
		if err = rows.Scan(&groupID, &count); err != nil {
			return nil, err
		}
		counts[groupID] = count
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return counts, nil
}

// GetAccountIDsByGroupIDs 获取多个分组的所有账号 ID（去重）
func (r *groupRepository) GetAccountIDsByGroupIDs(ctx context.Context, groupIDs []int64) ([]int64, error) {
	if len(groupIDs) == 0 {
		return nil, nil
	}

	rows, err := r.sql.QueryContext(
		ctx,
		"SELECT DISTINCT account_id FROM account_groups WHERE group_id = ANY($1) ORDER BY account_id",
		pq.Array(groupIDs),
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var accountIDs []int64
	for rows.Next() {
		var accountID int64
		if err := rows.Scan(&accountID); err != nil {
			return nil, err
		}
		accountIDs = append(accountIDs, accountID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return accountIDs, nil
}

// BindAccountsToGroup 将多个账号绑定到指定分组（批量插入，忽略已存在的绑定）
func (r *groupRepository) BindAccountsToGroup(ctx context.Context, groupID int64, accountIDs []int64) error {
	if len(accountIDs) == 0 {
		return nil
	}

	// 使用 INSERT ... ON CONFLICT DO NOTHING 忽略已存在的绑定
	_, err := r.sql.ExecContext(
		ctx,
		`INSERT INTO account_groups (account_id, group_id, priority, created_at)
		 SELECT unnest($1::bigint[]), $2, 50, NOW()
		 ON CONFLICT (account_id, group_id) DO NOTHING`,
		pq.Array(accountIDs),
		groupID,
	)
	if err != nil {
		return err
	}

	// 发送调度器事件
	if err := enqueueSchedulerOutbox(ctx, r.sql, service.SchedulerOutboxEventGroupChanged, nil, &groupID, nil); err != nil {
		logger.LegacyPrintf("repository.group", "[SchedulerOutbox] enqueue bind accounts to group failed: group=%d err=%v", groupID, err)
	}

	return nil
}

// UpdateSortOrders 批量更新分组排序
func (r *groupRepository) UpdateSortOrders(ctx context.Context, updates []service.GroupSortOrderUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	// 去重后保留最后一次排序值，避免重复 ID 造成 CASE 分支冲突。
	sortOrderByID := make(map[int64]int, len(updates))
	groupIDs := make([]int64, 0, len(updates))
	for _, u := range updates {
		if u.ID <= 0 {
			continue
		}
		if _, exists := sortOrderByID[u.ID]; !exists {
			groupIDs = append(groupIDs, u.ID)
		}
		sortOrderByID[u.ID] = u.SortOrder
	}
	if len(groupIDs) == 0 {
		return nil
	}

	// 与旧实现保持一致：任何不存在/已删除的分组都返回 not found，且不执行更新。
	var existingCount int
	if err := scanSingleRow(
		ctx,
		r.sql,
		`SELECT COUNT(*) FROM groups WHERE deleted_at IS NULL AND id = ANY($1)`,
		[]any{pq.Array(groupIDs)},
		&existingCount,
	); err != nil {
		return err
	}
	if existingCount != len(groupIDs) {
		return service.ErrGroupNotFound
	}

	args := make([]any, 0, len(groupIDs)*2+1)
	caseClauses := make([]string, 0, len(groupIDs))
	placeholder := 1
	for _, id := range groupIDs {
		caseClauses = append(caseClauses, fmt.Sprintf("WHEN $%d THEN $%d", placeholder, placeholder+1))
		args = append(args, id, sortOrderByID[id])
		placeholder += 2
	}
	args = append(args, pq.Array(groupIDs))

	query := fmt.Sprintf(`
		UPDATE groups
		SET sort_order = CASE id
			%s
			ELSE sort_order
		END
		WHERE deleted_at IS NULL AND id = ANY($%d)
	`, strings.Join(caseClauses, "\n\t\t\t"), placeholder)

	result, err := r.sql.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != int64(len(groupIDs)) {
		return service.ErrGroupNotFound
	}

	for _, id := range groupIDs {
		if err := enqueueSchedulerOutbox(ctx, r.sql, service.SchedulerOutboxEventGroupChanged, nil, &id, nil); err != nil {
			logger.LegacyPrintf("repository.group", "[SchedulerOutbox] enqueue group sort update failed: group=%d err=%v", id, err)
		}
	}
	return nil
}

// GetGroupMonitoringStats 获取所有分组的账户监控统计
func (r *groupRepository) GetGroupMonitoringStats(ctx context.Context) ([]service.GroupMonitoringStat, error) {
	query := `
		SELECT
			g.id,
			g.name,
			g.platform,
			g.rate_multiplier,
			g.sort_order,
			g.is_exclusive,
			g.subscription_type,
			COALESCE(gms.total_accounts, 0) as total_accounts,
			COALESCE(gms.normal_accounts, 0) as normal_accounts,
			COALESCE(gms.error_accounts, 0) as error_accounts,
			COALESCE(gms.ratelimit_accounts, 0) as ratelimit_accounts,
			COALESCE(gms.overload_accounts, 0) as overload_accounts,
			COALESCE(gms.disabled_accounts, 0) as disabled_accounts,
			COALESCE(gms.availability_rate, -1) as availability_rate,
			COALESCE(gms.cache_hit_rate, -1) as cache_hit_rate,
			COALESCE(gms.avg_response_time, 0) as avg_response_time
		FROM groups g
		LEFT JOIN group_monitoring_stats gms ON g.id = gms.group_id
		WHERE g.deleted_at IS NULL AND g.status = 'active'
		ORDER BY g.sort_order ASC, g.id ASC
	`

	rows, err := r.sql.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var stats []service.GroupMonitoringStat
	for rows.Next() {
		var stat service.GroupMonitoringStat
		err := rows.Scan(
			&stat.GroupID,
			&stat.GroupName,
			&stat.Platform,
			&stat.RateMultiplier,
			&stat.SortOrder,
			&stat.IsExclusive,
			&stat.SubscriptionType,
			&stat.TotalAccounts,
			&stat.NormalAccounts,
			&stat.ErrorAccounts,
			&stat.RateLimitAccounts,
			&stat.OverloadAccounts,
			&stat.DisabledAccounts,
			&stat.AvailabilityRate,
			&stat.CacheHitRate,
			&stat.AvgResponseTime,
		)
		if err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}

// UpsertGroupMonitoringStats 更新或插入分组监控统计
func (r *groupRepository) UpsertGroupMonitoringStats(ctx context.Context, stats []service.GroupMonitoringStat) error {
	if len(stats) == 0 {
		return nil
	}

	query := `
		INSERT INTO group_monitoring_stats (
			group_id, total_accounts, normal_accounts, error_accounts,
			ratelimit_accounts, overload_accounts, disabled_accounts,
			availability_rate, cache_hit_rate, avg_response_time, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		ON CONFLICT (group_id)
		DO UPDATE SET
			total_accounts = EXCLUDED.total_accounts,
			normal_accounts = EXCLUDED.normal_accounts,
			error_accounts = EXCLUDED.error_accounts,
			ratelimit_accounts = EXCLUDED.ratelimit_accounts,
			overload_accounts = EXCLUDED.overload_accounts,
			disabled_accounts = EXCLUDED.disabled_accounts,
			availability_rate = EXCLUDED.availability_rate,
			cache_hit_rate = EXCLUDED.cache_hit_rate,
			avg_response_time = EXCLUDED.avg_response_time,
			updated_at = NOW()
	`

	for _, stat := range stats {
		_, err := r.sql.ExecContext(ctx, query,
			stat.GroupID,
			stat.TotalAccounts,
			stat.NormalAccounts,
			stat.ErrorAccounts,
			stat.RateLimitAccounts,
			stat.OverloadAccounts,
			stat.DisabledAccounts,
			stat.AvailabilityRate,
			stat.CacheHitRate,
			stat.AvgResponseTime,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// InsertGroupMonitoringHistory 插入分组监控历史记录
func (r *groupRepository) InsertGroupMonitoringHistory(ctx context.Context, stats []service.GroupMonitoringStat) error {
	if len(stats) == 0 {
		return nil
	}

	query := `
		INSERT INTO group_monitoring_history (
			group_id, total_accounts, normal_accounts, error_accounts,
			ratelimit_accounts, overload_accounts, availability_rate, cache_hit_rate, recorded_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, date_trunc_immutable('minute', NOW()))
		ON CONFLICT (group_id, date_trunc_immutable('minute', recorded_at)) DO UPDATE SET
			total_accounts = EXCLUDED.total_accounts,
			normal_accounts = EXCLUDED.normal_accounts,
			error_accounts = EXCLUDED.error_accounts,
			ratelimit_accounts = EXCLUDED.ratelimit_accounts,
			overload_accounts = EXCLUDED.overload_accounts,
			availability_rate = EXCLUDED.availability_rate,
			cache_hit_rate = EXCLUDED.cache_hit_rate
	`

	for _, stat := range stats {
		_, err := r.sql.ExecContext(ctx, query,
			stat.GroupID,
			stat.TotalAccounts,
			stat.NormalAccounts,
			stat.ErrorAccounts,
			stat.RateLimitAccounts,
			stat.OverloadAccounts,
			stat.AvailabilityRate,
			stat.CacheHitRate,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetGroupMonitoringHistory 获取分组监控历史数据
func (r *groupRepository) GetGroupMonitoringHistory(ctx context.Context, groupID int64, limit int) ([]service.MonitoringHistoryPoint, error) {
	if limit <= 0 {
		limit = 100
	}

	// 先按时间倒序取最新的 N 条，然后在应用层反转为升序（用于图表展示）
	query := `
		SELECT
			EXTRACT(EPOCH FROM recorded_at)::bigint as recorded_at,
			availability_rate,
			cache_hit_rate
		FROM group_monitoring_history
		WHERE group_id = $1
		ORDER BY recorded_at DESC
		LIMIT $2
	`

	rows, err := r.sql.QueryContext(ctx, query, groupID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var points []service.MonitoringHistoryPoint
	for rows.Next() {
		var point service.MonitoringHistoryPoint
		if err := rows.Scan(&point.RecordedAt, &point.AvailabilityRate, &point.CacheHitRate); err != nil {
			return nil, err
		}
		points = append(points, point)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// 反转为升序（最旧到最新），用于图表从左到右展示
	for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
		points[i], points[j] = points[j], points[i]
	}

	return points, nil
}

// ComputeGroupMonitoringStats 从 accounts 表和 usage_logs 表实时计算分组监控统计
func (r *groupRepository) ComputeGroupMonitoringStats(ctx context.Context) ([]service.GroupMonitoringStat, error) {
	// 第一步：查询账户状态
	accountQuery := `
		SELECT
			g.id,
			g.name,
			g.platform,
			g.rate_multiplier,
			g.sort_order,
			g.is_exclusive,
			g.subscription_type,
			COUNT(DISTINCT ag.account_id) as total_accounts,
			COUNT(DISTINCT CASE
				WHEN a.schedulable = true AND a.status = 'active'
					AND (a.rate_limit_reset_at IS NULL OR a.rate_limit_reset_at <= NOW())
					AND (a.overload_until IS NULL OR a.overload_until <= NOW())
				THEN ag.account_id
			END) as normal_accounts,
			COUNT(DISTINCT CASE
				WHEN a.status = 'error'
				THEN ag.account_id
			END) as error_accounts,
			COUNT(DISTINCT CASE
				WHEN a.rate_limit_reset_at IS NOT NULL AND a.rate_limit_reset_at > NOW()
				THEN ag.account_id
			END) as ratelimit_accounts,
			COUNT(DISTINCT CASE
				WHEN a.overload_until IS NOT NULL AND a.overload_until > NOW()
				THEN ag.account_id
			END) as overload_accounts,
			COUNT(DISTINCT CASE
				WHEN a.schedulable = false OR a.status != 'active'
				THEN ag.account_id
			END) as disabled_accounts
		FROM groups g
		LEFT JOIN account_groups ag ON g.id = ag.group_id
		LEFT JOIN accounts a ON ag.account_id = a.id AND a.deleted_at IS NULL
		WHERE g.deleted_at IS NULL AND g.status = 'active'
		GROUP BY g.id, g.name, g.platform, g.rate_multiplier, g.sort_order, g.is_exclusive, g.subscription_type
		ORDER BY g.sort_order ASC, g.id ASC
	`

	rows, err := r.sql.QueryContext(ctx, accountQuery)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var stats []service.GroupMonitoringStat
	groupIDs := make([]int64, 0)
	for rows.Next() {
		var stat service.GroupMonitoringStat
		err := rows.Scan(
			&stat.GroupID,
			&stat.GroupName,
			&stat.Platform,
			&stat.RateMultiplier,
			&stat.SortOrder,
			&stat.IsExclusive,
			&stat.SubscriptionType,
			&stat.TotalAccounts,
			&stat.NormalAccounts,
			&stat.ErrorAccounts,
			&stat.RateLimitAccounts,
			&stat.OverloadAccounts,
			&stat.DisabledAccounts,
		)
		if err != nil {
			return nil, err
		}
		stat.AvailabilityRate = -1
		stat.CacheHitRate = -1
		stats = append(stats, stat)
		groupIDs = append(groupIDs, stat.GroupID)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(groupIDs) == 0 {
		return stats, nil
	}

	// 第二步：计算可用率（基于账户状态）
	for i := range stats {
		if stats[i].TotalAccounts > 0 {
			stats[i].AvailabilityRate = float64(stats[i].NormalAccounts) / float64(stats[i].TotalAccounts) * 100
		}
	}

	// 第三步：从 usage_logs 聚合最近 1 小时的缓存命中率和响应时间
	usageQuery := `
		SELECT
			group_id,
			COALESCE(SUM(cache_read_tokens), 0) as total_cache_read,
			COALESCE(SUM(input_tokens + cache_read_tokens), 0) as total_input_with_cache,
			COALESCE(AVG(duration_ms) FILTER (WHERE duration_ms > 0), 0) as avg_duration
		FROM usage_logs
		WHERE created_at >= NOW() - INTERVAL '1 hour'
			AND group_id = ANY($1)
		GROUP BY group_id
	`

	usageRows, err := r.sql.QueryContext(ctx, usageQuery, pq.Array(groupIDs))
	if err != nil {
		// 如果 usage_logs 查询失败，不影响账户状态统计
		return stats, nil
	}
	defer func() { _ = usageRows.Close() }()

	// 构建 group_id -> usage 数据的映射
	type usageData struct {
		cacheRead      int64
		inputWithCache int64
		avgDuration    float64
	}
	usageMap := make(map[int64]*usageData)

	for usageRows.Next() {
		var groupID int64
		var ud usageData
		err := usageRows.Scan(
			&groupID,
			&ud.cacheRead,
			&ud.inputWithCache,
			&ud.avgDuration,
		)
		if err != nil {
			continue
		}
		usageMap[groupID] = &ud
	}

	// 第四步：合并 usage 数据到 stats
	for i := range stats {
		ud, ok := usageMap[stats[i].GroupID]
		if !ok {
			continue
		}

		// 缓存命中率 = cache_read_tokens / (input_tokens + cache_read_tokens) × 100
		if ud.inputWithCache > 0 {
			stats[i].CacheHitRate = float64(ud.cacheRead) / float64(ud.inputWithCache) * 100
			if stats[i].CacheHitRate > 100 {
				stats[i].CacheHitRate = 100
			}
		}

		// 平均响应时间
		if ud.avgDuration > 0 {
			stats[i].AvgResponseTime = int(ud.avgDuration)
		}
	}

	return stats, nil
}
