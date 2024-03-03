package parse

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

func ReadExcel(path string) {
	file, err := excelize.OpenFile(path)

	if err != nil {
		fmt.Println(err)
		return
	}

	sheets := readSheets(file)

	//[1:] excludes the first sheet
	for sheetIndex, sheetName := range sheets[1:] {

		fmt.Printf("Sheet Name: %s, Index %d\n", sheetName, sheetIndex)

		//read all rows and columns of a sheet
		rows := readRows(sheetName, file)
		rows = rows
		cols := readColumns(sheetName, file)

		var SecondaryCostCentres []int

		//iterates over cells in the second column, skipping the first row
		for colCellIndex, colCell := range cols[1][1:] {
			if colCell != "" {
				SecondaryCostCentres = append(SecondaryCostCentres, colCellIndex+1)
			}
		}

		fmt.Println(SecondaryCostCentres)

		for _, secondaryCostCentre := range SecondaryCostCentres {
			var secondaryCostCentreName = cols[1][secondaryCostCentre]
			for colCellIndex, budgetLine := range cols[2][secondaryCostCentre+1:] {

				if budgetLine == "" {
					fmt.Print("\n")
					break
				} else {
					account := cols[3][colCellIndex+secondaryCostCentre+1]
					income := cols[4][colCellIndex+secondaryCostCentre+1]
					expense := cols[5][colCellIndex+secondaryCostCentre+1]
					comment := cols[6][colCellIndex+secondaryCostCentre+1]
					comment = comment
					fmt.Print(sheetName + "\t" + secondaryCostCentreName + "\t" + budgetLine + "\t" + account + "\t" + income + "\t" + expense)
					fmt.Print("\n")
				}

			}
		}
		fmt.Println()
	}

	if err := file.Close(); err != nil {
		fmt.Println(err)
	}
}

func readSheets(file *excelize.File) []string {

	return file.GetSheetList()

}

func readRows(sheetName string, file *excelize.File) [][]string {
	rows, err := file.GetRows(sheetName)

	if err != nil {
		fmt.Println(err)
	}

	return rows
}

func readColumns(sheetName string, file *excelize.File) [][]string {
	cols, err := file.GetCols(sheetName)

	if err != nil {
		fmt.Println(err)
	}

	return cols
}

func getSecondaryCostCentreByIndex(index int) string {
	return ""
}
