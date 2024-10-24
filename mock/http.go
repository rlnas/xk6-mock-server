// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package mock

import (
	"github.com/grafana/sobek"
	"go.k6.io/k6/js/common"
)

// XXX: add batch function support

var (
	urlFirstMethods  = []string{"get", "head", "post", "put", "patch", "options", "del"}
	urlSecondMethods = []string{"request", "asyncRequest"}
)

func (mod *Module) wrapHTTPExports(defaults *sobek.Object) {
	for _, method := range urlFirstMethods {
		mod.wrap(defaults, method, 0)
	}

	for _, method := range urlSecondMethods {
		mod.wrap(defaults, method, 1)
	}
}

func (mod *Module) parseBody(args []sobek.Value, index int) {
	// Extract request object and check type assertion
	reqObj, ok := args[index].(*sobek.Object)
	if !ok {
		//mod.logger.Error("Invalid request object: expected *sobek.Object")
		return // If the request object is invalid, skip body parsing silently
	}

	// Get the body, but don't enforce a type assertion
	bodyVal := reqObj.Get("body")

	// If there's no body or it's undefined, skip parsing
	if bodyVal == nil || bodyVal == sobek.Undefined() {
		return
	}

	// Check if the body is a string
	body, ok := bodyVal.Export().(string)
	if !ok {
		return // If the body isn't a string, skip parsing (optional behavior)
	}

	// No renaming: keep the body attribute and set the raw body
	reqObj.Set("body", mod.runtime().ToValue(body))
}

func (mod *Module) wrap(this *sobek.Object, method string, index int) {
	v := this.Get(method)

	callable, ok := sobek.AssertFunction(v)
	if !ok {
		mod.throwf("%s must be callable", errInvalidArg, method)
	}

	wrapper := func(call sobek.FunctionCall) sobek.Value {
		if len(call.Arguments) > index {
			mod.rewrite(call.Arguments, index)

			// Add body parsing here (new functionality)
			mod.parseBody(call.Arguments, index)
		}

		v, err := callable(mod.runtime().GlobalObject(), call.Arguments...)
		if err != nil {
			common.Throw(mod.runtime(), err)
		}

		return v
	}

	err := this.Set(method, mod.runtime().ToValue(wrapper))
	if err != nil {
		common.Throw(mod.runtime(), err)
	}
}
