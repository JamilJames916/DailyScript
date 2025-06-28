package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type CSVProcessor struct {
	headers []string
	records [][]string
}

func NewCSVProcessor() *CSVProcessor {
	return &CSVProcessor{
		headers: make([]string, 0),
		records: make([][]string, 0),
	}
}

func (cp *CSVProcessor) LoadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("CSV file is empty")
	}

	cp.headers = records[0]
	cp.records = records[1:]

	return nil
}

func (cp *CSVProcessor) SaveToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers
	if err := writer.Write(cp.headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	// Write records
	for _, record := range cp.records {
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}

func (cp *CSVProcessor) ToJSON() ([]byte, error) {
	var jsonData []map[string]interface{}

	for _, record := range cp.records {
		row := make(map[string]interface{})
		for i, value := range record {
			if i < len(cp.headers) {
				// Try to convert to number if possible
				if num, err := strconv.ParseFloat(value, 64); err == nil {
					row[cp.headers[i]] = num
				} else if value == "true" || value == "false" {
					row[cp.headers[i]] = value == "true"
				} else {
					row[cp.headers[i]] = value
				}
			}
		}
		jsonData = append(jsonData, row)
	}

	return json.MarshalIndent(jsonData, "", "  ")
}

func (cp *CSVProcessor) FromJSON(jsonData []byte) error {
	var data []map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if len(data) == 0 {
		return fmt.Errorf("JSON data is empty")
	}

	// Extract headers from first object
	cp.headers = make([]string, 0)
	for key := range data[0] {
		cp.headers = append(cp.headers, key)
	}

	// Convert data to CSV records
	cp.records = make([][]string, 0)
	for _, row := range data {
		record := make([]string, len(cp.headers))
		for i, header := range cp.headers {
			if value, exists := row[header]; exists {
				record[i] = fmt.Sprintf("%v", value)
			}
		}
		cp.records = append(cp.records, record)
	}

	return nil
}

func (cp *CSVProcessor) Filter(columnIndex int, filterValue string) *CSVProcessor {
	filtered := NewCSVProcessor()
	filtered.headers = make([]string, len(cp.headers))
	copy(filtered.headers, cp.headers)

	for _, record := range cp.records {
		if columnIndex < len(record) && strings.Contains(strings.ToLower(record[columnIndex]), strings.ToLower(filterValue)) {
			filtered.records = append(filtered.records, record)
		}
	}

	return filtered
}

func (cp *CSVProcessor) GetColumn(columnName string) ([]string, error) {
	columnIndex := -1
	for i, header := range cp.headers {
		if header == columnName {
			columnIndex = i
			break
		}
	}

	if columnIndex == -1 {
		return nil, fmt.Errorf("column '%s' not found", columnName)
	}

	column := make([]string, len(cp.records))
	for i, record := range cp.records {
		if columnIndex < len(record) {
			column[i] = record[columnIndex]
		}
	}

	return column, nil
}

func (cp *CSVProcessor) AddColumn(columnName string, values []string) error {
	if len(values) != len(cp.records) {
		return fmt.Errorf("number of values (%d) doesn't match number of records (%d)", len(values), len(cp.records))
	}

	cp.headers = append(cp.headers, columnName)
	for i, value := range values {
		cp.records[i] = append(cp.records[i], value)
	}

	return nil
}

func (cp *CSVProcessor) Statistics() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["total_rows"] = len(cp.records)
	stats["total_columns"] = len(cp.headers)
	stats["headers"] = cp.headers

	// Calculate statistics for each column
	columnStats := make(map[string]interface{})
	for _, header := range cp.headers {
		column, _ := cp.GetColumn(header)
		
		// Count non-empty values
		nonEmpty := 0
		for _, value := range column {
			if strings.TrimSpace(value) != "" {
				nonEmpty++
			}
		}
		
		columnStats[header] = map[string]interface{}{
			"non_empty": nonEmpty,
			"empty":     len(column) - nonEmpty,
		}
	}
	
	stats["column_statistics"] = columnStats
	return stats
}

