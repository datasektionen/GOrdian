package parse

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"strconv"
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

		//printerino
		//TODO remove
		for _, colCell := range cols[2][1:] {
			if colCell != "" {
				fmt.Print(sheetName + "\t" + colCell + "\t")
				fmt.Print("\n")
			}
		}

		fmt.Println(SecondaryCostCentres)

		for _, SecondaryCostCentre := range SecondaryCostCentres {
			for colCellIndex, colCell := range cols[2][SecondaryCostCentre+1:] {
				colCellIndex = colCellIndex
				if colCell == "" {
					break
				} else {
					fmt.Print(sheetName + "\t" + strconv.Itoa(SecondaryCostCentre) + "\t" + colCell + "\t")
					fmt.Print("\n")
				}

			}
		}

		for SecondaryCostCentreIndex, SecondaryCostCentre := range SecondaryCostCentres {
			var budgetLines []string
			fmt.Println(strconv.Itoa(SecondaryCostCentreIndex) + " " + cols[1][SecondaryCostCentre])
			for colIndex, col := range cols {
				if colIndex == 2 {
					budgetLines = append(budgetLines, col[SecondaryCostCentre+1])

				}

			}

		}
		fmt.Println()
	}

	readCell(file)

	if err := file.Close(); err != nil {
		fmt.Println(err)
	}
}

func readCell(file *excelize.File) {

	cell, err := file.GetCellValue("3 - DKM Detaljbudget", "C9")

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(cell)
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
