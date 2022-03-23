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
	"fmt"
	"testing"
)

func testLPad(t *testing.T, r int64, l int) {
	fmtd := fmt.Sprintf("%0" + fmt.Sprintf("%d", l) + "d", r)
	if LPad(r, l)  != fmtd {
		t.Errorf("\"%s\" != \"%s\"", LPad(r, l), fmtd)
	}
}

func TestLPad(t *testing.T) {
	testLPad(t, 12345, 12)
	testLPad(t, 1234567, 10)
}

func BenchmarkLPad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = LPad(123456789, 10)
	}
}


func BenchmarkFmt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("%010d", 123456789)
	}
}
