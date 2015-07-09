#!/bin/sh

# Runs the riscv64 test case located in TAVOR_FUZZ_FILE with the rocket-chip emulator
# The TOP variable must be set to the root of the riscv installation

if [ -z "$TOP" ]; then
	>&2 echo "The TOP variable must be set to the root of the riscv installation"
	exit 2
fi

if [ -z "$TAVOR_FUZZ_FILE" ]; then
	>&2 echo "The TAVOR_FUZZ_FILE variable must be set"
	exit 3
fi

if [ ! -f "$TAVOR_FUZZ_FILE" ]; then
	echo "File $TAVOR_FUZZ_FILE not found!"
	exit 4
fi

f=$(realpath $TAVOR_FUZZ_FILE)
f=${f%%.*}

# add test header and footer
echo "#include \"riscv_test.h\"
#include \"test_macros.h\"
RVTEST_RV64U
RVTEST_CODE_BEGIN
$(cat $f.S)
RVTEST_PASS
RVTEST_CODE_END" > $f.S

riscv64-unknown-elf-gcc -static -fpic -fvisibility=hidden -nostdlib -nostartfiles -Wa,-march=RVIMAFDXhwacha -I $TOP/riscv-tools/riscv-tests/env/p -I $TOP/riscv-tools/riscv-tests/isa/macros/scalar -T $TOP/riscv-tools/riscv-tests/env/p/link.ld "$f.S" -o "$f" && elf2hex 16 8192 "$f" > "$f.hex" && cd $TOP/rocket-chip/emulator; ./emulator-DefaultCPPConfig +dramsim +max-cycles=100000 +loadmem="$f.hex" +verbose none 2>&1

exit 0
