#!/bin/sh

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
echo "
#include \"riscv_test.h\"
#include \"test_macros.h\"
RVTEST_RV64S
RVTEST_CODE_BEGIN

#ifdef __MACHINE_MODE
  #define sscratch mscratch
  #define sstatus mstatus
  #define scause mcause
  #define sepc mepc
  #define stvec_handler mtvec_handler
#endif

$(cat $f)

RVTEST_PASS

  .align 3
stvec_handler:
  csrr t0, sepc
  addi t0, t0, 4
  csrw sepc, t0
  sret

RVTEST_CODE_END" > $f.S

riscv64-unknown-elf-gcc -static -fpic -fvisibility=hidden -nostdlib -nostartfiles -Wa,-march=RVIMAFDXhwacha -I $TOP/riscv-tools/riscv-tests/env/p -I $TOP/riscv-tools/riscv-tests/isa/macros/scalar -T $TOP/riscv-tools/riscv-tests/env/p/link.ld "$f.S" -o "$f.bin" \
	&& spike "$f.bin" \
	&& rm "$f.S" "$f.bin" \
	&& exit 0

exit 1
