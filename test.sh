#!/usr/bin/env bash
set -euo pipefail

# instantiate a success count to 0
# instantiate a failure count to 0
# loop through all p-expected-*.bin files in the current directory
# run ./c/build/c/basic-debugger on each file
# save the output to a file with the same name but with .txt extension
# run ./c/build/c/basic-debugger on the corresponding p-received-*.bin file (where the * is the same as in the p-expected-*.bin file)
# save the output to a file with the same name but with .txt extension
# compare the outputs
# if the outputs are different, print the difference and increment the failure count
# if the outputs are the same, print a message saying the outputs match and increment the success count
# after the loop, print the success and failure counts
# if the failure count is greater than 0, exit with a non-zero code
# add a trap to always clean up the .txt files after the script exits

set -euo pipefail

success=0
failure=0

# Run debugger and normalize first line to remove numeric byte count
run_and_normalize() {
    input_file="$1"
    output_file="$2"
    echo "Running debugger on $input_file"
    echo "  Saving to $output_file"
    ./c/build/c/basic-debugger "$input_file" > "$output_file"
    # Remove the first line of the file since it contains the numeric byte count and the file name
    sed -i '1d' "$output_file"
}

# Clean up .txt files on exit
cleanup() {
    rm -f p-expected-*.txt p-received-*.txt
}
trap cleanup EXIT

echo -e "\n----------------------------------"
echo -e "🧪 Running tests..."
echo -e "----------------------------------\n"

for expected_file in p-expected-*.bin; do
    # Skip literal pattern if no files found
    [[ -e "$expected_file" ]] || continue

    base="${expected_file#p-expected-}"
    base="${base%.bin}"

    expected_txt="p-expected-${base}.txt"
    received_bin="p-received-${base}.bin"
    received_txt="p-received-${base}.txt"

    if [[ ! -e "$received_bin" ]]; then
        echo "⚠️  Missing matching file: $received_bin"
        failure=$((failure + 1))
        continue
    fi

    # Run debugger on expected
    run_and_normalize "$expected_file" "$expected_txt"

    # Run debugger on received
    run_and_normalize "$received_bin" "$received_txt"

    # Compare outputs
    if diff_output=$(diff -u "$expected_txt" "$received_txt"); then
        echo -e "\033[32m✔️ Match for $base\033[0m"
        success=$((success + 1))
    else
        echo -e "\033[31m❌ Difference for $base:\033[0m"
        echo "  $diff_output"
        failure=$((failure + 1))
    fi

    echo -e "\n"
done

echo "----------------------------------"
echo "Successes: $success"
echo "Failures:  $failure"
echo "----------------------------------"

if [[ $failure -gt 0 ]]; then
    exit 1
fi

# print success in green
echo -e "\n\033[32m✅ Tests all passed!\033[0m\n"