func (cp *CSVProcessor) Print() {
	// Print headers
	fmt.Println(strings.Join(cp.headers, " | "))
	fmt.Println(strings.Repeat("-", len(strings.Join(cp.headers, " | "))))

	// Print records (limit to first 10 for readability)
	limit := len(cp.records)
	if limit > 10 {
		limit = 10
	}

	for i := 0; i < limit; i++ {
		fmt.Println(strings.Join(cp.records[i], " | "))
	}

	if len(cp.records) > 10 {
		fmt.Printf("... and %d more rows\n", len(cp.records)-10)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run csv-processor.go <command> [args...]")
		fmt.Println("Commands:")
		fmt.Println("  show <file.csv>                    - Display CSV content")
		fmt.Println("  to-json <file.csv> [output.json]   - Convert CSV to JSON")
		fmt.Println("  from-json <file.json> <output.csv> - Convert JSON to CSV")
		fmt.Println("  filter <file.csv> <column> <value> - Filter CSV by column value")
		fmt.Println("  column <file.csv> <column_name>    - Get specific column")
		fmt.Println("  stats <file.csv>                   - Show CSV statistics")
		os.Exit(1)
	}

	command := os.Args[1]
	processor := NewCSVProcessor()

	switch command {
	case "show":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run csv-processor.go show <file.csv>")
			os.Exit(1)
		}

		err := processor.LoadFromFile(os.Args[2])
		if err != nil {
			fmt.Printf("Error loading CSV: %v\n", err)
			os.Exit(1)
		}

		processor.Print()

	case "to-json":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run csv-processor.go to-json <file.csv> [output.json]")
			os.Exit(1)
		}

		err := processor.LoadFromFile(os.Args[2])
		if err != nil {
			fmt.Printf("Error loading CSV: %v\n", err)
			os.Exit(1)
		}

		jsonData, err := processor.ToJSON()
		if err != nil {
			fmt.Printf("Error converting to JSON: %v\n", err)
			os.Exit(1)
		}

		if len(os.Args) > 3 {
			// Save to file
			err = os.WriteFile(os.Args[3], jsonData, 0644)
			if err != nil {
				fmt.Printf("Error saving JSON file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("JSON saved to %s\n", os.Args[3])
		} else {
			// Print to stdout
			fmt.Println(string(jsonData))
		}

	case "from-json":
		if len(os.Args) < 4 {
			fmt.Println("Usage: go run csv-processor.go from-json <file.json> <output.csv>")
			os.Exit(1)
		}

		jsonData, err := os.ReadFile(os.Args[2])
		if err != nil {
			fmt.Printf("Error reading JSON file: %v\n", err)
			os.Exit(1)
		}

		err = processor.FromJSON(jsonData)
		if err != nil {
			fmt.Printf("Error converting from JSON: %v\n", err)
			os.Exit(1)
		}

		err = processor.SaveToFile(os.Args[3])
		if err != nil {
			fmt.Printf("Error saving CSV file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("CSV saved to %s\n", os.Args[3])

	case "filter":
		if len(os.Args) < 5 {
			fmt.Println("Usage: go run csv-processor.go filter <file.csv> <column_index> <value>")
			os.Exit(1)
		}

		err := processor.LoadFromFile(os.Args[2])
		if err != nil {
			fmt.Printf("Error loading CSV: %v\n", err)
			os.Exit(1)
		}

		columnIndex, err := strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Printf("Invalid column index: %v\n", err)
			os.Exit(1)
		}

		filtered := processor.Filter(columnIndex, os.Args[4])
		fmt.Printf("Filtered results (found %d rows):\n", len(filtered.records))
		filtered.Print()

	case "column":
		if len(os.Args) < 4 {
			fmt.Println("Usage: go run csv-processor.go column <file.csv> <column_name>")
			os.Exit(1)
		}

		err := processor.LoadFromFile(os.Args[2])
		if err != nil {
			fmt.Printf("Error loading CSV: %v\n", err)
			os.Exit(1)
		}

		column, err := processor.GetColumn(os.Args[3])
		if err != nil {
			fmt.Printf("Error getting column: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Column '%s':\n", os.Args[3])
		for idx, value := range column {
			fmt.Printf("%d: %s\n", idx+1, value)
		}

	case "stats":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run csv-processor.go stats <file.csv>")
			os.Exit(1)
		}

		err := processor.LoadFromFile(os.Args[2])
		if err != nil {
			fmt.Printf("Error loading CSV: %v\n", err)
			os.Exit(1)
		}

		stats := processor.Statistics()
		statsJSON, _ := json.MarshalIndent(stats, "", "  ")
		fmt.Println("CSV Statistics:")
		fmt.Println(string(statsJSON))

	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
