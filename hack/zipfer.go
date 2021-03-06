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
	"math/rand"
	"time"
)

func main() {
	src := rand.NewSource(time.Now().Unix())
	randObj := rand.New(src)
	size := 500000
	z := rand.NewZipf(randObj, 2.0, 1.0, uint64(size*8))
	counter := make(map[uint64]uint64)
	for i := 0; i < size*4; i++ {
		zOut := z.Uint64()
		counter[zOut]++
	}

	fmt.Println(counter)
}
