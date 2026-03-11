package search

import "testing"

func TestMatchesSeparatorAwareTokens(t *testing.T) {
	if !Matches("stock_ledger", "st ger") {
		t.Fatalf("expected tokenized partial query to match separator-based candidate")
	}
}

func TestMatchesIsCaseInsensitive(t *testing.T) {
	if !Matches("Feature-Branch", "feature bra") {
		t.Fatalf("expected case-insensitive multi-token match")
	}
}

func TestMatchesCompactQueryAgainstSeparatorName(t *testing.T) {
	if !Matches("Stock_Location", "stocklocation") {
		t.Fatalf("expected compact query to match separator-based candidate")
	}
}

func TestMatchesReturnsFalseWhenTokenMissing(t *testing.T) {
	if Matches("alpha_beta", "alpha zeta") {
		t.Fatalf("expected false when one query token is missing")
	}
}

func TestMatchRuneIndexesMarksContiguousMatch(t *testing.T) {
	indexes := MatchRuneIndexes("stock_ledger", "ledger")
	if len(indexes) == 0 {
		t.Fatalf("expected highlight indexes for contiguous match")
	}
	for _, expectedIndex := range []int{6, 7, 8, 9, 10, 11} {
		if !indexes[expectedIndex] {
			t.Fatalf("expected rune index %d to be highlighted", expectedIndex)
		}
	}
}

func TestMatchRuneIndexesReturnsNilForNoMatch(t *testing.T) {
	if indexes := MatchRuneIndexes("stock_ledger", "missing"); indexes != nil {
		t.Fatalf("expected nil indexes when there is no match")
	}
}

func TestMatchRuneIndexesCompactQueryAcrossSeparator(t *testing.T) {
	indexes := MatchRuneIndexes("Stock_Location", "stocklocation")
	if len(indexes) == 0 {
		t.Fatalf("expected highlight indexes for compact separator-spanning query")
	}
	for _, expectedIndex := range []int{0, 1, 2, 3, 4, 6, 7, 8, 9, 10, 11, 12, 13} {
		if !indexes[expectedIndex] {
			t.Fatalf("expected rune index %d to be highlighted", expectedIndex)
		}
	}
}
