package sqlite

import (
	"context"
	"testing"

	"github.com/salary-manager/backend/internal/model"
)

func newCountryRepoForTest(t *testing.T) *countryRepo {
	t.Helper()
	return &countryRepo{db: newTestDB(t)}
}

// ----- Create -----

func TestCountry_Create_Success(t *testing.T) {
	cr := newCountryRepoForTest(t)
	ctx := context.Background()

	c := &model.Country{Name: "United States", Code: "US", Currency: "USD"}
	if err := cr.Create(ctx, c); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if c.ID == 0 {
		t.Error("Create() should assign an ID")
	}
	if !c.IsActive {
		t.Error("Create() should default IsActive to true")
	}
	if c.CreatedAt.IsZero() || c.UpdatedAt.IsZero() {
		t.Error("Create() should set CreatedAt and UpdatedAt")
	}
}

func TestCountry_Create_NormalizesCodeAndCurrency(t *testing.T) {
	cr := newCountryRepoForTest(t)
	ctx := context.Background()

	c := &model.Country{Name: "United States", Code: "us", Currency: "usd"}
	if err := cr.Create(ctx, c); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if c.Code != "US" {
		t.Errorf("Code = %q, want uppercase US", c.Code)
	}
	if c.Currency != "USD" {
		t.Errorf("Currency = %q, want uppercase USD", c.Currency)
	}
}

func TestCountry_Create_DuplicateName(t *testing.T) {
	cr := newCountryRepoForTest(t)
	ctx := context.Background()

	if err := cr.Create(ctx, &model.Country{Name: "Brazil", Code: "BR", Currency: "BRL"}); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	err := cr.Create(ctx, &model.Country{Name: "Brazil", Code: "BZ", Currency: "BZD"})
	if err == nil {
		t.Fatal("expected duplicate-name insert to fail")
	}
}

func TestCountry_Create_DuplicateCode(t *testing.T) {
	cr := newCountryRepoForTest(t)
	ctx := context.Background()

	if err := cr.Create(ctx, &model.Country{Name: "United States", Code: "US", Currency: "USD"}); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	err := cr.Create(ctx, &model.Country{Name: "Different Name", Code: "us", Currency: "USD"})
	if err == nil {
		t.Fatal("expected duplicate-code insert to fail (uppercase normalisation)")
	}
}

// ----- GetByID -----

func TestCountry_GetByID_Success(t *testing.T) {
	cr := newCountryRepoForTest(t)
	ctx := context.Background()

	c := &model.Country{Name: "Germany", Code: "DE", Currency: "EUR"}
	if err := cr.Create(ctx, c); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := cr.GetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got == nil {
		t.Fatal("GetByID returned nil for an existing id")
	}
	if got.Name != "Germany" || got.Code != "DE" || got.Currency != "EUR" {
		t.Errorf("got %+v, want Germany/DE/EUR", got)
	}
}

func TestCountry_GetByID_NotFound(t *testing.T) {
	cr := newCountryRepoForTest(t)

	got, err := cr.GetByID(context.Background(), 9999)
	if err != nil {
		t.Fatalf("GetByID: unexpected error %v", err)
	}
	if got != nil {
		t.Errorf("GetByID for missing id should return nil, got %+v", got)
	}
}

// ----- GetByCode -----

