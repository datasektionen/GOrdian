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

		var SecondaryCostCentreIndices []int

		//get indices of all secondary cost centres
		SecondaryCostCentreIndices = findSecondaryCostCentreIndices(cols)

		//iterates over the found secondary cost centres indices
		for _, secondaryCostCentreIndex := range SecondaryCostCentreIndices {
			var secondaryCostCentreName = getSecondaryCostCentreByIndex(secondaryCostCentreIndex, cols)

			//loops over the column containing budget lines. Starting one row below the current sec cost centre
			for colCellIndex, budgetLine := range cols[2][secondaryCostCentreIndex+1:] {
				if budgetLine == "" {
					//When encountering an empty cell we have gone through all relevant budget lines
					fmt.Print("\n")
					break
				} else {
					account, income, expense, comment := getBudgetLineData(secondaryCostCentreIndex, colCellIndex, cols)
					//Print all relevant data
					//We already have sheetname from outmost loop and secondaryCostCentreName from inner loop
					//0s are gotten without kr for some reason
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

func findSecondaryCostCentreIndices(cols [][]string) []int {
	var SecondaryCostCentreIndices []int
	//iterates over cells in the second column, skipping the first row
	//finds all Secondary cost centres and appends their index to a slice
	for colCellIndex, colCell := range cols[1][1:] {
		if colCell != "" {
			//+1 is needed since 1st row is skipped
			SecondaryCostCentreIndices = append(SecondaryCostCentreIndices, colCellIndex+1)
		}
	}
	return SecondaryCostCentreIndices
}

func getSecondaryCostCentreByIndex(secondaryCostCentreIndex int, cols [][]string) string {
	//Secondary cost centres are always in the second column
	return cols[1][secondaryCostCentreIndex]
}

func getBudgetLineData(secondaryCostCentreIndex int, colCellIndex int, cols [][]string) (string, string, string, string) {

	//Find all relevant data
	//0s are gotten without kr for some reason
	account := cols[3][colCellIndex+secondaryCostCentreIndex+1]
	income := cols[4][colCellIndex+secondaryCostCentreIndex+1]
	expense := cols[5][colCellIndex+secondaryCostCentreIndex+1]
	comment := cols[7][colCellIndex+secondaryCostCentreIndex+1]
	return account, income, expense, comment
}
