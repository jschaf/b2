package mdext

import (
	"testing"

	"github.com/jschaf/jsc/pkg/markdown/mdtest"
	"github.com/jschaf/jsc/pkg/texts"
)

func TestNewTableExt(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			"single table",
			texts.Dedent(`
        | head 1 | head 2 |
        |--------|--------|
        | val 1  | val 2  |
		`),
			texts.Dedent(`
				<table>
					<thead>
					<tr>
						<th>head 1</th>
						<th>head 2</th>
					</tr>
					</thead>
					<tbody>
					<tr>
						<td>val 1</td>
						<td>val 2</td>
					</tr>
					</tbody>
				</table>
		`),
		},
		{
			"table with caption",
			texts.Dedent(`
        TABLE: caption
        | head 1 | head 2 |
        |--------|--------|
        | val 1  | val 2  |
		`),
			texts.Dedent(`
				<table>
          <caption><span class=table-caption-order>Table 1:</span> caption</caption>
					<thead>
					<tr>
						<th>head 1</th>
						<th>head 2</th>
					</tr>
					</thead>
					<tbody>
					<tr>
						<td>val 1</td>
						<td>val 2</td>
					</tr>
					</tbody>
				</table>
		`),
		},
		{
			"two tables with captions",
			texts.Dedent(`
        TABLE: caption 1
        | val 1  | val 2  |
        |--------|--------|

        TABLE: caption 2
        | val 3  | val 4  |
        |--------|--------|
		`),
			texts.Dedent(`
				<table>
          <caption><span class=table-caption-order>Table 1:</span> caption 1</caption>
					<thead>
					<tr>
						<th>val 1</th>
						<th>val 2</th>
					</tr>
					</thead>
				</table>

				<table>
          <caption><span class=table-caption-order>Table 2:</span> caption 2</caption>
					<thead>
					<tr>
						<th>val 3</th>
						<th>val 4</th>
					</tr>
					</thead>
				</table>
		`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, ctx := mdtest.NewTester(t, NewTableExt())
			doc := mdtest.MustParseMarkdown(t, md, ctx, tt.src)
			mdtest.AssertNoRenderDiff(t, doc, md, tt.src, tt.want)
		})
	}
}
