from itertools import product
import os

def generate_prefixed_combinations(filename, prefix):
    """
    Generate combinations of 7 digits and write them to a file with a given prefix.

    Args:
    - prefix (str): The prefix to prepend to each combination.
    """
    digits = '0123456789'
    try:
        with open(filename, 'w') as file:
            for combination in product(digits, repeat=7):
                file.write(f"{prefix}{''.join(combination)}\n")
        print(f"\033[92mSuccessfully created combinations with prefix '{prefix}' in file {filename}\033[0m")
    except IOError as e:
        print(f"\033[91mError writing to file {filename}: {e}\033[0m")

def main():
    try:
        GREEN = '\033[92m'  # ANSI escape sequence for green color
        RESET = '\033[0m'   # ANSI escape sequence to reset color
        BOLD = '\033[1m'    # ANSI escape sequence for bold text

        print(f"{BOLD}{GREEN}Welcome! This script generates combinations of 7 digits with a custom prefix.{RESET}")
        print(f"{BOLD}{GREEN}For example, entering '053' generates combinations like '0530000000', '0530000001', etc.{RESET}")

        prefix = input("Enter a prefix: ").strip()

        # Construct the filename based on the provided prefix
        output_filename = f"{prefix}-XXX-XXXX.txt"

        # Determine the directory of the script
        script_dir = os.path.dirname(os.path.abspath(__file__))
        output_directory = os.path.join(script_dir, 'numbers')

        # Create the 'numbers' directory if it doesn't exist
        os.makedirs(output_directory, exist_ok=True)

        output_filepath = os.path.join(output_directory, output_filename)

        # Check if file already exists
        if os.path.exists(output_filepath):
            print(f"\033[91mError: File '{output_filename}' already exists in the 'numbers' directory.\033[0m")
            return

        generate_prefixed_combinations(output_filepath, prefix)

    except OSError as e:
        print(f"\033[91mError: {e}\033[0m")
    except KeyboardInterrupt:
        print(f"\n{BOLD}\033[93mOperation aborted by user.\033[0m")
    except Exception as e:
        print(f"\033[91mAn unexpected error occurred: {e}\033[0m")

if __name__ == "__main__":
    main()
