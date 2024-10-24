// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package muxpress

import (
	"github.com/rlnas/xk6-mock-server/mock"
	"go.k6.io/k6/js/modules"
)

func init() { //nolint:gochecknoinits
	modules.Register("k6/x/mock", mock.New())
}
