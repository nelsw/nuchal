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
)

func CreateCharts() {

	page := components.NewPage()

	for _, play := range best {
		kline := createChart(play)
		addData(kline, play.Rates)
		page.AddCharts(kline)
	}

	name := fmt.Sprintf("./html/%s.html", target.ProductId)

	if f, err := os.Create(name); err != nil {
		panic(err)
	} else if err := page.Render(io.MultiWriter(f)); err != nil {
		panic(err)
	}

	fs := http.FileServer(http.Dir("html"))
	log.Println("running server at http://localhost:8089")
	log.Fatal(http.ListenAndServe("localhost:8089", logRequest(fs)))
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func addData(kline *charts.Kline, rates []Rate) {

	x := make([]string, 0)
	y := make([]opts.KlineData, 0)
	for i := 0; i < len(rates); i++ {
		x = append(x, rates[i].Time().String())
		y = append(y, opts.KlineData{Value: rates[i].Data()})
	}

	kline.SetXAxis(x).AddSeries("kline", y).
		SetSeriesOptions(
			charts.WithMarkPointStyleOpts(opts.MarkPointStyle{
				Label: &opts.Label{
					Show: true,
				},
			}),
			charts.WithItemStyleOpts(opts.ItemStyle{
				Color0:        "#ec0000",
				Color:       "#00da3c",
				BorderColor0:  "#8A0000",
				BorderColor: "#008F28",
			}),
		)
}

func createChart(play Play) *charts.Kline {
	kline := charts.NewKLine()
	kline.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: fmt.Sprintf("%f", play.Result),
			Subtitle: fmt.Sprintf("DURATION: %s\nENTER: %f\nEXIT: %f\n", play.Duration, play.Enter, play.Exit),
		}),
		charts.WithXAxisOpts(opts.XAxis{
			SplitNumber: 20,
		}),
		charts.WithYAxisOpts(opts.YAxis{
			SplitNumber: 10,
			Scale: true,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Start:      0,
			End:        100,
			XAxisIndex: []int{0},
		}),
	)
	return kline
}
