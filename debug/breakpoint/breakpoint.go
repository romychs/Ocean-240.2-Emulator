package breakpoint

import (
	"context"
	"fmt"
	"okemu/gval"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

const MaxBreakpoints = 256

const (
	BPTypeSimplePC   = iota // Simple PC=nn breakpoint
	BPTypeSimpleSP          // Simple SP>=nn breakpoint
	BPTypeExpression        // Complex expression breakpoint
)

type Breakpoint struct {
	addr      uint16
	cond      string
	eval      gval.Evaluable
	bpType    int
	pass      uint16
	passCount uint16
	enabled   bool
}

var andMatch = regexp.MustCompile(`\s+AND\s+`)
var orMatch = regexp.MustCompile(`\s+OR\s+`)

// var xorMatch = regexp.MustCompile(`\s+XOR\s+`)
var hexHMatch = regexp.MustCompile(`[[:xdigit:]]+H`)
var eqMatch = regexp.MustCompile(`[^=><!]=[^=]`)

func patchExpression(expr string) string {
	ex := strings.ToUpper(expr)
	ex = andMatch.ReplaceAllString(ex, " && ")
	ex = orMatch.ReplaceAllString(ex, " || ")
	//	ex = xorMatch.ReplaceAllString(ex, " ^ ")
	for {
		pos := hexHMatch.FindStringIndex(ex)
		if pos != nil && len(pos) == 2 {
			hex := "0x" + ex[pos[0]:pos[1]-1]
			ex = ex[:pos[0]] + hex + ex[pos[1]:]
		} else {
			break
		}
	}
	for {
		pos := eqMatch.FindStringIndex(ex)
		if pos != nil && len(pos) == 2 {
			ex = ex[:pos[0]+1] + "==" + ex[pos[1]-1:]
		} else {
			break
		}
	}
	ex = strings.ReplaceAll(ex, "NOT", "!")
	ex = strings.ReplaceAll(ex, "<>", "!=")
	return ex
}

var pcMatch = regexp.MustCompile(`^PC=[[:xdigit:]]+h$`)
var spMatch = regexp.MustCompile(`^SP>=[[:xdigit:]]+$`)

func getSecondUint16(param string, sep string) (uint16, error) {
	p := strings.Split(param, sep)
	v := p[1]
	base := 0
	if strings.HasSuffix(v, "h") || strings.HasSuffix(v, "H") {
		v = strings.TrimSuffix(v, "H")
		v = strings.TrimSuffix(v, "h")
		base = 16
	}
	a, e := strconv.ParseUint(v, base, 16)
	if e != nil {
		return 0, e
	}
	return uint16(a), nil
}

func NewBreakpoint(expr string) (*Breakpoint, error) {
	bp := Breakpoint{
		addr:      0,
		enabled:   false,
		passCount: 0,
		pass:      0,
		bpType:    BPTypeSimplePC,
	}

	// Check if BP is simple PC=addr
	expr = strings.TrimSpace(expr)
	bp.cond = expr
	pcMatched := pcMatch.MatchString(expr)
	spMatched := spMatch.MatchString(expr)

	if pcMatched {
		// PC=xxxxh
		bp.bpType = BPTypeSimplePC
		v, e := getSecondUint16(expr, "=")
		if e != nil {
			return nil, e
		}
		bp.addr = v
	} else if spMatched {
		// SP>=xxxx
		bp.bpType = BPTypeSimpleSP
		v, e := getSecondUint16(expr, "=")
		if e != nil {
			return nil, e
		}
		bp.addr = v
	} else {
		// complex expression
		bp.bpType = BPTypeExpression
		ex := patchExpression(expr)
		log.Debugf("Original Expression: '%s'", expr)
		log.Debugf(" Patched Expression: '%s'", ex)
		err := bp.SetExpression(ex)
		if err != nil {
			return nil, err
		}
	}
	return &bp, nil
}

func (b *Breakpoint) Enabled() bool {
	return b.enabled
}
func (b *Breakpoint) SetEnabled(enabled bool) {
	b.enabled = enabled
}
func (b *Breakpoint) PassCount() uint16 {
	return b.passCount
}
func (b *Breakpoint) SetPassCount(passCount uint16) {
	b.passCount = passCount
}
func (b *Breakpoint) Pass() uint16 {
	return b.pass
}
func (b *Breakpoint) SetPass(pass uint16) {
	b.pass = pass
}

func (b *Breakpoint) IncPass() {
	b.pass++
}

func (b *Breakpoint) Addr() uint16 {
	return b.addr
}
func (b *Breakpoint) SetAddr(addr uint16) {
	b.addr = addr
}
func (b *Breakpoint) Type() int {
	return b.bpType
}
func (b *Breakpoint) SetType(bpType int) {
	b.bpType = bpType
}

func getUint16(name string, ctx map[string]interface{}) uint16 {
	if v, ok := ctx[name]; ok {
		if v == nil {
			return 0
		}
		// most frequent case
		if v, ok := v.(uint16); ok {
			return v
		}
		// for less frequent cases
		switch value := v.(type) {
		case int:
			return uint16(value)
		case int8:
			return uint16(value)
		case int16:
			return uint16(value)
		case int32:
			return uint16(value)
		case int64:
			return uint16(value)
		case uint:
			return uint16(value)
		case uint8:
			return uint16(value)
		case uint32:
			return uint16(value)
		case uint64:
			return uint16(value)
		default:
			log.Errorf("Unknown type %v for variable %s", value, name)
			return 0
		}
	} else {
		log.Errorf("Variable %s not found in context!", name)
	}
	return 0
}

func (b *Breakpoint) Hit(ctx map[string]interface{}) bool {
	if !b.enabled {
		return false
	}
	if b.bpType == BPTypeSimplePC {
		pc := getUint16("PC", ctx)
		if pc == b.addr {
			log.Debugf("Breakpoint Hit PC=%04X", b.addr)
		}
		return pc == b.addr
	} else if b.bpType == BPTypeSimpleSP {
		sp := getUint16("SP", ctx)
		if sp >= b.addr {
			log.Debugf("Breakpoint Hit SP>=%04X", b.addr)
		}
		return sp >= b.addr
	}
	value, err := b.eval.EvalBool(context.Background(), ctx)
	if err != nil {
		fmt.Println(err)
	}
	return value
}

var language gval.Language

func init() {
	language = gval.NewLanguage(gval.Base(), gval.Arithmetic(), gval.Bitmask(), gval.PropositionalLogic())
}

func (b *Breakpoint) SetExpression(expression string) error {
	var err error
	b.eval, err = language.NewEvaluable(expression)
	if err != nil {
		log.Error("Illegal expression", err)
		return err
	}
	b.bpType = BPTypeExpression
	return nil
}

func (b *Breakpoint) Expression() string {
	return b.cond
}
