# First release of tavor-isa
[Here is](https://github.com/yblein/tavor-isa/tree/v0.1) the first release of my summer project! As a remainder this
project aims at fuzzing a CPU instruction set thanks to the Tavor framework.
This first release introduces a new format to describe instruction sets as well
as new fuzzing strategy extracting test programs from this specification.

## The format
The specification of an instruction set is basically composed of a set of
enhanced-assembly files together with a configuration file. For instance, one
can define the `addi` instructions in the file `I.S`:
```
addi @r, @r, $i12
```
where `@r` is a random register from the ones defined in the configuration file
and `$i12` is a 12-bit random integer.
The configuration file includes the list of assembly files (here only `I.S`)
and the list of registers:
```
instructions = ["I.S"]

[variables]
r = ["x0", "x1", ..., "x31"]
```

On the contraty to `@` values that are read from the configuration file, the special keys starting with a dollar sign (e.g., `$i12`) are predefined:
`$i` followed by a number `n` will produce a `n` bits random integer value.
Likewise, `$u` will produce an unsigned random integer.

You can find a more complete example in [example/riscv64](/example/riscv64). This
example is a beginning of specification for the RISCV64 instruction set,
including the base integer instruction set and the extension for multiplication
and division.

## Program generation
As already mentioned, the fuzzing step relies on the Tavor framework. This
framework is already able to generate random programs but we wanted to
generated programs targeting a specific coverage level. That is why I
developed a new strategy achieving a sort of "all-nodes" coverage. In the
previous example, the nodes are `addi `, `@r`, `, `, `@r`, `,` and `$i12`.
Because covering all the registers corresponding to `@r` and all the 12-bits
integers would produce lengthy and redundant programs, these values are
replaced with boundary values. Concretely, fuzzing the specification presented
in the previous section produces the following program:
```
addi x0, x0, -2048
addi x16, x16, -1
addi x31, x31, 0
addi x0, x0, 1
addi x0, x0, 2047
```
With only 5 instructions, this program covers several corner cases of the
`addi` instruction.

## Program execution
The RISCV64 example comes with a script that automatically executes the
generated program with the [spike emulator](https://github.com/riscv/riscv-isa-sim).
The test passes is the emulator terminates correctly. See the
[README](/README.md) for more information on building and running
tavor-isa.

## Limitations and Future Work
This work is a starting point and there is still a lot of work to do. Indeed,
the format doesn't allow to specify more complex instructions like labels and
jumps. Once this feature will be implemented, it should be possible to complete
the description of the RISCV64 instruction set.
With a complete description, the generated programs should be more exhaustive
and allow to assess the good working order of different emulators by comparing
their outputs.
