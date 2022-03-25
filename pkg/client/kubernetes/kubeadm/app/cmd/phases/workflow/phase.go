package workflow

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Phase provides an implementation of a workflow phase that allows
// creation of new phases by simply instantiating a variable of this type.
type Phase struct {
	// name of the phase.
	// Phase name should be unique among peer phases (phases belonging to
	// the same workflow or phases belonging to the same parent phase).
	Name string

	// Aliases returns the aliases for the phase.
	Aliases []string

	// Short description of the phase.
	Short string

	// Long returns the long description of the phase.
	Long string

	// Example returns the example for the phase.
	Example string

	// Hidden define if the phase should be hidden in the workflow help.
	// e.g. PrintFilesIfDryRunning phase in the kubeadm init workflow is candidate for being hidden to the users
	Hidden bool

	// Phases defines a nested, ordered sequence of phases.
	Phases []Phase

	// RunAllSiblings allows to assign to a phase the responsibility to
	// run all the sibling phases
	// Nb. phase marked as RunAllSiblings can not have Run functions
	RunAllSiblings bool

	// Run defines a function implementing the phase action.
	// It is recommended to implent type assertion, e.g. using golang type switch,
	// for validating the RunData type.
	Run func(data RunData) error

	// RunIf define a function that implements a condition that should be checked
	// before executing the phase action.
	// If this function return nil, the phase action is always executed.
	RunIf func(data RunData) (bool, error)

	// InheritFlags defines the list of flags that the cobra command generated for this phase should Inherit
	// from local flags defined in the parent command / or additional flags defined in the phase runner.
	// If the values is not set or empty, no flags will be assigned to the command
	// Nb. global flags are automatically inherited by nested cobra command
	InheritFlags []string

	// LocalFlags defines the list of flags that should be assigned to the cobra command generated
	// for this phase.
	// Nb. if two or phases have the same local flags, please consider using local flags in the parent command
	// or additional flags defined in the phase runner.
	LocalFlags *pflag.FlagSet

	// ArgsValidator defines the positional arg function to be used for validating args for this phase
	// If not set a phase will adopt the args of the top level command.
	ArgsValidator cobra.PositionalArgs
}

// AppendPhase adds the given phase to the nested, ordered sequence of phases.
func (t *Phase) AppendPhase(phase Phase) {
	t.Phases = append(t.Phases, phase)
}
