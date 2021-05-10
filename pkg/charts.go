package pkg

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
)

func ServeCharts(simulation Simulation) {

	fmt.Println("serving charts")

	page := components.NewPage()
	fmt.Println()
	fmt.Println("productId", simulation.ProductId)
	fmt.Println("     from", simulation.From)
	fmt.Println("       to", simulation.To)
	fmt.Println("scenarios", len(simulation.Scenarios))
	fmt.Println("      won", simulation.Won)
	fmt.Println("     lost", simulation.Lost)
	fmt.Println("   result", simulation.sum())
	fmt.Println("   volume", simulation.Vol)
	fmt.Println("   return", simulation.result())
	fmt.Println()

	sort.SliceStable(simulation.Scenarios, func(i, j int) bool {
		return simulation.Scenarios[i].Result > simulation.Scenarios[j].Result
	})

	for _, play := range simulation.Scenarios[:25] {
		kline := charts.NewKLine()
		t := fmt.Sprintf("RESULT: %f\tENTER: %f\tEXIT: %f\tPEAK: %f\t", play.Result, play.Market, play.Entry, play.Exit)

		kline.SetGlobalOptions(
			charts.WithTitleOpts(opts.Title{
				Title: t,
			}),
			charts.WithXAxisOpts(opts.XAxis{
				SplitNumber: 20,
			}),
			charts.WithYAxisOpts(opts.YAxis{
				SplitNumber: 10,
				Scale:       true,
			}),
			charts.WithDataZoomOpts(opts.DataZoom{
				Start:      0,
				End:        100,
				XAxisIndex: []int{0},
			}),
		)
		x := make([]string, 0)
		y := make([]opts.KlineData, 0)
		for i := 0; i < len(play.Rates); i++ {
			x = append(x, play.Rates[i].Time().String())
			y = append(y, opts.KlineData{Value: play.Rates[i].Data()})
		}

		kline.SetXAxis(x).AddSeries("kline", y).
			SetSeriesOptions(
				charts.WithMarkPointStyleOpts(opts.MarkPointStyle{
					Label: &opts.Label{
						Show: true,
					},
				}),
				charts.WithItemStyleOpts(opts.ItemStyle{
					Color0:       "#ec0000",
					Color:        "#00da3c",
					BorderColor0: "#8A0000",
					BorderColor:  "#008F28",
				}),
			)
		page.AddCharts(kline)
	}

	name := fmt.Sprintf("./html/%s.html", simulation.ProductId)

	if f, err := os.Create(name); err != nil {
		panic(err)
	} else if err := page.Render(io.MultiWriter(f)); err != nil {
		panic(err)
	}

	fs := http.FileServer(http.Dir("html"))
	fmt.Println("served charts at http://localhost:8089")
	log.Fatal(http.ListenAndServe("localhost:8089", logRequest(fs)))

}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}
