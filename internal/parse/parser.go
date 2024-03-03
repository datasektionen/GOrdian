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
		//finds all Secondary cost centres and appends their index to a slice
		for colCellIndex, colCell := range cols[1][1:] {
			if colCell != "" {
				SecondaryCostCentres = append(SecondaryCostCentres, colCellIndex+1)
			}
		}

		//iterates over the found secondary cost centres indices
		for _, secondaryCostCentre := range SecondaryCostCentres {
			var secondaryCostCentreName = cols[1][secondaryCostCentre]

			//loops over the column containing budget lines. Starting one row below the current sec cost centre
			for colCellIndex, budgetLine := range cols[2][secondaryCostCentre+1:] {

				if budgetLine == "" {
					//When encountering an empty cell we have gone through all relevant budget lines
					fmt.Print("\n")
					break
				} else {
					//Find and print all relevant data
					account := cols[3][colCellIndex+secondaryCostCentre+1]
					income := cols[4][colCellIndex+secondaryCostCentre+1]
					expense := cols[5][colCellIndex+secondaryCostCentre+1]
					comment := cols[7][colCellIndex+secondaryCostCentre+1]
					fmt.Print(sheetName + "\t")
					fmt.Print(secondaryCostCentreName + "\t")
					fmt.Print(budgetLine + "\t")
					fmt.Print(account + "\t")
					fmt.Print(income + "\t")
					fmt.Print(expense + "\t")
					fmt.Print(comment)
					fmt.Print("\n")
				}

			}
		}
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
