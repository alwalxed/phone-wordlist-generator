# Phonegen

Generate 7-digit number combinations with custom prefix.

## Installation

```bash
go install github.com/alwalxed/phonegen@latest
```

## Usage

```bash
phonegen
```

Enter your desired prefix when prompted:

```bash
Enter prefix: 057
Generated 10,000,000 combinations in /home/user/.wordlist-generator/057-XXX-XXXX.txt
Completed in: 345ms
```

Default output directory is ~/.wordlist-generator. Existing files will not be overwritten.
