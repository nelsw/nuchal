package model

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"time"
)

type Chart struct {
	Rates    []Rate
	Duration time.Duration
	Entry,
	Goal,
	Exit,
	Result,
	Size float64
	FoundExit bool
}

func (c Chart) Volume() float64 {
	return c.Entry * c.Size
}

func (c Chart) Kline() *charts.Kline {

	kline := charts.NewKLine()

	title := fmt.Sprintf("RESULT: %f\tBOUGHT: %f\tGOAL: %f\tSOLD: %f\tVOL: %f",
		c.Result, c.Entry, c.Goal, c.Exit, c.Volume())
	subtitle := fmt.Sprintf("Duration: %s\t Exited: %v\t", c.Duration.String(), c.FoundExit)

	kline.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    title,
			Subtitle: subtitle,
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
	for i := 0; i < len(c.Rates); i++ {
		x = append(x, c.Rates[i].Time().String())
		y = append(y, opts.KlineData{Value: c.Rates[i].Data()})
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
	return kline
}
