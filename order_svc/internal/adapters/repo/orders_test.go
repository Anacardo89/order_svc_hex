package repo

import (
	"context"
	"testing"

	"github.com/Anacardo89/order_svc_hex/order_svc/internal/core"
	"github.com/Anacardo89/order_svc_hex/order_svc/pkg/ptr"
	"github.com/Anacardo89/order_svc_hex/order_svc/pkg/testutils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderRepo_Create(t *testing.T) {
	ctx := context.Background()
	dbConn, err := testutils.ConnectTestDB(ctx, dsn)
	require.NoError(t, err)
	err = testutils.SeedTestDB(ctx, dbConn, seedPath)
	require.NoError(t, err)

	tests := []struct {
		name  string
		order *core.Order
	}{
		{
			name: "create pending order",
			order: &core.Order{
				ID:     uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
				Items:  map[string]int{"sku_1": 2, "sku_2": 1},
				Status: ptr.Ptr(core.StatusPending),
			},
		},
		{
			name: "create confirmed order",
			order: &core.Order{
				ID:     uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
				Items:  map[string]int{"sku_3": 5},
				Status: ptr.Ptr(core.StatusConfirmed),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = repo.Create(ctx, tt.order)
			require.NoError(t, err)

			savedOrder, err := repo.GetByID(ctx, tt.order.ID)
			require.NoError(t, err)

			assert.Equal(t, tt.order.ID, savedOrder.ID)
			assert.Equal(t, tt.order.Items, map[string]int(savedOrder.Items))
			assert.Equal(t, *tt.order.Status, core.Status(ptr.Val(savedOrder.Status)))
			assert.False(t, savedOrder.CreatedAt.IsZero(), "CreatedAt should be set")
			assert.False(t, savedOrder.UpdatedAt.IsZero(), "UpdatedAt should be set")
		})
	}
}

func TestOrderRepo_GetByID(t *testing.T) {
	ctx := context.Background()
	dbConn, err := testutils.ConnectTestDB(ctx, dsn)
	require.NoError(t, err)
	err = testutils.SeedTestDB(ctx, dbConn, seedPath)
	require.NoError(t, err)

	tests := []struct {
		name        string
		id          uuid.UUID
		expected    *core.Order
		expectError error
	}{
		{
			name: "get pending order",
			id:   uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			expected: &core.Order{
				ID:     uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				Items:  map[string]int{"sku_1": 2, "sku_2": 1},
				Status: ptr.Ptr(core.StatusPending),
			},
		},
		{
			name: "get confirmed order",
			id:   uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			expected: &core.Order{
				ID:     uuid.MustParse("22222222-2222-2222-2222-222222222222"),
				Items:  map[string]int{"sku_3": 5},
				Status: ptr.Ptr(core.StatusConfirmed),
			},
		},
		{
			name: "get failed order",
			id:   uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			expected: &core.Order{
				ID:     uuid.MustParse("33333333-3333-3333-3333-333333333333"),
				Items:  map[string]int{"sku_4": 1, "sku_5": 3},
				Status: ptr.Ptr(core.StatusFailed),
			},
		},
		{
			name:        "order not found",
			id:          uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
			expected:    nil,
			expectError: pgx.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := repo.GetByID(ctx, tt.id)
			if tt.expectError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectError)
				assert.Nil(t, order)
				return
			} else {
				require.NoError(t, err)
				require.NotNil(t, order)
			}

			assert.Equal(t, tt.expected.ID, order.ID)
			assert.Equal(t, tt.expected.Items, order.Items)
			assert.Equal(t, tt.expected.Status, order.Status)
			assert.False(t, order.CreatedAt.IsZero())
			assert.False(t, order.UpdatedAt.IsZero())
		})
	}
}

func TestOrderRepo_ListByStatus(t *testing.T) {
	ctx := context.Background()
	dbConn, err := testutils.ConnectTestDB(ctx, dsn)
	require.NoError(t, err)
	err = testutils.SeedTestDB(ctx, dbConn, seedPath)
	require.NoError(t, err)

	tests := []struct {
		name        string
		status      core.Status
		expected    []*core.Order
		expectError bool
	}{
		{
			name:   "pending orders",
			status: core.StatusPending,
			expected: []*core.Order{
				{
					ID:     uuid.MustParse("11111111-1111-1111-1111-111111111111"),
					Items:  map[string]int{"sku_1": 2, "sku_2": 1},
					Status: ptr.Ptr(core.StatusPending),
				},
			},
		},
		{
			name:   "confirmed orders",
			status: core.StatusConfirmed,
			expected: []*core.Order{
				{
					ID:     uuid.MustParse("22222222-2222-2222-2222-222222222222"),
					Items:  map[string]int{"sku_3": 5},
					Status: ptr.Ptr(core.StatusConfirmed),
				},
			},
		},
		{
			name:   "failed orders",
			status: core.StatusFailed,
			expected: []*core.Order{
				{
					ID:     uuid.MustParse("33333333-3333-3333-3333-333333333333"),
					Items:  map[string]int{"sku_4": 1, "sku_5": 3},
					Status: ptr.Ptr(core.StatusFailed),
				},
			},
		},
		{
			name:        "invalid status",
			status:      "archived",
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orders, err := repo.ListByStatus(ctx, tt.status)
			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, orders)
				return
			} else {
				require.NoError(t, err)
				require.Len(t, orders, len(tt.expected))
			}

			for i, expectedOrder := range tt.expected {
				got := orders[i]
				assert.Equal(t, expectedOrder.ID, got.ID)
				assert.Equal(t, expectedOrder.Items, got.Items)
				assert.Equal(t, expectedOrder.Status, got.Status)
				assert.False(t, got.CreatedAt.IsZero())
				assert.False(t, got.UpdatedAt.IsZero())
			}
		})
	}
}

func TestOrderRepo_UpdateStatus(t *testing.T) {
	ctx := context.Background()
	dbConn, err := testutils.ConnectTestDB(ctx, dsn)
	require.NoError(t, err)
	err = testutils.SeedTestDB(ctx, dbConn, seedPath)
	require.NoError(t, err)

	verifyQuery := `
		SELECT
			id,
			items,
			status,
			created_at,
			updated_at
		FROM orders
		WHERE id = $1
	;`

	tests := []struct {
		name        string
		id          uuid.UUID
		newStatus   core.Status
		expectError bool
	}{
		{
			name:      "update pending → confirmed",
			id:        uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			newStatus: core.StatusConfirmed,
		},
		{
			name:      "update confirmed → failed",
			id:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			newStatus: core.StatusFailed,
		},
		{
			name:        "non-existent order",
			id:          uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
			newStatus:   core.StatusPending,
			expectError: false,
		},
		{
			name:        "invalid status enum",
			id:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			newStatus:   "archived",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.UpdateStatus(ctx, tt.id, tt.newStatus)
			if tt.expectError {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			var updated Order
			err = dbConn.QueryRow(ctx, verifyQuery, tt.id).Scan(
				&updated.ID,
				&updated.Items,
				&updated.Status,
				&updated.CreatedAt,
				&updated.UpdatedAt,
			)
			if tt.name == "non-existent order" {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
				assert.Equal(t, string(tt.newStatus), ptr.Val(updated.Status))
			}
		})
	}
}
