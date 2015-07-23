#!/bin/sh

# currently not working (require priviledge instructions)
exit 200

# Runs the riscv64 test case given in argument
# The TOP variable must be set to the root of the riscv installation

if [ -z "$1" ]; then
	>&2 echo "usage: $0 test_file"
	exit 2
fi

if [ ! -f "$1" ]; then
	echo "File $1 not found!"
	exit 3
fi

f=$(realpath $1)

# add test header and footer
echo "#include \"riscv_test.h\"
#include \"test_macros.h\"
RVTEST_RV64U
RVTEST_CODE_BEGIN
$(cat $f)
RVTEST_PASS
RVTEST_CODE_END" > $f.S

riscv64-unknown-elf-gcc -static -fpic -fvisibility=hidden -nostdlib -nostartfiles -Wa,-march=RVIMAFDXhwacha -I $TOP/riscv-tools/riscv-tests/env/p -I $TOP/riscv-tools/riscv-tests/isa/macros/scalar -T $TOP/riscv-tools/riscv-tests/env/p/link.ld "$f.S" -o "$f.bin" \
	&& elf2hex 16 8192 "$f.bin" > "$f.hex" \
	&& cd $TOP/rocket-chip/emulator \
	&& ./emulator-DefaultCPPConfig +dramsim +max-cycles=100000 +loadmem="$f.hex" none \
	&& rm "$f.S" "$f.bin" "$f.hex" \
	&& exit 0

exit 1
