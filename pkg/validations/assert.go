// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package validations

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/vmware-tanzu/carvel-ytt/pkg/filepos"
	"github.com/vmware-tanzu/carvel-ytt/pkg/template"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamlmeta"
)

// Declare @assert/... annotation names
const (
	AnnotationAssertValidate    template.AnnotationName = "assert/validate"
	ValidationKwargWhen         string                  = "when"
	ValidationKwargWhenNullSkip string                  = "when_null_skip"
	ValidationKwargMinLength    string                  = "min_len"
	ValidationKwargMaxLength    string                  = "max_len"
	ValidationKwargMin          string                  = "min"
	ValidationKwargMax          string                  = "max"
	ValidationKwargNotNull      string                  = "not_null"
	ValidationKwargOneNotNull   string                  = "one_not_null"
)

// ProcessAssertValidateAnns checks Assert annotations on data values and stores them on a Node as Validations.
// Returns an error if any Assert annotations are malformed.
func ProcessAssertValidateAnns(rootNode yamlmeta.Node) error {
	if rootNode == nil {
		return nil
	}
	return yamlmeta.Walk(rootNode, &convertAssertAnnsToValidations{})
}

type convertAssertAnnsToValidations struct{}

// Visit if `node` is annotated with `@assert/validate` (AnnotationAssertValidate).
// Checks annotation, and stores the validation on Node's validations meta.
//
// This visitor returns and error if any assert annotation is not well-formed,
// otherwise, returns nil.
func (a *convertAssertAnnsToValidations) Visit(node yamlmeta.Node) error {
	nodeAnnotations := template.NewAnnotations(node)
	if !nodeAnnotations.Has(AnnotationAssertValidate) {
		return nil
	}
	switch node.(type) {
	case *yamlmeta.DocumentSet, *yamlmeta.Array, *yamlmeta.Map:
		return fmt.Errorf("Invalid @%s annotation - not supported on %s at %s", AnnotationAssertValidate, yamlmeta.TypeName(node), node.GetPosition().AsCompactString())
	default:
		validation, err := NewValidationFromValidationAnnotation(nodeAnnotations[AnnotationAssertValidate])
		if err != nil {
			return fmt.Errorf("Invalid @%s annotation - %s", AnnotationAssertValidate, err.Error())
		}
		// store rules in node's validations meta without overriding any existing rules
		Add(node, []NodeValidation{*validation})
	}

	return nil
}

// NewValidationFromValidationAnnotation creates a NodeValidation from the values provided in a validation annotation.
// If any value in the annotation is not well-formed, it returns an error.
func NewValidationFromValidationAnnotation(annotation template.NodeAnnotation) (*NodeValidation, error) {
	var rules []rule
	if len(annotation.Args) == 0 && len(annotation.Kwargs) == 0 {
		return nil, fmt.Errorf("expected annotation to have 2-tuple as argument(s), but found no arguments (by %s)", annotation.Position.AsCompactString())
	}
	for _, arg := range annotation.Args {
		ruleTuple, ok := arg.(starlark.Tuple)
		if !ok {
			return nil, fmt.Errorf("expected annotation to have 2-tuple as argument(s), but found: %s (by %s)", arg.String(), annotation.Position.AsCompactString())
		}
		if len(ruleTuple) != 2 {
			return nil, fmt.Errorf("expected 2-tuple, but found tuple with length %v (by %s)", len(ruleTuple), annotation.Position.AsCompactString())
		}

		message, ok := ruleTuple[0].(starlark.String)
		if !ok {
			return nil, fmt.Errorf("expected first item in the 2-tuple to be a string describing a valid value, but was %s (at %s)", ruleTuple[0].Type(), annotation.Position.AsCompactString())
		}

		assertion, ok := ruleTuple[1].(starlark.Callable)
		if !ok {
			var err error
			assertion, err = assertionFromCheckAttr(ruleTuple[1])
			if err != nil {
				return nil, fmt.Errorf("%s (at %s)", err, annotation.Position.AsCompactString())
			}
		}
		rules = append(rules, rule{
			msg:       message.GoString(),
			assertion: assertion,
		})
	}
	kwargs, err := newValidationKwargs(annotation.Kwargs, annotation.Position)
	if err != nil {
		return nil, err
	}

	rules = append(rules, kwargs.convertToRules()...)

	return &NodeValidation{rules, kwargs, annotation.Position}, nil
}

