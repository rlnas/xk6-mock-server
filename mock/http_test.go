// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package mock

import (
	"testing"

	"github.com/grafana/sobek"
	"github.com/stretchr/testify/assert"
)

func TestModuleWrap(t *testing.T) {
	t.Parallel()

	helper := newHelper(t)

	target := helper.vu.Runtime().NewObject()

	assert.NoError(t, target.Set("not_a_func", 1))

	assert.Panics(t, func() { helper.module.wrap(target, "not_a_func", 1) })

	var actual string

	method := func(loc string) {
		actual = loc
	}

	assert.NoError(t, target.Set("method", method))

	helper.module.lookup["https://example.com"] = "https://example.net"

	helper.module.wrap(target, "method", 0)

	callable, ok := sobek.AssertFunction(target.Get("method"))

	assert.True(t, ok)

	_, err := callable(sobek.Undefined(), helper.vu.Runtime().ToValue("https://example.com"))

	assert.NoError(t, err)
	assert.Equal(t, "https://example.net", actual)
}

func TestParseBodyWithStringBody(t *testing.T) {
	t.Parallel()

	helper := newHelper(t)

	// Create a mock request object with a valid string body
	reqObj := helper.vu.Runtime().NewObject()
	assert.NoError(t, reqObj.Set("body", "raw body content"))

	args := []sobek.Value{reqObj}

	// Call the parseBody function
	helper.module.parseBody(args, 0)

	// Check if the body is still "raw body content"
	body := reqObj.Get("body").Export().(string)
	assert.Equal(t, "raw body content", body)
}

func TestParseBodyWithUndefinedBody(t *testing.T) {
	t.Parallel()

	helper := newHelper(t)

	// Create a mock request object with an undefined body
	reqObj := helper.vu.Runtime().NewObject()
	assert.NoError(t, reqObj.Set("body", sobek.Undefined()))

	args := []sobek.Value{reqObj}

	// Call the parseBody function
	helper.module.parseBody(args, 0)

	// Ensure the body is still undefined (i.e., no processing was done)
	body := reqObj.Get("body")
	assert.Equal(t, sobek.Undefined(), body)
}

func TestParseBodyWithNonStringBody(t *testing.T) {
	t.Parallel()

	helper := newHelper(t)

	// Create a mock request object with a non-string body (e.g., an object)
	reqObj := helper.vu.Runtime().NewObject()
	objBody := helper.vu.Runtime().NewObject()
	assert.NoError(t, reqObj.Set("body", objBody))

	args := []sobek.Value{reqObj}

	// Call the parseBody function
	helper.module.parseBody(args, 0)

	// Ensure the body is still the object (i.e., no processing was done)
	body := reqObj.Get("body")
	assert.Equal(t, objBody, body)
}

func TestParseBodyWhenBodyNotPresent(t *testing.T) {
	t.Parallel()

	helper := newHelper(t)

	// Create a mock request object without setting a body
	reqObj := helper.vu.Runtime().NewObject()

	args := []sobek.Value{reqObj}

	// Call the parseBody function
	helper.module.parseBody(args, 0)

	// Ensure that the body is still nil or undefined
	body := reqObj.Get("body")
	assert.Equal(t, nil, body) // or sobek.Undefined(), depending on the environment
}

func TestParseBodyWithXMLBody(t *testing.T) {
	t.Parallel()

	helper := newHelper(t)

	// Create a mock request object with an XML body
	reqObj := helper.vu.Runtime().NewObject()
	xmlBody := `<person><name>John Doe</name><age>30</age></person>`
	assert.NoError(t, reqObj.Set("body", xmlBody))

	args := []sobek.Value{reqObj}

	// Call the parseBody function to handle the XML body
	helper.module.parseBody(args, 0)

	// Check if the body is still the XML string (or assert if further parsing is required)
	body := reqObj.Get("body").Export().(string)
	assert.Equal(t, xmlBody, body)
}

func TestModuleWrapWithXMLBody(t *testing.T) {
	t.Parallel()

	helper := newHelper(t)

	target := helper.vu.Runtime().NewObject()

	// Create the method to simulate an HTTP POST request with an XML body
	var actualBody string
	method := func(body string) {
		actualBody = body
	}

	assert.NoError(t, target.Set("postMethod", method))

	// Wrap the method
	helper.module.wrap(target, "postMethod", 0)

	callable, ok := sobek.AssertFunction(target.Get("postMethod"))
	assert.True(t, ok)

	// XML body to be passed
	xmlBody := `<person><name>John Doe</name><age>30</age></person>`

	// Call the wrapped method with the XML body as argument
	_, err := callable(sobek.Undefined(), helper.vu.Runtime().ToValue(xmlBody))
	assert.NoError(t, err)

	// Check that the XML body was passed correctly
	assert.Equal(t, xmlBody, actualBody)
}

func TestModuleWrapWithJSONBody(t *testing.T) {
	t.Parallel()

	helper := newHelper(t)

	target := helper.vu.Runtime().NewObject()

	// Create the method to simulate an HTTP POST request with a JSON body
	var actualBody string
	method := func(body string) {
		actualBody = body
	}

	assert.NoError(t, target.Set("postMethod", method))

	// Wrap the method
	helper.module.wrap(target, "postMethod", 0)

	callable, ok := sobek.AssertFunction(target.Get("postMethod"))
	assert.True(t, ok)

	// JSON body to be passed
	jsonBody := `{"name": "John Doe", "age": 30}`

	// Call the wrapped method with the JSON body as argument
	_, err := callable(sobek.Undefined(), helper.vu.Runtime().ToValue(jsonBody))
	assert.NoError(t, err)

	// Check that the JSON body was passed correctly
	assert.Equal(t, jsonBody, actualBody)
}

