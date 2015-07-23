# tavor-isa

## Compilation
```
go build
```

## Running the riscv64 example
Generate a test program:
```
./tavor-isa example/riscv64/config.toml
```

Automatically execute the generated program with the [spike emulator](https://github.com/riscv/riscv-isa-sim):
```
export TOP=/root/of/riscv/install
./tavor-isa --exec example/riscv64/run_spike.sh example/riscv64/config.toml
```
