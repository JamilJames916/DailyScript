# Function to reverse each word in a sentence
function Invoke-ReverseWords {
    [CmdletBinding()]
    param (
        [Parameter(ValueFromPipeline=$true, Position=0)]
        [string]$sentence
    )
    
    process {
        $words = $sentence -split '\s+'
        $reversedWords = foreach ($word in $words) {
            -join ($word.ToCharArray() | Select-Object -Property @{Name='Char';Expression={$_}} | Sort-Object -Property @{Expression={$_.Char};Descending=$true} | ForEach-Object -Process {$_.Char})
        }
        $reversedSentence = $reversedWords -join ' '
        Write-Output $reversedSentence
    }
}

# Function to reverse a sentence
function Invoke-ReverseSentence {
    [CmdletBinding()]
    param (
        [Parameter(ValueFromPipeline=$true, Position=0)]
        [string]$sentence
    )
    
    process {
        $reversedSentence = Reverse-Words -sentence $sentence
        Write-Output $reversedSentence
    }
}

# Example usage:
# Reverse a specific sentence
$testSentence = "This is how you can improve your PowerShell skills."
$reversedTestSentence = Reverse-Sentence -sentence $testSentence
Write-Output "Original sentence: $testSentence"
Write-Output "Reversed sentence with reversed words: $reversedTestSentence"

# Example with pipeline input
"Another example sentence." | Reverse-Sentence
