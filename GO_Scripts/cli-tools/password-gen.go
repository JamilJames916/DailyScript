package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
)

type PasswordGenerator struct {
	length      int
	includeUpper bool
	includeLower bool
	includeDigits bool
	includeSymbols bool
	excludeSimilar bool
	customChars string
}

const (
	lowerChars = "abcdefghijklmnopqrstuvwxyz"
	upperChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digitChars = "0123456789"
	symbolChars = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	similarChars = "il1Lo0O"
)

func NewPasswordGenerator() *PasswordGenerator {
	return &PasswordGenerator{
		length: 12,
		includeUpper: true,
		includeLower: true,
		includeDigits: true,
		includeSymbols: false,
		excludeSimilar: false,
		customChars: "",
	}
}

func (pg *PasswordGenerator) SetLength(length int) *PasswordGenerator {
	if length < 1 {
		length = 1
	}
	pg.length = length
	return pg
}

func (pg *PasswordGenerator) IncludeUppercase(include bool) *PasswordGenerator {
	pg.includeUpper = include
	return pg
}

func (pg *PasswordGenerator) IncludeLowercase(include bool) *PasswordGenerator {
	pg.includeLower = include
	return pg
}

func (pg *PasswordGenerator) IncludeDigits(include bool) *PasswordGenerator {
	pg.includeDigits = include
	return pg
}

func (pg *PasswordGenerator) IncludeSymbols(include bool) *PasswordGenerator {
	pg.includeSymbols = include
	return pg
}

func (pg *PasswordGenerator) ExcludeSimilar(exclude bool) *PasswordGenerator {
	pg.excludeSimilar = exclude
	return pg
}

func (pg *PasswordGenerator) SetCustomChars(chars string) *PasswordGenerator {
	pg.customChars = chars
	return pg
}

func (pg *PasswordGenerator) buildCharset() string {
	var charset string
	
	if pg.customChars != "" {
		charset = pg.customChars
	} else {
		if pg.includeLower {
			charset += lowerChars
		}
		if pg.includeUpper {
			charset += upperChars
		}
		if pg.includeDigits {
			charset += digitChars
		}
		if pg.includeSymbols {
			charset += symbolChars
		}
	}
	
	if pg.excludeSimilar {
		for _, char := range similarChars {
			charset = strings.ReplaceAll(charset, string(char), "")
		}
	}
	
	return charset
}

func (pg *PasswordGenerator) Generate() (string, error) {
	charset := pg.buildCharset()
	
	if len(charset) == 0 {
		return "", fmt.Errorf("no characters available for password generation")
	}
	
	password := make([]byte, pg.length)
	charsetLen := big.NewInt(int64(len(charset)))
	
	for i := range password {
		randomIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		password[i] = charset[randomIndex.Int64()]
	}
	
	return string(password), nil
}

func (pg *PasswordGenerator) GenerateMultiple(count int) ([]string, error) {
	passwords := make([]string, count)
	
	for i := 0; i < count; i++ {
		password, err := pg.Generate()
		if err != nil {
			return nil, err
		}
		passwords[i] = password
	}
	
	return passwords, nil
}

func (pg *PasswordGenerator) CheckStrength(password string) map[string]interface{} {
	strength := make(map[string]interface{})
	
	length := len(password)
	strength["length"] = length
	
	hasLower := strings.ContainsAny(password, lowerChars)
	hasUpper := strings.ContainsAny(password, upperChars)
	hasDigits := strings.ContainsAny(password, digitChars)
	hasSymbols := strings.ContainsAny(password, symbolChars)
	
	strength["has_lowercase"] = hasLower
	strength["has_uppercase"] = hasUpper
	strength["has_digits"] = hasDigits
	strength["has_symbols"] = hasSymbols
	
	// Calculate complexity score
	complexity := 0
	if hasLower { complexity++ }
	if hasUpper { complexity++ }
	if hasDigits { complexity++ }
	if hasSymbols { complexity++ }
	
	strength["complexity"] = complexity
	
	// Calculate strength level
	var level string
	if length < 6 {
		level = "Very Weak"
	} else if length < 8 {
		level = "Weak"
	} else if length < 12 {
		if complexity >= 3 {
			level = "Medium"
		} else {
			level = "Weak"
		}
	} else if length < 16 {
		if complexity >= 3 {
			level = "Strong"
		} else {
			level = "Medium"
		}
	} else {
		if complexity >= 4 {
			level = "Very Strong"
		} else if complexity >= 3 {
			level = "Strong"
		} else {
			level = "Medium"
		}
	}
	
	strength["level"] = level
	
	return strength
}

