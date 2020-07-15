package cmd

import (
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/cronexpr"
	"testing"
)

func TestFlagParsing(t *testing.T) {
	cases := []struct {
		flags  []string
		result map[string]*Expr
	}{
		{
			flags: []string{
				"--cmd", "first.command", "--sched", "* * * * *",
				"--cmd", "second.command", "--sched", "1 * * * *",
				"--cmd", "third.command", "--sched", "2 * * * *",
			},
			result: map[string]*Expr{
				"first.command": {
					RawExpr: "* * * * *",
					Expr:    cronexpr.MustParse("* * * * *"),
				},
				"second.command": {
					RawExpr: "1 * * * *",
					Expr:    cronexpr.MustParse("1 * * * *"),
				},
				"third.command": {
					RawExpr: "2 * * * *",
					Expr:    cronexpr.MustParse("2 * * * *"),
				},
			},
		},
	}

	for idx, test := range cases {
		err := manager.Flags().Parse(test.flags)
		if err != nil {
			t.Fatalf("case %d: failed to parse flags. %s", idx, err)
		}

		if !cmp.Equal(test.result, expressions, cmp.AllowUnexported(cronexpr.Expression{})) {
			t.Fatalf("case %d: result mismatch. %s", idx, cmp.Diff(test.result, expressions, cmp.AllowUnexported(cronexpr.Expression{})))
		}
	}
}
