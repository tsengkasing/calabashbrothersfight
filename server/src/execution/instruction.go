package execution

import (
	"reflect"
)

type Instruction interface {
	GetCode() string
	GetDescription() string
	GetName() string
	GetExpandInstructions() []Instruction
	Execute(gc *GlobalContext, tc *ThreadContext)
	IsBlocking(gc *GlobalContext, tc *ThreadContext) bool
}

type baseInstruction struct {
	Code               string
	Description        string
	Name               string
	ExpandInstructions []Instruction
}

func (i *baseInstruction) GetCode() string {
	return i.Code
}

func (i *baseInstruction) GetDescription() string {
	return i.Description
}

func (i *baseInstruction) GetName() string {
	return i.Name
}

func (i *baseInstruction) GetExpandInstructions() []Instruction {
	return i.ExpandInstructions
}

func (i *baseInstruction) IsBlocking(gc *GlobalContext, tc *ThreadContext) bool {
	return false
}

func moveToNextInstruction(tc *ThreadContext) {
	if tc.Expanded {
		tc.ExpProgramCounter++

		expandInsLen := len(tc.Instructions[tc.ProgramCounter].GetExpandInstructions())
		// End of expandInstruction
		if tc.ExpProgramCounter >= expandInsLen {
			tc.Expanded = false
			tc.ProgramCounter++
			tc.ExpProgramCounter = 0
		}
	} else {
		tc.ProgramCounter++
		tc.ExpProgramCounter = 0
	}
}

type CommentInstruction struct {
	baseInstruction
}

type DummyInstruction struct {
	baseInstruction
}

type IfInstruction struct {
	baseInstruction
	exp Expression
}

type EndIfInstruction struct {
	baseInstruction
}

type ExpandableInstruction struct {
	baseInstruction
}

type AtomicAssignmentToTemp struct {
	baseInstruction
	expr Expression
}

type AtomicAssignmentFromTemp struct {
	baseInstruction
	expr Expression
}

func NewExpandableInstruction(name string, expandIns []Instruction) ExpandableInstruction {
	return ExpandableInstruction{}
}

func NewAtomicAssignFromTemp(exprLeft Expression) AtomicAssignmentFromTemp {
	base := baseInstruction{
		Code:        InstructionExpr(AssignmentExpr(exprLeft, &NewVariableExpression("temp"))),
		Description: "Atomic assign temp to left expr",
	}
	return AtomicAssignmentFromTemp{base, exprLeft}
}

func NewAtomicAssignToTemp(exprRight Expression) AtomicAssignmentToTemp {
	base := baseInstruction{
		Code:        InstructionExpr(AssignmentExpr(&NewVariableExpression("temp"), exprRight)),
		Description: "Atomic assign right expr to temp",
	}
	return AtomicAssignmentToTemp{base, exprRight}
}

func NewAssignmentStatment(varName string, exp Expression) ExpandableInstruction {
	expInsList := []Instruction{
		&NewAtomicAssignToTemp(exp),
		&NewAtomicAssignFromTemp(&NewVariableExpression(varName)),
	}
	
	expIns := NewExpandableInstruction(
		AssignmentExpr(&NewVariableExpression(varName), exp), expInsList)

	return expIns
}

func NewStartIfStatment(exp Expression, name string) IfInstruction {
	base := baseInstruction{
		Code:        IfStart() + AddBraces(exp) + Then(),
		Description: "If statement",
		Name:        name,
	}
	return IfInstruction{base, exp}
}

func NewEndIfStatment(name string) EndIfInstruction {
	base := baseInstruction{
		Code:        End(),
		Description: "End if statement",
		Name:        name,
	}
	return EndIfInstruction{base}
}

func (e *AtomicAssignmentFromTemp) Execute(gc *GlobalContext, tc *ThreadContext) {
	gc.values[string(e.expr.Evaluate(gc, tc))] =
		GlobalStateType{value: tc.TempVariable, name: e.expr.GetName()}
}

func (e *AtomicAssignmentToTemp) Execute(gc *GlobalContext, tc *ThreadContext) {
	// TODO: Do we really need to evaluate?
	tc.TempVariable = e.expr.Evaluate(gc, tc)
	moveToNextInstruction(tc)
}

func (e *EndIfInstruction) Execute(gc *GlobalContext, tc *ThreadContext) {
	moveToNextInstruction(tc)
}

func (c *CommentInstruction) Execute(gc *GlobalContext, tc *ThreadContext) {
	moveToNextInstruction(tc)
}

func (d *DummyInstruction) Execute(gc *GlobalContext, tc *ThreadContext) {
	moveToNextInstruction(tc)
}

func (i *IfInstruction) Execute(gc *GlobalContext, tc *ThreadContext) {
	if (i.exp.Evaluate(gc, tc)).(bool) {
		moveToNextInstruction(tc)
	} else {
		matchingEndIf := findMatchingInsIndex(tc, i.GetName(), reflect.TypeOf(EndIfInstruction{}))
		goToInstruction(tc, matchingEndIf)
	}
}

func goToInstruction(context *ThreadContext, num int) {
	context.ProgramCounter = num
	context.ExpProgramCounter = 0
}

func findMatchingInsIndex(context *ThreadContext, name string, tp reflect.Type) int {
	for i, ins := range context.Instructions {
		if reflect.TypeOf(ins) == tp && ins.GetName() == name {
			return i
		}
	}
	panic("挂了吧")
	return -1
}

func IfStart() string {
	return "if"
}

func AddBraces(expCode Expression) string {
	return " (" + expCode.GetCode() + ") "
}

func Then() string {
	return "{"
}

func End() string {
	return "}"
}

func InstructionExpr(code string) string {
	return code + ";"
}

func AssignmentExpr(left Expression, right Expression) string {
	return BinaryOperationCode(left, right, "=")
}
