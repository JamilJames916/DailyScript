# The contains_value function checks if a given array contains a specific value.

def contains_value(arr, value):
    return value in arr

# Example
print(contains_value([1, 2, 3, 4, 5], 3)) # True

# Example
print(contains_value([1, 2, 3, 4, 5], 6)) # False