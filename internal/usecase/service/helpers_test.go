package service

import (
	"github.com/stretchr/testify/assert"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/domain/vo"
	"testing"
	"time"
)

func TestSortReports(t *testing.T) {
	t.Parallel()

	type args struct {
		reports []*models.ClientReport
		sortBy  string
	}

	tests := []struct {
		name      string
		args      args
		wantOrder []string
		wantErr   assert.ErrorAssertionFunc
	}{
		{
			name: "SortByOrdersDescending",
			args: args{
				reports: []*models.ClientReport{
					{UserID: "u1", TotalOrders: 1},
					{UserID: "u3", TotalOrders: 3},
					{UserID: "u2", TotalOrders: 3},
				},
				sortBy: "orders",
			},

			wantOrder: []string{"u2", "u3", "u1"},
			wantErr:   assert.NoError,
		},
		{
			name: "SortBySumDescending",
			args: args{
				reports: []*models.ClientReport{
					{UserID: "a", TotalPurchaseSum: models.PriceKopecks(100)},
					{UserID: "b", TotalPurchaseSum: models.PriceKopecks(300)},
					{UserID: "c", TotalPurchaseSum: models.PriceKopecks(300)},
				},
				sortBy: "sum",
			},

			wantOrder: []string{"b", "c", "a"},
			wantErr:   assert.NoError,
		},
		{
			name: "InvalidSortOption",
			args: args{
				reports: []*models.ClientReport{
					{UserID: "x", TotalOrders: 1},
					{UserID: "y", TotalOrders: 2},
				},
				sortBy: "unknown",
			},
			wantOrder: []string{"x", "y"},
			wantErr:   assert.Error,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// клонирую слайс, чтобы избежать изменения данных между тестами
			reports := make([]*models.ClientReport, len(tt.args.reports))
			copy(reports, tt.args.reports)

			err := sortReports(reports, tt.args.sortBy)
			tt.wantErr(t, err)

			if err == nil {
				var gotOrder []string
				for _, r := range reports {
					gotOrder = append(gotOrder, r.UserID)
				}
				assert.Equal(t, tt.wantOrder, gotOrder)
			}
		})
	}
}

func TestValidateAccept(t *testing.T) {
	t.Parallel()

	type args struct {
		orderID string
		userID  string
		exp     time.Time
		weight  float64
		wantErr assert.ErrorAssertionFunc
	}

	now := time.Now()
	future := now.Add(1 * time.Hour)
	past := now.Add(-1 * time.Hour)

	tests := []struct {
		name string
		args args
	}{
		{
			name: "EmptyOrderID",
			args: args{"", "u1", future, 1.0, assert.Error},
		},
		{
			name: "EmptyUserID",
			args: args{"o1", "", future, 1.0, assert.Error},
		},
		{
			name: "ExpiredExpiry",
			args: args{"o1", "u1", past, 1.0, assert.Error},
		},
		{
			name: "NonPositiveWeight",
			args: args{"o1", "u1", future, 0, assert.Error},
		},
		{
			name: "AllValid",
			args: args{"o1", "u1", future, 2.5, assert.NoError},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateAccept(tt.args.orderID, tt.args.userID, tt.args.exp, tt.args.weight)
			tt.args.wantErr(t, err)
		})
	}
}

func TestValidateReturn(t *testing.T) {
	t.Parallel()

	type args struct {
		order    *models.Order
		wantErr  assert.ErrorAssertionFunc
		contains string
	}

	now := time.Now()
	notExpired := now.Add(1 * time.Hour)
	expired := now.Add(-1 * time.Hour)

	tests := []struct {
		name string
		args args
	}{
		{
			name: "IssuedStatus",
			args: args{
				&models.Order{ID: "o1", Status: models.StatusIssued, ExpiresAt: expired},
				assert.Error,
				"cannot return an issued order",
			},
		},
		{
			name: "NotExpiredYet",
			args: args{
				&models.Order{ID: "o2", Status: models.StatusAccepted, ExpiresAt: notExpired},
				assert.Error,
				"storage period not expired yet",
			},
		},
		{
			name: "ExpiredAndAccepted",
			args: args{
				&models.Order{ID: "o3", Status: models.StatusAccepted, ExpiresAt: expired},
				assert.NoError,
				"",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateReturn(tt.args.order)
			tt.args.wantErr(t, err)
			if tt.args.contains != "" && err != nil {
				assert.Contains(t, err.Error(), tt.args.contains)
			}
		})
	}
}

func TestValidateIssue(t *testing.T) {
	t.Parallel()

	type args struct {
		order    *models.Order
		userID   string
		now      time.Time
		wantErr  assert.ErrorAssertionFunc
		contains string
	}

	baseExp := time.Now().Add(1 * time.Hour)

	tests := []struct {
		name string
		args args
	}{
		{
			name: "WrongUser",
			args: args{
				&models.Order{ID: "o1", UserID: "u2", Status: models.StatusAccepted, ExpiresAt: baseExp},
				"u1", time.Now(),
				assert.Error, "order belongs to another user",
			},
		},
		{
			name: "WrongStatus",
			args: args{
				&models.Order{ID: "o2", UserID: "u1", Status: models.StatusIssued, ExpiresAt: baseExp},
				"u1", time.Now(),
				assert.Error, "order not in accepted status",
			},
		},
		{
			name: "Expired",
			args: args{
				&models.Order{ID: "o3", UserID: "u1", Status: models.StatusAccepted, ExpiresAt: time.Now().Add(-time.Minute)},
				"u1", time.Now(),
				assert.Error, "storage period expired",
			},
		},
		{
			name: "AllValid",
			args: args{
				&models.Order{ID: "o4", UserID: "u1", Status: models.StatusAccepted, ExpiresAt: time.Now().Add(time.Hour)},
				"u1", time.Now(),
				assert.NoError, "",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateIssue(tt.args.order, tt.args.userID, tt.args.now)
			tt.args.wantErr(t, err)
			if tt.args.contains != "" && err != nil {
				assert.Contains(t, err.Error(), tt.args.contains)
			}
		})
	}
}

