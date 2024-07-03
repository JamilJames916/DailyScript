# Define the game's board: an array of numbers from 1 to 9
$numbers = 1..9

# Function to display the current board state
function Show-Board {
    $board = ($numbers | ForEach-Object { if ($_ -ne 0) { $_ } else { "_" } }) -join " "
    Write-Host "Board: $board"
}

# Function to roll dice
function Invoke-DiceRoll {
    $dice1 = Get-Random -Minimum 1 -Maximum 7
    $dice2 = Get-Random -Minimum 1 -Maximum 7
    $total = $dice1 + $dice2
    Write-Host "You rolled: $dice1 and $dice2 totaling $total"
    return $total
}

# Check if the game is over
function Test-GameOver {
    if ($numbers -eq 0) {
        Write-Host "Congratulations, you shut the box!"
        return $true
    }
    return $false
}

# Player's turn
function Invoke-PlayerTurn {
    param ($roll)
    do {
        Display-Board
        $choice = Read-Host "Enter numbers that add up to $roll (separated by spaces)"
        $selections = $choice -split ' ' | ForEach-Object { [int]$_ }
        $sum = $selections | Measure-Object -Sum | Select-Object -ExpandProperty Sum

        if ($sum -eq $roll) {
            foreach ($num in $selections) {
                $index = [array]::IndexOf($numbers, $num) # Find index of $num in $numbers
                if ($index -ne -1) {
                    $numbers[$index] = 0 # Set the found number to 0 to "Shut the box"
                }
            }
            break
        }
        else {
            Write-Host "The numbers you entered do not add up to $roll. Please try again."
        }
    } while ($true)
}

# Main game loop
function Start-Game {
    do {
        $roll = Roll-Dice
        if (-not (Game-Over)) {
            Player-Turn $roll
        }
        else {
            break
        }
    } while ($true)
}

# Start the game
Start-Game