func TestModuleWrapWithFullXMLRequestInfo(t *testing.T) {
	t.Parallel()

	helper := newHelper(t)

	target := helper.vu.Runtime().NewObject()

	// Create a method to capture the full request information
	var capturedRequest struct {
		Method  string
		Headers map[string]string
		Body    string
	}

	// Simulate capturing the full request data
	method := func(method string, headers map[string]string, body string) {
		capturedRequest.Method = method
		capturedRequest.Headers = headers
		capturedRequest.Body = body
	}

	assert.NoError(t, target.Set("postMethod", method))

	// Wrap the method
	helper.module.wrap(target, "postMethod", 0)

	callable, ok := sobek.AssertFunction(target.Get("postMethod"))
	assert.True(t, ok)

	// XML body to be passed
	xmlBody := `<person><name>John Doe</name><age>30</age></person>`

	// Define headers to simulate an HTTP POST request
	headers := map[string]string{
		"Content-Type": "application/xml",
		"Accept":       "application/xml",
	}

	// Call the wrapped method with the method type, headers, and XML body as arguments
	_, err := callable(
		sobek.Undefined(),
		helper.vu.Runtime().ToValue("POST"),
		helper.vu.Runtime().ToValue(headers),
		helper.vu.Runtime().ToValue(xmlBody),
	)
	assert.NoError(t, err)

	// Check if the method is POST
	assert.Equal(t, "POST", capturedRequest.Method)

	// Check if the headers are correctly set
	assert.Equal(t, "application/xml", capturedRequest.Headers["Content-Type"])
	assert.Equal(t, "application/xml", capturedRequest.Headers["Accept"])

	// Check if the XML body is correctly passed
	assert.Equal(t, xmlBody, capturedRequest.Body)
}

func TestModuleWrapWithFullJSONRequestInfo(t *testing.T) {
	t.Parallel()

	helper := newHelper(t)

	target := helper.vu.Runtime().NewObject()

	// Create a method to capture the full request information
	var capturedRequest struct {
		Method  string
		Headers map[string]string
		Body    string
	}

	// Simulate capturing the full request data
	method := func(method string, headers map[string]string, body string) {
		capturedRequest.Method = method
		capturedRequest.Headers = headers
		capturedRequest.Body = body
	}

	assert.NoError(t, target.Set("postMethod", method))

	// Wrap the method
	helper.module.wrap(target, "postMethod", 0)

	callable, ok := sobek.AssertFunction(target.Get("postMethod"))
	assert.True(t, ok)

	// JSON body to be passed
	jsonBody := `{"name": "John Doe", "age": 30}`

	// Define headers to simulate an HTTP POST request
	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}

	// Call the wrapped method with the method type, headers, and JSON body as arguments
	_, err := callable(
		sobek.Undefined(),
		helper.vu.Runtime().ToValue("POST"),
		helper.vu.Runtime().ToValue(headers),
		helper.vu.Runtime().ToValue(jsonBody),
	)
	assert.NoError(t, err)

	// Check if the method is POST
	assert.Equal(t, "POST", capturedRequest.Method)

	// Check if the headers are correctly set
	assert.Equal(t, "application/json", capturedRequest.Headers["Content-Type"])
	assert.Equal(t, "application/json", capturedRequest.Headers["Accept"])

	// Check if the JSON body is correctly passed
	assert.Equal(t, jsonBody, capturedRequest.Body)
}

func TestModuleWrapWithXMLBodyNoHeaders(t *testing.T) {
	t.Parallel()

	helper := newHelper(t)

	target := helper.vu.Runtime().NewObject()

	// Create a method to capture the full request information (without headers)
	var capturedRequest struct {
		Method  string
		Headers map[string]string
		Body    string
	}

	// Simulate capturing the full request data (without headers)
	method := func(method string, headers map[string]string, body string) {
		capturedRequest.Method = method
		capturedRequest.Headers = headers
		capturedRequest.Body = body
	}

	assert.NoError(t, target.Set("postMethod", method))

	// Wrap the method
	helper.module.wrap(target, "postMethod", 0)

	callable, ok := sobek.AssertFunction(target.Get("postMethod"))
	assert.True(t, ok)

	// XML body to be passed (no headers)
	xmlBody := `<person><name>John Doe</name><age>30</age></person>`

	// Call the wrapped method with the method type and XML body, but without headers
	_, err := callable(
		sobek.Undefined(),
		helper.vu.Runtime().ToValue("POST"),
		helper.vu.Runtime().ToValue(map[string]string{}), // No headers
		helper.vu.Runtime().ToValue(xmlBody),
	)
	assert.NoError(t, err)

	// Check if the method is POST
	assert.Equal(t, "POST", capturedRequest.Method)

	// Check that no headers were passed
	assert.Empty(t, capturedRequest.Headers)

	// Check if the XML body is correctly passed
	assert.Equal(t, xmlBody, capturedRequest.Body)
}