func assertionFromCheckAttr(value starlark.Value) (starlark.Callable, error) {
	val, hasAttrs := value.(starlark.HasAttrs)
	if !hasAttrs {
		return nil, fmt.Errorf("expected second item in the 2-tuple to be an assertion function, but was %s", value.Type())
	}

	checkAttr, err := val.Attr("check")
	if err != nil {
		return nil, fmt.Errorf("expected second item in the 2-tuple to be an assertion function or assertion object, but was %s", value.Type())
	}

	assertionFunc, ok := checkAttr.(starlark.Callable)
	if !ok {
		return nil, fmt.Errorf("expected assertion object with assertion function \"check()\" , but was %s", checkAttr.Type())
	}

	return assertionFunc, nil
}

// newValidationKwargs takes the keyword arguments from a Validation annotation,
// and makes sure they are well-formed.
func newValidationKwargs(kwargs []starlark.Tuple, annPos *filepos.Position) (validationKwargs, error) {
	var processedKwargs validationKwargs
	for _, value := range kwargs {
		kwargName := string(value[0].(starlark.String))
		switch kwargName {
		case ValidationKwargWhen:
			v, ok := value[1].(starlark.Callable)
			if !ok {
				return validationKwargs{}, fmt.Errorf("expected keyword argument %q to be a function, but was %s (at %s)", ValidationKwargWhen, value[1].Type(), annPos.AsCompactString())
			}
			processedKwargs.when = v
		case ValidationKwargWhenNullSkip:
			v, ok := value[1].(starlark.Bool)
			if !ok {
				return validationKwargs{}, fmt.Errorf("expected keyword argument %q to be a boolean, but was %s (at %s)", ValidationKwargWhenNullSkip, value[1].Type(), annPos.AsCompactString())
			}
			b := bool(v)
			processedKwargs.whenNullSkip = &b
		case ValidationKwargMinLength:
			v, err := starlark.NumberToInt(value[1])
			if err != nil {
				return validationKwargs{}, fmt.Errorf("expected keyword argument %q to be a number, but was %s (at %s)", ValidationKwargMinLength, value[1].Type(), annPos.AsCompactString())
			}
			processedKwargs.minLength = &v
		case ValidationKwargMaxLength:
			v, err := starlark.NumberToInt(value[1])
			if err != nil {
				return validationKwargs{}, fmt.Errorf("expected keyword argument %q to be a number, but was %s (at %s)", ValidationKwargMaxLength, value[1].Type(), annPos.AsCompactString())
			}
			processedKwargs.maxLength = &v
		case ValidationKwargMin:
			processedKwargs.min = value[1]
		case ValidationKwargMax:
			processedKwargs.max = value[1]
		case ValidationKwargNotNull:
			v, ok := value[1].(starlark.Bool)
			if !ok {
				return validationKwargs{}, fmt.Errorf("expected keyword argument %q to be a boolean, but was %s (at %s)", ValidationKwargNotNull, value[1].Type(), annPos.AsCompactString())
			}
			processedKwargs.notNull = bool(v)
		case ValidationKwargOneNotNull:
			switch v := value[1].(type) {
			case starlark.Bool:
				if v {
					processedKwargs.oneNotNull = v
				} else {
					return validationKwargs{}, fmt.Errorf("one_not_null= cannot be False")
				}
			case starlark.Sequence:
				processedKwargs.oneNotNull = v
			default:
				return validationKwargs{}, fmt.Errorf("expected True or a sequence of keys, but was a \"%s\"", value[1].Type())
			}
		default:
			return validationKwargs{}, fmt.Errorf("unknown keyword argument %q (at %s)", kwargName, annPos.AsCompactString())
		}
	}
	return processedKwargs, nil
}
