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

package sim

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/render"
	"github.com/nelsw/nuchal/pkg/util"
	"io"
	"os"
	"sort"
	"strings"
)

func newPage(productID, symbol, dir string, charts []Chart) error {

	if len(charts) < 1 {
		return nil
	}

	page := &components.Page{}
	page.Assets.InitAssets()
	page.Renderer = render.NewPageRender(page, page.Validate)
	page.Layout = components.PageFlexLayout
	page.PageTitle = "nuchal  " + symbol + "  " + strings.TrimSpace(util.GetCurrency(productID))

	sort.SliceStable(charts, func(i, j int) bool {
		return charts[i].result() > charts[j].result()
	})

	for _, s := range charts {
		page.AddCharts(s.kline())
	}

	if err := util.MakePath("html/" + productID); err != nil {
		return err
	} else if f, err := os.Create(fmt.Sprintf("./html/%s/%s.html", productID, dir)); err != nil {
		return err
	} else if err := page.Render(io.MultiWriter(f)); err != nil {
		return err
	}

	return nil
}