func printUsage() {
	fmt.Println("Usage: go run password-gen.go [options]")
	fmt.Println("Options:")
	fmt.Println("  -l, --length <n>      Password length (default: 12)")
	fmt.Println("  -c, --count <n>       Number of passwords to generate (default: 1)")
	fmt.Println("  --no-upper           Exclude uppercase letters")
	fmt.Println("  --no-lower           Exclude lowercase letters")
	fmt.Println("  --no-digits          Exclude digits")
	fmt.Println("  --symbols            Include symbols")
	fmt.Println("  --exclude-similar    Exclude similar characters (il1Lo0O)")
	fmt.Println("  --custom <chars>     Use custom character set")
	fmt.Println("  --check <password>   Check password strength")
	fmt.Println("  --help               Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run password-gen.go")
	fmt.Println("  go run password-gen.go -l 16 --symbols")
	fmt.Println("  go run password-gen.go -c 5 -l 8")
	fmt.Println("  go run password-gen.go --check mypassword123")
}

func main() {
	if len(os.Args) == 1 {
		// Default behavior
		generator := NewPasswordGenerator()
		password, err := generator.Generate()
		if err != nil {
			fmt.Printf("Error generating password: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(password)
		return
	}

	generator := NewPasswordGenerator()
	count := 1
	checkPassword := ""
	
	// Parse command line arguments
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		
		switch arg {
		case "-l", "--length":
			if i+1 < len(os.Args) {
				length, err := strconv.Atoi(os.Args[i+1])
				if err != nil {
					fmt.Printf("Invalid length: %s\n", os.Args[i+1])
					os.Exit(1)
				}
				generator.SetLength(length)
				i++
			} else {
				fmt.Println("Length value required")
				os.Exit(1)
			}
			
		case "-c", "--count":
			if i+1 < len(os.Args) {
				c, err := strconv.Atoi(os.Args[i+1])
				if err != nil {
					fmt.Printf("Invalid count: %s\n", os.Args[i+1])
					os.Exit(1)
				}
				count = c
				i++
			} else {
				fmt.Println("Count value required")
				os.Exit(1)
			}
			
		case "--no-upper":
			generator.IncludeUppercase(false)
			
		case "--no-lower":
			generator.IncludeLowercase(false)
			
		case "--no-digits":
			generator.IncludeDigits(false)
			
		case "--symbols":
			generator.IncludeSymbols(true)
			
		case "--exclude-similar":
			generator.ExcludeSimilar(true)
			
		case "--custom":
			if i+1 < len(os.Args) {
				generator.SetCustomChars(os.Args[i+1])
				i++
			} else {
				fmt.Println("Custom character set required")
				os.Exit(1)
			}
			
		case "--check":
			if i+1 < len(os.Args) {
				checkPassword = os.Args[i+1]
				i++
			} else {
				fmt.Println("Password to check required")
				os.Exit(1)
			}
			
		case "--help":
			printUsage()
			return
			
		default:
			fmt.Printf("Unknown option: %s\n", arg)
			printUsage()
			os.Exit(1)
		}
	}
	
	if checkPassword != "" {
		// Check password strength
		strength := generator.CheckStrength(checkPassword)
		
		fmt.Printf("Password: %s\n", checkPassword)
		fmt.Printf("Length: %d\n", strength["length"])
		fmt.Printf("Strength Level: %s\n", strength["level"])
		fmt.Printf("Complexity Score: %d/4\n", strength["complexity"])
		fmt.Printf("Has Lowercase: %t\n", strength["has_lowercase"])
		fmt.Printf("Has Uppercase: %t\n", strength["has_uppercase"])
		fmt.Printf("Has Digits: %t\n", strength["has_digits"])
		fmt.Printf("Has Symbols: %t\n", strength["has_symbols"])
		
		return
	}
	
	// Generate passwords
	if count == 1 {
		password, err := generator.Generate()
		if err != nil {
			fmt.Printf("Error generating password: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(password)
	} else {
		passwords, err := generator.GenerateMultiple(count)
		if err != nil {
			fmt.Printf("Error generating passwords: %v\n", err)
			os.Exit(1)
		}
		
		for i, password := range passwords {
			fmt.Printf("%d: %s\n", i+1, password)
		}
	}
}
