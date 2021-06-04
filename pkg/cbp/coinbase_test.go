/*
 *
 * Copyright © 2021 Connor Van Elswyk ConnorVanElswyk@gmail.com
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package cbp

import (
	"github.com/nelsw/nuchal/pkg/config"
	"github.com/nelsw/nuchal/pkg/util"
	"github.com/nelsw/nuchal/test"
	"testing"
)

func TestApi_GetProducts(t *testing.T) {
	ses, err := config.NewSession(test.Usd, test.Size, test.Gain, test.Loss, test.Delta)
	if err != nil {
		panic(err)
	}
	if p, err := ses.GetActivePositions(); err != nil {
		t.Error(err)
	} else {
		util.PrintlnPrettyJson(p)
	}
}
