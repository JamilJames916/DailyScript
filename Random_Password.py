#3import random
# import string 
'''
def generate_password(length):
    characters = string.ascii_letters + string.punctuation
    password = ''.join(random.choice(characters) for _ in range(length))
    return password

password_length = 12
print("Generated password:", generate_password(password_length))
'''
import random
import string

# Function to generate random words
def generate_random_word(word_length):
    letters = string.ascii_lowercase
    return ''.join(random.choice(letters) for _ in range(word_length))

# Function to generate a password with random words and symbols
def generate_password(num_words, word_length, num_symbols):
    words = [generate_random_word(word_length) for _ in range(num_words)]
    password_base = '-'.join(words)
    
    symbols = string.punctuation.replace('&', '')
    random_symbols = ''.join(random.choice(symbols) for _ in range(num_symbols))
    
    password = password_base + random_symbols
    return password

# Parameters
num_words = 3
word_length = 5  # Length of each random word
num_symbols = 4  # Number of random symbols to append

# Generate and print the password
password = generate_password(num_words, word_length, num_symbols)
print("Generated password:", password)


