# Intermediate level 
# Given a Value like this:
# $i = 2568
#What is the total sum of all the individual integers? That is to say 
# 2 + 5 + 6 + 8 
# Write a function that will accept a value and return the sum.

function Get-DigitSum {
    param (
        [int]$number
    )
    # Convert the number to a string to easily access each digit
    $digits = $number.ToString().ToCharArray()
    # Initialize sum to 0
    $sum = 0
    # Loop through each digit and add it to the sum
    foreach ($digit in $digits) {
        # Convert char to correct integer
        $intDigit = [int]::Parse($digit)
        $sum += $intDigit
        # Write-Output "Adding $intDigit, sum so far: $sum"
    }
    # Return the total sum
    return $sum

}

# Example usage 

$i = 2568
$result = Get-DigitSum -number $i
Write-Output "The sum of the digits in $i is: $result"
