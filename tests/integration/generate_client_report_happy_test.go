package integration

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func (s *TestSuite) TestGenerateClientReport_Happy() {
	s.loadFixtures()

	ctx := context.Background()

	data, err := s.svc.GenerateClientReportByte(ctx, "orders")
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), data)

	f, err := excelize.OpenReader(bytes.NewReader(data))
	require.NoError(s.T(), err)

	sheet := "ClientsReport"
	require.Contains(s.T(), f.GetSheetList(), sheet)

	rows, err := f.GetRows(sheet)
	require.NoError(s.T(), err)

	expectedHeaders := []string{
		"UserID", "Total Orders", "Returned Orders", "Total Purchase Sum (₽)",
	}
	assert.Equal(s.T(), expectedHeaders, rows[0])

	dataMap := make(map[string][]string)
	for _, row := range rows[1:] {
		if len(row) >= 4 {
			dataMap[row[0]] = row
		}
	}

	r200, ok := dataMap["200"]
	require.True(s.T(), ok, "missing row for user 200")
	assert.Equal(s.T(), "3", r200[1])
	assert.Equal(s.T(), "0", r200[2])
	assert.Equal(s.T(), "1500", r200[3])
}
