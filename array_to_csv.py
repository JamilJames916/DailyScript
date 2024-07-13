# The array_to_csv function converts an array to a comma-separated string. 

def array_to_csv(arr):
    return ','.join(map(str, arr))
# Example usage
print(array_to_csv([1, 2, 3, 4, 5]))
