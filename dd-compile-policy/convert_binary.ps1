# Convert binary file to C header array (mimics 'xxd -i' on Unix systems)
# Usage: ./convert_binary.ps1 <input_binary_file> <output_header_file>

param(
    [Parameter(Mandatory=$true)]
    [string]$InputFile,
    
    [Parameter(Mandatory=$true)]
    [string]$OutputFile
)

# Read binary file as bytes
$bytes = Get-Content $InputFile -Encoding Byte

# Convert bytes to hex array format
$hexArray = $bytes | ForEach-Object { "0x{0:x2}" -f $_ }
$hexString = $hexArray -join ", "

# Generate C header content
$header = "unsigned char schema_policy_bfbs[] = { " + $hexString + " };"
$length = "unsigned int schema_policy_bfbs_len = " + $bytes.Length + ";"

# Write to output file
$header + [System.Environment]::NewLine + $length | Out-File $OutputFile -Encoding ascii
