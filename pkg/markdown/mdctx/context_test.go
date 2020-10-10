package mdctx

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yuin/goldmark/parser"
)

type CtxOp = func(parser.Context)

func TestPushError(t *testing.T) {
	errCmp := cmp.Comparer(func(err1, err2 error) bool {
		return err1.Error() == err2.Error()
	})
	push := func(errs ...error) CtxOp {
		return func(pc parser.Context) {
			for _, err := range errs {
				PushError(pc, err)
			}
		}
	}
	popExpect := func(wantErrs ...error) CtxOp {
		return func(pc parser.Context) {
			errs := PopErrors(pc)
			if diff := cmp.Diff(errs, wantErrs, errCmp); diff != "" {
				t.Errorf("PopErrors mismatch (-want +got):\n%s", diff)
			}
		}
	}
	err1 := errors.New("alpha")
	err2 := errors.New("bravo")
	tests := []struct {
		name string
		ops  []CtxOp
	}{
		{"empty pop", []CtxOp{popExpect()}},
		{"1 elem push-pop", []CtxOp{push(err1), popExpect(err1)}},
		{"2 elem push-pop", []CtxOp{push(err1, err2), popExpect(err1, err2)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := parser.NewContext()
			for _, op := range tt.ops {
				op(ctx)
			}
		})
	}
}

func TestAddFeature(t *testing.T) {
	ctx := parser.NewContext()
	if GetFeatures(ctx).Len() != 0 {
		t.Fatalf("Expected features to be empty but had %v", GetFeatures(ctx))
	}
	AddFeature(ctx, FeatureKatex)
	if diff := cmp.Diff(GetFeatures(ctx).Slice(), []Feature{FeatureKatex}); diff != "" {
		t.Fatalf("expected features to only contain %s; but got:\n%s", FeatureKatex, diff)
	}
	AddFeature(ctx, FeatureKatex)
	if diff := cmp.Diff(GetFeatures(ctx).Slice(), []Feature{FeatureKatex}); diff != "" {
		t.Fatalf("expected features to only contain one %s; but got:\n%s", FeatureKatex, diff)
	}
}
