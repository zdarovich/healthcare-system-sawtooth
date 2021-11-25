package main

import (
	"encoding/csv"
	"github.com/go-echarts/go-echarts/v2/components"
	"io"
	"os"
	"sort"
	"strconv"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

const (
	Bar1Name string = "1 Client"
	Bar2Name string = "3 Client"
	Bar3Name string = "5 Client"
)

func generateBarItemsFloat(data []float64) []opts.BarData {
	items := make([]opts.BarData, 0)
	for _, v := range data {
		items = append(items, opts.BarData{Value: v})
	}
	return items
}

func generateBarItemsInt(data []int) []opts.BarData {
	items := make([]opts.BarData, 0)
	for _, v := range data {
		items = append(items, opts.BarData{Value: v})
	}
	return items
}

func main() {

	filePath := "test/report/2021-11-25T15:57:04-Test_1_User_Register.csv"

	_, ytps1, ylat1, ymem1 := GetData(filePath)

	filePath = "test/report/2021-11-25T16:02:18-Test_3_User_Register.csv"

	x2, ytps3, ylat3, ymem3 := GetData(filePath)

	filePath = "test/report/2021-11-25T15:58:42-Test_5_User_Register.csv"

	_, ytps5, ylat5, ymem5 := GetData(filePath)

	page := components.NewPage()
	page.AddCharts(
		Get1510ClientThroughputSendRatechart(x2, ytps1, ytps3, ytps5),
		Get1510ClientLatencySendRatechart(x2, ylat1, ylat3, ylat5),
		Get1510ClientMemorySendRatechart(x2, ymem1, ymem3, ymem5),
	)
	f2, _ := os.Create("resources/charts/1_3_5_user_register_chart_bar.html")
	page.Render(io.MultiWriter(f2))
}

func GetData(filePath string) ([]int, []float64, []float64, []int) {
	f2, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer f2.Close()
	csvReader := csv.NewReader(f2)
	records, err := csvReader.ReadAll()
	if err != nil {
		panic(err)
	}
	var xAxis []int
	var yAxisTps []float64
	var yAxisLatency []float64
	var yAxisMemory []int
	for _, r := range records {
		latency := r[1]
		tps := r[2]
		memory := r[3]
		sendRate := r[4]

		tpsF, _ := strconv.ParseFloat(tps, 64)

		sendRateF, _ := strconv.ParseFloat(sendRate, 64)
		latencyF, _ := strconv.ParseFloat(latency, 64)
		memoryI, _ := strconv.Atoi(memory)

		xAxis = append(xAxis, int(sendRateF))
		yAxisTps = append(yAxisTps, tpsF)
		yAxisMemory = append(yAxisMemory, memoryI)
		yAxisLatency = append(yAxisLatency, latencyF)
	}
	sort.Ints(xAxis)

	return xAxis, yAxisTps, yAxisLatency, yAxisMemory
}

func Get1510ClientThroughputSendRatechart(xAxis1 []int, yAxisData1, yAxisData5, yAxisData10 []float64) *charts.Bar {
	// create a new bar instance
	bar := charts.NewBar()
	// set some global options like Title/Legend/ToolTip or anything else
	bar.SetGlobalOptions(
		charts.WithLegendOpts(opts.Legend{
			Show: true,
		}),
		charts.WithTitleOpts(opts.Title{
			Title: "User register throughput performance",
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "3300px",
			Height: "600px",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Send rate(bytes/seconds)",
			SplitLine: &opts.SplitLine{
				Show: true,
			},
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "Throughput(tps)",
			SplitLine: &opts.SplitLine{
				Show: true,
			},
		}),
	)

	// Put data into instance
	bar.SetXAxis(xAxis1).
		AddSeries(Bar1Name, generateBarItemsFloat(yAxisData1)).
		AddSeries(Bar2Name, generateBarItemsFloat(yAxisData5)).
		AddSeries(Bar3Name, generateBarItemsFloat(yAxisData10)).
		SetSeriesOptions()

	return bar
}

func Get1510ClientLatencySendRatechart(xAxis1 []int, yAxisData1, yAxisData5, yAxisData10 []float64) *charts.Bar {
	// create a new bar instance
	bar := charts.NewBar()
	// set some global options like Title/Legend/ToolTip or anything else
	bar.SetGlobalOptions(
		charts.WithLegendOpts(opts.Legend{
			Show: true,
		}),
		charts.WithTitleOpts(opts.Title{
			Title: "User register latency performance",
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "3300px",
			Height: "600px",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Send rate(bytes/seconds)",
			SplitLine: &opts.SplitLine{
				Show: true,
			},
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "Latency(seconds)",
			SplitLine: &opts.SplitLine{
				Show: true,
			},
		}),
	)

	bar.SetXAxis(xAxis1).
		AddSeries(Bar1Name, generateBarItemsFloat(yAxisData1)).
		AddSeries(Bar2Name, generateBarItemsFloat(yAxisData5)).
		AddSeries(Bar3Name, generateBarItemsFloat(yAxisData10)).
		SetSeriesOptions()
	return bar
}

func Get1510ClientMemorySendRatechart(xAxis1 []int, yAxisData1, yAxisData5, yAxisData10 []int) *charts.Bar {
	// create a new bar instance
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithLegendOpts(opts.Legend{
			Show: true,
		}),
		charts.WithTitleOpts(opts.Title{
			Title: "User register memory performance",
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "3300px",
			Height: "600px",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Send rate(bytes/seconds)",
			SplitLine: &opts.SplitLine{
				Show: true,
			},
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "Memory(bytes)",
			SplitLine: &opts.SplitLine{
				Show: true,
			},
		}),
	)

	bar.SetXAxis(xAxis1).
		AddSeries(Bar1Name, generateBarItemsInt(yAxisData1)).
		AddSeries(Bar2Name, generateBarItemsInt(yAxisData5)).
		AddSeries(Bar3Name, generateBarItemsInt(yAxisData10)).
		SetSeriesOptions()
	return bar
}
