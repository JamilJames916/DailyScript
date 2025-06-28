package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
)

type JSONProcessor struct {
	data interface{}
}

func NewJSONProcessor() *JSONProcessor {
	return &JSONProcessor{}
}

func (jp *JSONProcessor) LoadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&jp.data)
}

func (jp *JSONProcessor) LoadFromString(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), &jp.data)
}

func (jp *JSONProcessor) SaveToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(jp.data)
}

func (jp *JSONProcessor) PrettyPrint() {
	jsonBytes, err := json.MarshalIndent(jp.data, "", "  ")
	if err != nil {
		fmt.Printf("Error formatting JSON: %v\n", err)
		return
	}
	fmt.Println(string(jsonBytes))
}

func (jp *JSONProcessor) GetKeys() []string {
	return jp.getKeysRecursive(jp.data, "")
}

func (jp *JSONProcessor) getKeysRecursive(data interface{}, prefix string) []string {
	var keys []string

	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			fullKey := key
			if prefix != "" {
				fullKey = prefix + "." + key
			}
			keys = append(keys, fullKey)

			// Recursively get nested keys
			nestedKeys := jp.getKeysRecursive(value, fullKey)
			keys = append(keys, nestedKeys...)
		}
	case []interface{}:
		for i, item := range v {
			indexKey := fmt.Sprintf("[%d]", i)
			if prefix != "" {
				indexKey = prefix + indexKey
			}
			keys = append(keys, indexKey)

			nestedKeys := jp.getKeysRecursive(item, indexKey)
			keys = append(keys, nestedKeys...)
		}
	}

	return keys
}

func (jp *JSONProcessor) GetValue(path string) interface{} {
	return jp.getValueByPath(jp.data, path)
}

func (jp *JSONProcessor) getValueByPath(data interface{}, path string) interface{} {
	if path == "" {
		return data
	}

	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		// Handle array access
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			// Extract array name and index
			arrayName := part[:strings.Index(part, "[")]
			indexStr := part[strings.Index(part, "[")+1 : strings.Index(part, "]")]

			if arrayName != "" {
				if m, ok := current.(map[string]interface{}); ok {
					current = m[arrayName]
				} else {
					return nil
				}
			}

			if arr, ok := current.([]interface{}); ok {
				var index int
				fmt.Sscanf(indexStr, "%d", &index)
				if index >= 0 && index < len(arr) {
					current = arr[index]
				} else {
					return nil
				}
			} else {
				return nil
			}
		} else {
			if m, ok := current.(map[string]interface{}); ok {
				current = m[part]
			} else {
				return nil
			}
		}
	}

	return current
}

func (jp *JSONProcessor) Filter(filterFunc func(key string, value interface{}) bool) map[string]interface{} {
	result := make(map[string]interface{})
	jp.filterRecursive(jp.data, "", result, filterFunc)
	return result
}

func (jp *JSONProcessor) filterRecursive(data interface{}, prefix string, result map[string]interface{}, filterFunc func(string, interface{}) bool) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			fullKey := key
			if prefix != "" {
				fullKey = prefix + "." + key
			}

			if filterFunc(fullKey, value) {
				result[fullKey] = value
			}

			jp.filterRecursive(value, fullKey, result, filterFunc)
		}
	}
}

func (jp *JSONProcessor) Statistics() map[string]interface{} {
	stats := make(map[string]interface{})
	
	keys := jp.GetKeys()
	stats["total_keys"] = len(keys)
	
	typeCount := make(map[string]int)
	jp.countTypes(jp.data, typeCount)
	stats["type_distribution"] = typeCount
	
	stats["max_depth"] = jp.getMaxDepth(jp.data, 0)
	
	return stats
}

