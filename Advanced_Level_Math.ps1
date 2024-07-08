# Given value like this:
# $a = 2,5,6
# What are all the unique non-sums using all possible combinations. For example:
# 2
# 5
# 6
# 2+5 = 7
# 2+5+6 = 13
# 2+6 = 8
# 5+6 = 11
# Write a Powershel function or set of functions 
# that will return a unique, and ordered list of sums based on the array. 
# The array should have no more than 9 elements and use values 1-9 in the array.
# Bonus Elements 
# Add parameter validation on the array length.
# Include Verbose output to show the operation.

function  Get-UniqueSums {
    param (
        [Parameter(Mandatory=$true)]
        [ValidateCount(1, 9)]
        [ValidateRange(1, 9)]
        [int[]]$Numbers
    )
    


    process {
        Write-Verbose "Input Numbers: $($Numbers -join ',')"

        # Function to get all unique combinations of the input array
        function Get-Combinations {
            param (
                [int[]]$Array
            )

            $combinations = @()

            for ($i = 1; $i -le [math]::Pow(2, $Array.Length) - 1; $i++) {
                $currentcombination = @()
                for ($j = 0; $j -lt $Array.Length; $j++) {
                    if ($i -band [math]::Pow(2, $j)) {
                    $currentcombination += $Array[$j]

                    }
                }   
                $combinations += ,$currentcombination
                
            }

            return $combinations
        }

        # Get all unique combinations
        $combinations = Get-Combinations -Array $Numbers
        Write-Verbose "Total Combinations: $($combinations.Count)"

        # Calculate sums from combinations
        $sums = @()
        foreach ($combination in $combinations) {
            $sum = ($combination | Measure-Object -Sum).Sum
            Write-Verbose "Combination: $($combination -join ',') => Sum: $sum"
            $sums += $sum
        
        }

        # Get unique and ordered sums
        $uniqueSums = $sums | Sort-Object -Unique
        return $uniqueSums
    }
}

# Example Usage:
$a = 2, 5, 6
Get-UniqueSums -Numbers $a -Verbose