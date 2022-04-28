// Copyright (c) 2022 Viant Inc.
//
//    Licensed under the Apache License, Version 2.0 (the "License"); you may
//    not use this file except in compliance with the License. You may obtain
//    a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
//    WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
//    License for the specific language governing permissions and limitations
//    under the License.
package hack

import (
	"math"
)

func LPad(i int64, l int) string {
	m := int64(math.Pow(10, float64(l-1)))
	bs := make([]byte, l, l)
	ri := int64(i)
	for p := 0; p < l; p++ {
		c := ri / m
		bs[p] = byte(c + '0')
		ri = ri % m
		m = m / 10
	}

	return string(bs)
}
