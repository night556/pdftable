package example

import (
	"fmt"

	"github.com/night556/pdftable"
	"github.com/signintech/gopdf"
)

var defaultPageSize = gopdf.Rect{
	W: gopdf.PageSizeA4Landscape.W,
	H: 200,
}
var contentPageW = defaultPageSize.W - 60

func BaseExample() {
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{
		PageSize: defaultPageSize,
	})
	pdf.SetMargins(30, 30, 30, 30)
	pdf.AddPage()
	pdf.AddTTFFont("songti", "./STSongti-SC-Regular-07.ttf")
	pdf.SetFont("songti", "", 11)
	pdf.CellWithOption(&gopdf.Rect{W: contentPageW, H: 19}, "测试打印 PDF 表格", gopdf.CellOption{
		Align: gopdf.Center,
		Float: gopdf.Bottom,
	})
	pdf.SetLineWidth(0.5)
	t := pdftable.NewTable(&pdftable.TableConfig{
		Pdf:             pdf,
		DefaultFontSize: 10,
		Page:            defaultPageSize,
	})
	t.SetHead([]string{"head1", "head2", "head3", "head4", "head5"})
	for i := 0; i < 10; i++ {
		r := &pdftable.RowData{
			Value: []string{fmt.Sprintf("data1-%d", i)},
		}
		t.AddRowData(r, nil)
		for j := 0; j < 3; j++ {
			r2 := &pdftable.RowData{
				Value: []string{
					fmt.Sprintf("data2-%d", j),
					fmt.Sprintf("data3-%d", j),
				},
			}
			r.AddSubRowData(r2, nil)
			for k := 0; k < 4; k++ {
				r3 := &pdftable.RowData{
					Value: []string{
						fmt.Sprintf("data4-%d", k),
						fmt.Sprintf("data5-%d", k),
					},
				}
				r2.AddSubRowData(r3, nil)
			}
		}
	}
	t.Draw()
	pdf.WritePdf("example.pdf")
}
