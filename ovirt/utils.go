// Copyright (C) 2018 Joey Ma <majunjiev@gmail.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt

import (
	"strings"
)

func extractSemanticVerion(version string) (major, minor, patch string) {
	vs := strings.Split(version, ".")
	switch len(vs) {
	case 1:
		return vs[0], "", ""
	case 2:
		return vs[0], vs[1], ""
	case 3:
		return vs[0], vs[1], vs[2]
	default:
		return "", "", ""
	}
}
