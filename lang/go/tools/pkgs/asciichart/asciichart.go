// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2019-2022 Intel Corporation
//
// Modified in 2019 from https://github.com/guptarohit/asciigraph

package asciichart

import (
	"bytes"
	"fmt"
	"math"
	"strings"
)

// Chart - ascii chart structure
type Chart struct {
	config PlotConfig
}

const (
	asciichartLog = "AsciiChartLogID"
)

// New Chart structure
func New() *Chart {

	chart := &Chart{}

	chart.config.Precision = 2
	chart.config.AddColor = true

	return chart
}

func createEmpty(rows, width int) [][]string {
	var plot [][]string

	// initialise empty 2D grid
	for i := 0; i < rows+1; i++ {
		var line []string

		for j := 0; j < width; j++ {
			line = append(line, " ")
		}
		plot = append(plot, line)
	}

	return plot
}

// Plot returns ascii graph for a series.
func (ac *Chart) Plot(series []float64) string {
	var logMaximum float64

	if len(series) <= 0 {
		return ""
	}
	config := &ac.config

	if config.Width > 0 {
		series = interpolateArray(series, config.Width)
	}

	minimum, maximum := minMaxFloat64Slice(series)

	if config.Min != 0.0 && config.Min < minimum {
		minimum = config.Min
	}
	if config.Max != 0.0 && config.Max > maximum {
		maximum = config.Max
	}

	interval := math.Abs(maximum - minimum)

	if config.Height <= 0 {
		if int(interval) <= 0 {
			config.Height = int(interval * math.Pow10(int(math.Ceil(-math.Log10(interval)))))
		} else {
			config.Height = int(interval)
		}
	}

	if config.Offset <= 0 {
		config.Offset = 3
	}

	var ratio float64
	if interval != 0 {
		ratio = float64(config.Height) / interval
	} else {
		ratio = 1
	}
	min2 := round(minimum * ratio)
	max2 := round(maximum * ratio)

	intmin2 := int(min2)
	intmax2 := int(max2)

	rows := int(math.Abs(float64(intmax2 - intmin2)))
	width := len(series) + config.Offset

	plot := createEmpty(rows, width)

	precision := config.Precision
	logMaximum = math.Log10(math.Max(math.Abs(maximum), math.Abs(minimum))) //to find number of zeros after decimal
	if minimum == float64(0) && maximum == float64(0) {
		logMaximum = float64(-1)
	}

	if logMaximum < 0 {
		// negative log
		if math.Mod(logMaximum, 1) != 0 {
			// non-zero digits after decimal
			precision = precision + int(math.Abs(logMaximum))
		} else {
			precision = precision + int(math.Abs(logMaximum)-1.0)
		}
	} else if logMaximum > 2 {
		precision = 0
	}

	maxNumLength := len(fmt.Sprintf("%0.*f", precision, maximum))
	minNumLength := len(fmt.Sprintf("%0.*f", precision, minimum))
	maxWidth := int(math.Max(float64(maxNumLength), float64(minNumLength)))

	if config.FieldWidth > 0 && maxWidth != config.FieldWidth {
		maxWidth = config.FieldWidth
	}

	// axis and labels
	for y := intmin2; y < intmax2+1; y++ {
		var magnitude float64
		if rows > 0 {
			magnitude = maximum - (float64(y-intmin2) * interval / float64(rows))
		} else {
			magnitude = float64(y)
		}

		label := fmt.Sprintf("%s%*.*f%s", ac.LabelColor(), maxWidth+1, precision, magnitude, ac.EndColor())
		w := y - intmin2
		h := int(math.Max(float64(config.Offset)-float64(len(label)), 0))

		plot[w][h] = label
		if y == 0 {
			plot[w][config.Offset-1] = fmt.Sprintf("%s┼%s", ac.TickColor(), ac.EndColor())
		} else {
			plot[w][config.Offset-1] = fmt.Sprintf("%s┤%s", ac.TickColor(), ac.EndColor())
		}
	}

	y0 := int(round(series[0]*ratio) - min2)
	var y1 int

	plot[rows-y0][config.Offset-1] = fmt.Sprintf("%s┼%s", ac.TickColor(), ac.EndColor()) // first value

	for x := 0; x < len(series)-1; x++ { // plot the line
		y0 = int(round(series[x+0]*ratio) - float64(intmin2))
		y1 = int(round(series[x+1]*ratio) - float64(intmin2))
		if y0 == y1 {
			plot[rows-y0][x+config.Offset] = fmt.Sprintf("%s─%s", ac.LineColor(), ac.EndColor())
		} else {
			if y0 > y1 {
				plot[rows-y1][x+config.Offset] = fmt.Sprintf("%s╰%s", ac.LineColor(), ac.EndColor())
				plot[rows-y0][x+config.Offset] = fmt.Sprintf("%s╮%s", ac.LineColor(), ac.EndColor())
			} else {
				plot[rows-y1][x+config.Offset] = fmt.Sprintf("%s╭%s", ac.LineColor(), ac.EndColor())
				plot[rows-y0][x+config.Offset] = fmt.Sprintf("%s╯%s", ac.LineColor(), ac.EndColor())
			}

			start := int(math.Min(float64(y0), float64(y1))) + 1
			end := int(math.Max(float64(y0), float64(y1)))
			for y := start; y < end; y++ {
				plot[rows-y][x+config.Offset] = fmt.Sprintf("%s│%s", ac.LineColor(), ac.EndColor())
			}
		}
	}

	// join columns
	var lines bytes.Buffer
	for h, horizontal := range plot {
		if h != 0 {
			lines.WriteRune('\n')
		}
		for _, v := range horizontal {
			lines.WriteString(v)
		}
	}

	// add caption if not empty
	if config.Caption != "" {
		lines.WriteRune('\n')
		lines.WriteString(strings.Repeat(" ", config.Offset+maxWidth+2))
		lines.WriteString(fmt.Sprintf("%s%s%s", ac.CaptionColor(), config.Caption, ac.EndColor()))
	}

	return lines.String()
}
