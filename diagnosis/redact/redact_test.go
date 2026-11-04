package redact

import (
	"strings"
	"testing"
)

func TestSQLRedactsLiteralsCommentsAndArguments(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		secrets []string
	}{
		{"values", "SELECT * FROM orders WHERE id = 987654 AND email = 'private@example.com' AND enabled = TRUE", []string{"987654", "private@example.com", "true"}},
		{"escaped", `SELECT * FROM t WHERE s = 'can''t reveal' AND x = 0xdeadbeef AND n = -789123`, []string{"can", "reveal", "deadbeef", "789123"}},
		{"dates and json", `SELECT * FROM t WHERE d = DATE '2025-03-25' AND j = '{"private":"payload"}'`, []string{"2025-03-25", "private", "payload"}},
		{"in list", "SELECT * FROM t WHERE a IN (781237, 781238, 781239)", []string{"781237", "781238", "781239", "..."}},
		{"comment", "SELECT /* comment secret */ * FROM t WHERE x = 'secretvalue'", []string{"secret", "comment"}},
		{"prepared", "SELECT * FROM t WHERE a = ? [arguments: 918273, 'privatearg']", []string{"918273", "privatearg", "arguments"}},
		{"hint value", "SELECT /*+ MAX_EXECUTION_TIME(918273) USE_INDEX(t, idx) */ * FROM t WHERE a = 712345", []string{"918273", "712345"}},
		{"like escape", "SELECT * FROM t WHERE x LIKE 'private' ESCAPE '!'", []string{"private", "'!'"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SQL(tc.input)
			if err != nil {
				t.Fatalf("redact SQL: %v", err)
			}
			for _, secret := range tc.secrets {
				if strings.Contains(strings.ToLower(got), strings.ToLower(secret)) {
					t.Fatalf("redacted SQL contains %q: %s", secret, got)
				}
			}
			if !strings.Contains(got, "?") {
				t.Fatalf("no placeholder in %q", got)
			}
		})
	}
}

func TestSQLFailsClosed(t *testing.T) {
	for _, input := range []string{
		"SELECT * FROM t WHERE value = 'unfinished-secret",
		"SELECT * FROM t WHERE x = 1; SELECT 'second-secret'",
		`SELECT "ambiguous-secret" FROM t`,
		"SELECT * FROM t WHERE x = ? [arguments: unfinished-secret",
	} {
		result, err := SQL(input)
		if err == nil || result != "" || strings.Contains(err.Error(), "secret") || DisplaySQL(input) != Omitted {
			t.Fatalf("expected safe omission, got %q, %v", result, err)
		}
	}
}

func TestListMarkerPreservesQuotedIdentifiers(t *testing.T) {
	got := replaceListMarkers("select `name ... value` from `t` where `x` in ( ... )")
	if got != "select `name ... value` from `t` where `x` in ( ? )" {
		t.Fatalf("unexpected list replacement: %s", got)
	}
}

func TestBindingRedactsBothStatements(t *testing.T) {
	got, err := Binding("select * from t where id = 234567", "select /*+ USE_INDEX(t, idx) */ * from t where id = 234567")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "234567") || strings.Count(got, "?") != 2 || !strings.Contains(strings.ToLower(got), "use_index") {
		t.Fatalf("unexpected redacted binding: %s", got)
	}
	if got := DisplayBinding("select * from t where x='secret", "select * from t"); got != Omitted {
		t.Fatalf("malformed binding not omitted: %s", got)
	}
}

func TestPlanRedactsUnknownAndParameterFields(t *testing.T) {
	input := "id\testRows\testCost\ttask\taccess object\toperator info\tunknown\n" +
		"└─IndexRangeScan_14\t1.00\t12.34\tcop[tikv]\ttable:t\trange:[987654,987654], eq(email, 'private')\tsecret"
	got, err := Plan(input)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"└─IndexRangeScan_14", "1.00", "12.34", "cop[tikv]", "table:t", "range:[?,?], eq(?, ?)", "\t?"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
	for _, secret := range []string{"987654", "private", "secret"} {
		if strings.Contains(got, secret) {
			t.Fatalf("plan leaks %q: %q", secret, got)
		}
	}
}

func TestPlanFailsClosedForUnknownLayout(t *testing.T) {
	for _, input := range []string{"secret plan", "id\ttask\nIndexScan_1\troot\textra secret", "task\nsecret"} {
		if got := DisplayPlan(input); got != Omitted {
			t.Fatalf("expected omission, got %q", got)
		}
	}
}

func TestPlanHandlesFormattingTabAndEmptyLastColumn(t *testing.T) {
	got, err := Plan("\tid\ttask\toperator info\n\tPoint_Get_1\troot\t")
	if err != nil || !strings.Contains(got, "Point_Get_1\troot\t") {
		t.Fatalf("unexpected plan %q: %v", got, err)
	}
}

func TestPlanPreservesIdentifiersAndRedactsRanges(t *testing.T) {
	plan := "id\ttask\taccess object\toperator info\n" +
		"IndexRangeScan_4\tcop[tikv]\ttable:orders, index:idx_customer(customer_id, status)\trange:[123 456,123 456], keep order:false"
	got, err := PlanWithSQL(plan, "SELECT customer_id, status FROM orders WHERE customer_id = 123", "app")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"table:orders, index:idx_customer(customer_id, status)", "range:[?,?], keep order:false"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
	for _, secret := range []string{"123", "456"} {
		if strings.Contains(got, secret) {
			t.Fatalf("plan leaks %q: %s", secret, got)
		}
	}
}

func TestPlanExpressionUsesOnlyKnownSQLIdentifiers(t *testing.T) {
	plan := "id\ttask\toperator info\nSelection_1\troot\t" +
		"eq(app.orders.customer_id, 123), not(isnull(app.orders.customer_id)), eq(secret_domain, 'private')"
	got, err := PlanWithSQL(plan, "SELECT customer_id FROM orders WHERE customer_id=123", "app")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"eq(app.orders.customer_id, ?)", "not(isnull(app.orders.customer_id))", "eq(?, ?)"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
	for _, secret := range []string{"123", "private", "secret_domain"} {
		if strings.Contains(got, secret) {
			t.Fatalf("plan leaks %q: %s", secret, got)
		}
	}
}

func TestPlanMasksUnknownAndMalformedFragments(t *testing.T) {
	for _, value := range []string{
		"range:[\"secret,a\",\"secret,b\"], keep order:true",
		"range:[\"secret,a\",\"secret,b\"], (999,1000], desc",
		"table:t, unknown:secret", "custom(secret)",
		"range:[secret, secret", "eq(t.c, 'unfinished secret)",
		"table_name:[\"secret.example.com\"]",
	} {
		got := redactOperatorInfo(value, nil)
		for _, secret := range []string{"secret", "999", "1000"} {
			if strings.Contains(got, secret) {
				t.Errorf("input %q leaked %q: %q", value, secret, got)
			}
		}
	}
}

func TestAccessObjectRejectsUnknownFragments(t *testing.T) {
	got := redactAccessObject("table:orders, index:idx_customer(customer_id), extra:private")
	if got != "table:orders, index:idx_customer(customer_id), ?" {
		t.Fatalf("unexpected access object: %s", got)
	}
}
