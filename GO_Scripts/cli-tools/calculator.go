package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Calculator provides basic arithmetic operations
type Calculator struct {
	memory float64
	history []string
}

func NewCalculator() *Calculator {
	return &Calculator{
		memory: 0,
		history: make([]string, 0),
	}
}

func (c *Calculator) Add(a, b float64) float64 {
	result := a + b
	c.addToHistory(fmt.Sprintf("%.2f + %.2f = %.2f", a, b, result))
	return result
}

func (c *Calculator) Subtract(a, b float64) float64 {
	result := a - b
	c.addToHistory(fmt.Sprintf("%.2f - %.2f = %.2f", a, b, result))
	return result
}

func (c *Calculator) Multiply(a, b float64) float64 {
	result := a * b
	c.addToHistory(fmt.Sprintf("%.2f × %.2f = %.2f", a, b, result))
	return result
}

func (c *Calculator) Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}
	result := a / b
	c.addToHistory(fmt.Sprintf("%.2f ÷ %.2f = %.2f", a, b, result))
	return result, nil
}

func (c *Calculator) Power(base, exponent float64) float64 {
	result := 1.0
	if exponent == 0 {
		return 1
	}
	
	// Simple power implementation
	for i := 0; i < int(exponent); i++ {
		result *= base
	}
	
	c.addToHistory(fmt.Sprintf("%.2f^%.0f = %.2f", base, exponent, result))
	return result
}

func (c *Calculator) Sqrt(n float64) (float64, error) {
	if n < 0 {
		return 0, fmt.Errorf("square root of negative number")
	}
	
	// Newton's method for square root
	if n == 0 {
		return 0, nil
	}
	
	x := n
	for i := 0; i < 10; i++ { // 10 iterations should be enough for precision
		x = (x + n/x) / 2
	}
	
	c.addToHistory(fmt.Sprintf("√%.2f = %.2f", n, x))
	return x, nil
}

func (c *Calculator) StoreMemory(value float64) {
	c.memory = value
	c.addToHistory(fmt.Sprintf("Memory stored: %.2f", value))
}

func (c *Calculator) RecallMemory() float64 {
	c.addToHistory(fmt.Sprintf("Memory recalled: %.2f", c.memory))
	return c.memory
}

func (c *Calculator) ClearMemory() {
	c.memory = 0
	c.addToHistory("Memory cleared")
}

func (c *Calculator) addToHistory(operation string) {
	c.history = append(c.history, operation)
	if len(c.history) > 50 { // Keep only last 50 operations
		c.history = c.history[1:]
	}
}

func (c *Calculator) ShowHistory() {
	fmt.Println("\n=== Calculation History ===")
	if len(c.history) == 0 {
		fmt.Println("No calculations yet")
		return
	}
	
	for i, operation := range c.history {
		fmt.Printf("%d. %s\n", i+1, operation)
	}
}

func (c *Calculator) ClearHistory() {
	c.history = make([]string, 0)
	fmt.Println("History cleared")
}

// Expression evaluator for simple expressions
func (c *Calculator) EvaluateExpression(expr string) (float64, error) {
	expr = strings.ReplaceAll(expr, " ", "")
	
	// Simple expression parser for basic operations
	// This is a basic implementation - for complex expressions, use a proper parser
	
	var result float64
	var operator string
	var number strings.Builder
	
	for i, char := range expr {
		switch char {
		case '+', '-', '*', '/', '^':
			if number.Len() > 0 {
				num, err := strconv.ParseFloat(number.String(), 64)
				if err != nil {
					return 0, fmt.Errorf("invalid number: %s", number.String())
				}
				
				if operator == "" {
					result = num
				} else {
					switch operator {
					case "+":
						result = c.Add(result, num)
					case "-":
						result = c.Subtract(result, num)
					case "*":
						result = c.Multiply(result, num)
					case "/":
						var err error
						result, err = c.Divide(result, num)
						if err != nil {
							return 0, err
						}
					case "^":
						result = c.Power(result, num)
					}
				}
				number.Reset()
			}
			operator = string(char)
		default:
			number.WriteRune(char)
		}
		
		// Handle last number
		if i == len(expr)-1 && number.Len() > 0 {
			num, err := strconv.ParseFloat(number.String(), 64)
			if err != nil {
				return 0, fmt.Errorf("invalid number: %s", number.String())
			}
			
			if operator == "" {
				result = num
			} else {
				switch operator {
				case "+":
					result = c.Add(result, num)
				case "-":
					result = c.Subtract(result, num)
				case "*":
					result = c.Multiply(result, num)
				case "/":
					var err error
					result, err = c.Divide(result, num)
					if err != nil {
						return 0, err
					}
				case "^":
					result = c.Power(result, num)
				}
			}
		}
	}
	
	return result, nil
}

func printMenu() {
	fmt.Println("\n=== Go Calculator ===")
	fmt.Println("Operations:")
	fmt.Println("  1. Addition (a + b)")
	fmt.Println("  2. Subtraction (a - b)")
	fmt.Println("  3. Multiplication (a * b)")
	fmt.Println("  4. Division (a / b)")
	fmt.Println("  5. Power (a ^ b)")
	fmt.Println("  6. Square root (√a)")
	fmt.Println("  7. Expression (e.g., 2+3*4)")
	fmt.Println("Memory:")
	fmt.Println("  m. Store to memory")
	fmt.Println("  r. Recall from memory")
	fmt.Println("  c. Clear memory")
	fmt.Println("Other:")
	fmt.Println("  h. Show history")
	fmt.Println("  x. Clear history")
	fmt.Println("  q. Quit")
	fmt.Print("\nEnter choice: ")
}