func TestValidateClientReturn(t *testing.T) {
	t.Parallel()

	type args struct {
		order    *models.Order
		userID   string
		now      time.Time
		wantErr  assert.ErrorAssertionFunc
		contains string
	}

	issuedAt := time.Now().Add(-24 * time.Hour)
	tooLate := time.Now().Add(-49 * time.Hour)

	tests := []struct {
		name string
		args args
	}{
		{
			name: "WrongUser",
			args: args{
				&models.Order{ID: "o1", UserID: "u2", Status: models.StatusIssued, IssuedAt: &issuedAt},
				"u1", time.Now(), assert.Error, "order belongs to another user",
			},
		},
		{
			name: "NotIssued",
			args: args{
				&models.Order{ID: "o2", UserID: "u1", Status: models.StatusAccepted, IssuedAt: &issuedAt},
				"u1", time.Now(), assert.Error, "order not in issued status",
			},
		},
		{
			name: "WindowExpired",
			args: args{
				&models.Order{ID: "o3", UserID: "u1", Status: models.StatusIssued, IssuedAt: &tooLate},
				"u1", time.Now(), assert.Error, "return window expired",
			},
		},
		{
			name: "AllValid",
			args: args{
				&models.Order{ID: "o4", UserID: "u1", Status: models.StatusIssued, IssuedAt: &issuedAt},
				"u1", time.Now(), assert.NoError, "",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateClientReturn(tt.args.order, tt.args.userID, tt.args.now)
			tt.args.wantErr(t, err)
			if tt.args.contains != "" && err != nil {
				assert.Contains(t, err.Error(), tt.args.contains)
			}
		})
	}
}

func TestSortOrders(t *testing.T) {
	t.Parallel()

	now := time.Now()
	o1 := &models.Order{ID: "a", CreatedAt: now.Add(-2 * time.Hour)}
	o2 := &models.Order{ID: "b", CreatedAt: now.Add(-1 * time.Hour)}
	o3 := &models.Order{ID: "c", CreatedAt: now}

	tests := []struct {
		name     string
		inputIDs []string
		list     []*models.Order
		wantIDs  []string
	}{
		{
			name:     "DifferentTimestamps",
			inputIDs: []string{"c", "a", "b"},
			list:     []*models.Order{o3, o1, o2},
			wantIDs:  []string{"a", "b", "c"},
		},
		{
			name:     "EqualTimestamps",
			inputIDs: []string{"y", "x"},
			list: []*models.Order{
				{ID: "y", CreatedAt: now},
				{ID: "x", CreatedAt: now},
			},
			wantIDs: []string{"x", "y"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			list := make([]*models.Order, len(tt.list))
			copy(list, tt.list)

			sortOrders(list)

			var got []string
			for _, o := range list {
				got = append(got, o.ID)
			}
			assert.Equal(t, tt.wantIDs, got)
		})
	}
}

func TestPaginate(t *testing.T) {
	t.Parallel()

	type args struct {
		list  []int
		lastN int
		pg    vo.Pagination
	}
	type want struct {
		sublist []int
		total   int
	}

	full := []int{1, 2, 3, 4, 5}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "LastN_LessThanTotal",
			args: args{list: full, lastN: 2, pg: vo.Pagination{}},
			want: want{sublist: []int{4, 5}, total: 5},
		},
		{
			name: "LastN_EqualsTotal_NoPagination",
			args: args{list: full, lastN: 5, pg: vo.Pagination{Page: 0, Limit: 0}},
			want: want{sublist: full, total: 5},
		},
		{
			name: "LastN_EqualsTotal_WithPagination",
			args: args{list: full, lastN: 5, pg: vo.Pagination{Page: 1, Limit: 2}},
			want: want{sublist: []int{1, 2}, total: 5},
		},
		{
			name: "Pagination_Page1Limit2",
			args: args{list: full, lastN: 0, pg: vo.Pagination{Page: 1, Limit: 2}},
			want: want{sublist: []int{1, 2}, total: 5},
		},
		{
			name: "Pagination_Page2Limit2",
			args: args{list: full, lastN: 0, pg: vo.Pagination{Page: 2, Limit: 2}},
			want: want{sublist: []int{3, 4}, total: 5},
		},
		{
			name: "Pagination_Page3Limit2_Partial",
			args: args{list: full, lastN: 0, pg: vo.Pagination{Page: 3, Limit: 2}},
			want: want{sublist: []int{5}, total: 5},
		},
		{
			name: "Pagination_PageTooHigh",
			args: args{list: full, lastN: 0, pg: vo.Pagination{Page: 4, Limit: 2}},
			want: want{sublist: []int{}, total: 5},
		},
		{
			name: "NoLastN_NoPagination",
			args: args{list: full, lastN: 0, pg: vo.Pagination{Page: 0, Limit: 0}},
			want: want{sublist: full, total: 5},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, total := paginate[int](tt.args.list, tt.args.lastN, tt.args.pg)
			assert.Equal(t, tt.want.total, total)
			assert.Equal(t, tt.want.sublist, got)
		})
	}
}
