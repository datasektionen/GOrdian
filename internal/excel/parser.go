package excel

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

type BudgetLine struct {
	costCentreName          string
	costCentreType          string
	secondaryCostCentreName string
	budgetLineName          string
	budgetLineAccount       string
	budgetLineIncome        string
	budgetLineExpense       string
	budgetLineComment       string
}

const (
	columnSecondaryCostCentres = 1
	columnBudgetLineNames      = 2
	columnAccount              = 3
	columnIncome               = 4
	columnExpense              = 5
	columnComment              = 7
	costCentreNameCell         = "A2"
	costCentreTypeCell         = "A3"
)

func ReadExcel(path string) ([]BudgetLine, error) {
	file, err := excelize.OpenFile(path)

	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			_ = fmt.Errorf("error closing Excel file: %v", err)
		}
	}()

	sheets := readSheets(file)

	var budgetLines []BudgetLine

	//[1:] excludes the first sheet
	for sheetIndex, sheetName := range sheets[1:] {

		fmt.Printf("Sheet Name: %s, Index %d\n", sheetName, sheetIndex)

		//read all rows and columns of a sheet
		_, err := readRows(sheetName, file)
		if err != nil {
			return nil, fmt.Errorf("failed to read rows for sheet %s: %v", sheetName, err)
		}

		cols, err := readColumns(sheetName, file)
		if err != nil {
			return nil, fmt.Errorf("failed to read columns for sheet %s: %v", sheetName, err)
		}

		costCentreName, err := readCell(file, sheetName, costCentreNameCell)
		if err != nil {
			return nil, fmt.Errorf("failed to read cell for sheet %s: %v", sheetName, err)
		}

		costCentreType, err := readCell(file, sheetName, costCentreTypeCell)
		if err != nil {
			return nil, fmt.Errorf("failed to read cell for sheet %s: %v", sheetName, err)
		}

		//get indices of all secondary cost centres
		var SecondaryCostCentreIndices []int
		SecondaryCostCentreIndices = findSecondaryCostCentreIndices(cols)

		//iterates over the found secondary cost centres indices
		for _, secondaryCostCentreIndex := range SecondaryCostCentreIndices {
			secondaryCostCentreName := getSecondaryCostCentreByIndex(secondaryCostCentreIndex, cols)

			//loops over the column containing budget lines. Starting one row below the current sec cost centre
			for colCellIndex, budgetLineName := range cols[columnBudgetLineNames][secondaryCostCentreIndex+1:] {
				if budgetLineName == "" {
					//When encountering an empty cell we have gone through all relevant budget lines
					fmt.Print("\n")
					break
				} else {
					budgetLineAccount, budgetLineIncome, budgetLineExpense, budgetLineComment :=
						getBudgetLineData(secondaryCostCentreIndex, colCellIndex, cols)

					//create BudgetLine object
					budgetLine := BudgetLine{
						costCentreName:          costCentreName,
						costCentreType:          costCentreType,
						secondaryCostCentreName: secondaryCostCentreName,
						budgetLineName:          budgetLineName,
						budgetLineAccount:       budgetLineAccount,
						budgetLineIncome:        budgetLineIncome,
						budgetLineExpense:       budgetLineExpense,
						budgetLineComment:       budgetLineComment,
					}
					//append to budgetLines slice
					budgetLines = append(budgetLines, budgetLine)

					//Print all relevant data
					//We already have sheetName from outmost loop and secondaryCostCentreName from inner loop
					//0s are gotten without kr for some reason
					fmt.Print(costCentreName + "\t")
					fmt.Print(costCentreType + "\t")
					fmt.Print(secondaryCostCentreName + "\t")
					fmt.Print(budgetLineName + "\t")
					fmt.Print(budgetLineAccount + "\t")
					fmt.Print(budgetLineIncome + "\t")
					fmt.Print(budgetLineExpense + "\t")
					fmt.Print(budgetLineComment)
					fmt.Print("\n")
				}
			}
		}
	}
	return budgetLines, nil
}

func readCell(file *excelize.File, sheetName string, targetCell string) (string, error) {

	cellContent, err := file.GetCellValue(sheetName, targetCell)

	if err != nil {
		return "", fmt.Errorf("failed to read cell: %v", err)
	}
	return cellContent, err
}

func readSheets(file *excelize.File) []string {
	return file.GetSheetList()
}

func readRows(sheetName string, file *excelize.File) ([][]string, error) {
	rows, err := file.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows: %v", err)
	}
	return rows, err
}

func readColumns(sheetName string, file *excelize.File) ([][]string, error) {
	cols, err := file.GetCols(sheetName)
	if err != nil {
		return nil, fmt.Errorf("error closing Excel file: %v", err)
	}
	return cols, err
}

// findSecondaryCostCentreIndices finds indices of secondary cost centers
func findSecondaryCostCentreIndices(cols [][]string) []int {
	var SecondaryCostCentreIndices []int
	//iterates over cells in the second column, skipping the first row
	//finds all Secondary cost centres and appends their index to a slice
	for colCellIndex, colCell := range cols[columnSecondaryCostCentres][1:] {
		if colCell != "" {
			//+1 is needed since 1st row is skipped
			SecondaryCostCentreIndices = append(SecondaryCostCentreIndices, colCellIndex+1)
		}
	}
	return SecondaryCostCentreIndices
}

// getSecondaryCostCentreByIndex retrieves secondary cost center name by index
func getSecondaryCostCentreByIndex(secondaryCostCentreIndex int, cols [][]string) string {
	//Secondary cost centres are always in the second column
	return cols[columnSecondaryCostCentres][secondaryCostCentreIndex]
}

func getBudgetLineData(secondaryCostCentreIndex int, colCellIndex int, cols [][]string) (string, string, string, string) {

	//Find all relevant data
	//0s are gotten without kr for some reason
	account := cols[columnAccount][colCellIndex+secondaryCostCentreIndex+1]
	income := cols[columnIncome][colCellIndex+secondaryCostCentreIndex+1]
	expense := cols[columnExpense][colCellIndex+secondaryCostCentreIndex+1]
	comment := cols[columnComment][colCellIndex+secondaryCostCentreIndex+1]
	return account, income, expense, comment
}