func (jp *JSONProcessor) countTypes(data interface{}, typeCount map[string]int) {
	switch v := data.(type) {
	case map[string]interface{}:
		typeCount["object"]++
		for _, value := range v {
			jp.countTypes(value, typeCount)
		}
	case []interface{}:
		typeCount["array"]++
		for _, item := range v {
			jp.countTypes(item, typeCount)
		}
	case string:
		typeCount["string"]++
	case float64:
		typeCount["number"]++
	case bool:
		typeCount["boolean"]++
	case nil:
		typeCount["null"]++
	}
}

func (jp *JSONProcessor) getMaxDepth(data interface{}, currentDepth int) int {
	maxDepth := currentDepth

	switch v := data.(type) {
	case map[string]interface{}:
		for _, value := range v {
			depth := jp.getMaxDepth(value, currentDepth+1)
			if depth > maxDepth {
				maxDepth = depth
			}
		}
	case []interface{}:
		for _, item := range v {
			depth := jp.getMaxDepth(item, currentDepth+1)
			if depth > maxDepth {
				maxDepth = depth
			}
		}
	}

	return maxDepth
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run json-processor.go <command> [args...]")
		fmt.Println("Commands:")
		fmt.Println("  pretty <file.json>           - Pretty print JSON file")
		fmt.Println("  keys <file.json>             - List all keys in JSON")
		fmt.Println("  get <file.json> <path>       - Get value at specific path")
		fmt.Println("  stats <file.json>            - Show JSON statistics")
		fmt.Println("  validate <file.json>         - Validate JSON format")
		os.Exit(1)
	}

	command := os.Args[1]
	processor := NewJSONProcessor()

	switch command {
	case "pretty":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run json-processor.go pretty <file.json>")
			os.Exit(1)
		}
		
		err := processor.LoadFromFile(os.Args[2])
		if err != nil {
			fmt.Printf("Error loading JSON: %v\n", err)
			os.Exit(1)
		}
		
		processor.PrettyPrint()

	case "keys":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run json-processor.go keys <file.json>")
			os.Exit(1)
		}
		
		err := processor.LoadFromFile(os.Args[2])
		if err != nil {
			fmt.Printf("Error loading JSON: %v\n", err)
			os.Exit(1)
		}
		
		keys := processor.GetKeys()
		sort.Strings(keys)
		
		fmt.Printf("Found %d keys:\n", len(keys))
		for _, key := range keys {
			fmt.Printf("  %s\n", key)
		}

	case "get":
		if len(os.Args) < 4 {
			fmt.Println("Usage: go run json-processor.go get <file.json> <path>")
			fmt.Println("Example: go run json-processor.go get data.json user.name")
			os.Exit(1)
		}
		
		err := processor.LoadFromFile(os.Args[2])
		if err != nil {
			fmt.Printf("Error loading JSON: %v\n", err)
			os.Exit(1)
		}
		
		path := os.Args[3]
		value := processor.GetValue(path)
		
		if value != nil {
			if reflect.TypeOf(value).Kind() == reflect.Map || reflect.TypeOf(value).Kind() == reflect.Slice {
				jsonBytes, _ := json.MarshalIndent(value, "", "  ")
				fmt.Println(string(jsonBytes))
			} else {
				fmt.Printf("%v\n", value)
			}
		} else {
			fmt.Printf("Path '%s' not found\n", path)
		}

	case "stats":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run json-processor.go stats <file.json>")
			os.Exit(1)
		}
		
		err := processor.LoadFromFile(os.Args[2])
		if err != nil {
			fmt.Printf("Error loading JSON: %v\n", err)
			os.Exit(1)
		}
		
		stats := processor.Statistics()
		fmt.Println("JSON Statistics:")
		
		statsBytes, _ := json.MarshalIndent(stats, "", "  ")
		fmt.Println(string(statsBytes))

	case "validate":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run json-processor.go validate <file.json>")
			os.Exit(1)
		}
		
		err := processor.LoadFromFile(os.Args[2])
		if err != nil {
			fmt.Printf("❌ Invalid JSON: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Println("✅ Valid JSON")

	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
