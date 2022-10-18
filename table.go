package pdftable

import (
	"errors"
	"math"

	"github.com/signintech/gopdf"
)

type TableConfig struct {
	// gopdf 对象
	Pdf *gopdf.GoPdf
	// 表格内容默认大小
	DefaultFontSize float64
	DefaultMarginH  float64
	DefaultMarginW  float64
	// 列宽
	HeadsW []float64
	// 新页规格，列宽继承初始页面
	Page gopdf.Rect
}

type Table struct {
	config   *TableConfig
	rowDatas []*RowData
}

func NewTable(config *TableConfig) *Table {
	if config.DefaultFontSize == 0 {
		panic("need DefaultFontSize")
	}
	if config.Pdf == nil {
		panic("need Pdf")
	}
	return &Table{
		config: config,
	}
}

// SetHead 变更默认marginW
func (t *Table) SetHead(heads []string) {
	var sum float64
	for _, v := range heads {
		w, _ := t.config.Pdf.MeasureTextWidth(v)
		t.config.HeadsW = append(t.config.HeadsW, w)
		sum += w
	}
	if t.config.DefaultMarginW == 0 {
		t.config.DefaultMarginW = (t.config.Page.W - sum - t.config.Pdf.MarginLeft() - t.config.Pdf.MarginRight()) / float64(len(heads))
	}
	for i := range t.config.HeadsW {
		t.config.HeadsW[i] += t.config.DefaultMarginW
	}
	t.rowDatas = append(t.rowDatas, &RowData{
		Value:  heads,
		config: t.config,
	})
}

func (t *Table) AddRowData(r *RowData, config *TableConfig) {
	if config != nil {
		r.config = config
	} else {
		r.config = t.config
	}
	t.rowDatas = append(t.rowDatas, r)
}

func (t *Table) Draw() error {
	t.config.Pdf.SetFontSize(t.config.DefaultFontSize)
	if t.config.HeadsW == nil {
		return errors.New("unspecified column width")
	}
	for _, v := range t.rowDatas {
		_, pageBreak, rowData := v.Draw(t.config.Pdf.GetX(), t.config.Pdf.GetY(), 0, 0)
		if pageBreak {
			if rowData == nil {
				rowData = v
			}
			t.config.Pdf.AddPage()
			rowData.Draw(t.config.Pdf.GetX(), t.config.Pdf.GetY(), 0, 0)
		}
	}
	return nil
}

// DrawHead 手动渲染头部列表，并返回宽度
func (t *Table) DrawHead(f func() []float64) {
	if t.config.HeadsW != nil {
		panic("You have specified the column width")
	}
	headw := f()
	t.config.HeadsW = headw
}

type RowData struct {
	Value        []string
	subTableData []*RowData
	config       *TableConfig
}

// Draw
// sx, sy 表示该单元的标准起始位置
// minhigh 表示父单元所需最小高度，用于父子单元高度一致
// currentIndex 表示当前渲染的单元格序号
// currentIndex 表示当前行数据的起始列号（从0开始）
// 返回这一行的最低高度，便于父格子高度调整
// 返回是否分页，便于父格子渲染
// 返回切分后的RowData，用于多次渲染
func (r *RowData) Draw(sx float64, sy float64, minhigh float64, currentIndex int) (float64, bool, *RowData) {
	Align := gopdf.Center | gopdf.Middle
	var (
		// 自身最小高度
		minHigh     float64
		minDrawHigh float64
	)
	// 返回值
	var (
		splitR     *RowData
		pageBreakP bool
	)
	// 获取最低高度
	for i, v := range r.Value {
		index := currentIndex + i
		w, _ := r.config.Pdf.MeasureTextWidth(v)
		lineNum := w/r.config.HeadsW[index] + 1
		minHigh = float64(lineNum) * (r.config.DefaultMarginH + r.config.DefaultFontSize)
	}
	// 同步父子单元高度 该高度需要返回
	minHigh = math.Max(minHigh, minhigh)
	// 分页
	if r.config.Page.H-r.config.Pdf.MarginBottom()-sy < minHigh {
		// 如果该记录无法画出，则需要返回父单元，通知分页，父单元内容需要提前绘制
		// 终止该记录绘制
		pageBreakP = true
		return 0, pageBreakP, splitR
	}
	// 渲染子表格
	// 返回每条子记录高度
	// Y 表示子记录渲染完之后画笔位置
	// 初始化画笔位置 -XY
	X := sx
	Y := sy
	if r.subTableData != nil {
		// 更新子单元初始画笔位置-X
		for i := currentIndex; i < currentIndex+len(r.Value); i++ {
			X += r.config.HeadsW[i]
		}
		// 开始绘制子行
		for i, v := range r.subTableData {
			// 绘制
			h, pageBreak, splitSubR := v.Draw(X, Y, minHigh, currentIndex+len(r.Value))
			// 更新下一子行画笔位置 - Y
			Y += h

			// 如果子行分页,画出本行并将后续返回新的数据
			if pageBreak {
				pageBreakP = pageBreak
				splitR = &RowData{
					Value:  make([]string, len(r.Value)),
					config: r.config,
				}
				if splitSubR != nil {
					splitR.subTableData = append(splitR.subTableData, splitSubR)
				}
				if i+1 < len(r.subTableData) {
					i = i + 1
				}
				splitR.subTableData = append(splitR.subTableData, r.subTableData[i:]...)
				break
			}
		}
	}
	minDrawHigh = math.Max(minHigh, Y-sy)
	// 正常绘制
	r.config.Pdf.SetX(sx)
	r.config.Pdf.SetY(sy)
	for i, v := range r.Value {
		r.config.Pdf.CellWithOption(&gopdf.Rect{W: r.config.HeadsW[currentIndex+i], H: minDrawHigh}, v, gopdf.CellOption{
			Align:  Align,
			Float:  gopdf.Right,
			Border: gopdf.AllBorders,
		})
	}
	r.config.Pdf.Br(minDrawHigh)
	return minDrawHigh, pageBreakP, splitR
}

func (r *RowData) AddSubRowData(subr *RowData, config *TableConfig) {
	if config != nil {
		subr.config = config
	} else {
		subr.config = r.config
	}
	r.subTableData = append(r.subTableData, subr)
}

func NewRowData(r *RowData, c *TableConfig) {
	r.config = c
}
