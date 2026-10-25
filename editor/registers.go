package editor

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/malivvan/readline/color"
	"github.com/malivvan/readline/completion"
)

type Buffers struct {
	num      map[int][]rune
	alpha    map[rune][]rune
	ro       map[rune][]rune
	waiting  bool
	selected bool
	active   rune
	mutex    *sync.Mutex

	in     io.Reader
	out    io.Writer
	errOut io.Writer
}

// SetIO sets the streams used when spawning the system editor.
// A nil stream restores the corresponding process standard stream.
func (reg *Buffers) SetIO(in io.Reader, out, err io.Writer) {
	reg.in = in
	reg.out = out
	reg.errOut = err
}

func (reg *Buffers) input() io.Reader {
	if reg.in != nil {
		return reg.in
	}

	return os.Stdin
}

func (reg *Buffers) output() io.Writer {
	if reg.out != nil {
		return reg.out
	}

	return os.Stdout
}

func (reg *Buffers) errorOutput() io.Writer {
	if reg.errOut != nil {
		return reg.errOut
	}

	return os.Stderr
}

var (
	tallyRegisters = 10
	alphaRegisters = 52
)

func NewBuffers() *Buffers {
	return &Buffers{
		num:   make(map[int][]rune, tallyRegisters),
		alpha: make(map[rune][]rune, alphaRegisters),
		ro:    map[rune][]rune{},
		mutex: &sync.Mutex{},
	}
}

func (reg *Buffers) Active() []rune {
	defer reg.Reset()

	if !reg.waiting && !reg.selected {
		return reg.GetKill()
	}

	return reg.Get(reg.active)
}

func (reg *Buffers) Complete() completion.Values {

	numRegs := reg.completeNumRegs()
	alphaRegs := reg.completeAlphaRegs()

	vals := make([]completion.Candidate, 0, len(numRegs)+len(alphaRegs))
	vals = append(vals, numRegs...)
	vals = append(vals, alphaRegs...)

	comps := completion.AddRaw(vals)
	if comps.NoSort == nil {
		comps.NoSort = make(map[string]bool)
	}

	comps.NoSort["*"] = true

	if comps.ListLong == nil {
		comps.ListLong = make(map[string]bool)
	}

	comps.ListLong["*"] = true

	hint := color.Bold + color.FgBlue + "(registers)"

	if len(vals) == 0 {
		hint += " - empty -"
	}

	comps.Messages.Add(hint)

	return comps
}

func (reg *Buffers) Get(register rune) []rune {
	if register == 0 {
		return reg.GetKill()
	}

	num, err := strconv.Atoi(string(register))
	if err == nil {
		return reg.num[num]
	}

	if buf, found := reg.alpha[register]; found {
		return buf
	}

	if buf, found := reg.ro[register]; found {
		return buf
	}

	return nil
}

func (reg *Buffers) GetKill() []rune {
	if len(reg.num) == 0 {
		return nil
	}

	return reg.num[0]
}

func (reg *Buffers) IsSelected() (name string, selected bool) {
	return string(reg.active), reg.selected
}

func (reg *Buffers) Pop() []rune {
	buf, found := reg.num[0]
	if !found {
		return nil
	}

	for i := 1; i < tallyRegisters; i++ {
		if val, ok := reg.num[i]; ok {
			reg.num[i-1] = val
		} else {
			delete(reg.num, i-1)
		}
	}

	delete(reg.num, tallyRegisters-1)

	return []rune(string(buf))
}

func (reg *Buffers) Reset() {
	reg.active = 0
	reg.waiting = false
	reg.selected = false
}

func (reg *Buffers) SetActive(register rune) {
	defer func() {

		reg.waiting = false
		reg.selected = true
	}()

	num, err := strconv.Atoi(string(register))
	if err == nil && num < 10 {
		reg.active = register
		return
	}

	if _, found := reg.ro[register]; found {
		reg.active = register

		return
	}

	reg.active = register
}

func (reg *Buffers) Write(content ...rune) {
	buf := string(content)

	defer reg.Reset()

	if len(content) == 0 || buf == "" {
		return
	}

	if reg.selected {
		reg.WriteTo(reg.active, []rune(buf)...)
	} else {
		reg.writeNum(-1, []rune(buf))
	}
}

func (reg *Buffers) WriteTo(register rune, content ...rune) {
	buf := string(content)

	if len(content) == 0 || buf == "" {
		return
	}

	if register == 0 {
		reg.writeNum(0, []rune(buf))
		return
	}

	num, err := strconv.Atoi(string(register))
	if err == nil && num > 0 && num < tallyRegisters {
		reg.writeNum(num, []rune(buf))
		return
	}

	if unicode.IsLetter(register) {
		reg.writeAlpha(register, []rune(buf))
		return
	}
}

func (reg *Buffers) completeAlphaRegs() []completion.Candidate {
	tag := color.Dim + "alpha ([a-z], [A-Z])" + color.Reset

	lett := make([]rune, 0, len(reg.alpha))
	for slot := range reg.alpha {
		lett = append(lett, slot)
	}

	sort.Slice(lett, func(i, j int) bool { return lett[i] < lett[j] })

	regs := make([]completion.Candidate, 0, len(lett))

	for _, letter := range lett {
		buf := reg.alpha[letter]
		display := strings.ReplaceAll(string(buf), "\n", ` `)

		comp := completion.Candidate{
			Tag:     tag,
			Value:   string(buf),
			Display: fmt.Sprintf("%s\"%s%s %s", color.Dim, string(letter), color.DimReset, display),
		}

		regs = append(regs, comp)
	}

	return regs
}

func (reg *Buffers) completeNumRegs() []completion.Candidate {
	tag := color.Dim + "num ([0-9])" + color.Reset

	nums := make([]int, 0, len(reg.num))
	for reg := range reg.num {
		nums = append(nums, reg)
	}

	sort.Ints(nums)

	regs := make([]completion.Candidate, 0, len(nums))

	for _, num := range nums {
		buf := reg.num[num]
		display := strings.ReplaceAll(string(buf), "\n", ` `)

		comp := completion.Candidate{
			Tag:     tag,
			Value:   string(buf),
			Display: fmt.Sprintf("%s\"%d%s %s", color.Dim, num, color.DimReset, display),
		}

		regs = append(regs, comp)
	}

	return regs
}

func (reg *Buffers) writeAlpha(register rune, buf []rune) {
	appendRegs := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	appended := false

	for _, char := range appendRegs {
		if char == register {
			register = unicode.ToLower(register)
			_, exists := reg.alpha[register]

			if exists {
				reg.alpha[register] = append(reg.alpha[register], buf...)
			} else {
				reg.alpha[register] = buf
			}

			appended = true
		}
	}

	if !appended {
		reg.alpha[register] = buf
	}
}

func (reg *Buffers) writeNum(register int, buf []rune) {

	if register > tallyRegisters-1 {
		return
	}

	if register > 0 {
		reg.num[register] = buf

		return
	}

	for i := tallyRegisters - 1; i > 0; i-- {
		if val, ok := reg.num[i-1]; ok {
			reg.num[i] = val
		} else {
			delete(reg.num, i)
		}
	}

	reg.num[0] = append([]rune{}, buf...)
}
