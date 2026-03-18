package breakpoint

import (
	"regexp"
	"testing"
)

const expr1 = "PC=00100h && SP>=256"
const expr2 = "PC=00115h"
const expr3 = "SP>=1332"

var ctx = map[string]interface{}{
	"PC": 0x100,
	"A":  0x55,
	"SP": 0x200,
}

const exprRep = "PC=00115h and (B=5 or  BC = 5)"
const exprDst = "PC==0x00115 && (B==5 || BC == 5)"

func Test_PatchExpression(t *testing.T) {
	ex := patchExpression(exprRep)
	if ex != exprDst {
		t.Errorf("Patched expression does not match\n     got: %s\nexpected: %s", ex, exprDst)
	}

}

func Test_ComplexExpr(t *testing.T) {

	b, e := NewBreakpoint(expr1)
	//e := b.SetExpression(exp1)
	if e != nil {
		t.Error(e)
	} else if b != nil {
		b.enabled = true
		if !b.Hit(ctx) {
			t.Errorf("Breakpoint not hit")
		}
	}

}

const expSimplePC = "PC=00119h"

func Test_BPSetPC(t *testing.T) {
	b, e := NewBreakpoint(expSimplePC)
	if e != nil {
		t.Error(e)
	} else if b != nil {
		if b.bpType != BPTypeSimplePC {
			t.Errorf("Breakpoint type does not match BPTypeSimplePC")
		}
		b.enabled = true
		if b.Hit(ctx) {
			t.Errorf("Breakpoint hit but will not!")
		}
	}
}

func Test_MatchSP(t *testing.T) {

	pcMatch := regexp.MustCompile(`SP>=[[:xdigit:]]+$`)
	matched := pcMatch.MatchString(expr3)
	if !matched {
		t.Errorf("SP>=XXXXh not matched")
	}

}

func Test_GetCtx(t *testing.T) {
	pc := getUint16("PC", ctx)
	if pc != 0x100 {
		t.Errorf("PC value not found in context")
	}
}