func getFloat64Input(prompt string) (float64, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	return strconv.ParseFloat(input, 64)
}

func getStringInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func main() {
	if len(os.Args) > 1 {
		// Command line mode
		calc := NewCalculator()
		
		switch os.Args[1] {
		case "add":
			if len(os.Args) != 4 {
				fmt.Println("Usage: go run calculator.go add <a> <b>")
				os.Exit(1)
			}
			a, _ := strconv.ParseFloat(os.Args[2], 64)
			b, _ := strconv.ParseFloat(os.Args[3], 64)
			result := calc.Add(a, b)
			fmt.Printf("%.2f + %.2f = %.2f\n", a, b, result)
			
		case "subtract":
			if len(os.Args) != 4 {
				fmt.Println("Usage: go run calculator.go subtract <a> <b>")
				os.Exit(1)
			}
			a, _ := strconv.ParseFloat(os.Args[2], 64)
			b, _ := strconv.ParseFloat(os.Args[3], 64)
			result := calc.Subtract(a, b)
			fmt.Printf("%.2f - %.2f = %.2f\n", a, b, result)
			
		case "multiply":
			if len(os.Args) != 4 {
				fmt.Println("Usage: go run calculator.go multiply <a> <b>")
				os.Exit(1)
			}
			a, _ := strconv.ParseFloat(os.Args[2], 64)
			b, _ := strconv.ParseFloat(os.Args[3], 64)
			result := calc.Multiply(a, b)
			fmt.Printf("%.2f × %.2f = %.2f\n", a, b, result)
			
		case "divide":
			if len(os.Args) != 4 {
				fmt.Println("Usage: go run calculator.go divide <a> <b>")
				os.Exit(1)
			}
			a, _ := strconv.ParseFloat(os.Args[2], 64)
			b, _ := strconv.ParseFloat(os.Args[3], 64)
			result, err := calc.Divide(a, b)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("%.2f ÷ %.2f = %.2f\n", a, b, result)
			
		case "expr":
			if len(os.Args) != 3 {
				fmt.Println("Usage: go run calculator.go expr \"2+3*4\"")
				os.Exit(1)
			}
			result, err := calc.EvaluateExpression(os.Args[2])
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("%s = %.2f\n", os.Args[2], result)
			
		default:
			fmt.Println("Usage: go run calculator.go [add|subtract|multiply|divide|expr] <args>")
			fmt.Println("Or run without arguments for interactive mode")
		}
		
		return
	}

	// Interactive mode
	calc := NewCalculator()
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Welcome to Go Calculator!")
	fmt.Println("Type 'q' to quit")

	for {
		printMenu()
		
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			a, err1 := getFloat64Input("Enter first number: ")
			b, err2 := getFloat64Input("Enter second number: ")
			
			if err1 != nil || err2 != nil {
				fmt.Println("Invalid input. Please enter valid numbers.")
				continue
			}
			
			result := calc.Add(a, b)
			fmt.Printf("Result: %.2f\n", result)

		case "2":
			a, err1 := getFloat64Input("Enter first number: ")
			b, err2 := getFloat64Input("Enter second number: ")
			
			if err1 != nil || err2 != nil {
				fmt.Println("Invalid input. Please enter valid numbers.")
				continue
			}
			
			result := calc.Subtract(a, b)
			fmt.Printf("Result: %.2f\n", result)

		case "3":
			a, err1 := getFloat64Input("Enter first number: ")
			b, err2 := getFloat64Input("Enter second number: ")
			
			if err1 != nil || err2 != nil {
				fmt.Println("Invalid input. Please enter valid numbers.")
				continue
			}
			
			result := calc.Multiply(a, b)
			fmt.Printf("Result: %.2f\n", result)

		case "4":
			a, err1 := getFloat64Input("Enter first number: ")
			b, err2 := getFloat64Input("Enter second number: ")
			
			if err1 != nil || err2 != nil {
				fmt.Println("Invalid input. Please enter valid numbers.")
				continue
			}
			
			result, err := calc.Divide(a, b)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Result: %.2f\n", result)
			}

		case "5":
			base, err1 := getFloat64Input("Enter base: ")
			exponent, err2 := getFloat64Input("Enter exponent: ")
			
			if err1 != nil || err2 != nil {
				fmt.Println("Invalid input. Please enter valid numbers.")
				continue
			}
			
			result := calc.Power(base, exponent)
			fmt.Printf("Result: %.2f\n", result)

		case "6":
			n, err := getFloat64Input("Enter number: ")
			
			if err != nil {
				fmt.Println("Invalid input. Please enter a valid number.")
				continue
			}
			
			result, err := calc.Sqrt(n)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Result: %.2f\n", result)
			}

		case "7":
			expr := getStringInput("Enter expression (e.g., 2+3*4): ")
			
			result, err := calc.EvaluateExpression(expr)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Result: %.2f\n", result)
			}

		case "m":
			value, err := getFloat64Input("Enter value to store in memory: ")
			if err != nil {
				fmt.Println("Invalid input. Please enter a valid number.")
				continue
			}
			calc.StoreMemory(value)

		case "r":
			value := calc.RecallMemory()
			fmt.Printf("Memory value: %.2f\n", value)

		case "c":
			calc.ClearMemory()

		case "h":
			calc.ShowHistory()

		case "x":
			calc.ClearHistory()

		case "q":
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}
