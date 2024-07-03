# Function to reverse a string
function Invoke-StringReverse {
    param (
        [string]$inputString
    )
    $charArray = $inputString.ToCharArray()
    [array]::Reverse($charArray)
    return -join $charArray
}

# Input word
$word = Read-Host -Prompt "Enter a word"

# Print word in reverse
$reversedWord = Invoke-StringReverse -inputString $word
Write-Output "Original word: $word"
Write-Output "Reversed word: $reversedWord"