func TestCountry_GetByCode_CaseInsensitive(t *testing.T) {
	cr := newCountryRepoForTest(t)
	ctx := context.Background()
	if err := cr.Create(ctx, &model.Country{Name: "India", Code: "IN", Currency: "INR"}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	for _, code := range []string{"IN", "in", "In", "iN"} {
		got, err := cr.GetByCode(context.Background(), code)
		if err != nil {
			t.Fatalf("GetByCode(%q): %v", code, err)
		}
		if got == nil || got.Name != "India" {
			t.Errorf("GetByCode(%q) did not find India", code)
		}
	}
}

func TestCountry_GetByCode_NotFound(t *testing.T) {
	cr := newCountryRepoForTest(t)

	got, err := cr.GetByCode(context.Background(), "ZZ")
	if err != nil {
		t.Fatalf("GetByCode: %v", err)
	}
	if got != nil {
		t.Errorf("GetByCode for missing code should return nil, got %+v", got)
	}
}

// ----- List -----

func TestCountry_List_EmptyDatabase(t *testing.T) {
	cr := newCountryRepoForTest(t)

	out, err := cr.List(context.Background(), model.CountryListRequest{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 0 {
		t.Errorf("Total = %d, want 0", out.Total)
	}
	if out.Countries == nil {
		t.Error("List should return an empty slice, not nil")
	}
	if len(out.Countries) != 0 {
		t.Errorf("Countries = %v, want empty", out.Countries)
	}
}

func TestCountry_List_AllRecords_NoLimit(t *testing.T) {
	cr := newCountryRepoForTest(t)
	ctx := context.Background()
	for _, c := range []model.Country{
		{Name: "Argentina", Code: "AR", Currency: "ARS"},
		{Name: "Brazil", Code: "BR", Currency: "BRL"},
		{Name: "Chile", Code: "CL", Currency: "CLP"},
	} {
		c := c
		if err := cr.Create(ctx, &c); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	out, err := cr.List(ctx, model.CountryListRequest{}) // Limit=0 -> all
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 3 || len(out.Countries) != 3 {
		t.Errorf("Total=%d, returned=%d, want 3 of each", out.Total, len(out.Countries))
	}
}

func TestCountry_List_SortedByName(t *testing.T) {
	cr := newCountryRepoForTest(t)
	ctx := context.Background()
	for _, c := range []model.Country{
		{Name: "Zambia", Code: "ZM", Currency: "ZMW"},
		{Name: "Argentina", Code: "AR", Currency: "ARS"},
		{Name: "Mexico", Code: "MX", Currency: "MXN"},
	} {
		c := c
		if err := cr.Create(ctx, &c); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	out, err := cr.List(ctx, model.CountryListRequest{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	want := []string{"Argentina", "Mexico", "Zambia"}
	for i, c := range out.Countries {
		if c.Name != want[i] {
			t.Errorf("Countries[%d].Name = %q, want %q", i, c.Name, want[i])
		}
	}
}

func TestCountry_List_PaginationWindow(t *testing.T) {
	cr := newCountryRepoForTest(t)
	ctx := context.Background()
	// Insert 5 countries; sorted alphabetically: A, B, C, D, E
	for _, name := range []string{"Eire", "Denmark", "Cuba", "Belgium", "Argentina"} {
		c := &model.Country{Name: name, Code: name[:2], Currency: "XXX"}
		if err := cr.Create(ctx, c); err != nil {
			t.Fatalf("Create %s: %v", name, err)
		}
	}

	out, err := cr.List(ctx, model.CountryListRequest{
		Pagination: model.Pagination{Limit: 2, Offset: 1},
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 5 {
		t.Errorf("Total = %d, want 5 (count is unaffected by pagination)", out.Total)
	}
	if len(out.Countries) != 2 {
		t.Fatalf("page size = %d, want 2", len(out.Countries))
	}
	// Sorted: Argentina, Belgium, Cuba, Denmark, Eire — offset 1, limit 2 -> Belgium, Cuba
	if out.Countries[0].Name != "Belgium" || out.Countries[1].Name != "Cuba" {
		t.Errorf("page contents = %v, want [Belgium Cuba]", []string{out.Countries[0].Name, out.Countries[1].Name})
	}
	if out.Limit != 2 || out.Offset != 1 {
		t.Errorf("echoed pagination = limit=%d offset=%d, want 2/1", out.Limit, out.Offset)
	}
}

func TestCountry_List_OffsetBeyondEnd_ReturnsEmpty(t *testing.T) {
	cr := newCountryRepoForTest(t)
	ctx := context.Background()
	if err := cr.Create(ctx, &model.Country{Name: "Solo", Code: "SO", Currency: "SOX"}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	out, err := cr.List(ctx, model.CountryListRequest{
		Pagination: model.Pagination{Limit: 10, Offset: 50},
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 1 {
		t.Errorf("Total = %d, want 1", out.Total)
	}
	if len(out.Countries) != 0 {
		t.Errorf("expected empty page beyond end of data, got %d rows", len(out.Countries))
	}
}

func TestCountry_List_NormalizesNegativePagination(t *testing.T) {
	cr := newCountryRepoForTest(t)
	ctx := context.Background()
	if err := cr.Create(ctx, &model.Country{Name: "Solo", Code: "SO", Currency: "SOX"}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	out, err := cr.List(ctx, model.CountryListRequest{
		Pagination: model.Pagination{Limit: -10, Offset: -5},
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 1 || len(out.Countries) != 1 {
		t.Errorf("negative pagination should be clamped and return all rows; got total=%d returned=%d",
			out.Total, len(out.Countries))
	}
}

func TestCountry_List_ExcludesInactiveByDefault(t *testing.T) {
	cr := newCountryRepoForTest(t)
	ctx := context.Background()
	active := &model.Country{Name: "Active Land", Code: "AL", Currency: "ALX"}
	inactive := &model.Country{Name: "Old Land", Code: "OL", Currency: "OLX"}
	for _, c := range []*model.Country{active, inactive} {
		if err := cr.Create(ctx, c); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}
	deactivateCountry(t, cr.db, inactive.ID)

	out, err := cr.List(ctx, model.CountryListRequest{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 1 || out.Countries[0].Name != "Active Land" {
		t.Errorf("active filter failed: total=%d countries=%v", out.Total, out.Countries)
	}
}

func TestCountry_List_IncludeInactive(t *testing.T) {
	cr := newCountryRepoForTest(t)
	ctx := context.Background()
	a := &model.Country{Name: "Alpha", Code: "AA", Currency: "AAX"}
	b := &model.Country{Name: "Beta", Code: "BB", Currency: "BBX"}
	for _, c := range []*model.Country{a, b} {
		if err := cr.Create(ctx, c); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}
	deactivateCountry(t, cr.db, b.ID)

	out, err := cr.List(ctx, model.CountryListRequest{IncludeInactive: true})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 2 || len(out.Countries) != 2 {
		t.Errorf("IncludeInactive=true should return both rows, got total=%d returned=%d",
			out.Total, len(out.Countries))
	}
}
